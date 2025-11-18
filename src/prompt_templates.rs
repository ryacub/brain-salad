//! Advanced Prompt Engineering System with Dynamic Templates
//!
//! This module provides dynamic prompt construction, few-shot examples,
//! and chain-of-thought reasoning for improved LLM analysis quality.

use crate::commands::analyze_llm::LlmProvider;
use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use std::sync::OnceLock;

/// Prompt template for different idea types and contexts
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct PromptTemplate {
    /// Unique template identifier
    pub id: String,
    /// Template name/description
    pub name: String,
    /// Idea type this template is optimized for
    pub idea_type: String,
    /// The actual prompt template with placeholders
    pub template: String,
    /// Few-shot examples specific to this template
    pub examples: Vec<FewShotExample>,
    /// Chain-of-thought steps to include
    pub cot_steps: Vec<String>,
    /// Template version for tracking improvements
    pub version: String,
    /// Performance metrics for this template
    pub performance: TemplatePerformance,
}

/// Few-shot example for prompt engineering
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct FewShotExample {
    /// Example idea text
    pub idea: String,
    /// Expected analysis pattern
    pub expected_pattern: String,
    /// Explanation of why this example is relevant
    pub explanation: String,
    /// Difficulty level (easy, medium, hard)
    pub difficulty: String,
    /// Domain tags for categorization
    pub tags: Vec<String>,
}

/// Template performance metrics
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct TemplatePerformance {
    /// Number of times this template was used
    pub usage_count: u64,
    /// Average quality score of responses
    pub avg_quality_score: f64,
    /// Average confidence level
    pub avg_confidence: f64,
    /// Success rate (responses that passed validation)
    pub success_rate: f64,
    /// Average response time in milliseconds
    pub avg_response_time_ms: u64,
}

impl Default for TemplatePerformance {
    fn default() -> Self {
        Self {
            usage_count: 0,
            avg_quality_score: 0.0,
            avg_confidence: 0.0,
            success_rate: 0.0,
            avg_response_time_ms: 0,
        }
    }
}

/// Example library for few-shot learning
#[derive(Debug, Clone)]
pub struct ExampleLibrary {
    examples: HashMap<String, Vec<FewShotExample>>,
}

impl ExampleLibrary {
    /// Create a new example library
    pub fn new() -> Self {
        let mut library = Self {
            examples: HashMap::new(),
        };

        // Initialize with curated examples
        library.initialize_default_examples();
        library
    }

    /// Initialize default few-shot examples
    fn initialize_default_examples(&mut self) {
        // Technical/Development examples
        self.add_examples("technical", vec![
            FewShotExample {
                idea: "Create a mobile app for tracking daily habits".to_string(),
                expected_pattern: "High domain expertise, strong AI alignment, good execution support, moderate revenue potential".to_string(),
                explanation: "Technical project that leverages existing mobile development skills".to_string(),
                difficulty: "medium".to_string(),
                tags: vec!["mobile".to_string(), "productivity".to_string(), "technical".to_string()],
            },
            FewShotExample {
                idea: "Build a Rust-based CLI tool for telos-idea-matrix automation".to_string(),
                expected_pattern: "Very high domain expertise, excellent AI alignment, strong execution support, low revenue potential".to_string(),
                explanation: "Directly leverages existing tech stack and domain knowledge".to_string(),
                difficulty: "easy".to_string(),
                tags: vec!["rust".to_string(), "cli".to_string(), "automation".to_string()],
            },
        ]);

        // Business/Startup examples
        self.add_examples("business", vec![
            FewShotExample {
                idea: "Start a consulting business for AI implementation strategies".to_string(),
                expected_pattern: "High domain expertise, good AI alignment, strong execution support, high revenue potential".to_string(),
                explanation: "Leverages AI expertise while creating value for others".to_string(),
                difficulty: "hard".to_string(),
                tags: vec!["consulting".to_string(), "ai".to_string(), "business".to_string()],
            },
        ]);

        // Content/Creative examples
        self.add_examples("content", vec![
            FewShotExample {
                idea: "Write a technical blog series about Rust performance optimization".to_string(),
                expected_pattern: "High domain expertise, good AI alignment, moderate execution support, moderate revenue potential".to_string(),
                explanation: "Builds on technical expertise and creates content assets".to_string(),
                difficulty: "medium".to_string(),
                tags: vec!["blog".to_string(), "technical".to_string(), "content".to_string()],
            },
        ]);

        // Learning/Research examples
        self.add_examples("learning", vec![
            FewShotExample {
                idea: "Master advanced Rust concurrency patterns through project-based learning".to_string(),
                expected_pattern: "High domain expertise, excellent AI alignment, moderate execution support, low revenue potential".to_string(),
                explanation: "Direct skill development that enhances existing capabilities".to_string(),
                difficulty: "medium".to_string(),
                tags: vec!["rust".to_string(), "learning".to_string(), "concurrency".to_string()],
            },
        ]);
    }

    /// Add examples for a specific idea type
    pub fn add_examples(&mut self, idea_type: &str, examples: Vec<FewShotExample>) {
        self.examples.insert(idea_type.to_string(), examples);
    }

    /// Get relevant examples for an idea type
    pub fn get_examples(&self, idea_type: &str, max_examples: usize) -> Vec<&FewShotExample> {
        self.examples
            .get(idea_type)
            .map(|examples| {
                let mut examples: Vec<_> = examples.iter().collect();
                // Sort by difficulty (easy first for learning)
                examples.sort_by(
                    |a, b| match (a.difficulty.as_str(), b.difficulty.as_str()) {
                        ("easy", _) => std::cmp::Ordering::Less,
                        (_, "easy") => std::cmp::Ordering::Greater,
                        ("medium", "hard") => std::cmp::Ordering::Less,
                        ("hard", "medium") => std::cmp::Ordering::Greater,
                        _ => std::cmp::Ordering::Equal,
                    },
                );
                examples.into_iter().take(max_examples).collect()
            })
            .unwrap_or_default()
    }
}

/// Dynamic prompt builder that adapts to idea context
pub struct DynamicPromptBuilder {
    templates: HashMap<String, PromptTemplate>,
    example_library: ExampleLibrary,
    default_template: PromptTemplate,
}

impl DynamicPromptBuilder {
    /// Create a new dynamic prompt builder
    pub fn new() -> Self {
        let mut builder = Self {
            templates: HashMap::new(),
            example_library: ExampleLibrary::new(),
            default_template: Self::create_default_template(),
        };

        // Initialize specialized templates
        builder.initialize_templates();
        builder
    }

    /// Initialize default templates
    fn initialize_templates(&mut self) {
        // Technical ideas template
        self.add_template(PromptTemplate {
            id: "technical_v1".to_string(),
            name: "Technical Project Analysis".to_string(),
            idea_type: "technical".to_string(),
            template: Self::get_technical_template(),
            examples: self
                .example_library
                .get_examples("technical", 2)
                .into_iter()
                .cloned()
                .collect(),
            cot_steps: vec![
                "First, analyze the technical feasibility and required skills".to_string(),
                "Next, evaluate alignment with current tech stack and goals".to_string(),
                "Then, assess execution complexity and resource requirements".to_string(),
                "Finally, consider learning value and long-term benefits".to_string(),
            ],
            version: "1.0".to_string(),
            performance: TemplatePerformance::default(),
        });

        // Business ideas template
        self.add_template(PromptTemplate {
            id: "business_v1".to_string(),
            name: "Business Venture Analysis".to_string(),
            idea_type: "business".to_string(),
            template: Self::get_business_template(),
            examples: self
                .example_library
                .get_examples("business", 2)
                .into_iter()
                .cloned()
                .collect(),
            cot_steps: vec![
                "First, analyze market need and opportunity size".to_string(),
                "Next, evaluate required resources and timeline".to_string(),
                "Then, assess risks and competitive advantages".to_string(),
                "Finally, consider alignment with personal goals and values".to_string(),
            ],
            version: "1.0".to_string(),
            performance: TemplatePerformance::default(),
        });

        // Content ideas template
        self.add_template(PromptTemplate {
            id: "content_v1".to_string(),
            name: "Content Creation Analysis".to_string(),
            idea_type: "content".to_string(),
            template: Self::get_content_template(),
            examples: self
                .example_library
                .get_examples("content", 2)
                .into_iter()
                .cloned()
                .collect(),
            cot_steps: vec![
                "First, evaluate content value and target audience".to_string(),
                "Next, assess creation effort and required skills".to_string(),
                "Then, consider distribution and impact potential".to_string(),
                "Finally, evaluate alignment with expertise and goals".to_string(),
            ],
            version: "1.0".to_string(),
            performance: TemplatePerformance::default(),
        });

        // Learning ideas template
        self.add_template(PromptTemplate {
            id: "learning_v1".to_string(),
            name: "Learning Project Analysis".to_string(),
            idea_type: "learning".to_string(),
            template: Self::get_learning_template(),
            examples: self
                .example_library
                .get_examples("learning", 2)
                .into_iter()
                .cloned()
                .collect(),
            cot_steps: vec![
                "First, evaluate learning value and skill relevance".to_string(),
                "Next, assess time commitment and difficulty".to_string(),
                "Then, consider application to current or future projects".to_string(),
                "Finally, evaluate enjoyment and motivation factors".to_string(),
            ],
            version: "1.0".to_string(),
            performance: TemplatePerformance::default(),
        });
    }

    /// Create the default template
    fn create_default_template() -> PromptTemplate {
        PromptTemplate {
            id: "default_v1".to_string(),
            name: "General Purpose Analysis".to_string(),
            idea_type: "general".to_string(),
            template: Self::get_default_template(),
            examples: vec![],
            cot_steps: vec![
                "Analyze the idea's feasibility and requirements".to_string(),
                "Evaluate alignment with personal goals and values".to_string(),
                "Assess potential benefits and drawbacks".to_string(),
                "Consider immediate next steps and timeline".to_string(),
            ],
            version: "1.0".to_string(),
            performance: TemplatePerformance::default(),
        }
    }

    /// Add a template to the builder
    pub fn add_template(&mut self, template: PromptTemplate) {
        self.templates.insert(template.idea_type.clone(), template);
    }

    /// Build a dynamic prompt for the given idea
    pub fn build_prompt(
        &self,
        idea: &str,
        idea_type: &str,
        provider: &LlmProvider,
        include_examples: bool,
        include_cot: bool,
    ) -> String {
        let template = self
            .templates
            .get(idea_type)
            .unwrap_or(&self.default_template);

        let mut prompt = template.template.clone();

        // Add few-shot examples if requested
        if include_examples && !template.examples.is_empty() {
            prompt.push_str("\n\n### EXAMPLE ANALYSES:\n\n");
            for (i, example) in template.examples.iter().enumerate() {
                prompt.push_str(&format!(
                    "Example {}:\nIdea: {}\nExpected Pattern: {}\nExplanation: {}\n\n",
                    i + 1,
                    example.idea,
                    example.expected_pattern,
                    example.explanation
                ));
            }
        }

        // Add chain-of-thought steps if requested
        if include_cot && !template.cot_steps.is_empty() {
            prompt.push_str("\n### ANALYSIS APPROACH:\n\n");
            prompt.push_str("Please follow this step-by-step approach:\n");
            for (i, step) in template.cot_steps.iter().enumerate() {
                prompt.push_str(&format!("{}. {}\n", i + 1, step));
            }
            prompt.push('\n');
        }

        // Replace placeholders
        prompt = prompt.replace("{IDEA}", idea);
        prompt = prompt.replace("{IDEA_TYPE}", idea_type);
        prompt = prompt.replace("{PROVIDER}", &provider.provider_type());

        prompt
    }

    /// Get template performance metrics
    pub fn get_template_performance(&self, idea_type: &str) -> Option<&TemplatePerformance> {
        self.templates.get(idea_type).map(|t| &t.performance)
    }

    /// Update template performance
    pub fn update_template_performance(
        &mut self,
        idea_type: &str,
        quality_score: f64,
        confidence: f64,
        success: bool,
        response_time_ms: u64,
    ) {
        if let Some(template) = self.templates.get_mut(idea_type) {
            let perf = &mut template.performance;
            perf.usage_count += 1;

            // Update running averages
            let n = perf.usage_count as f64;
            perf.avg_quality_score = (perf.avg_quality_score * (n - 1.0) + quality_score) / n;
            perf.avg_confidence = (perf.avg_confidence * (n - 1.0) + confidence) / n;
            perf.success_rate =
                (perf.success_rate * (n - 1.0) + if success { 1.0 } else { 0.0 }) / n;
            perf.avg_response_time_ms = (perf.avg_response_time_ms as f64 * (n - 1.0)
                + response_time_ms as f64) as u64
                / n as u64;
        }
    }

    /// Get the best performing template for an idea type
    pub fn get_best_template(&self, idea_type: &str) -> Option<&PromptTemplate> {
        self.templates.get(idea_type)
    }

    // Template definitions
    fn get_default_template() -> String {
        r#"You are an expert idea evaluator specializing in Telos-aligned decision analysis.

Your task is to analyze the following idea using the structured evaluation framework below.

IDEA TO ANALYZE: {IDEA}

## Evaluation Framework

### 1. Mission Alignment (Maximum 4.00 points)
- Domain Expertise (1.20 max): How well does this leverage existing skills and knowledge?
- AI Alignment (1.50 max): How does this align with AI/automation goals?
- Execution Support (0.80 max): How well does this support current projects and workflows?
- Revenue Potential (0.50 max): What are the immediate/short-term revenue opportunities?

### 2. Anti-Challenge Patterns (Maximum 3.50 points)
- Avoid Context-Switching (1.20 max): How focused is this idea on single-threaded execution?
- Rapid Prototyping (1.00 max): How quickly can this be tested and validated?
- Accountability (0.80 max): How does this support public commitment and delivery?
- Income Anxiety (0.50 max): How does this address income stability concerns?

### 3. Strategic Fit (Maximum 2.50 points)
- Stack Compatibility (1.00 max): How well does this fit with current tech stack?
- Shipping Habit (0.80 max): How well does this support regular shipping and delivery?
- Public Accountability (0.40 max): What are the opportunities for public sharing?
- Revenue Testing (0.30 max): How easily can revenue potential be tested?

## Analysis Instructions

Provide your analysis as a JSON object with the following structure:

```json
{
  "scores": {
    "Mission Alignment": {
      "Domain Expertise": <score 0.00-1.20>,
      "AI Alignment": <score 0.00-1.50>,
      "Execution Support": <score 0.00-0.80>,
      "Revenue Potential": <score 0.00-0.50>,
      "category_total": <sum of above>
    },
    "Anti-Challenge Patterns": {
      "Avoid Context-Switching": <score 0.00-1.20>,
      "Rapid Prototyping": <score 0.00-1.00>,
      "Accountability": <score 0.00-0.80>,
      "Income Anxiety": <score 0.00-0.50>,
      "category_total": <sum of above>
    },
    "Strategic Fit": {
      "Stack Compatibility": <score 0.00-1.00>,
      "Shipping Habit": <score 0.00-0.80>,
      "Public Accountability": <score 0.00-0.40>,
      "Revenue Testing": <score 0.00-0.30>,
      "category_total": <sum of above>
    }
  },
  "weighted_totals": {
    "Mission Alignment": <mission_total * 0.4>,
    "Anti-Challenge Patterns": <anti_challenge_total * 0.35>,
    "Strategic Fit": <strategic_total * 0.25>
  },
  "final_score": <sum of weighted totals>,
  "recommendation": "Priority" | "Good" | "Consider" | "Avoid",
  "explanations": {
    "Domain Expertise": "<detailed explanation>",
    "AI Alignment": "<detailed explanation>",
    "Execution Support": "<detailed explanation>",
    "Revenue Potential": "<detailed explanation>",
    "Avoid Context-Switching": "<detailed explanation>",
    "Rapid Prototyping": "<detailed explanation>",
    "Accountability": "<detailed explanation>",
    "Income Anxiety": "<detailed explanation>",
    "Stack Compatibility": "<detailed explanation>",
    "Shipping Habit": "<detailed explanation>",
    "Public Accountability": "<detailed explanation>",
    "Revenue Testing": "<detailed explanation>"
  }
}
```

Ensure all scores:
1. Use exactly 2 decimal places
2. Are within the specified ranges
3. The category totals equal the sum of their components
4. Weighted totals are calculated correctly (40%, 35%, 25%)
5. The final score equals the sum of weighted totals

Provide specific, actionable explanations for each score component."#
            .to_string()
    }

    fn get_technical_template() -> String {
        r#"You are an expert technical evaluator specializing in software development and AI projects.

IDEA TO ANALYZE: {IDEA}
IDEA TYPE: Technical Project

## Technical Project Analysis Framework

Focus on technical feasibility, skill development, and implementation efficiency.

### Technical Expertise Assessment (Domain Expertise - 1.20 max)
- Current skill stack alignment
- Learning curve and required knowledge gaps
- Technical complexity and implementation challenges
- Leverage of existing codebases and tools

### AI/ML Alignment (AI Alignment - 1.50 max)
- Opportunities for automation and AI integration
- Alignment with AI learning goals
- Potential for AI-enhanced features
- Data collection and model training opportunities

### Implementation Strategy (Execution Support - 0.80 max)
- Integration with existing projects and workflows
- Tooling and infrastructure requirements
- Development timeline and milestones
- Dependencies and external requirements

### Monetization Strategy (Revenue Potential - 0.50 max)
- Direct revenue opportunities (SaaS, tools, services)
- Indirect value (portfolio, skills, networking)
- Market demand and competitive landscape
- Time to market and MVP potential

Provide detailed technical analysis with specific implementation considerations."#.to_string()
    }

    fn get_business_template() -> String {
        r#"You are an expert business analyst specializing in digital ventures and consulting services.

IDEA TO ANALYZE: {IDEA}
IDEA TYPE: Business Venture

## Business Venture Analysis Framework

Focus on market opportunity, resource requirements, and revenue potential.

### Market Analysis (Domain Expertise - 1.20 max)
- Market size and growth potential
- Target customer identification
- Competitive landscape and differentiation
- Personal expertise and industry knowledge

### Technology Leverage (AI Alignment - 1.50 max)
- AI/automation opportunities for efficiency
- Technology-enabled competitive advantages
- Scalability and systems thinking
- Digital transformation potential

### Resource Requirements (Execution Support - 0.80 max)
- Capital investment needs
- Team and talent requirements
- Timeline to market/launch
- Risk factors and mitigation strategies

### Revenue Model (Revenue Potential - 0.50 max)
- Pricing strategy and revenue streams
- Customer acquisition cost
- Lifetime value projections
- Break-even analysis

Focus on practical business considerations and realistic implementation planning."#.to_string()
    }

    fn get_content_template() -> String {
        r#"You are an expert content strategist specializing in technical and educational content.

IDEA TO ANALYZE: {IDEA}
IDEA TYPE: Content Creation

## Content Creation Analysis Framework

Focus on value creation, audience impact, and content leverage.

### Expertise Leverage (Domain Expertise - 1.20 max)
- Subject matter expertise and credibility
- Unique perspective or insights
- Research and knowledge requirements
- Content differentiation opportunities

### AI-Enhanced Creation (AI Alignment - 1.50 max)
- AI tools for content creation and optimization
- Automation opportunities for content production
- Personal AI branding and thought leadership
- Content distribution and AI-powered reach

### Production Strategy (Execution Support - 0.80 max)
- Content creation workflow and timeline
- Required tools and resources
- Distribution channels and platforms
- Consistency and scheduling considerations

### Monetization Pathways (Revenue Potential - 0.50 max)
- Direct monetization (courses, consulting, sponsorships)
- Indirect value (networking, opportunities, authority)
- Content reuse and repurposing potential
- Audience building and long-term value

Emphasize sustainable content creation and audience building strategies."#
            .to_string()
    }

    fn get_learning_template() -> String {
        r#"You are an expert learning strategist specializing in skill development and knowledge acquisition.

IDEA TO ANALYZE: {IDEA}
IDEA TYPE: Learning Project

## Learning Project Analysis Framework

Focus on skill value, learning efficiency, and practical application.

### Skill Value (Domain Expertise - 1.20 max)
- Relevance to current and future goals
- Market demand and career impact
- Knowledge transfer and applicability
- Foundation for advanced learning

### Technology Integration (AI Alignment - 1.50 max)
- AI-powered learning tools and methods
- Automation of learning processes
- Data-driven progress tracking
- Future-proof technology skills

### Learning Strategy (Execution Support - 0.80 max)
- Project-based vs. theoretical learning
- Time commitment and scheduling
- Resource requirements and access
- Milestone and progress tracking

### Career Impact (Revenue Potential - 0.50 max)
- Immediate job application opportunities
- Salary increase potential
- Freelance/consulting opportunities
- Long-term career trajectory impact

Focus on practical, applicable learning with clear ROI on time investment."#.to_string()
    }
}

impl Default for DynamicPromptBuilder {
    fn default() -> Self {
        Self::new()
    }
}

/// Global prompt builder instance
static PROMPT_BUILDER: OnceLock<DynamicPromptBuilder> = OnceLock::new();

/// Get the global prompt builder instance
pub fn get_prompt_builder() -> &'static DynamicPromptBuilder {
    PROMPT_BUILDER.get_or_init(DynamicPromptBuilder::new)
}

/// Classification for idea types based on content analysis
pub fn classify_idea_type(idea: &str) -> String {
    let idea_lower = idea.to_lowercase();

    // Technical/Development ideas
    if idea_lower.contains("app")
        || idea_lower.contains("software")
        || idea_lower.contains("code")
        || idea_lower.contains("programming")
        || idea_lower.contains("development")
        || idea_lower.contains("tech")
        || idea_lower.contains("api")
        || idea_lower.contains("tool")
        || idea_lower.contains("rust")
        || idea_lower.contains("javascript")
        || idea_lower.contains("python")
        || idea_lower.contains("database")
    {
        return "technical".to_string();
    }

    // Business/Startup ideas
    if idea_lower.contains("business")
        || idea_lower.contains("startup")
        || idea_lower.contains("company")
        || idea_lower.contains("service")
        || idea_lower.contains("product")
        || idea_lower.contains("market")
        || idea_lower.contains("consulting")
        || idea_lower.contains("freelance")
        || idea_lower.contains("agency")
        || idea_lower.contains("saas")
    {
        return "business".to_string();
    }

    // Content/Creative ideas
    if idea_lower.contains("blog")
        || idea_lower.contains("content")
        || idea_lower.contains("video")
        || idea_lower.contains("write")
        || idea_lower.contains("creative")
        || idea_lower.contains("art")
        || idea_lower.contains("book")
        || idea_lower.contains("course")
        || idea_lower.contains("tutorial")
        || idea_lower.contains("podcast")
    {
        return "content".to_string();
    }

    // Learning/Research ideas
    if idea_lower.contains("learn")
        || idea_lower.contains("study")
        || idea_lower.contains("research")
        || idea_lower.contains("course")
        || idea_lower.contains("book")
        || idea_lower.contains("education")
        || idea_lower.contains("master")
        || idea_lower.contains("skill")
        || idea_lower.contains("training")
        || idea_lower.contains("certification")
    {
        return "learning".to_string();
    }

    // Personal/Productivity ideas
    if idea_lower.contains("habit")
        || idea_lower.contains("personal")
        || idea_lower.contains("productivity")
        || idea_lower.contains("health")
        || idea_lower.contains("fitness")
        || idea_lower.contains("life")
        || idea_lower.contains("routine")
        || idea_lower.contains("system")
    {
        return "personal".to_string();
    }

    "general".to_string()
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::commands::analyze_llm::LlmProvider;

    #[test]
    fn test_idea_type_classification() {
        assert_eq!(
            classify_idea_type("Create a mobile app for tracking habits"),
            "technical"
        );
        assert_eq!(
            classify_idea_type("Start a consulting business"),
            "business"
        );
        assert_eq!(
            classify_idea_type("Write a blog about Rust programming"),
            "content"
        );
        assert_eq!(
            classify_idea_type("Master advanced Rust patterns"),
            "learning"
        );
        assert_eq!(classify_idea_type("Build a morning routine"), "personal");
        assert_eq!(classify_idea_type("Random idea"), "general");
    }

    #[test]
    fn test_dynamic_prompt_builder() {
        let builder = DynamicPromptBuilder::new();

        let prompt = builder.build_prompt(
            "Create a Rust CLI tool",
            "technical",
            &LlmProvider::Ollama,
            false,
            false,
        );

        assert!(prompt.contains("Rust CLI tool"));
        assert!(prompt.contains("IDEA TYPE: Technical Project"));
        assert!(!prompt.contains("EXAMPLE ANALYSES"));
        assert!(!prompt.contains("ANALYSIS APPROACH"));
    }

    #[test]
    fn test_prompt_with_examples_and_cot() {
        let builder = DynamicPromptBuilder::new();

        let prompt = builder.build_prompt(
            "Learn advanced Rust",
            "learning",
            &LlmProvider::Claude,
            true,
            true,
        );

        assert!(prompt.contains("Learn advanced Rust"));
        assert!(prompt.contains("EXAMPLE ANALYSES"));
        assert!(prompt.contains("ANALYSIS APPROACH"));
        assert!(prompt.contains("step-by-step approach"));
    }

    #[test]
    fn test_example_library() {
        let library = ExampleLibrary::new();

        let examples = library.get_examples("technical", 1);
        assert!(!examples.is_empty());

        let examples = library.get_examples("nonexistent", 5);
        assert!(examples.is_empty());
    }

    #[test]
    fn test_template_performance_tracking() {
        let mut builder = DynamicPromptBuilder::new();

        // Update performance metrics
        builder.update_template_performance("technical", 0.8, 0.9, true, 500);
        builder.update_template_performance("technical", 0.6, 0.7, false, 800);

        let perf = builder.get_template_performance("technical").unwrap();
        assert_eq!(perf.usage_count, 2);
        assert!(perf.avg_quality_score > 0.6 && perf.avg_quality_score < 0.8);
        assert!(perf.avg_confidence > 0.7 && perf.avg_confidence < 0.9);
        assert_eq!(perf.success_rate, 0.5);
    }

    #[test]
    fn test_provider_type_inclusion() {
        let builder = DynamicPromptBuilder::new();

        let prompt =
            builder.build_prompt("Test idea", "general", &LlmProvider::OpenAi, false, false);

        assert!(prompt.contains("OpenAI"));
    }
}
