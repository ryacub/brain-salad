use thiserror::Error;

/// Input validation related errors
#[derive(Error, Debug)]
pub enum ValidationError {
    #[error("Empty {field} - {field} cannot be empty")]
    EmptyField { field: String },

    #[error("{field} is too long: {length} characters (max: {max_len})")]
    TooLong {
        field: String,
        length: usize,
        max_len: usize,
    },

    #[error("{field} is too short: {length} characters (min: {min_len})")]
    TooShort {
        field: String,
        length: usize,
        min_len: usize,
    },

    #[error("Invalid {field}: {value} - {reason}")]
    InvalidValue {
        field: String,
        value: String,
        reason: String,
    },

    #[error("Invalid {field} format: {reason}")]
    InvalidFormat { field: String, reason: String },

    #[error("Invalid UTF-8 sequence in {field}")]
    InvalidUtf8 { field: String },

    #[error("Unsafe content detected in {field}: {reason}")]
    UnsafeContent {
        field: String,
        reason: String,
        pattern: Option<String>,
    },

    #[error("Path traversal attempt detected: {path}")]
    PathTraversal { path: String },

    #[error("Invalid file extension: {extension}")]
    InvalidExtension {
        extension: String,
        allowed: Vec<String>,
    },

    #[error("Request size exceeds limit: {size} bytes (max: {max_size} bytes)")]
    RequestTooLarge { size: usize, max_size: usize },

    #[error("Rate limit exceeded for {identifier}: {current}/{max} requests")]
    RateLimitExceeded {
        identifier: String,
        current: u32,
        max: u32,
        reset_time: Option<std::time::Instant>,
    },

    #[error("Invalid JSON in {field}: {source}")]
    InvalidJson {
        field: String,
        #[source]
        source: serde_json::Error,
    },

    #[error("Invalid date/time format in {field}: {value} - {reason}")]
    InvalidDateTime {
        field: String,
        value: String,
        reason: String,
    },

    #[error("Invalid UUID format in {field}: {value}")]
    InvalidUuid { field: String, value: String },

    #[error("Invalid email address in {field}: {value}")]
    InvalidEmail { field: String, value: String },

    #[error("Invalid URL in {field}: {value} - {reason}")]
    InvalidUrl {
        field: String,
        value: String,
        reason: String,
    },

    #[error("Forbidden content detected: {content_type}")]
    ForbiddenContent {
        content_type: String,
        detected_content: String,
    },

    #[error("Cross-site scripting (XSS) attempt detected")]
    XssAttempt { field: String, pattern: String },

    #[error("SQL injection attempt detected")]
    SqlInjectionAttempt { pattern: String },

    #[error("Command injection attempt detected")]
    CommandInjectionAttempt { pattern: String },

    #[error("Invalid characters in {field}: {characters}")]
    InvalidCharacters {
        field: String,
        characters: String,
        allowed: Option<String>,
    },
}

impl ValidationError {
    /// Create an empty field error
    pub fn empty_field(field: impl Into<String>) -> Self {
        Self::EmptyField {
            field: field.into(),
        }
    }

    /// Create a too long field error
    pub fn too_long(field: impl Into<String>, length: usize, max_len: usize) -> Self {
        Self::TooLong {
            field: field.into(),
            length,
            max_len,
        }
    }

    /// Create a too short field error
    pub fn too_short(field: impl Into<String>, length: usize, min_len: usize) -> Self {
        Self::TooShort {
            field: field.into(),
            length,
            min_len,
        }
    }

    /// Create an invalid value error
    pub fn invalid_value(
        field: impl Into<String>,
        value: impl Into<String>,
        reason: impl Into<String>,
    ) -> Self {
        Self::InvalidValue {
            field: field.into(),
            value: value.into(),
            reason: reason.into(),
        }
    }

    /// Create an invalid format error
    pub fn invalid_format(field: impl Into<String>, reason: impl Into<String>) -> Self {
        Self::InvalidFormat {
            field: field.into(),
            reason: reason.into(),
        }
    }

    /// Create an invalid UTF-8 error
    pub fn invalid_utf8(field: impl Into<String>) -> Self {
        Self::InvalidUtf8 {
            field: field.into(),
        }
    }

    /// Create an unsafe content error
    pub fn unsafe_content(
        field: impl Into<String>,
        reason: impl Into<String>,
        pattern: Option<String>,
    ) -> Self {
        Self::UnsafeContent {
            field: field.into(),
            reason: reason.into(),
            pattern,
        }
    }

    /// Create a path traversal error
    pub fn path_traversal(path: impl Into<String>) -> Self {
        Self::PathTraversal { path: path.into() }
    }

    /// Create an invalid extension error
    pub fn invalid_extension(extension: impl Into<String>, allowed: Vec<String>) -> Self {
        Self::InvalidExtension {
            extension: extension.into(),
            allowed,
        }
    }

    /// Create a request too large error
    pub fn request_too_large(size: usize, max_size: usize) -> Self {
        Self::RequestTooLarge { size, max_size }
    }

    /// Create a rate limit exceeded error
    pub fn rate_limit_exceeded(
        identifier: impl Into<String>,
        current: u32,
        max: u32,
        reset_time: Option<std::time::Instant>,
    ) -> Self {
        Self::RateLimitExceeded {
            identifier: identifier.into(),
            current,
            max,
            reset_time,
        }
    }

    /// Create an invalid JSON error
    pub fn invalid_json(field: impl Into<String>, source: serde_json::Error) -> Self {
        Self::InvalidJson {
            field: field.into(),
            source,
        }
    }

    /// Create an invalid date/time error
    pub fn invalid_date_time(
        field: impl Into<String>,
        value: impl Into<String>,
        reason: impl Into<String>,
    ) -> Self {
        Self::InvalidDateTime {
            field: field.into(),
            value: value.into(),
            reason: reason.into(),
        }
    }

    /// Create an invalid UUID error
    pub fn invalid_uuid(field: impl Into<String>, value: impl Into<String>) -> Self {
        Self::InvalidUuid {
            field: field.into(),
            value: value.into(),
        }
    }

    /// Create an invalid email error
    pub fn invalid_email(field: impl Into<String>, value: impl Into<String>) -> Self {
        Self::InvalidEmail {
            field: field.into(),
            value: value.into(),
        }
    }

    /// Create an invalid URL error
    pub fn invalid_url(
        field: impl Into<String>,
        value: impl Into<String>,
        reason: impl Into<String>,
    ) -> Self {
        Self::InvalidUrl {
            field: field.into(),
            value: value.into(),
            reason: reason.into(),
        }
    }

    /// Create a forbidden content error
    pub fn forbidden_content(
        content_type: impl Into<String>,
        detected_content: impl Into<String>,
    ) -> Self {
        Self::ForbiddenContent {
            content_type: content_type.into(),
            detected_content: detected_content.into(),
        }
    }

    /// Create an XSS attempt error
    pub fn xss_attempt(field: impl Into<String>, pattern: impl Into<String>) -> Self {
        Self::XssAttempt {
            field: field.into(),
            pattern: pattern.into(),
        }
    }

    /// Create a SQL injection attempt error
    pub fn sql_injection_attempt(pattern: impl Into<String>) -> Self {
        Self::SqlInjectionAttempt {
            pattern: pattern.into(),
        }
    }

    /// Create a command injection attempt error
    pub fn command_injection_attempt(pattern: impl Into<String>) -> Self {
        Self::CommandInjectionAttempt {
            pattern: pattern.into(),
        }
    }

    /// Create an invalid characters error
    pub fn invalid_characters(
        field: impl Into<String>,
        characters: impl Into<String>,
        allowed: Option<String>,
    ) -> Self {
        Self::InvalidCharacters {
            field: field.into(),
            characters: characters.into(),
            allowed,
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_empty_field_error() {
        let error = ValidationError::empty_field("content");
        assert!(matches!(error, ValidationError::EmptyField { .. }));

        let error_str = error.to_string();
        assert!(error_str.contains("content"));
        assert!(error_str.contains("cannot be empty"));
    }

    #[test]
    fn test_too_long_error() {
        let error = ValidationError::too_long("idea", 1500, 1000);

        if let ValidationError::TooLong {
            field,
            length,
            max_len,
        } = error
        {
            assert_eq!(field, "idea");
            assert_eq!(length, 1500);
            assert_eq!(max_len, 1000);
        } else {
            panic!("Expected TooLong error");
        }
    }

    #[test]
    fn test_invalid_value_error() {
        let error = ValidationError::invalid_value("score", "15.5", "must be between 0-10");

        if let ValidationError::InvalidValue {
            field,
            value,
            reason,
        } = error
        {
            assert_eq!(field, "score");
            assert_eq!(value, "15.5");
            assert_eq!(reason, "must be between 0-10");
        } else {
            panic!("Expected InvalidValue error");
        }
    }

    #[test]
    fn test_path_traversal_error() {
        let error = ValidationError::path_traversal("../../../etc/passwd");
        assert!(matches!(error, ValidationError::PathTraversal { .. }));

        let error_str = error.to_string();
        assert!(error_str.contains("../../../etc/passwd"));
        assert!(error_str.contains("Path traversal"));
    }

    #[test]
    fn test_rate_limit_exceeded_error() {
        let reset_time = std::time::Instant::now() + std::time::Duration::from_secs(60);
        let error = ValidationError::rate_limit_exceeded("user123", 61, 60, Some(reset_time));

        if let ValidationError::RateLimitExceeded {
            identifier,
            current,
            max,
            ..
        } = error
        {
            assert_eq!(identifier, "user123");
            assert_eq!(current, 61);
            assert_eq!(max, 60);
        } else {
            panic!("Expected RateLimitExceeded error");
        }
    }

    #[test]
    fn test_sql_injection_attempt_error() {
        let error = ValidationError::sql_injection_attempt("DROP TABLE users;");
        assert!(matches!(error, ValidationError::SqlInjectionAttempt { .. }));

        // Check that the pattern is stored in the error
        if let ValidationError::SqlInjectionAttempt { pattern } = error {
            assert_eq!(pattern, "DROP TABLE users;");
        } else {
            panic!("Expected SqlInjectionAttempt variant");
        }

        // The display message just says "SQL injection attempt detected"
        let error2 = ValidationError::sql_injection_attempt("test");
        let error_str = error2.to_string();
        assert!(error_str.contains("SQL injection"));
    }
}
