use thiserror::Error;

/// Scoring engine and analysis related errors
#[derive(Error, Debug)]
pub enum ScoringError {
    #[error("Empty idea content - cannot analyze empty ideas")]
    EmptyContent,

    #[error("Idea content too long: {length} characters (max: {max_len})")]
    TooLong { length: usize, max_len: usize },

    #[error("Idea content too short: {length} characters (min: {min_len})")]
    TooShort { length: usize, min_len: usize },

    #[error("Pattern compilation failed for '{pattern}': {source}")]
    PatternCompilationFailed {
        pattern: String,
        #[source]
        source: regex::Error,
    },

    #[error("Pattern matching failed: {source}")]
    PatternMatching {
        #[source]
        source: regex::Error,
    },

    #[error("Score calculation overflow - score exceeded bounds")]
    ScoreOverflow,

    #[error("Invalid score range: {score} (must be between 0.0 and 10.0)")]
    InvalidScoreRange { score: f64 },

    #[error("Invalid weight configuration: {details}")]
    InvalidWeights {
        details: String,
        weights: Option<(f64, f64, f64)>, // (mission, anti_challenge, strategic)
    },

    #[error("Telos configuration loading failed: {source}")]
    TelosConfigFailed {
        #[source]
        source: Box<dyn std::error::Error + Send + Sync>,
    },

    #[error("Telos file parsing failed: {file} - {reason}")]
    TelosParsingFailed {
        file: String,
        reason: String,
        line: Option<usize>,
    },

    #[error("Component scoring failed: {component}")]
    ComponentScoring {
        component: String,
        #[source]
        source: Box<dyn std::error::Error + Send + Sync>,
    },

    #[error("Pattern detection failed: {pattern_type}")]
    PatternDetection {
        pattern_type: String,
        #[source]
        source: Box<dyn std::error::Error + Send + Sync>,
    },

    #[error("AI enhancement failed: {reason}")]
    AiEnhancement {
        reason: String,
        #[source]
        source: Option<Box<dyn std::error::Error + Send + Sync>>,
    },

    #[error("Score serialization failed: {source}")]
    ScoreSerialization {
        #[from]
        source: serde_json::Error,
    },

    #[error("Recommendation generation failed: {score} - {reason}")]
    RecommendationGeneration { score: f64, reason: String },

    #[error("Invalid pattern configuration: {details}")]
    InvalidPatternConfig { details: String },

    #[error("Cache operation failed: {operation}")]
    CacheError {
        operation: String,
        #[source]
        source: Box<dyn std::error::Error + Send + Sync>,
    },

    #[error("Scoring engine not initialized")]
    NotInitialized,

    #[error("Configuration validation failed: {field} - {reason}")]
    ConfigValidation { field: String, reason: String },
}

impl ScoringError {
    /// Create a too long content error
    pub fn too_long(length: usize, max_len: usize) -> Self {
        Self::TooLong { length, max_len }
    }

    /// Create a too short content error
    pub fn too_short(length: usize, min_len: usize) -> Self {
        Self::TooShort { length, min_len }
    }

    /// Create a pattern compilation error
    pub fn pattern_compilation(pattern: impl Into<String>, source: regex::Error) -> Self {
        Self::PatternCompilationFailed {
            pattern: pattern.into(),
            source,
        }
    }

    /// Create an invalid score range error
    pub fn invalid_score_range(score: f64) -> Self {
        Self::InvalidScoreRange { score }
    }

    /// Create an invalid weights error
    pub fn invalid_weights(details: impl Into<String>, weights: (f64, f64, f64)) -> Self {
        Self::InvalidWeights {
            details: details.into(),
            weights: Some(weights),
        }
    }

    /// Create a Telos parsing error
    pub fn telos_parsing_failed(file: impl Into<String>, reason: impl Into<String>) -> Self {
        Self::TelosParsingFailed {
            file: file.into(),
            reason: reason.into(),
            line: None,
        }
    }

    /// Create a Telos parsing error with line number
    pub fn telos_parsing_failed_with_line(
        file: impl Into<String>,
        reason: impl Into<String>,
        line: usize,
    ) -> Self {
        Self::TelosParsingFailed {
            file: file.into(),
            reason: reason.into(),
            line: Some(line),
        }
    }

    /// Create a component scoring error
    pub fn component_scoring(
        component: impl Into<String>,
        source: Box<dyn std::error::Error + Send + Sync>,
    ) -> Self {
        Self::ComponentScoring {
            component: component.into(),
            source,
        }
    }

    /// Create a pattern detection error
    pub fn pattern_detection(
        pattern_type: impl Into<String>,
        source: Box<dyn std::error::Error + Send + Sync>,
    ) -> Self {
        Self::PatternDetection {
            pattern_type: pattern_type.into(),
            source,
        }
    }

    /// Create an AI enhancement error
    pub fn ai_enhancement(reason: impl Into<String>) -> Self {
        Self::AiEnhancement {
            reason: reason.into(),
            source: None,
        }
    }

    /// Create an AI enhancement error with source
    pub fn ai_enhancement_with_source(
        reason: impl Into<String>,
        source: Box<dyn std::error::Error + Send + Sync>,
    ) -> Self {
        Self::AiEnhancement {
            reason: reason.into(),
            source: Some(source),
        }
    }

    /// Create a recommendation generation error
    pub fn recommendation_generation(score: f64, reason: impl Into<String>) -> Self {
        Self::RecommendationGeneration {
            score,
            reason: reason.into(),
        }
    }

    /// Create an invalid pattern config error
    pub fn invalid_pattern_config(details: impl Into<String>) -> Self {
        Self::InvalidPatternConfig {
            details: details.into(),
        }
    }

    /// Create a cache error
    pub fn cache_error(
        operation: impl Into<String>,
        source: Box<dyn std::error::Error + Send + Sync>,
    ) -> Self {
        Self::CacheError {
            operation: operation.into(),
            source,
        }
    }

    /// Create a configuration validation error
    pub fn config_validation(field: impl Into<String>, reason: impl Into<String>) -> Self {
        Self::ConfigValidation {
            field: field.into(),
            reason: reason.into(),
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_empty_content_error() {
        let error = ScoringError::EmptyContent;
        assert!(matches!(error, ScoringError::EmptyContent));
        assert!(error.to_string().contains("empty ideas"));
    }

    #[test]
    fn test_too_long_error() {
        let error = ScoringError::too_long(150000, 100000);

        if let ScoringError::TooLong { length, max_len } = error {
            assert_eq!(length, 150000);
            assert_eq!(max_len, 100000);
        } else {
            panic!("Expected TooLong error");
        }

        let error_str = error.to_string();
        assert!(error_str.contains("150000"));
        assert!(error_str.contains("100000"));
    }

    #[test]
    fn test_pattern_compilation_error() {
        let regex_error = regex::Error::Syntax("unclosed character class in regex".to_string());
        let error = ScoringError::pattern_compilation("invalid[regex", regex_error);

        if let ScoringError::PatternCompilationFailed { pattern, .. } = error {
            assert_eq!(pattern, "invalid[regex");
        } else {
            panic!("Expected PatternCompilationFailed error");
        }
    }

    #[test]
    fn test_invalid_score_range() {
        let error = ScoringError::invalid_score_range(15.5);

        if let ScoringError::InvalidScoreRange { score } = error {
            assert_eq!(score, 15.5);
        } else {
            panic!("Expected InvalidScoreRange error");
        }

        let error_str = error.to_string();
        assert!(error_str.contains("15.5"));
        assert!(error_str.contains("0.0 and 10.0"));
    }

    #[test]
    fn test_telos_parsing_error() {
        let error = ScoringError::telos_parsing_failed("telos.md", "Invalid YAML");

        if let ScoringError::TelosParsingFailed { file, reason, line } = error {
            assert_eq!(file, "telos.md");
            assert_eq!(reason, "Invalid YAML");
            assert!(line.is_none());
        } else {
            panic!("Expected TelosParsingFailed error");
        }
    }

    #[test]
    fn test_telos_parsing_error_with_line() {
        let error = ScoringError::telos_parsing_failed_with_line("telos.md", "Syntax error", 25);

        if let ScoringError::TelosParsingFailed { file, reason, line } = error {
            assert_eq!(file, "telos.md");
            assert_eq!(reason, "Syntax error");
            assert_eq!(line, Some(25));
        } else {
            panic!("Expected TelosParsingFailed error");
        }
    }

    #[test]
    fn test_recommendation_generation_error() {
        let error = ScoringError::recommendation_generation(3.2, "Score too low");

        if let ScoringError::RecommendationGeneration { score, reason } = error {
            assert_eq!(score, 3.2);
            assert_eq!(reason, "Score too low");
        } else {
            panic!("Expected RecommendationGeneration error");
        }
    }
}
