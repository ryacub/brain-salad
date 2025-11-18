//! Comprehensive input validation for the telos-idea-matrix application.
//!
//! This module provides validation functions for all user inputs including:
//! - Idea content validation (length, characters, security)
//! - File path security validation
//! - Configuration validation
//! - Database parameter validation
//! - Command-line argument validation

use crate::errors::{Result, ValidationError};
use regex::Regex;
use std::path::{Path, PathBuf};

/// Validation configuration parameters
#[derive(Debug, Clone)]
pub struct ValidationConfig {
    /// Maximum length for idea content
    pub max_idea_length: usize,
    /// Minimum length for idea content
    pub min_idea_length: usize,
    /// Maximum length for file paths
    pub max_path_length: usize,
    /// Allowed file extensions for imports/exports
    pub allowed_file_extensions: Vec<String>,
    /// Whether to allow Unicode characters in ideas
    pub allow_unicode: bool,
    /// Maximum number of ideas to process in batch operations
    pub max_batch_size: usize,
    /// Maximum score value
    pub max_score_value: f64,
    /// Minimum score value
    pub min_score_value: f64,
}

impl Default for ValidationConfig {
    fn default() -> Self {
        Self {
            max_idea_length: 5000,
            min_idea_length: 3,
            max_path_length: 4096,
            allowed_file_extensions: vec![
                "txt".to_string(),
                "json".to_string(),
                "csv".to_string(),
                "md".to_string(),
            ],
            allow_unicode: true,
            max_batch_size: 1000,
            max_score_value: 10.0,
            min_score_value: 0.0,
        }
    }
}

/// Validator for user inputs
#[derive(Debug, Clone)]
pub struct InputValidator {
    config: ValidationConfig,
    // Pre-compiled regex patterns for performance
    xss_pattern: Regex,
    sql_injection_pattern: Regex,
    path_traversal_pattern: Regex,
    dangerous_chars_pattern: Regex,
}

impl InputValidator {
    /// Create a new validator with default configuration
    pub fn new() -> Self {
        Self::with_config(ValidationConfig::default())
    }

    /// Create a new validator with custom configuration
    pub fn with_config(config: ValidationConfig) -> Self {
        Self {
            config,
            xss_pattern: Regex::new(r"(?i)<script[^>]*>.*?</script>|javascript:|on\w+\s*=")
                .expect("Invalid XSS regex"),
            sql_injection_pattern: Regex::new(
                r"(?i)(union|select|insert|update|delete|drop|create|alter|exec|execute)\s+",
            )
            .expect("Invalid SQL injection regex"),
            path_traversal_pattern: Regex::new(r"\.\.[\\/]|[\\/]\.\.[\\/]|[\\/]\.\.$")
                .expect("Invalid path traversal regex"),
            dangerous_chars_pattern: Regex::new(r"[\x00-\x08\x0B\x0C\x0E-\x1F\x7F]")
                .expect("Invalid dangerous characters regex"),
        }
    }

    /// Validate idea content
    pub fn validate_idea(&self, content: &str) -> Result<()> {
        // Check for empty content
        if content.trim().is_empty() {
            return Err(ValidationError::empty_field("idea").into());
        }

        // Length validation
        let content_len = content.chars().count();
        if content_len < self.config.min_idea_length {
            return Err(ValidationError::too_short(
                "idea",
                content_len,
                self.config.min_idea_length,
            )
            .into());
        }

        if content_len > self.config.max_idea_length {
            return Err(ValidationError::too_long(
                "idea",
                content_len,
                self.config.max_idea_length,
            )
            .into());
        }

        // UTF-8 validation - all Rust strings are valid UTF-8, so this check is not needed
        // If we were dealing with bytes, we would check here
        if content.contains('\0') {
            return Err(ValidationError::invalid_characters(
                "idea",
                "null bytes",
                Some("valid characters only".to_string()),
            )
            .into());
        }

        // Unicode validation
        if !self.config.allow_unicode && !content.is_ascii() {
            return Err(ValidationError::invalid_value(
                "idea",
                content,
                "Unicode characters not allowed",
            )
            .into());
        }

        // Security checks
        self.check_security_threats("idea", content)?;

        // Control characters check
        if self.dangerous_chars_pattern.is_match(content) {
            return Err(ValidationError::invalid_characters(
                "idea",
                "control characters",
                Some("printable characters only".to_string()),
            )
            .into());
        }

        Ok(())
    }

    /// Validate file path for security
    pub fn validate_file_path(&self, path: &str) -> Result<PathBuf> {
        // Empty path check
        if path.trim().is_empty() {
            return Err(ValidationError::empty_field("file_path").into());
        }

        // Length check
        if path.len() > self.config.max_path_length {
            return Err(ValidationError::too_long(
                "file_path",
                path.len(),
                self.config.max_path_length,
            )
            .into());
        }

        // Path traversal protection
        if self.path_traversal_pattern.is_match(path) {
            crate::logging::log_security_event(
                "path_traversal_attempt",
                crate::logging::SecurityEventSeverity::High,
                &format!("Path traversal attempt detected: {}", path),
                None,
                None,
            );
            return Err(ValidationError::path_traversal(path.to_string()).into());
        }

        // Normalize the path
        let normalized = Path::new(path).canonicalize().map_err(|e| {
            ValidationError::invalid_format("file_path", format!("Invalid path: {}", e))
        })?;

        // Check if path exists (for operations that require it)
        if !normalized.exists() {
            return Err(
                ValidationError::invalid_value("file_path", path, "Path does not exist").into(),
            );
        }

        // Check if it's a directory when a file is expected
        if normalized.is_dir() {
            return Err(
                ValidationError::invalid_value("file_path", path, "Path is a directory").into(),
            );
        }

        // File extension validation
        if let Some(extension) = normalized.extension().and_then(|ext| ext.to_str()) {
            if !self
                .config
                .allowed_file_extensions
                .contains(&extension.to_lowercase())
            {
                return Err(ValidationError::invalid_extension(
                    extension.to_string(),
                    self.config.allowed_file_extensions.clone(),
                )
                .into());
            }
        }

        Ok(normalized)
    }

    /// Validate file path for output (doesn't need to exist yet)
    pub fn validate_output_path(&self, path: &str) -> Result<PathBuf> {
        // Empty path check
        if path.trim().is_empty() {
            return Err(ValidationError::empty_field("output_path").into());
        }

        // Length check
        if path.len() > self.config.max_path_length {
            return Err(ValidationError::too_long(
                "output_path",
                path.len(),
                self.config.max_path_length,
            )
            .into());
        }

        // Path traversal protection
        if self.path_traversal_pattern.is_match(path) {
            crate::logging::log_security_event(
                "path_traversal_attempt",
                crate::logging::SecurityEventSeverity::High,
                &format!("Path traversal attempt detected: {}", path),
                None,
                None,
            );
            return Err(ValidationError::path_traversal(path.to_string()).into());
        }

        let path_buf = PathBuf::from(path);

        // Check parent directory exists or can be created
        if let Some(parent) = path_buf.parent() {
            if !parent.exists() {
                return Err(ValidationError::invalid_value(
                    "output_path",
                    path,
                    "Parent directory does not exist",
                )
                .into());
            }
        }

        // File extension validation
        if let Some(extension) = path_buf.extension().and_then(|ext| ext.to_str()) {
            if !self
                .config
                .allowed_file_extensions
                .contains(&extension.to_lowercase())
            {
                return Err(ValidationError::invalid_extension(
                    extension.to_string(),
                    self.config.allowed_file_extensions.clone(),
                )
                .into());
            }
        }

        Ok(path_buf)
    }

    /// Validate score value
    pub fn validate_score(&self, score: f64, field_name: &str) -> Result<()> {
        if score < self.config.min_score_value || score > self.config.max_score_value {
            return Err(ValidationError::invalid_value(
                field_name,
                score.to_string(),
                format!(
                    "must be between {} and {}",
                    self.config.min_score_value, self.config.max_score_value
                ),
            )
            .into());
        }

        // Check for NaN or infinite
        if !score.is_finite() {
            return Err(ValidationError::invalid_value(
                field_name,
                score.to_string(),
                "must be a finite number",
            )
            .into());
        }

        Ok(())
    }

    /// Validate limit parameter for queries
    pub fn validate_limit(&self, limit: usize) -> Result<()> {
        if limit == 0 {
            return Err(ValidationError::invalid_value(
                "limit",
                limit.to_string(),
                "must be greater than 0",
            )
            .into());
        }

        if limit > self.config.max_batch_size {
            return Err(
                ValidationError::too_long("limit", limit, self.config.max_batch_size).into(),
            );
        }

        Ok(())
    }

    /// Validate UUID string
    pub fn validate_uuid(&self, uuid_str: &str) -> Result<()> {
        if uuid_str.trim().is_empty() {
            return Err(ValidationError::empty_field("id").into());
        }

        uuid::Uuid::parse_str(uuid_str)
            .map_err(|_| ValidationError::invalid_uuid("id", uuid_str.to_string()))?;

        Ok(())
    }

    /// Validate date/time string format
    pub fn validate_datetime(&self, datetime_str: &str) -> Result<()> {
        if datetime_str.trim().is_empty() {
            return Err(ValidationError::empty_field("datetime").into());
        }

        // Try parsing as ISO 8601 first
        chrono::DateTime::parse_from_rfc3339(datetime_str)
            .or_else(|_| chrono::DateTime::parse_from_str(datetime_str, "%Y-%m-%d %H:%M:%S"))
            .map_err(|_| {
                ValidationError::invalid_date_time(
                    "datetime",
                    datetime_str,
                    "must be in ISO 8601 or YYYY-MM-DD HH:MM:SS format",
                )
            })?;

        Ok(())
    }

    /// Validate email address format
    pub fn validate_email(&self, email: &str) -> Result<()> {
        if email.trim().is_empty() {
            return Err(ValidationError::empty_field("email").into());
        }

        // Basic email validation regex
        let email_regex = Regex::new(r"^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$")
            .expect("Invalid email regex");

        if !email_regex.is_match(email) {
            return Err(ValidationError::invalid_email("email", email.to_string()).into());
        }

        Ok(())
    }

    /// Validate URL format
    pub fn validate_url(&self, url: &str) -> Result<()> {
        if url.trim().is_empty() {
            return Err(ValidationError::empty_field("url").into());
        }

        if !url.starts_with("http://") && !url.starts_with("https://") {
            return Err(ValidationError::invalid_url(
                "url",
                url,
                "must start with http:// or https://",
            )
            .into());
        }

        url::Url::parse(url)
            .map_err(|_| ValidationError::invalid_url("url", url, "invalid URL format"))?;

        Ok(())
    }

    /// Validate batch size for operations
    pub fn validate_batch_size(&self, batch_size: usize) -> Result<()> {
        if batch_size == 0 {
            return Err(ValidationError::invalid_value(
                "batch_size",
                batch_size.to_string(),
                "must be greater than 0",
            )
            .into());
        }

        if batch_size > self.config.max_batch_size {
            return Err(ValidationError::too_long(
                "batch_size",
                batch_size,
                self.config.max_batch_size,
            )
            .into());
        }

        Ok(())
    }

    /// Validate JSON content
    pub fn validate_json(&self, json_str: &str) -> Result<()> {
        if json_str.trim().is_empty() {
            return Err(ValidationError::empty_field("json").into());
        }

        // Try to parse as JSON
        serde_json::from_str::<serde_json::Value>(json_str)
            .map_err(|e| ValidationError::invalid_json("json", e))?;

        Ok(())
    }

    /// Validate that string contains no SQL injection attempts
    pub fn validate_no_sql_injection(&self, input: &str, field_name: &str) -> Result<()> {
        if self.sql_injection_pattern.is_match(input) {
            crate::logging::log_security_event(
                "sql_injection_attempt",
                crate::logging::SecurityEventSeverity::High,
                &format!("SQL injection attempt in {}: {}", field_name, input),
                None,
                None,
            );
            return Err(ValidationError::sql_injection_attempt(input.to_string()).into());
        }

        Ok(())
    }

    /// Validate that string contains no XSS attempts
    pub fn validate_no_xss(&self, input: &str, field_name: &str) -> Result<()> {
        if self.xss_pattern.is_match(input) {
            crate::logging::log_security_event(
                "xss_attempt",
                crate::logging::SecurityEventSeverity::Medium,
                &format!("XSS attempt in {}: {}", field_name, input),
                None,
                None,
            );
            return Err(
                ValidationError::xss_attempt(field_name.to_string(), input.to_string()).into(),
            );
        }

        Ok(())
    }

    /// Comprehensive security threat detection
    fn check_security_threats(&self, field_name: &str, content: &str) -> Result<()> {
        // Check for SQL injection
        if self.sql_injection_pattern.is_match(content) {
            crate::logging::log_security_event(
                "sql_injection_attempt",
                crate::logging::SecurityEventSeverity::High,
                &format!("SQL injection attempt detected in {}", field_name),
                None,
                None,
            );
            return Err(ValidationError::sql_injection_attempt(content.to_string()).into());
        }

        // Check for XSS
        if self.xss_pattern.is_match(content) {
            crate::logging::log_security_event(
                "xss_attempt",
                crate::logging::SecurityEventSeverity::Medium,
                &format!("XSS attempt detected in {}", field_name),
                None,
                None,
            );
            return Err(
                ValidationError::xss_attempt(field_name.to_string(), content.to_string()).into(),
            );
        }

        // Check for command injection patterns
        if content.contains(";")
            && (content.contains("rm ") || content.contains("del ") || content.contains("format "))
        {
            crate::logging::log_security_event(
                "command_injection_attempt",
                crate::logging::SecurityEventSeverity::High,
                &format!("Command injection attempt detected in {}", field_name),
                None,
                None,
            );
            return Err(ValidationError::command_injection_attempt(content.to_string()).into());
        }

        Ok(())
    }

    /// Validate search query parameters
    pub fn validate_search_query(&self, query: &str) -> Result<()> {
        if query.trim().is_empty() {
            return Err(ValidationError::empty_field("search_query").into());
        }

        // Length check for search queries
        if query.len() > 200 {
            return Err(ValidationError::too_long("search_query", query.len(), 200).into());
        }

        // Security checks
        self.check_security_threats("search_query", query)?;

        Ok(())
    }
}

impl Default for InputValidator {
    fn default() -> Self {
        Self::new()
    }
}

/// Utility function to validate common command line arguments
pub fn validate_common_args(idea: Option<&str>, limit: Option<usize>) -> Result<()> {
    let validator = InputValidator::new();

    if let Some(idea_content) = idea {
        validator.validate_idea(idea_content)?;
    }

    if let Some(limit_value) = limit {
        validator.validate_limit(limit_value)?;
    }

    Ok(())
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_valid_idea_validation() {
        let validator = InputValidator::new();
        assert!(validator
            .validate_idea("This is a valid idea with sufficient length")
            .is_ok());
    }

    #[test]
    fn test_empty_idea_validation() {
        let validator = InputValidator::new();
        assert!(validator.validate_idea("").is_err());
        assert!(validator.validate_idea("   ").is_err());
    }

    #[test]
    fn test_idea_length_validation() {
        let validator = InputValidator::new();

        // Too short
        assert!(validator.validate_idea("Hi").is_err());

        // Too long (create a very long string)
        let long_idea = "a".repeat(6000);
        assert!(validator.validate_idea(&long_idea).is_err());
    }

    #[test]
    fn test_path_traversal_prevention() {
        let validator = InputValidator::new();

        // These should all fail
        let malicious_paths = [
            "../../../etc/passwd",
            "/etc/passwd/../../../secret",
            "..\\..\\windows\\system32",
            "folder/../../../etc/passwd",
        ];

        for path in malicious_paths {
            assert!(validator.validate_file_path(path).is_err());
        }
    }

    #[test]
    fn test_score_validation() {
        let validator = InputValidator::new();

        // Valid scores
        assert!(validator.validate_score(0.0, "test_score").is_ok());
        assert!(validator.validate_score(5.5, "test_score").is_ok());
        assert!(validator.validate_score(10.0, "test_score").is_ok());

        // Invalid scores
        assert!(validator.validate_score(-1.0, "test_score").is_err());
        assert!(validator.validate_score(15.0, "test_score").is_err());
        assert!(validator.validate_score(f64::NAN, "test_score").is_err());
        assert!(validator
            .validate_score(f64::INFINITY, "test_score")
            .is_err());
    }

    #[test]
    fn test_sql_injection_detection() {
        let validator = InputValidator::new();

        let malicious_inputs = [
            "'; DROP TABLE users; --",
            "UNION SELECT * FROM users",
            "'; UPDATE users SET password=''; --",
        ];

        for input in malicious_inputs {
            assert!(validator.validate_idea(input).is_err());
        }
    }

    #[test]
    fn test_xss_detection() {
        let validator = InputValidator::new();

        let malicious_inputs = [
            "<script>alert('xss')</script>",
            "javascript:alert('xss')",
            "onclick=alert('xss')",
        ];

        for input in malicious_inputs {
            assert!(validator.validate_idea(input).is_err());
        }
    }

    #[test]
    fn test_uuid_validation() {
        let validator = InputValidator::new();

        // Valid UUID
        let valid_uuid = "550e8400-e29b-41d4-a716-446655440000";
        assert!(validator.validate_uuid(valid_uuid).is_ok());

        // Invalid UUID
        let invalid_uuid = "not-a-uuid";
        assert!(validator.validate_uuid(invalid_uuid).is_err());
    }

    #[test]
    fn test_email_validation() {
        let validator = InputValidator::new();

        // Valid emails
        let valid_emails = [
            "test@example.com",
            "user.name@domain.co.uk",
            "user+tag@example.org",
        ];

        for email in valid_emails {
            assert!(validator.validate_email(email).is_ok());
        }

        // Invalid emails
        let invalid_emails = [
            "not-an-email",
            "@domain.com",
            "user@",
            "user..name@domain.com",
        ];

        for email in invalid_emails {
            assert!(validator.validate_email(email).is_err());
        }
    }

    #[test]
    fn test_url_validation() {
        let validator = InputValidator::new();

        // Valid URLs
        assert!(validator.validate_url("https://example.com").is_ok());
        assert!(validator.validate_url("http://localhost:8080").is_ok());

        // Invalid URLs
        assert!(validator.validate_url("not-a-url").is_err());
        assert!(validator.validate_url("ftp://example.com").is_err());
        assert!(validator.validate_url("").is_err());
    }
}
