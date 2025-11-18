use crate::errors::{ApplicationError, Result, ScoringError};
use regex::Regex;
use serde::{Deserialize, Serialize};

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Score {
    pub mission: MissionScores,
    pub anti_challenge: AntiChallengeScores,
    pub strategic: StrategicScores,
    pub raw_score: f64,
    pub final_score: f64,
    pub recommendation: Recommendation,
    pub scoring_details: Vec<String>,
    pub explanations: std::collections::HashMap<String, String>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct MissionScores {
    pub domain_expertise: f64,  // 0-1.2 points max
    pub ai_alignment: f64,      // 0-1.5 points max
    pub execution_support: f64, // 0-0.8 points max
    pub revenue_potential: f64, // 0-0.5 points max
    pub total: f64,             // max 4.0 points
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct AntiChallengeScores {
    pub context_switching: f64, // 0-1.2 points max
    pub rapid_prototyping: f64, // 0-1.0 points max
    pub accountability: f64,    // 0-0.8 points max
    pub income_anxiety: f64,    // 0-0.5 points max
    pub total: f64,             // max 3.5 points
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct StrategicScores {
    pub stack_compatibility: f64,   // 0-1.0 points max
    pub shipping_habit: f64,        // 0-0.8 points max
    pub public_accountability: f64, // 0-0.4 points max
    pub revenue_testing: f64,       // 0-0.3 points max
    pub total: f64,                 // max 2.5 points
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum Recommendation {
    Priority,
    Good,
    Consider,
    Avoid,
}

impl Recommendation {
    pub fn emoji(&self) -> &'static str {
        match self {
            Recommendation::Priority => "🔥",
            Recommendation::Good => "✅",
            Recommendation::Consider => "⚠️",
            Recommendation::Avoid => "🚫",
        }
    }

    pub fn text(&self) -> &'static str {
        match self {
            Recommendation::Priority => "PRIORITIZE NOW",
            Recommendation::Good => "GOOD ALIGNMENT",
            Recommendation::Consider => "CONSIDER LATER",
            Recommendation::Avoid => "AVOID FOR NOW",
        }
    }
}

#[derive(Clone)]
pub struct ScoringEngine {
    telos_config: TelosConfig,
    patterns: ScoringPatterns,
}

#[derive(Debug, Clone)]
pub struct TelosConfig {
    pub current_stack: Vec<String>,
    pub domain_keywords: Vec<String>,
    pub income_deadline: String,
    pub active_goals: Vec<String>,
    pub active_strategies: Vec<String>,
    pub challenges: Vec<String>,
}

#[derive(Debug, Clone)]
struct ScoringPatterns {
    ai_positive: Vec<Regex>,
    ai_negative: Vec<Regex>,
    shipping_positive: Vec<Regex>,
    shipping_negative: Vec<Regex>,
    income_positive: Vec<Regex>,
    domain_positive: Vec<Regex>,
    stack_violations: Vec<Regex>,
    perfectionism_negative: Vec<Regex>,
    perfectionism_positive: Vec<Regex>,
    accountability_positive: Vec<Regex>,
    procrastination_negative: Vec<Regex>,
}

impl ScoringEngine {
    pub async fn new() -> Result<Self> {
        // For backward compatibility during transition, try loading with defaults first
        // but log that this usage is deprecated
        eprintln!("Warning: Using deprecated ScoringEngine::new(), please use ScoringEngine::with_config() instead");
        let config_paths = match crate::config::ConfigPaths::load() {
            Ok(paths) => paths,
            Err(e) => {
                eprintln!(
                    "Warning: Failed to load configuration, using defaults: {}",
                    e
                );
                return Ok(ScoringEngine::with_default_config().await);
            }
        };

        Self::with_config(&config_paths).await
    }

    pub async fn with_config(config_paths: &crate::config::ConfigPaths) -> Result<Self> {
        let telos_config = Self::load_telos_config_from_path(config_paths).await?;
        let patterns = Self::compile_patterns()?;

        Ok(ScoringEngine {
            telos_config,
            patterns,
        })
    }

    /// Create a ScoringEngine with a specific TelosConfig (mainly for testing)
    pub fn from_telos_config(telos_config: TelosConfig) -> Result<Self> {
        let patterns = Self::compile_patterns()?;
        Ok(ScoringEngine {
            telos_config,
            patterns,
        })
    }

    async fn load_telos_config_from_path(
        config_paths: &crate::config::ConfigPaths,
    ) -> Result<TelosConfig> {
        // Create TelosParser from config
        let parser = crate::telos::TelosParser::from_config(config_paths);

        // Try to parse from actual Telos.md first
        match parser.parse_for_scoring().await {
            Ok(config) => Ok(config),
            Err(e) => {
                // Log the error and fallback to hardcoded values from Ray's Telos
                eprintln!(
                    "Warning: Failed to load Telos config from {:?}, using defaults: {}",
                    config_paths.telos_file, e
                );
                Ok(TelosConfig {
                    current_stack: vec![
                        "python".to_string(),
                        "langchain".to_string(),
                        "openai".to_string(),
                        "gpt".to_string(),
                        "api".to_string(),
                        "streamlit".to_string(),
                        "web app".to_string(),
                    ],
                    domain_keywords: vec![
                        "hotel".to_string(),
                        "hospitality".to_string(),
                        "hilton".to_string(),
                        "mobile".to_string(),
                        "android".to_string(),
                        "app".to_string(),
                    ],
                    income_deadline: "2026-01-15".to_string(),
                    active_goals: vec![
                        "Generate First AI-Related Income".to_string(),
                        "Ship 2 Public AI Projects".to_string(),
                        "Build Working Personal Augmentation System".to_string(),
                        "Consistent Public Building".to_string(),
                    ],
                    active_strategies: vec![
                        "One Stack, One Month".to_string(),
                        "Shitty First Draft".to_string(),
                        "Public Accountability".to_string(),
                        "Revenue Testing".to_string(),
                    ],
                    challenges: vec![
                        "Chronic Context-Switching".to_string(),
                        "Fear of Imperfect Work".to_string(),
                        "No External Accountability".to_string(),
                        "Unclear Path to Income".to_string(),
                    ],
                })
            }
        }
    }

    async fn with_default_config() -> Self {
        // Only used as fallback when config loading fails
        let patterns = match Self::compile_patterns() {
            Ok(p) => p,
            Err(_) => ScoringPatterns::default_patterns(),
        };

        ScoringEngine {
            telos_config: TelosConfig {
                current_stack: vec![
                    "python".to_string(),
                    "langchain".to_string(),
                    "openai".to_string(),
                    "gpt".to_string(),
                    "api".to_string(),
                    "streamlit".to_string(),
                    "web app".to_string(),
                ],
                domain_keywords: vec![
                    "hotel".to_string(),
                    "hospitality".to_string(),
                    "hilton".to_string(),
                    "mobile".to_string(),
                    "android".to_string(),
                    "app".to_string(),
                ],
                income_deadline: "2026-01-15".to_string(),
                active_goals: vec![
                    "Generate First AI-Related Income".to_string(),
                    "Ship 2 Public AI Projects".to_string(),
                    "Build Working Personal Augmentation System".to_string(),
                    "Consistent Public Building".to_string(),
                ],
                active_strategies: vec![
                    "One Stack, One Month".to_string(),
                    "Shitty First Draft".to_string(),
                    "Public Accountability".to_string(),
                    "Revenue Testing".to_string(),
                ],
                challenges: vec![
                    "Chronic Context-Switching".to_string(),
                    "Fear of Imperfect Work".to_string(),
                    "No External Accountability".to_string(),
                    "Unclear Path to Income".to_string(),
                ],
            },
            patterns,
        }
    }

    fn compile_patterns() -> Result<ScoringPatterns> {
        macro_rules! compile_regex {
            ($pattern:expr) => {
                Regex::new($pattern).map_err(|e| ScoringError::pattern_compilation($pattern, e))
            };
        }

        Ok(ScoringPatterns {
            ai_positive: vec![
                compile_regex!(
                    r"\b(build|create|develop|implement|ship)\s+.*\b(ai|AI|artificial intelligence)\b"
                )?,
                compile_regex!(
                    r"\b(ai|AI|artificial intelligence)\s+(tool|system|app|application|platform)\b"
                )?,
                compile_regex!(r"\b(automate|automation)\s+.*\b(with|using)\s+.*\b(ai|AI)\b")?,
            ],
            ai_negative: vec![
                compile_regex!(
                    r"\b(learn|study|explore|research|understand)\s+.*\b(ai|AI|machine learning|ML)\b"
                )?,
                compile_regex!(r"\b(ai|AI)\s+(tutorial|course|book|documentation)\b")?,
            ],
            shipping_positive: vec![
                compile_regex!(
                    r"\b(ship|launch|release|publish|deploy)\s+(by|before|on)\s+.*\d{4}-\d{2}-\d{2}"
                )?,
                compile_regex!(
                    r"\b(MVP|prototype|v1|minimum.*viable|basic.*version|simple.*implementation)\b"
                )?,
                compile_regex!(r"\b(this\sweek|by\s.*\d{1,2}|next\sweek)\b")?,
            ],
            shipping_negative: vec![
                compile_regex!(
                    r"\b(plan|design|outline|specification|research)\s+(fully|completely|thoroughly)\b"
                )?,
                compile_regex!(
                    r"\b(perfect|complete|comprehensive|full-featured)\s+(implementation|solution|system)\b"
                )?,
            ],
            income_positive: vec![
                compile_regex!(r"\$\d+")?, // Any dollar amount
                compile_regex!(
                    r"\b(freelance|consult|sell|monetize|revenue|income|client|customer)\b"
                )?,
                compile_regex!(r"\b(paid|charging|pricing|business)\b")?,
            ],
            domain_positive: vec![
                compile_regex!(r"\b(hotel|hospitality|Hilton|guest|reservation|booking)\b")?,
                compile_regex!(r"\b(mobile|Android|iOS|app|application)\b")?,
                compile_regex!(r"\b(software|development|programming|code|tech)\b")?,
            ],
            stack_violations: vec![
                compile_regex!(
                    r"\b(Rust|JavaScript|TypeScript|React|Vue|Angular|Flutter|Swift|Kotlin|Go|Elixir|Phoenix)\b"
                )?,
                compile_regex!(r"\b(mobile|android|ios)\s+(app|development)\b")?,
                compile_regex!(
                    r"\b(new|latest|just\sdiscovered|came\sacross)\s+(framework|library|technology|tech)\b"
                )?,
            ],
            perfectionism_negative: vec![
                compile_regex!(
                    r"\b(complete|comprehensive|full-featured|enterprise-grade|production-ready|scalable|robust|perfect|flawless)\b"
                )?,
                compile_regex!(
                    r"\b(build.*from\sscratch|custom.*implementation|reimplement.*wheel|own.*version)\b"
                )?,
            ],
            perfectionism_positive: vec![
                compile_regex!(
                    r"\b(MVP|prototype|v1|minimum.*viable|basic.*version|simple.*implementation|quick.*prototype)\b"
                )?,
                compile_regex!(r"\b(start.*simple|begin.*basic|iterative|feedback.*loop)\b")?,
            ],
            accountability_positive: vec![
                compile_regex!(
                    r"\b(public|share|demo|launch|release|publish|GitHub|Twitter|show.*someone)\b"
                )?,
                compile_regex!(
                    r"\b(client.*deadline|customer.*need|business.*requirement|deliverable)\b"
                )?,
                compile_regex!(r"\b(team|collaborate|with.*someone|for.*someone)\b")?,
            ],
            procrastination_negative: vec![
                compile_regex!(
                    r"\b(learn|study|explore|research|investigate|understand|master|get.*familiar)\b.*(before.*start|then.*build|once.*I.*know)\b"
                )?,
                compile_regex!(
                    r"\b(need.*to|have.*to|should.*probably|might.*as.*well)\s+(learn|study|research|explore)\b"
                )?,
                compile_regex!(
                    r"\b(someday|eventually|one\sday|when.*I.*have.*time|after.*I.*finish)\b"
                )?,
            ],
        })
    }

    pub fn calculate_score(&self, idea: &str) -> Result<Score> {
        // Validate input
        if idea.trim().is_empty() {
            return Err(ApplicationError::Scoring(ScoringError::EmptyContent));
        }

        if idea.len() > 100000 {
            return Err(ApplicationError::Scoring(ScoringError::too_long(
                idea.len(),
                100000,
            )));
        }

        let idea_lower = idea.to_lowercase();

        // Mission Alignment (40% weight)
        let mission = self.score_mission(&idea_lower, idea);

        // Anti-Challenge (35% weight)
        let anti_challenge = self.score_anti_challenge(&idea_lower, idea);

        // Strategic Fit (25% weight)
        let strategic = self.score_strategic(&idea_lower, idea);

        // Calculate totals
        let raw_score = mission.total + anti_challenge.total + strategic.total;
        let final_score = (raw_score / 10.0) * 10.0; // Scale to 0-10

        // Validate final score is within expected range
        if !(0.0..=10.0).contains(&final_score) {
            return Err(ApplicationError::Scoring(
                ScoringError::invalid_score_range(final_score),
            ));
        }

        let recommendation = match final_score {
            s if s >= 8.5 => Recommendation::Priority,
            s if s >= 7.0 => Recommendation::Good,
            s if s >= 5.0 => Recommendation::Consider,
            _ => Recommendation::Avoid,
        };

        let mut scoring_details = Vec::new();
        scoring_details.push(format!("Mission Alignment: {:.2}/4.00", mission.total));
        scoring_details.push(format!("Anti-Challenge: {:.2}/3.50", anti_challenge.total));
        scoring_details.push(format!("Strategic Fit: {:.2}/2.50", strategic.total));

        // Generate detailed explanations for each sub-criterion
        let explanations =
            self.generate_explanations(idea, &idea_lower, &mission, &anti_challenge, &strategic);

        Ok(Score {
            mission,
            anti_challenge,
            strategic,
            raw_score,
            final_score,
            recommendation,
            scoring_details,
            explanations,
        })
    }

    // Deprecated: Use calculate_score instead which returns a Result
    // This method is kept for backward compatibility but should not be used
    #[deprecated(
        since = "0.2.0",
        note = "Use calculate_score instead which returns a proper Result"
    )]
    #[allow(dead_code)]
    pub fn calculate_score_unsafe(&self, idea: &str) -> Score {
        // If this is ever called, return a default Score rather than panicking
        self.calculate_score(idea).unwrap_or_else(|_| Score {
            mission: MissionScores {
                domain_expertise: 0.0,
                ai_alignment: 0.0,
                execution_support: 0.0,
                revenue_potential: 0.0,
                total: 0.0,
            },
            anti_challenge: AntiChallengeScores {
                context_switching: 0.0,
                rapid_prototyping: 0.0,
                accountability: 0.0,
                income_anxiety: 0.0,
                total: 0.0,
            },
            strategic: StrategicScores {
                stack_compatibility: 0.0,
                shipping_habit: 0.0,
                public_accountability: 0.0,
                revenue_testing: 0.0,
                total: 0.0,
            },
            raw_score: 0.0,
            final_score: 0.0,
            recommendation: Recommendation::Avoid,
            scoring_details: vec!["Error: Scoring failed".to_string()],
            explanations: std::collections::HashMap::new(),
        })
    }

    fn score_mission(&self, idea_lower: &str, idea: &str) -> MissionScores {
        // Domain Expertise Leverage (1.2 points max)
        let domain_expertise = self.score_domain_expertise(idea_lower, idea);

        // AI Systems/Building Alignment (1.5 points max)
        let ai_alignment = self.score_ai_alignment(idea_lower);

        // Shipping Bias (0.8 points max)
        let execution_support = self.score_execution_support(idea_lower);

        // Revenue Potential (0.5 points max)
        let revenue_potential = self.score_revenue_potential(idea_lower);

        MissionScores {
            domain_expertise,
            ai_alignment,
            execution_support,
            revenue_potential,
            total: domain_expertise + ai_alignment + execution_support + revenue_potential,
        }
    }

    fn score_anti_challenge(&self, idea_lower: &str, idea: &str) -> AntiChallengeScores {
        // Tech Stack Continuity (1.2 points max)
        let context_switching = self.score_context_switching(idea_lower, idea);

        // Rapid Prototyping Design (1.0 points max)
        let rapid_prototyping = self.score_rapid_prototyping(idea_lower);

        // Built-in Accountability (0.8 points max)
        let accountability = self.score_accountability(idea_lower);

        // Income Anxiety Relief (0.5 points max)
        let income_anxiety = self.score_income_anxiety(idea_lower);

        AntiChallengeScores {
            context_switching,
            rapid_prototyping,
            accountability,
            income_anxiety,
            total: context_switching + rapid_prototyping + accountability + income_anxiety,
        }
    }

    fn score_strategic(&self, idea_lower: &str, idea: &str) -> StrategicScores {
        // Execution Compatibility (1.0 points max)
        let stack_compatibility = self.score_stack_compatibility(idea_lower, idea);

        // Compounding Benefits (0.8 points max)
        let shipping_habit = self.score_shipping_habit(idea_lower);

        // Validation Speed (0.4 points max)
        let public_accountability = self.score_public_accountability(idea_lower);

        // Scalability Potential (0.3 points max)
        let revenue_testing = self.score_revenue_testing(idea_lower);

        StrategicScores {
            stack_compatibility,
            shipping_habit,
            public_accountability,
            revenue_testing,
            total: stack_compatibility + shipping_habit + public_accountability + revenue_testing,
        }
    }

    fn score_dimension(
        &self,
        positive_patterns: &[Regex],
        negative_patterns: &[Regex],
        text: &str,
        max_points: f64,
    ) -> f64 {
        let positive_matches: usize = positive_patterns
            .iter()
            .map(|p| p.find_iter(text).count())
            .sum();

        let negative_matches: usize = negative_patterns
            .iter()
            .map(|p| p.find_iter(text).count())
            .sum();

        if negative_matches > 0 {
            (max_points * 0.2).min(max_points) // Heavy penalty for negative matches
        } else if positive_matches > 0 {
            max_points // Full points for positive matches
        } else {
            max_points * 0.5 // Neutral score
        }
    }

    fn score_domain_expertise(&self, idea_lower: &str, _idea: &str) -> f64 {
        // 0.90-1.20: Directly uses 80%+ existing skills (n8n, Python, AI automation, Android dev concepts)
        // 0.60-0.89: Uses 50-79% existing skills; minor learning required
        // 0.30-0.59: Uses 30-49% existing skills; significant new learning needed
        // 0.00-0.29: Requires mostly new skills outside current domain

        let mut matching_skills = 0;
        let total_skills = self.telos_config.current_stack.len();

        for tech in &self.telos_config.current_stack {
            if idea_lower.contains(tech) {
                matching_skills += 1;
            }
        }

        if total_skills == 0 {
            return 0.3; // Default to minimal if no skills defined
        }

        let match_ratio = matching_skills as f64 / total_skills as f64;

        if match_ratio >= 0.8 {
            0.9 + ((match_ratio - 0.8) * 0.75) // 0.9-1.2 range: 0.3 points over 0.2 ratio
        } else if match_ratio >= 0.5 {
            0.6 + ((match_ratio - 0.5) * 0.967) // 0.6-0.89 range: 0.29 points over 0.3 ratio
        } else if match_ratio >= 0.3 {
            0.3 + ((match_ratio - 0.3) * 0.967) // 0.3-0.59 range: 0.29 points over 0.2 ratio
        } else {
            match_ratio * 1.033 // 0.0-0.29 range: 0.29 points over 0.3 ratio
        }
    }

    fn score_ai_alignment(&self, idea_lower: &str) -> f64 {
        // 1.20-1.50: Core product IS AI automation/systems (e.g., AI agents, automation pipelines)
        // 0.80-1.19: AI is a significant component but not the core offering
        // 0.40-0.79: AI is auxiliary or optional to the main offering
        // 0.00-0.39: Minimal or no AI component

        let ai_core_keywords = [
            r"\b(artificial intelligence|AI|machine learning|ML|LLM|GPT|AI agent|AI system|automation|automated system|AI pipeline|AI workflow|AI infrastructure)\b",
            r"\b(build|create|develop)\s+.*\b(AI|artificial intelligence|machine learning|automation)\b",
            r"\b(AI|automation)\s+(tool|platform|system|agent|assistant|service|product|solution)\b",
        ];

        let ai_significant_keywords = [
            r"\b(integrate|use|implement|leverage|utilize)\s+.*\b(AI|artificial intelligence|GPT|ML|machine learning|LLM)\b",
            r"\b(with|using|powered by|driven by)\s+.*\b(AI|artificial intelligence|GPT|ML|machine learning|LLM)\b",
        ];

        let ai_auxiliary_keywords = [
            r"\b(AI|artificial intelligence|GPT|ML|LLM)\b",
            r"\b(smart|intelligent|automated)\s+(feature|functionality|process|aspect|component)\b",
        ];

        let mut core_matches = 0;
        for pattern in &ai_core_keywords {
            if let Ok(re) = Regex::new(pattern) {
                if re.is_match(idea_lower) {
                    core_matches += 1;
                }
            }
        }

        if core_matches > 0 {
            return 1.2 + ((core_matches as f64).min(3.0) * 0.1); // 1.2-1.5 range (max 3 matches * 0.1 each)
        }

        let mut significant_matches = 0;
        for pattern in &ai_significant_keywords {
            if let Ok(re) = Regex::new(pattern) {
                if re.is_match(idea_lower) {
                    significant_matches += 1;
                }
            }
        }

        if significant_matches > 0 {
            return 0.8 + ((significant_matches as f64).min(4.0) * 0.0975); // 0.8-1.19 range (max 4 matches * 0.0975 each)
        }

        let mut auxiliary_matches = 0;
        for pattern in &ai_auxiliary_keywords {
            if let Ok(re) = Regex::new(pattern) {
                if re.is_match(idea_lower) {
                    auxiliary_matches += 1;
                }
            }
        }

        if auxiliary_matches > 0 {
            return 0.4 + ((auxiliary_matches as f64).min(4.0) * 0.0975); // 0.4-0.79 range (max 4 matches * 0.0975 each)
        }

        0.0 // No AI component
    }

    fn score_execution_support(&self, idea_lower: &str) -> f64 {
        // 0.65-0.80: Idea has clear deliverable within 30 days; success = shipped product
        // 0.45-0.64: Deliverable within 60 days with defined MVP scope
        // 0.25-0.44: Longer timeline (90+ days) or unclear MVP definition
        // 0.00-0.24: Primarily learning-focused with no concrete deliverable

        let mvp_keywords = [
            r"\b(MVP|minimum viable product|prototype|v1|basic version|simple implementation|proof of concept|POC|minimum viable|basic implementation|first version)\b",
            r"\b(30\s*days?|1\s*month|within 1\s*month|by next month|within 30 days)\b",
            r"\b(60\s*days?|2\s*months|within 2\s*months|within 60 days)\b",
            r"\b(90\s*days?|3\s*months|within 3\s*months|within 90 days)\b",
        ];

        let learning_keywords = [
            r"\b(learn|study|research|explore|understand|master|get familiar|tutorial|course|certification|how to learn|learning journey)\b",
            r"\b(how to|learning|study guide|research project|educational|for learning|to learn|about learning)\b",
        ];

        let mut mvp_matches = 0;
        for pattern in &mvp_keywords {
            if let Ok(re) = Regex::new(pattern) {
                if re.is_match(idea_lower) {
                    mvp_matches += 1;
                }
            }
        }

        let mut learning_matches = 0;
        for pattern in &learning_keywords {
            if let Ok(re) = Regex::new(pattern) {
                if re.is_match(idea_lower) {
                    learning_matches += 1;
                }
            }
        }

        if learning_matches > 0 && mvp_matches == 0 {
            return (0.0_f64).max(0.24 - (learning_matches as f64 * 0.05)); // Learning-focused
        }

        if idea_lower.contains("30 day")
            || idea_lower.contains("1 month")
            || idea_lower.contains("within 1 month")
            || idea_lower.contains("within 30 days")
        {
            return 0.65 + ((mvp_matches as f64).min(3.0) * 0.05); // 0.65-0.80 range (max 3 matches * 0.05 each)
        } else if idea_lower.contains("60 day")
            || idea_lower.contains("2 months")
            || idea_lower.contains("within 60 days")
        {
            return 0.45 + ((mvp_matches as f64).min(4.0) * 0.0475); // 0.45-0.64 range (max 4 matches * 0.0475 each)
        } else if idea_lower.contains("90 day")
            || idea_lower.contains("3 months")
            || idea_lower.contains("within 90 days")
        {
            return 0.25 + ((mvp_matches as f64).min(4.0) * 0.0475); // 0.25-0.44 range (max 4 matches * 0.0475 each)
        } else if mvp_matches > 0 {
            return 0.45 + ((mvp_matches as f64).min(7.0) * 0.0214); // Somewhere in the middle (max 7 matches * 0.0214 each)
        }

        0.25 // Default to middle range if unclear
    }

    fn score_revenue_potential(&self, idea_lower: &str) -> f64 {
        // 0.40-0.50: Clear monetization model; proven market willing to pay ($1K-$2.5K/month target)
        // 0.25-0.39: Plausible monetization; requires validation; likely $500-$1K/month
        // 0.10-0.24: Speculative monetization; unclear market willingness to pay
        // 0.00-0.09: No clear revenue path or ad-based/very low revenue model

        let high_revenue_keywords = [
            r"\$(1000|1500|2000|2500|3000|4000|5000)[+]?|([1-9][0-9])00\s*(per month|monthly|a month|per\s+month|revenue target)",
            r"\b(subscription|SaaS|recurring revenue|monthly|recurring payment|retainer|recurring billing)\b",
            r"\b(business|enterprise|B2B|client|customer|service|consulting|sales|revenue|income|monetization|pricing|pricing model)\b",
        ];

        let medium_revenue_keywords = [
            r"\$(500|600|700|800|900)[+]?|([5-9][0-9])0\s*(per month|monthly|a month|revenue|income)",
            r"\b(freelance|project based|one time|custom|bespoke|contract|consulting|hourly|per project)\b",
        ];

        let low_revenue_keywords = [
            r"\b(ads|advertising|affiliate|sponsor|donation|tips?|gratuity|optional payment)\b",
            r"\b(free|gratis|no cost|open source|community|public|free to use|ad supported)\b",
        ];

        let mut high_matches = 0;
        for pattern in &high_revenue_keywords {
            if let Ok(re) = Regex::new(pattern) {
                if re.is_match(idea_lower) {
                    high_matches += 1;
                }
            }
        }

        if high_matches > 0 {
            return 0.4 + ((high_matches as f64).min(2.0) * 0.05); // 0.4-0.5 range (max 2 matches * 0.05 each)
        }

        let mut medium_matches = 0;
        for pattern in &medium_revenue_keywords {
            if let Ok(re) = Regex::new(pattern) {
                if re.is_match(idea_lower) {
                    medium_matches += 1;
                }
            }
        }

        if medium_matches > 0 {
            return 0.25 + ((medium_matches as f64).min(2.0) * 0.07); // 0.25-0.39 range (max 2 matches * 0.07 each)
        }

        let mut low_matches = 0;
        for pattern in &low_revenue_keywords {
            if let Ok(re) = Regex::new(pattern) {
                if re.is_match(idea_lower) {
                    low_matches += 1;
                }
            }
        }

        if low_matches > 0 {
            return (0.0_f64).max(0.24 - (low_matches as f64 * 0.08)); // 0.0-0.24 range (higher penalty for low revenue)
        }

        0.1 // Default to speculative if unclear
    }

    fn score_context_switching(&self, idea_lower: &str, idea: &str) -> f64 {
        // 0.95-1.20: Uses 90%+ current stack (n8n/Python/AI APIs); no new languages/platforms
        // 0.95-1.20: Uses 90%+ current stack (n8n/Python/AI APIs); no new languages/platforms
        // 0.65-0.94: Uses 70-89% current stack; minor additions (new library, familiar framework)
        // 0.30-0.64: Requires 50%+ new stack elements (new language, platform, or paradigm)
        // 0.00-0.29: Complete stack switch (e.g., mobile dev, game development, hardware)

        let stack_violation_keywords = [
            r"\b(Rust|JavaScript|TypeScript|React|Vue|Angular|Flutter|Swift|Kotlin|Go|Elixir|Phoenix|Unity|Unreal|C\+\+|C#|Java|PHP|Ruby|Dart|Node\.js|Express|Django|Rails|Laravel|Spring|ASP\.NET)\b",
            r"\b(mobile|android|ios)\s+(app|development|native|mobile dev|mobile development)\b",
            r"\b(new|latest|just\sdiscovered|came\sacross|want to try|experiment with)\s+(framework|library|technology|tech|language|platform|tool)\b",
            r"\b(game|gaming|VR|AR|hardware|embedded|IoT|blockchain|web3|smart contract|solidity|ethereum)\b",
        ];

        // Check for stack violations first
        for pattern in &stack_violation_keywords {
            if let Ok(re) = Regex::new(pattern) {
                if re.is_match(idea) {
                    return (0.0_f64).max(0.29 - (re.find_iter(idea).count() as f64 * 0.05));
                    // 0.0-0.29 range
                }
            }
        }

        // Count matching skills
        let mut matching_skills = 0;
        let total_skills = self.telos_config.current_stack.len();

        for tech in &self.telos_config.current_stack {
            if idea_lower.contains(tech) {
                matching_skills += 1;
            }
        }

        if total_skills == 0 {
            return 0.65; // Default to middle if no skills defined
        }

        let match_ratio = matching_skills as f64 / total_skills as f64;

        if match_ratio >= 0.9 {
            0.95 + ((match_ratio - 0.9) * 0.833) // 0.95-1.2 range: 0.25 points over 0.1 ratio
        } else if match_ratio >= 0.7 {
            0.65 + ((match_ratio - 0.7) * 0.967) // 0.65-0.94 range: 0.29 points over 0.2 ratio
        } else if match_ratio >= 0.5 {
            0.3 + ((match_ratio - 0.5) * 0.68) // 0.3-0.64 range: 0.34 points over 0.2 ratio
        } else {
            match_ratio * 0.58 // 0.0-0.29 range: 0.29 points over 0.5 ratio
        }
    }

    fn score_rapid_prototyping(&self, idea_lower: &str) -> f64 {
        // 0.80-1.00: Can ship functional MVP in 1-2 weeks; inherently iterative (SaaS, automation)
        // 0.55-0.79: MVP possible in 3-4 weeks; some iteration possible
        // 0.25-0.54: Requires 6+ weeks for minimal viable version; high quality bar needed
        // 0.00-0.24: Inherently perfection-dependent (content creation, courses, books)

        let rapid_keywords = [
            r"\b(1\s*week|2\s*weeks|within 2 weeks|quick|fast|rapid|immediate|soon|early|early prototype|early MVP)\b",
            r"\b(SaaS|automation|tool|web app|API|dashboard|script|bot|workflow|pipeline|integration|automation tool|webhook|trigger|action)\b",
        ];

        let moderate_keywords = [
            r"\b(3\s*weeks|4\s*weeks|1\s*month|within 1 month|within 4 weeks|moderate timeline)\b",
            r"\b(app|application|platform|system|software|web application|mobile app|desktop app)\b",
        ];

        let slow_keywords = [
            r"\b(6\s*weeks|8\s*weeks|2\s*months|long|extended|comprehensive|complete|thorough|detailed|extensive)\b",
            r"\b(course|curriculum|book|guide|manual|documentation|tutorial|content|article|blog|writing|ebook|course creation)\b",
        ];

        let mut rapid_matches = 0;
        for pattern in &rapid_keywords {
            if let Ok(re) = Regex::new(pattern) {
                if re.is_match(idea_lower) {
                    rapid_matches += 1;
                }
            }
        }

        if rapid_matches > 0 {
            return 0.8 + ((rapid_matches as f64).min(2.0) * 0.1); // 0.8-1.0 range (max 2 matches * 0.1 each)
        }

        let mut moderate_matches = 0;
        for pattern in &moderate_keywords {
            if let Ok(re) = Regex::new(pattern) {
                if re.is_match(idea_lower) {
                    moderate_matches += 1;
                }
            }
        }

        if moderate_matches > 0 {
            return 0.55 + ((moderate_matches as f64).min(3.0) * 0.08); // 0.55-0.79 range (max 3 matches * 0.08 each)
        }

        let mut slow_matches = 0;
        for pattern in &slow_keywords {
            if let Ok(re) = Regex::new(pattern) {
                if re.is_match(idea_lower) {
                    slow_matches += 1;
                }
            }
        }

        if slow_matches > 0 {
            return 0.25 + ((slow_matches as f64).min(6.0) * 0.0483); // 0.25-0.54 range (max 6 matches * 0.0483 each)
        }

        0.55 // Default to moderate
    }

    fn score_accountability(&self, idea_lower: &str) -> f64 {
        // 0.65-0.80: Paying customers or public commitments with consequences (e.g., pre-sales, cohort)
        // 0.45-0.64: Strong accountability structure (accountability partner, public building, deadlines)
        // 0.20-0.44: Weak accountability (social media updates, personal goals)
        // 0.00-0.19: No external accountability; purely self-motivated

        let strong_accountability_keywords = [
            r"\b(customer|client|pre-order|pre-sale|cohort|group|team|partner|collaborator|stakeholder|sponsor|investor|backer)\b",
            r"\b(pay|payment|fee|subscription|revenue|income|sales|contract|agreement|commitment with consequences)\b",
        ];

        let moderate_accountability_keywords = [
            r"\b(accountability partner|buddy|check-in|deadline|commitment|promise|public|tweet|post|update|build in public|progress report)\b",
            r"\b(weekly|daily|regular)\s+check-in|report|update|progress tracking|milestone|deadline",
            r"\b(public commitment|social commitment|shared goal|community accountability)\b",
        ];

        let weak_accountability_keywords = [
            r"\b(personal goal|self-imposed|my plan|I want|I will|personal project|for myself|private project)\b",
            r"\b(maybe|perhaps|might|could|someday|eventually|when I have time)\b",
        ];

        let mut strong_matches = 0;
        for pattern in &strong_accountability_keywords {
            if let Ok(re) = Regex::new(pattern) {
                if re.is_match(idea_lower) {
                    strong_matches += 1;
                }
            }
        }

        if strong_matches > 0 {
            return 0.65 + ((strong_matches as f64).min(3.0) * 0.05); // 0.65-0.8 range (max 3 matches * 0.05 each)
        }

        let mut moderate_matches = 0;
        for pattern in &moderate_accountability_keywords {
            if let Ok(re) = Regex::new(pattern) {
                if re.is_match(idea_lower) {
                    moderate_matches += 1;
                }
            }
        }

        if moderate_matches > 0 {
            return 0.45 + ((moderate_matches as f64).min(3.0) * 0.063); // 0.45-0.64 range (max 3 matches * 0.063 each)
        }

        let mut weak_matches = 0;
        for pattern in &weak_accountability_keywords {
            if let Ok(re) = Regex::new(pattern) {
                if re.is_match(idea_lower) {
                    weak_matches += 1;
                }
            }
        }

        if weak_matches > 0 {
            return 0.2 + ((weak_matches as f64).min(3.0) * 0.08); // 0.2-0.44 range (max 3 matches * 0.08 each)
        }

        0.1 // Default to low accountability
    }

    fn score_income_anxiety(&self, idea_lower: &str) -> f64 {
        // 0.40-0.50: First revenue possible within 30 days; path to consistent monthly income
        // 0.25-0.39: First revenue within 60 days; recurring revenue model
        // 0.10-0.24: First revenue 90+ days; project-based or unpredictable income
        // 0.00-0.09: Revenue 6+ months away or highly uncertain

        let fast_revenue_keywords = [
            r"\b(30\s*days?|1\s*month|within 1\s*month|immediate|quick|fast|soon|early|first month|within 30 days)\s*(revenue|income|payment|money|cash|profit|first dollar|first sale|first payment)\b",
            r"\b(recurring|subscription|monthly|consistent|regular|predictable)\s*(revenue|income|payment|cash flow)\b",
        ];

        let moderate_revenue_keywords = [
            r"\b(60\s*days?|2\s*months|within 2\s*months|within 60 days|relatively quick|soon after)\s*(revenue|income|payment|money|cash|profit|first revenue)\b",
        ];

        let slow_revenue_keywords = [
            r"\b(90\s*days?|3\s*months|6\s*months|long term|extended|eventually|after launch|post-launch)\s*(revenue|income|payment|money|cash|profit)\b",
        ];

        let uncertain_revenue_keywords = [
            r"\b(maybe|perhaps|possibly|eventually|someday|uncertain|unknown|maybe someday|potentially|hypothetically)\s*(revenue|income|payment|money|cash|profit)\b",
            r"\b(ad|advertising|affiliate|tips?|gratuity|optional|voluntary|donation|crowdfunding)\s*(revenue|income|monetization)\b",
        ];

        let mut fast_matches = 0;
        for pattern in &fast_revenue_keywords {
            if let Ok(re) = Regex::new(pattern) {
                if re.is_match(idea_lower) {
                    fast_matches += 1;
                }
            }
        }

        if fast_matches > 0 {
            return 0.4 + ((fast_matches as f64).min(2.0) * 0.05); // 0.4-0.5 range (max 2 matches * 0.05 each)
        }

        let mut moderate_matches = 0;
        for pattern in &moderate_revenue_keywords {
            if let Ok(re) = Regex::new(pattern) {
                if re.is_match(idea_lower) {
                    moderate_matches += 1;
                }
            }
        }

        if moderate_matches > 0 {
            return 0.25 + ((moderate_matches as f64).min(2.0) * 0.07); // 0.25-0.39 range (max 2 matches * 0.07 each)
        }

        let mut slow_matches = 0;
        for pattern in &slow_revenue_keywords {
            if let Ok(re) = Regex::new(pattern) {
                if re.is_match(idea_lower) {
                    slow_matches += 1;
                }
            }
        }

        if slow_matches > 0 {
            return 0.1 + ((slow_matches as f64).min(3.0) * 0.047); // 0.1-0.24 range (max 3 matches * 0.047 each)
        }

        let mut uncertain_matches = 0;
        for pattern in &uncertain_revenue_keywords {
            if let Ok(re) = Regex::new(pattern) {
                if re.is_match(idea_lower) {
                    uncertain_matches += 1;
                }
            }
        }

        if uncertain_matches > 0 {
            return (0.0_f64).max(0.09 - (uncertain_matches as f64 * 0.01)); // 0.0-0.09 range
        }

        0.1 // Default to slow/uncertain
    }

    fn score_stack_compatibility(&self, idea_lower: &str, _idea: &str) -> f64 {
        // 0.80-1.00: Enables 4+ hour flow sessions; clear systematic execution path
        // 0.55-0.79: Allows 2-3 hour focus blocks; mostly systematic with some ambiguity
        // 0.25-0.54: Requires frequent context switching or has unclear execution steps
        // 0.00-0.24: Inherently fragmented work or chaotic/creative process

        let flow_keywords = [
            r"\b(4\s*hour|4\+ hours?|long session|extended focus|deep work|flow state|uninterrupted|consecutive|long block|focus block)\b",
            r"\b(systematic|structured|methodical|step by step|clear path|defined process|routine|habit|automated|pipeline|workflow)\b",
        ];

        let moderate_focus_keywords = [
            r"\b(2\s*hour|3\s*hour|2-3 hours?|focus block|concentrated work|longer session)\b",
            r"\b(mostly systematic|partially structured|some ambiguity|partially clear|semi-structured)\b",
        ];

        let context_switch_keywords = [
            r"\b(frequent|often|multiple|various|different|switching|juggling|multitasking)\s+(context|task|focus|attention|technology|stack|tool)\b",
            r"\b(unclear|ambiguous|chaotic|creative|artistic|spontaneous|flexible|iterative|experimental)\s+(process|execution|workflow|development|building)\b",
        ];

        let mut flow_matches = 0;
        for pattern in &flow_keywords {
            if let Ok(re) = Regex::new(pattern) {
                if re.is_match(idea_lower) {
                    flow_matches += 1;
                }
            }
        }

        if flow_matches > 0 {
            return 0.8 + ((flow_matches as f64).min(2.0) * 0.1); // 0.8-1.0 range (max 2 matches * 0.1 each)
        }

        let mut moderate_matches = 0;
        for pattern in &moderate_focus_keywords {
            if let Ok(re) = Regex::new(pattern) {
                if re.is_match(idea_lower) {
                    moderate_matches += 1;
                }
            }
        }

        if moderate_matches > 0 {
            return 0.55 + ((moderate_matches as f64).min(3.0) * 0.08); // 0.55-0.79 range (max 3 matches * 0.08 each)
        }

        let mut context_switch_matches = 0;
        for pattern in &context_switch_keywords {
            if let Ok(re) = Regex::new(pattern) {
                if re.is_match(idea_lower) {
                    context_switch_matches += 1;
                }
            }
        }

        if context_switch_matches > 0 {
            return 0.25 + ((context_switch_matches as f64).min(6.0) * 0.0483); // 0.25-0.54 range (max 6 matches * 0.0483 each)
        }

        // Default to stack compatibility based on technology match
        let mut matching_skills = 0;
        for tech in &self.telos_config.current_stack {
            if idea_lower.contains(tech) {
                matching_skills += 1;
            }
        }

        if matching_skills > 0 {
            0.6 + ((matching_skills as f64).min(4.0) * 0.1) // Higher if using current stack (max 4 matches * 0.1 each)
        } else {
            0.3 // Lower if not using current stack
        }
    }

    fn score_shipping_habit(&self, idea_lower: &str) -> f64 {
        // 0.65-0.80: Creates reusable systems, code, or processes for future projects
        // 0.45-0.64: Some reusable components; partial knowledge transfer
        // 0.20-0.44: Minimal reusability; mostly project-specific work
        // 0.00-0.19: Purely one-off effort with no future leverage

        let reusability_keywords = [
            r"\b(reusable|library|module|component|template|framework|system|process|tool|utility|building block|boilerplate|starter|base code|reusable component|modular|plugin|package|API wrapper|SDK|reusable system|reusable process)\b",
            r"\b(can be used|reused|repurposed|adapted|applied to|for future|reusable|modular design|modular architecture|reusable architecture)\b",
        ];

        let partial_reusability_keywords = [
            r"\b(similar|pattern|approach|method|technique|knowledge|skill|experience|insight)\s+(transfer|applicable|useful|helpful|applicable to|transferable|carry over)\b",
            r"\b(learning|insight|understanding|knowledge gained)\s+(applicable|transferable|useful|valuable for|helpful for)\b",
        ];

        let one_off_keywords = [
            r"\b(one-off|single|unique|specific|custom|bespoke|one time|just for this|only for this|one-time|single use|project specific|unique to this|custom solution)\b",
            r"\b(not reusable|not transferable|project specific|unique to|one time use|not modular|not reusable|not modular design)\b",
        ];

        let mut reusable_matches = 0;
        for pattern in &reusability_keywords {
            if let Ok(re) = Regex::new(pattern) {
                if re.is_match(idea_lower) {
                    reusable_matches += 1;
                }
            }
        }

        if reusable_matches > 0 {
            return 0.65 + ((reusable_matches as f64).min(3.0) * 0.05); // 0.65-0.8 range (max 3 matches * 0.05 each)
        }

        let mut partial_matches = 0;
        for pattern in &partial_reusability_keywords {
            if let Ok(re) = Regex::new(pattern) {
                if re.is_match(idea_lower) {
                    partial_matches += 1;
                }
            }
        }

        if partial_matches > 0 {
            return 0.45 + ((partial_matches as f64).min(3.0) * 0.062); // 0.45-0.64 range (max 3 matches * 0.062 each)
        }

        let mut one_off_matches = 0;
        for pattern in &one_off_keywords {
            if let Ok(re) = Regex::new(pattern) {
                if re.is_match(idea_lower) {
                    one_off_matches += 1;
                }
            }
        }

        if one_off_matches > 0 {
            return (0.0_f64).max(0.19 - (one_off_matches as f64 * 0.01)); // 0.0-0.19 range
        }

        0.45 // Default to middle
    }

    fn score_public_accountability(&self, idea_lower: &str) -> f64 {
        // 0.32-0.40: Can validate core assumption within 1-2 weeks (landing page, calls, prototype)
        // 0.22-0.31: Validation possible in 3-4 weeks
        // 0.10-0.21: Requires 6-8 weeks to validate
        // 0.00-0.09: Validation requires 2+ months or full product build

        let fast_validation_keywords = [
            r"\b(1\s*week|2\s*weeks|within 2 weeks|quick|fast|rapid|early|early validation|early feedback|fast validation)\b",
            r"\b(landing page|prototype|mockup|survey|user research|interview|call|feedback|test|experiment|validation test|market research|customer discovery|user testing|A/B test|feedback loop|early feedback)\b",
        ];

        let moderate_validation_keywords = [
            r"\b(3\s*week|4\s*weeks|within 4 weeks|moderate|reasonable time|within month)\b",
            r"\b(test|experiment|trial|pilot|beta|alpha|user testing|feedback collection|market validation|user feedback|customer feedback)\b",
        ];

        let slow_validation_keywords = [
            r"\b(6\s*weeks|8\s*weeks|2\s*months|long|extended|full build|complete|after full development)\b",
            r"\b(full product|complete system|finished|100%|fully built|production ready|production version|complete implementation)\b",
        ];

        let mut fast_matches = 0;
        for pattern in &fast_validation_keywords {
            if let Ok(re) = Regex::new(pattern) {
                if re.is_match(idea_lower) {
                    fast_matches += 1;
                }
            }
        }

        if fast_matches > 0 {
            return 0.32 + ((fast_matches as f64).min(2.0) * 0.04); // 0.32-0.4 range (max 2 matches * 0.04 each)
        }

        let mut moderate_matches = 0;
        for pattern in &moderate_validation_keywords {
            if let Ok(re) = Regex::new(pattern) {
                if re.is_match(idea_lower) {
                    moderate_matches += 1;
                }
            }
        }

        if moderate_matches > 0 {
            return 0.22 + ((moderate_matches as f64).min(3.0) * 0.03); // 0.22-0.31 range (max 3 matches * 0.03 each)
        }

        let mut slow_matches = 0;
        for pattern in &slow_validation_keywords {
            if let Ok(re) = Regex::new(pattern) {
                if re.is_match(idea_lower) {
                    slow_matches += 1;
                }
            }
        }

        if slow_matches > 0 {
            return 0.1 + ((slow_matches as f64).min(3.0) * 0.037); // 0.1-0.21 range (max 3 matches * 0.037 each)
        }

        0.15 // Default to middle
    }

    fn score_revenue_testing(&self, idea_lower: &str) -> f64 {
        // 0.24-0.30: SaaS/product model; same work serves multiple customers
        // 0.16-0.23: Hybrid model; some leverage (templates, done-for-you services)
        // 0.08-0.15: Primarily service-based; limited leverage
        // 0.00-0.07: Pure time-for-money consulting with no scalability

        let saas_keywords = [
            r"\b(SaaS|software as a service|subscription|recurring|product|platform|system|tool|app|application|web app|digital product|online service)\b",
            r"\b(multiple customers|many users|scale|scaling|leverage|automated|passive income|recurring revenue|multi-tenant|user base|customer base)\b",
        ];

        let hybrid_keywords = [
            r"\b(template|course|ebook|done for you|DFY|package|bundle|kit|blueprint|course|guide|training|workshop|consulting package|digital asset|digital template)\b",
            r"\b(reusable|repurposable|multi use|leverage|some scale|multiple uses|replicable|scalable service|hybrid model)\b",
        ];

        let service_keywords = [
            r"\b(consulting|freelance|contract|hourly|per project|custom work|one on one|personalized|bespoke|custom solution|time for money|billable hour)\b",
            r"\b(time for money|billable|per hour|per day|per project|custom development|personal service|one-off service)\b",
        ];

        let mut saas_matches = 0;
        for pattern in &saas_keywords {
            if let Ok(re) = Regex::new(pattern) {
                if re.is_match(idea_lower) {
                    saas_matches += 1;
                }
            }
        }

        if saas_matches > 0 {
            return 0.24 + ((saas_matches as f64).min(2.0) * 0.03); // 0.24-0.3 range (max 2 matches * 0.03 each)
        }

        let mut hybrid_matches = 0;
        for pattern in &hybrid_keywords {
            if let Ok(re) = Regex::new(pattern) {
                if re.is_match(idea_lower) {
                    hybrid_matches += 1;
                }
            }
        }

        if hybrid_matches > 0 {
            return 0.16 + ((hybrid_matches as f64).min(2.0) * 0.035); // 0.16-0.23 range (max 2 matches * 0.035 each)
        }

        let mut service_matches = 0;
        for pattern in &service_keywords {
            if let Ok(re) = Regex::new(pattern) {
                if re.is_match(idea_lower) {
                    service_matches += 1;
                }
            }
        }

        if service_matches > 0 {
            return 0.08 + ((service_matches as f64).min(2.0) * 0.035); // 0.08-0.15 range (max 2 matches * 0.035 each)
        }

        0.05 // Default to low scalability
    }

    fn generate_explanations(
        &self,
        idea: &str,
        idea_lower: &str,
        mission: &MissionScores,
        anti_challenge: &AntiChallengeScores,
        strategic: &StrategicScores,
    ) -> std::collections::HashMap<String, String> {
        let mut explanations = std::collections::HashMap::new();

        // Mission Alignment explanations
        explanations.insert(
            "Domain Expertise".to_string(),
            self.explain_domain_expertise(idea, idea_lower, mission.domain_expertise),
        );
        explanations.insert(
            "AI Alignment".to_string(),
            self.explain_ai_alignment(idea_lower, mission.ai_alignment),
        );
        explanations.insert(
            "Execution Support".to_string(),
            self.explain_execution_support(idea_lower, mission.execution_support),
        );
        explanations.insert(
            "Revenue Potential".to_string(),
            self.explain_revenue_potential(idea_lower, mission.revenue_potential),
        );

        // Anti-Challenge Patterns explanations
        explanations.insert(
            "Avoid Context-Switching".to_string(),
            self.explain_context_switching(idea, idea_lower, anti_challenge.context_switching),
        );
        explanations.insert(
            "Rapid Prototyping".to_string(),
            self.explain_rapid_prototyping(idea_lower, anti_challenge.rapid_prototyping),
        );
        explanations.insert(
            "Accountability".to_string(),
            self.explain_accountability(idea_lower, anti_challenge.accountability),
        );
        explanations.insert(
            "Income Anxiety".to_string(),
            self.explain_income_anxiety(idea_lower, anti_challenge.income_anxiety),
        );

        // Strategic Fit explanations
        explanations.insert(
            "Stack Compatibility".to_string(),
            self.explain_stack_compatibility(idea, idea_lower, strategic.stack_compatibility),
        );
        explanations.insert(
            "Shipping Habit".to_string(),
            self.explain_shipping_habit(idea_lower, strategic.shipping_habit),
        );
        explanations.insert(
            "Public Accountability".to_string(),
            self.explain_public_accountability(idea_lower, strategic.public_accountability),
        );
        explanations.insert(
            "Revenue Testing".to_string(),
            self.explain_revenue_testing(idea_lower, strategic.revenue_testing),
        );

        explanations
    }

    fn explain_domain_expertise(&self, _idea: &str, idea_lower: &str, score: f64) -> String {
        if score >= 0.9 {
            format!("High domain expertise leverage ({} points): Idea directly leverages 80%+ of existing skills like n8n, Python, AI automation, or Android dev concepts.",
                    format!("{:.2}", score))
        } else if score >= 0.6 {
            format!("Medium domain expertise leverage ({} points): Idea uses 50-79% of existing skills; minor learning required.",
                    format!("{:.2}", score))
        } else if score >= 0.3 {
            format!("Low domain expertise leverage ({} points): Idea uses 30-49% of existing skills; significant new learning needed.",
                    format!("{:.2}", score))
        } else {
            format!("Minimal domain expertise leverage ({} points): Idea requires mostly new skills outside current domain.",
                    format!("{:.2}", score))
        }
    }

    fn explain_ai_alignment(&self, idea_lower: &str, score: f64) -> String {
        if score >= 1.2 {
            format!("Core AI product ({} points): Core product IS AI automation/systems (e.g., AI agents, automation pipelines).",
                    format!("{:.2}", score))
        } else if score >= 0.8 {
            format!("Significant AI component ({} points): AI is a significant component but not the core offering.",
                    format!("{:.2}", score))
        } else if score >= 0.4 {
            format!("Auxiliary AI component ({} points): AI is auxiliary or optional to the main offering.",
                    format!("{:.2}", score))
        } else {
            format!(
                "Minimal AI component ({} points): Little to no AI component in the idea.",
                format!("{:.2}", score)
            )
        }
    }

    fn explain_execution_support(&self, idea_lower: &str, score: f64) -> String {
        if score >= 0.65 {
            format!("Strong shipping bias ({} points): Idea has clear deliverable within 30 days; success = shipped product.",
                    format!("{:.2}", score))
        } else if score >= 0.45 {
            format!("Moderate shipping bias ({} points): Deliverable within 60 days with defined MVP scope.",
                    format!("{:.2}", score))
        } else if score >= 0.25 {
            format!("Weak shipping bias ({} points): Longer timeline (90+ days) or unclear MVP definition.",
                    format!("{:.2}", score))
        } else {
            format!("Learning-focused ({} points): Primarily learning-focused with no concrete deliverable.",
                    format!("{:.2}", score))
        }
    }

    fn explain_revenue_potential(&self, idea_lower: &str, score: f64) -> String {
        if score >= 0.4 {
            format!("High revenue potential ({} points): Clear monetization model; proven market willing to pay ($1K-$2.5K/month target).",
                    format!("{:.2}", score))
        } else if score >= 0.25 {
            format!("Moderate revenue potential ({} points): Plausible monetization; requires validation; likely $500-$1K/month.",
                    format!("{:.2}", score))
        } else if score >= 0.1 {
            format!("Speculative revenue potential ({} points): Speculative monetization; unclear market willingness to pay.",
                    format!("{:.2}", score))
        } else {
            format!("Low revenue potential ({} points): No clear revenue path or ad-based/very low revenue model.",
                    format!("{:.2}", score))
        }
    }

    fn explain_context_switching(&self, idea: &str, idea_lower: &str, score: f64) -> String {
        if score >= 0.95 {
            format!("Strong tech stack continuity ({} points): Uses 90%+ current stack (n8n/Python/AI APIs); no new languages/platforms.",
                    format!("{:.2}", score))
        } else if score >= 0.65 {
            format!("Moderate tech stack continuity ({} points): Uses 70-89% current stack; minor additions (new library, familiar framework).",
                    format!("{:.2}", score))
        } else if score >= 0.3 {
            format!("Weak tech stack continuity ({} points): Requires 50%+ new stack elements (new language, platform, or paradigm).",
                    format!("{:.2}", score))
        } else {
            format!("Tech stack switch detected ({} points): Complete stack switch (e.g., mobile dev, game development, hardware).",
                    format!("{:.2}", score))
        }
    }

    fn explain_rapid_prototyping(&self, idea_lower: &str, score: f64) -> String {
        if score >= 0.8 {
            format!("Rapid prototyping enabled ({} points): Can ship functional MVP in 1-2 weeks; inherently iterative (SaaS, automation).",
                    format!("{:.2}", score))
        } else if score >= 0.55 {
            format!("Moderate rapid prototyping ({} points): MVP possible in 3-4 weeks; some iteration possible.",
                    format!("{:.2}", score))
        } else if score >= 0.25 {
            format!("Slow rapid prototyping ({} points): Requires 6+ weeks for minimal viable version; high quality bar needed.",
                    format!("{:.2}", score))
        } else {
            format!("Perfectionism risk ({} points): Inherently perfection-dependent (content creation, courses, books).",
                    format!("{:.2}", score))
        }
    }

    fn explain_accountability(&self, idea_lower: &str, score: f64) -> String {
        if score >= 0.65 {
            format!("Strong built-in accountability ({} points): Paying customers or public commitments with consequences (e.g., pre-sales, cohort).",
                    format!("{:.2}", score))
        } else if score >= 0.45 {
            format!("Moderate built-in accountability ({} points): Strong accountability structure (accountability partner, public building, deadlines).",
                    format!("{:.2}", score))
        } else if score >= 0.2 {
            format!("Weak built-in accountability ({} points): Weak accountability (social media updates, personal goals).",
                    format!("{:.2}", score))
        } else {
            format!("No built-in accountability ({} points): No external accountability; purely self-motivated.",
                    format!("{:.2}", score))
        }
    }

    fn explain_income_anxiety(&self, idea_lower: &str, score: f64) -> String {
        if score >= 0.4 {
            format!("Fast income anxiety relief ({} points): First revenue possible within 30 days; path to consistent monthly income.",
                    format!("{:.2}", score))
        } else if score >= 0.25 {
            format!("Moderate income anxiety relief ({} points): First revenue within 60 days; recurring revenue model.",
                    format!("{:.2}", score))
        } else if score >= 0.1 {
            format!("Slow income anxiety relief ({} points): First revenue 90+ days; project-based or unpredictable income.",
                    format!("{:.2}", score))
        } else {
            format!(
                "Income anxiety remains ({} points): Revenue 6+ months away or highly uncertain.",
                format!("{:.2}", score)
            )
        }
    }

    fn explain_stack_compatibility(&self, idea: &str, idea_lower: &str, score: f64) -> String {
        if score >= 0.8 {
            format!("High execution compatibility ({} points): Enables 4+ hour flow sessions; clear systematic execution path.",
                    format!("{:.2}", score))
        } else if score >= 0.55 {
            format!("Moderate execution compatibility ({} points): Allows 2-3 hour focus blocks; mostly systematic with some ambiguity.",
                    format!("{:.2}", score))
        } else if score >= 0.25 {
            format!("Low execution compatibility ({} points): Requires frequent context switching or has unclear execution steps.",
                    format!("{:.2}", score))
        } else {
            format!("Fragmented work ({} points): Inherently fragmented work or chaotic/creative process.",
                    format!("{:.2}", score))
        }
    }

    fn explain_shipping_habit(&self, idea_lower: &str, score: f64) -> String {
        if score >= 0.65 {
            format!("High compounding benefits ({} points): Creates reusable systems, code, or processes for future projects.",
                    format!("{:.2}", score))
        } else if score >= 0.45 {
            format!("Some compounding benefits ({} points): Some reusable components; partial knowledge transfer.",
                    format!("{:.2}", score))
        } else if score >= 0.2 {
            format!("Low compounding benefits ({} points): Minimal reusability; mostly project-specific work.",
                    format!("{:.2}", score))
        } else {
            format!(
                "No compounding ({} points): Purely one-off effort with no future leverage.",
                format!("{:.2}", score)
            )
        }
    }

    fn explain_public_accountability(&self, idea_lower: &str, score: f64) -> String {
        if score >= 0.32 {
            format!("Fast validation possible ({} points): Can validate core assumption within 1-2 weeks (landing page, calls, prototype).",
                    format!("{:.2}", score))
        } else if score >= 0.22 {
            format!(
                "Moderate validation speed ({} points): Validation possible in 3-4 weeks.",
                format!("{:.2}", score)
            )
        } else if score >= 0.1 {
            format!(
                "Slow validation speed ({} points): Requires 6-8 weeks to validate.",
                format!("{:.2}", score)
            )
        } else {
            format!("Slow validation speed ({} points): Validation requires 2+ months or full product build.",
                    format!("{:.2}", score))
        }
    }

    fn explain_revenue_testing(&self, idea_lower: &str, score: f64) -> String {
        if score >= 0.24 {
            format!("High scalability potential ({} points): SaaS/product model; same work serves multiple customers.",
                    format!("{:.2}", score))
        } else if score >= 0.16 {
            format!("Some scalability potential ({} points): Hybrid model; some leverage (templates, done-for-you services).",
                    format!("{:.2}", score))
        } else if score >= 0.08 {
            format!(
                "Low scalability potential ({} points): Primarily service-based; limited leverage.",
                format!("{:.2}", score)
            )
        } else {
            format!("No scalability potential ({} points): Pure time-for-money consulting with no scalability.",
                    format!("{:.2}", score))
        }
    }

    /// Generate detailed explanations matching the prompt format
    pub fn generate_detailed_explanations(
        &self,
        idea: &str,
        idea_lower: &str,
        score: &Score,
    ) -> std::collections::HashMap<String, String> {
        let mut explanations = std::collections::HashMap::new();

        // Mission Alignment explanations with detailed justifications
        explanations.insert(
            "Domain Expertise".to_string(),
            format!(
                "Domain Expertise Leverage: {:.2} points - {}",
                score.mission.domain_expertise,
                self.explain_domain_expertise(idea, idea_lower, score.mission.domain_expertise)
            ),
        );

        explanations.insert(
            "AI Alignment".to_string(),
            format!(
                "AI Systems/Building Alignment: {:.2} points - {}",
                score.mission.ai_alignment,
                self.explain_ai_alignment(idea_lower, score.mission.ai_alignment)
            ),
        );

        explanations.insert(
            "Execution Support".to_string(),
            format!(
                "Shipping Bias: {:.2} points - {}",
                score.mission.execution_support,
                self.explain_execution_support(idea_lower, score.mission.execution_support)
            ),
        );

        explanations.insert(
            "Revenue Potential".to_string(),
            format!(
                "Revenue Potential: {:.2} points - {}",
                score.mission.revenue_potential,
                self.explain_revenue_potential(idea_lower, score.mission.revenue_potential)
            ),
        );

        // Anti-Challenge Patterns explanations with detailed justifications
        explanations.insert(
            "Avoid Context-Switching".to_string(),
            format!(
                "Tech Stack Continuity: {:.2} points - {}",
                score.anti_challenge.context_switching,
                self.explain_context_switching(
                    idea,
                    idea_lower,
                    score.anti_challenge.context_switching
                )
            ),
        );

        explanations.insert(
            "Rapid Prototyping".to_string(),
            format!(
                "Rapid Prototyping Design: {:.2} points - {}",
                score.anti_challenge.rapid_prototyping,
                self.explain_rapid_prototyping(idea_lower, score.anti_challenge.rapid_prototyping)
            ),
        );

        explanations.insert(
            "Accountability".to_string(),
            format!(
                "Built-in Accountability: {:.2} points - {}",
                score.anti_challenge.accountability,
                self.explain_accountability(idea_lower, score.anti_challenge.accountability)
            ),
        );

        explanations.insert(
            "Income Anxiety".to_string(),
            format!(
                "Income Anxiety Relief: {:.2} points - {}",
                score.anti_challenge.income_anxiety,
                self.explain_income_anxiety(idea_lower, score.anti_challenge.income_anxiety)
            ),
        );

        // Strategic Fit explanations with detailed justifications
        explanations.insert(
            "Stack Compatibility".to_string(),
            format!(
                "Execution Compatibility: {:.2} points - {}",
                score.strategic.stack_compatibility,
                self.explain_stack_compatibility(
                    idea,
                    idea_lower,
                    score.strategic.stack_compatibility
                )
            ),
        );

        explanations.insert(
            "Shipping Habit".to_string(),
            format!(
                "Compounding Benefits: {:.2} points - {}",
                score.strategic.shipping_habit,
                self.explain_shipping_habit(idea_lower, score.strategic.shipping_habit)
            ),
        );

        explanations.insert(
            "Public Accountability".to_string(),
            format!(
                "Validation Speed: {:.2} points - {}",
                score.strategic.public_accountability,
                self.explain_public_accountability(
                    idea_lower,
                    score.strategic.public_accountability
                )
            ),
        );

        explanations.insert(
            "Revenue Testing".to_string(),
            format!(
                "Scalability Potential: {:.2} points - {}",
                score.strategic.revenue_testing,
                self.explain_revenue_testing(idea_lower, score.strategic.revenue_testing)
            ),
        );

        explanations
    }
}

impl ScoringPatterns {
    fn default_patterns() -> Self {
        // Create default empty patterns to avoid compilation errors
        // Since regex compilation can fail, we use empty vectors as fallback
        ScoringPatterns {
            ai_positive: vec![],
            ai_negative: vec![],
            shipping_positive: vec![],
            shipping_negative: vec![],
            income_positive: vec![],
            domain_positive: vec![],
            stack_violations: vec![],
            perfectionism_negative: vec![],
            perfectionism_positive: vec![],
            accountability_positive: vec![],
            procrastination_negative: vec![],
        }
    }
}
