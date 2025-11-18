pub mod circuit_breaker;
pub mod database;
pub mod scoring;
pub mod security;
pub mod validation;

use thiserror::Error;

/// Main application error type that encompasses all possible errors
#[derive(Error, Debug)]
pub enum ApplicationError {
    #[error("Database operation failed: {0}")]
    Database(#[from] DatabaseError),

    #[error("Scoring operation failed: {0}")]
    Scoring(#[from] ScoringError),

    #[error("Validation failed: {0}")]
    Validation(#[from] validation::ValidationError),

    #[error("Security violation: {0}")]
    Security(#[from] security::SecurityError),

    #[error("Configuration error: {0}")]
    Configuration(String),

    #[error("I/O error: {0}")]
    Io(#[from] std::io::Error),

    #[error("AI service error: {0}")]
    AiService(#[from] circuit_breaker::CircuitBreakerError),

    #[error("Telos parsing error: {0}")]
    TelosParsing(String),

    #[error("Pattern detection error: {0}")]
    PatternDetection(String),

    #[error("CLI error: {0}")]
    Cli(String),

    #[error("Operation cancelled: {context}")]
    OperationCancelled { context: String },

    #[error("Operation timeout after {timeout_ms}ms: {context}")]
    OperationTimeout { timeout_ms: u64, context: String },

    #[error("Graceful shutdown in progress: {operation}")]
    ShutdownInProgress { operation: String },

    #[error("Dialog error: {0}")]
    Dialog(#[from] dialoguer::Error),

    #[error("Generic error: {0}")]
    Generic(#[from] anyhow::Error),

    #[error("Serialization error: {0}")]
    Serialization(#[from] serde_json::Error),
}

// Re-export error types for convenience
pub use circuit_breaker::CircuitBreakerError;
pub use database::DatabaseError;
pub use scoring::ScoringError;
pub use validation::ValidationError;

/// Result type alias for the entire application
pub type Result<T> = std::result::Result<T, ApplicationError>;

impl ApplicationError {
    /// Create a new validation error
    pub fn validation(message: impl Into<String>) -> Self {
        // Create a validation error using invalid_format which takes field and reason
        Self::Validation(validation::ValidationError::invalid_format(
            "general",
            message.into(),
        ))
    }

    /// Create a new validation error with field context
    pub fn validation_with_field(message: impl Into<String>, field: impl Into<String>) -> Self {
        Self::Validation(validation::ValidationError::invalid_format(
            field.into(),
            message.into(),
        ))
    }

    /// Create an operation cancelled error
    pub fn operation_cancelled(context: impl Into<String>) -> Self {
        Self::OperationCancelled {
            context: context.into(),
        }
    }

    /// Create an operation timeout error
    pub fn operation_timeout(timeout_ms: u64, context: impl Into<String>) -> Self {
        Self::OperationTimeout {
            timeout_ms,
            context: context.into(),
        }
    }

    /// Create a shutdown in progress error
    pub fn shutdown_in_progress(operation: impl Into<String>) -> Self {
        Self::ShutdownInProgress {
            operation: operation.into(),
        }
    }

    /// Add context to an existing error
    pub fn with_context(self, message: impl Into<String>) -> Self {
        // For now, we'll wrap the error in a generic error with context
        // This could be improved with more sophisticated error chaining
        let other = self;
        Self::Generic(anyhow::anyhow!("{}: {}", message.into(), other))
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_validation_error_creation() {
        let error = ApplicationError::validation("Test error");
        assert!(matches!(error, ApplicationError::Validation(_)));

        let error_str = error.to_string();
        assert!(error_str.contains("Test error"));
    }

    #[test]
    fn test_validation_error_with_field() {
        let error = ApplicationError::validation_with_field("Empty content", "content");
        assert!(matches!(error, ApplicationError::Validation(_)));
    }
}
