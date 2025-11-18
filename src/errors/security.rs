use thiserror::Error;

/// Security-related errors
#[derive(Error, Debug)]
pub enum SecurityError {
    #[error("Unauthorized access attempt: {action}")]
    Unauthorized {
        action: String,
        user: Option<String>,
        ip_address: Option<String>,
    },

    #[error("Authentication failed: {reason}")]
    Authentication {
        reason: String,
        #[source]
        source: Option<Box<dyn std::error::Error + Send + Sync>>,
    },

    #[error("Authorization failed: {permission} required")]
    Authorization {
        permission: String,
        resource: String,
        user: Option<String>,
    },

    #[error("CORS policy violation: {origin}")]
    CorsViolation {
        origin: String,
        allowed_origins: Vec<String>,
    },

    #[error("CSRF token invalid or missing")]
    CsrfTokenInvalid {
        source: Option<Box<dyn std::error::Error + Send + Sync>>,
    },

    #[error("Rate limit exceeded: {details}")]
    RateLimit {
        details: String,
        identifier: Option<String>,
        reset_after: Option<std::time::Duration>,
    },

    #[error("IP address blocked: {ip_address} - {reason}")]
    IpBlocked {
        ip_address: String,
        reason: String,
        permanent: bool,
    },

    #[error("File access denied: {path}")]
    FileAccessDenied {
        path: String,
        operation: String,
        reason: String,
    },

    #[error("Malicious request detected: {attack_type}")]
    MaliciousRequest {
        attack_type: String,
        details: String,
        source_ip: Option<String>,
        user_agent: Option<String>,
    },

    #[error("File upload security violation: {reason}")]
    FileUploadViolation {
        reason: String,
        filename: Option<String>,
        file_type: Option<String>,
        file_size: Option<usize>,
    },

    #[error("Invalid SSL/TLS configuration: {reason}")]
    InvalidTlsConfig {
        reason: String,
        #[source]
        source: Option<Box<dyn std::error::Error + Send + Sync>>,
    },

    #[error("Session expired or invalid: {reason}")]
    SessionInvalid {
        reason: String,
        session_id: Option<String>,
    },

    #[error("Permission denied: {resource} requires {permission}")]
    PermissionDenied {
        resource: String,
        permission: String,
        user: Option<String>,
    },

    #[error("Security policy violation: {policy} - {reason}")]
    PolicyViolation {
        policy: String,
        reason: String,
        severity: SecuritySeverity,
    },

    #[error("Audit log failure: {reason}")]
    AuditLogFailure {
        reason: String,
        #[source]
        source: Option<Box<dyn std::error::Error + Send + Sync>>,
    },

    #[error("Encryption/Decryption error: {operation}")]
    EncryptionError {
        operation: String,
        #[source]
        source: Option<Box<dyn std::error::Error + Send + Sync>>,
    },

    #[error("Certificate validation failed: {reason}")]
    CertificateValidation {
        reason: String,
        certificate: Option<String>,
    },

    #[error("Insecure configuration detected: {setting} = {value}")]
    InsecureConfig {
        setting: String,
        value: String,
        recommendation: String,
    },

    #[error("Data leakage risk: {data_type} in {context}")]
    DataLeakageRisk {
        data_type: String,
        context: String,
        risk_level: SecuritySeverity,
    },

    #[error("Password policy violation: {reason}")]
    PasswordPolicy {
        reason: String,
        user: Option<String>,
    },
}

/// Security severity levels for security events
#[derive(Debug, Clone, PartialEq, Eq, PartialOrd, Ord)]
pub enum SecuritySeverity {
    Low,
    Medium,
    High,
    Critical,
}

impl SecuritySeverity {
    pub fn as_str(&self) -> &'static str {
        match self {
            SecuritySeverity::Low => "low",
            SecuritySeverity::Medium => "medium",
            SecuritySeverity::High => "high",
            SecuritySeverity::Critical => "critical",
        }
    }
}

impl std::fmt::Display for SecuritySeverity {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        write!(f, "{}", self.as_str())
    }
}

impl SecurityError {
    /// Create an unauthorized error
    pub fn unauthorized(action: impl Into<String>) -> Self {
        Self::Unauthorized {
            action: action.into(),
            user: None,
            ip_address: None,
        }
    }

    /// Create an unauthorized error with user and IP context
    pub fn unauthorized_with_context(
        action: impl Into<String>,
        user: impl Into<String>,
        ip_address: impl Into<String>,
    ) -> Self {
        Self::Unauthorized {
            action: action.into(),
            user: Some(user.into()),
            ip_address: Some(ip_address.into()),
        }
    }

    /// Create an authentication error
    pub fn authentication(reason: impl Into<String>) -> Self {
        Self::Authentication {
            reason: reason.into(),
            source: None,
        }
    }

    /// Create an authentication error with source
    pub fn authentication_with_source(
        reason: impl Into<String>,
        source: Box<dyn std::error::Error + Send + Sync>,
    ) -> Self {
        Self::Authentication {
            reason: reason.into(),
            source: Some(source),
        }
    }

    /// Create an authorization error
    pub fn authorization(permission: impl Into<String>, resource: impl Into<String>) -> Self {
        Self::Authorization {
            permission: permission.into(),
            resource: resource.into(),
            user: None,
        }
    }

    /// Create an authorization error with user context
    pub fn authorization_with_user(
        permission: impl Into<String>,
        resource: impl Into<String>,
        user: impl Into<String>,
    ) -> Self {
        Self::Authorization {
            permission: permission.into(),
            resource: resource.into(),
            user: Some(user.into()),
        }
    }

    /// Create a CORS violation error
    pub fn cors_violation(origin: impl Into<String>, allowed_origins: Vec<String>) -> Self {
        Self::CorsViolation {
            origin: origin.into(),
            allowed_origins,
        }
    }

    /// Create a rate limit error
    pub fn rate_limit(
        details: impl Into<String>,
        identifier: Option<String>,
        reset_after: Option<std::time::Duration>,
    ) -> Self {
        Self::RateLimit {
            details: details.into(),
            identifier,
            reset_after,
        }
    }

    /// Create an IP blocked error
    pub fn ip_blocked(
        ip_address: impl Into<String>,
        reason: impl Into<String>,
        permanent: bool,
    ) -> Self {
        Self::IpBlocked {
            ip_address: ip_address.into(),
            reason: reason.into(),
            permanent,
        }
    }

    /// Create a file access denied error
    pub fn file_access_denied(
        path: impl Into<String>,
        operation: impl Into<String>,
        reason: impl Into<String>,
    ) -> Self {
        Self::FileAccessDenied {
            path: path.into(),
            operation: operation.into(),
            reason: reason.into(),
        }
    }

    /// Create a malicious request error
    pub fn malicious_request(
        attack_type: impl Into<String>,
        details: impl Into<String>,
        source_ip: Option<String>,
        user_agent: Option<String>,
    ) -> Self {
        Self::MaliciousRequest {
            attack_type: attack_type.into(),
            details: details.into(),
            source_ip,
            user_agent,
        }
    }

    /// Create a file upload violation error
    pub fn file_upload_violation(
        reason: impl Into<String>,
        filename: Option<String>,
        file_type: Option<String>,
        file_size: Option<usize>,
    ) -> Self {
        Self::FileUploadViolation {
            reason: reason.into(),
            filename,
            file_type,
            file_size,
        }
    }

    /// Create an insecure configuration error
    pub fn insecure_config(
        setting: impl Into<String>,
        value: impl Into<String>,
        recommendation: impl Into<String>,
    ) -> Self {
        Self::InsecureConfig {
            setting: setting.into(),
            value: value.into(),
            recommendation: recommendation.into(),
        }
    }

    /// Create a data leakage risk error
    pub fn data_leakage_risk(
        data_type: impl Into<String>,
        context: impl Into<String>,
        risk_level: SecuritySeverity,
    ) -> Self {
        Self::DataLeakageRisk {
            data_type: data_type.into(),
            context: context.into(),
            risk_level,
        }
    }

    /// Create a policy violation error
    pub fn policy_violation(
        policy: impl Into<String>,
        reason: impl Into<String>,
        severity: SecuritySeverity,
    ) -> Self {
        Self::PolicyViolation {
            policy: policy.into(),
            reason: reason.into(),
            severity,
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::time::Duration;

    #[test]
    fn test_unauthorized_error() {
        let error = SecurityError::unauthorized("Delete all ideas");
        assert!(matches!(error, SecurityError::Unauthorized { .. }));

        let error_str = error.to_string();
        assert!(error_str.contains("Delete all ideas"));
        assert!(error_str.contains("Unauthorized access"));
    }

    #[test]
    fn test_unauthorized_with_context() {
        let error = SecurityError::unauthorized_with_context(
            "Access admin panel",
            "user123",
            "192.168.1.100",
        );

        if let SecurityError::Unauthorized {
            action,
            user,
            ip_address,
        } = error
        {
            assert_eq!(action, "Access admin panel");
            assert_eq!(user.as_deref(), Some("user123"));
            assert_eq!(ip_address.as_deref(), Some("192.168.1.100"));
        } else {
            panic!("Expected Unauthorized error");
        }
    }

    #[test]
    fn test_rate_limit_error() {
        let reset_time = Duration::from_secs(300);
        let error = SecurityError::rate_limit(
            "Too many requests per minute",
            Some("user123".to_string()),
            Some(reset_time),
        );

        if let SecurityError::RateLimit {
            details,
            identifier,
            reset_after,
        } = error
        {
            assert_eq!(details, "Too many requests per minute");
            assert_eq!(identifier.as_deref(), Some("user123"));
            assert_eq!(reset_after, Some(reset_time));
        } else {
            panic!("Expected RateLimit error");
        }
    }

    #[test]
    fn test_ip_blocked_error() {
        let error = SecurityError::ip_blocked("192.168.1.100", "Malicious activity", false);

        if let SecurityError::IpBlocked {
            ip_address,
            reason,
            permanent,
        } = error
        {
            assert_eq!(ip_address, "192.168.1.100");
            assert_eq!(reason, "Malicious activity");
            assert!(!permanent);
        } else {
            panic!("Expected IpBlocked error");
        }
    }

    #[test]
    fn test_malicious_request_error() {
        let error = SecurityError::malicious_request(
            "SQL Injection",
            "DROP TABLE users;",
            Some("192.168.1.100".to_string()),
            Some("Mozilla/5.0".to_string()),
        );

        if let SecurityError::MaliciousRequest {
            attack_type,
            details,
            source_ip,
            user_agent,
        } = error
        {
            assert_eq!(attack_type, "SQL Injection");
            assert_eq!(details, "DROP TABLE users;");
            assert_eq!(source_ip.as_deref(), Some("192.168.1.100"));
            assert_eq!(user_agent.as_deref(), Some("Mozilla/5.0"));
        } else {
            panic!("Expected MaliciousRequest error");
        }
    }

    #[test]
    fn test_insecure_config_error() {
        let error = SecurityError::insecure_config(
            "debug_mode",
            "true",
            "Disable debug mode in production",
        );

        if let SecurityError::InsecureConfig {
            setting,
            value,
            recommendation,
        } = error
        {
            assert_eq!(setting, "debug_mode");
            assert_eq!(value, "true");
            assert_eq!(recommendation, "Disable debug mode in production");
        } else {
            panic!("Expected InsecureConfig error");
        }
    }

    #[test]
    fn test_security_severity() {
        assert_eq!(SecuritySeverity::Low.as_str(), "low");
        assert_eq!(SecuritySeverity::Medium.as_str(), "medium");
        assert_eq!(SecuritySeverity::High.as_str(), "high");
        assert_eq!(SecuritySeverity::Critical.as_str(), "critical");

        assert!(SecuritySeverity::High > SecuritySeverity::Medium);
        assert!(SecuritySeverity::Critical > SecuritySeverity::High);

        let display_str = SecuritySeverity::Medium.to_string();
        assert_eq!(display_str, "medium");
    }

    #[test]
    fn test_policy_violation_error() {
        let error = SecurityError::policy_violation(
            "No weak passwords",
            "Password '123456' is too weak",
            SecuritySeverity::High,
        );

        if let SecurityError::PolicyViolation {
            policy,
            reason,
            severity,
        } = error
        {
            assert_eq!(policy, "No weak passwords");
            assert_eq!(reason, "Password '123456' is too weak");
            assert_eq!(severity, SecuritySeverity::High);
        } else {
            panic!("Expected PolicyViolation error");
        }
    }
}
