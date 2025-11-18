use serde::{Deserialize, Serialize};

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct PatternMatch {
    pub pattern_type: PatternType,
    pub severity: Severity,
    pub matches: Vec<String>,
    pub message: String,
    pub suggestion: Option<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum PatternType {
    ContextSwitching,
    Perfectionism,
    Procrastination,
    AccountabilityAvoidance,
    ScopeCreep,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum Severity {
    Critical,
    High,
    Medium,
    Low,
    Positive,
}

impl PatternType {
    pub fn emoji(&self) -> &'static str {
        match self {
            PatternType::ContextSwitching => "🔄",
            PatternType::Perfectionism => "⚡",
            PatternType::Procrastination => "🕰️",
            PatternType::AccountabilityAvoidance => "👤",
            PatternType::ScopeCreep => "📏",
        }
    }

    pub fn title(&self) -> &'static str {
        match self {
            PatternType::ContextSwitching => "Context-Switching",
            PatternType::Perfectionism => "Perfectionism",
            PatternType::Procrastination => "Procrastination",
            PatternType::AccountabilityAvoidance => "Accountability Avoidance",
            PatternType::ScopeCreep => "Scope Creep",
        }
    }
}

impl Severity {
    pub fn emoji(&self) -> &'static str {
        match self {
            Severity::Critical => "🔴",
            Severity::High => "🟠",
            Severity::Medium => "🟡",
            Severity::Low => "🟢",
            Severity::Positive => "✅",
        }
    }
}

#[derive(Clone)]
pub struct PatternDetector {
    current_stack: Vec<String>,
}

impl PatternDetector {
    pub fn new() -> Self {
        let current_stack = vec![
            "python".to_string(),
            "langchain".to_string(),
            "openai".to_string(),
            "gpt".to_string(),
            "api".to_string(),
            "streamlit".to_string(),
            "web app".to_string(),
        ];

        Self { current_stack }
    }

    pub fn detect_patterns(&self, idea: &str) -> Vec<PatternMatch> {
        let idea_lower = idea.to_lowercase();

        // Use iterator to collect patterns efficiently
        [
            self.detect_context_switching(idea, &idea_lower),
            self.detect_perfectionism(idea, &idea_lower),
            self.detect_procrastination(idea, &idea_lower),
            self.detect_accountability(idea, &idea_lower),
        ]
        .into_iter()
        .flatten()
        .collect()
    }

    fn detect_context_switching(&self, _idea: &str, idea_lower: &str) -> Vec<PatternMatch> {
        let mut patterns = Vec::with_capacity(2); // Pre-allocate capacity

        // Simple keyword detection for context-switching
        if idea_lower.contains("rust")
            || idea_lower.contains("javascript")
            || idea_lower.contains("react")
        {
            patterns.push(PatternMatch {
                pattern_type: PatternType::ContextSwitching,
                severity: Severity::High,
                matches: vec!["New tech stack detected".to_string()],
                message: "Context-switching risk detected".to_string(),
                suggestion: Some(
                    "Focus on current stack (Python + LangChain + OpenAI)".to_string(),
                ),
            });
        }

        // Check if using current stack (positive pattern)
        if self
            .current_stack
            .iter()
            .any(|tech| idea_lower.contains(tech))
        {
            patterns.push(PatternMatch {
                pattern_type: PatternType::ContextSwitching,
                severity: Severity::Positive,
                matches: Vec::new(), // Avoid allocating empty vec with macro
                message: "Staying focused on current tech stack".to_string(),
                suggestion: None,
            });
        }

        patterns
    }

    fn detect_perfectionism(&self, idea: &str, _idea_lower: &str) -> Vec<PatternMatch> {
        let mut patterns = Vec::new();

        // Check for scope creep
        if idea.to_lowercase().contains("comprehensive") || idea.to_lowercase().contains("complete")
        {
            patterns.push(PatternMatch {
                pattern_type: PatternType::Perfectionism,
                severity: Severity::High,
                matches: vec!["Scope creep keyword detected".to_string()],
                message: "Scope creep risk - over-engineering detected".to_string(),
                suggestion: Some("Define v1 scope and postpone advanced features".to_string()),
            });
        }

        patterns
    }

    fn detect_procrastination(&self, _idea: &str, idea_lower: &str) -> Vec<PatternMatch> {
        let mut patterns = Vec::new();

        // Check for consumption traps
        if idea_lower.contains("learn")
            && (idea_lower.contains("before") || idea_lower.contains("then"))
        {
            patterns.push(PatternMatch {
                pattern_type: PatternType::Procrastination,
                severity: Severity::Critical,
                matches: vec!["Learning before building detected".to_string()],
                message: "Consumption trap - learning before building".to_string(),
                suggestion: Some(
                    "Build first, learn as needed. Don't get stuck in tutorial loop.".to_string(),
                ),
            });
        }

        patterns
    }

    fn detect_accountability(&self, _idea: &str, idea_lower: &str) -> Vec<PatternMatch> {
        let mut patterns = Vec::new();

        // Check for isolation patterns
        if idea_lower.contains("just for me") || idea_lower.contains("personal project") {
            patterns.push(PatternMatch {
                pattern_type: PatternType::AccountabilityAvoidance,
                severity: Severity::Medium,
                matches: vec!["Solo-only project".to_string()],
                message: "Solo-only project - no external accountability".to_string(),
                suggestion: Some("Add public component or external deadline".to_string()),
            });
        }

        // Check for accountability signals (positive)
        if idea_lower.contains("public")
            || idea_lower.contains("share")
            || idea_lower.contains("github")
        {
            patterns.push(PatternMatch {
                pattern_type: PatternType::AccountabilityAvoidance,
                severity: Severity::Positive,
                matches: vec!["Public accountability component detected".to_string()],
                message: "External accountability component detected".to_string(),
                suggestion: None,
            });
        }

        patterns
    }
}
