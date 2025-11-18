use thiserror::Error;

/// Circuit breaker and external service related errors
#[derive(Error, Debug)]
pub enum CircuitBreakerError {
    #[error("Circuit breaker is open for service: {service_name}")]
    CircuitOpen {
        service_name: String,
        last_failure_time: std::time::SystemTime,
        failure_count: u32,
        retry_after: Option<std::time::Duration>,
    },

    #[error("Service unavailable: {service_name} - {reason}")]
    ServiceUnavailable {
        service_name: String,
        reason: String,
        #[source]
        source: Option<Box<dyn std::error::Error + Send + Sync>>,
    },

    #[error("External service timeout: {service_name} after {timeout_ms}ms")]
    ServiceTimeout {
        service_name: String,
        timeout_ms: u64,
        operation: String,
    },

    #[error("Rate limit exceeded for service: {service_name} - {details}")]
    RateLimitExceeded {
        service_name: String,
        details: String,
        retry_after: Option<std::time::Duration>,
        current_limit: u32,
        window_size: std::time::Duration,
    },

    #[error("Service response malformed: {service_name}")]
    MalformedResponse {
        service_name: String,
        response_preview: String,
        #[source]
        source: Option<Box<dyn std::error::Error + Send + Sync>>,
    },

    #[error("Service authentication failed: {service_name}")]
    AuthenticationFailed {
        service_name: String,
        auth_type: String,
        details: String,
    },

    #[error("Service returned unexpected status: {service_name} - {status}")]
    UnexpectedStatus {
        service_name: String,
        status: u16,
        response_body: Option<String>,
    },

    #[error("Network error for service: {service_name} - {error}")]
    NetworkError {
        service_name: String,
        error: String,
        #[source]
        source: Option<Box<dyn std::error::Error + Send + Sync>>,
    },

    #[error("Configuration error for service: {service_name} - {details}")]
    ServiceConfigurationError {
        service_name: String,
        details: String,
        config_field: Option<String>,
    },

    #[error("Service dependency missing: {service_name} requires {dependency}")]
    DependencyMissing {
        service_name: String,
        dependency: String,
    },

    #[error("Service response too large: {service_name} - {size} bytes (max: {max_size})")]
    ResponseTooLarge {
        service_name: String,
        size: usize,
        max_size: usize,
    },

    #[error("Service quota exceeded: {service_name} - {quota_type}")]
    QuotaExceeded {
        service_name: String,
        quota_type: String,
        current_usage: u64,
        limit: u64,
        reset_time: Option<std::time::SystemTime>,
    },

    #[error("Invalid request to service: {service_name} - {field}")]
    InvalidRequest {
        service_name: String,
        field: String,
        value: String,
        reason: String,
    },

    #[error("Service maintenance mode: {service_name} - {message}")]
    MaintenanceMode {
        service_name: String,
        message: String,
        estimated_downtime: Option<std::time::Duration>,
    },

    #[error("Service version incompatible: {service_name} - {version}")]
    VersionIncompatible {
        service_name: String,
        version: String,
        required_version: String,
    },

    #[error("Service health check failed: {service_name} - {check_name}")]
    HealthCheckFailed {
        service_name: String,
        check_name: String,
        details: String,
    },

    #[error("DNS resolution failed: {service_name} - {hostname}")]
    DnsResolutionFailed {
        service_name: String,
        hostname: String,
        #[source]
        source: Option<Box<dyn std::error::Error + Send + Sync>>,
    },

    #[error("Connection refused: {service_name} at {address}")]
    ConnectionRefused {
        service_name: String,
        address: String,
        #[source]
        source: Option<Box<dyn std::error::Error + Send + Sync>>,
    },
}

impl CircuitBreakerError {
    /// Create a circuit open error
    pub fn circuit_open(
        service_name: impl Into<String>,
        last_failure_time: std::time::SystemTime,
        failure_count: u32,
        retry_after: Option<std::time::Duration>,
    ) -> Self {
        Self::CircuitOpen {
            service_name: service_name.into(),
            last_failure_time,
            failure_count,
            retry_after,
        }
    }

    /// Create a service unavailable error
    pub fn service_unavailable(service_name: impl Into<String>, reason: impl Into<String>) -> Self {
        Self::ServiceUnavailable {
            service_name: service_name.into(),
            reason: reason.into(),
            source: None,
        }
    }

    /// Create a service unavailable error with source
    pub fn service_unavailable_with_source(
        service_name: impl Into<String>,
        reason: impl Into<String>,
        source: Box<dyn std::error::Error + Send + Sync>,
    ) -> Self {
        Self::ServiceUnavailable {
            service_name: service_name.into(),
            reason: reason.into(),
            source: Some(source),
        }
    }

    /// Create a service timeout error
    pub fn service_timeout(
        service_name: impl Into<String>,
        timeout_ms: u64,
        operation: impl Into<String>,
    ) -> Self {
        Self::ServiceTimeout {
            service_name: service_name.into(),
            timeout_ms,
            operation: operation.into(),
        }
    }

    /// Create a rate limit exceeded error
    pub fn rate_limit_exceeded(
        service_name: impl Into<String>,
        details: impl Into<String>,
        retry_after: Option<std::time::Duration>,
        current_limit: u32,
        window_size: std::time::Duration,
    ) -> Self {
        Self::RateLimitExceeded {
            service_name: service_name.into(),
            details: details.into(),
            retry_after,
            current_limit,
            window_size,
        }
    }

    /// Create a malformed response error
    pub fn malformed_response(
        service_name: impl Into<String>,
        response_preview: impl Into<String>,
    ) -> Self {
        Self::MalformedResponse {
            service_name: service_name.into(),
            response_preview: response_preview.into(),
            source: None,
        }
    }

    /// Create a malformed response error with source
    pub fn malformed_response_with_source(
        service_name: impl Into<String>,
        response_preview: impl Into<String>,
        source: Box<dyn std::error::Error + Send + Sync>,
    ) -> Self {
        Self::MalformedResponse {
            service_name: service_name.into(),
            response_preview: response_preview.into(),
            source: Some(source),
        }
    }

    /// Create an authentication failed error
    pub fn authentication_failed(
        service_name: impl Into<String>,
        auth_type: impl Into<String>,
        details: impl Into<String>,
    ) -> Self {
        Self::AuthenticationFailed {
            service_name: service_name.into(),
            auth_type: auth_type.into(),
            details: details.into(),
        }
    }

    /// Create an unexpected status error
    pub fn unexpected_status(
        service_name: impl Into<String>,
        status: u16,
        response_body: Option<String>,
    ) -> Self {
        Self::UnexpectedStatus {
            service_name: service_name.into(),
            status,
            response_body,
        }
    }

    /// Create a network error
    pub fn network_error(service_name: impl Into<String>, error: impl Into<String>) -> Self {
        Self::NetworkError {
            service_name: service_name.into(),
            error: error.into(),
            source: None,
        }
    }

    /// Create a network error with source
    pub fn network_error_with_source(
        service_name: impl Into<String>,
        error: impl Into<String>,
        source: Box<dyn std::error::Error + Send + Sync>,
    ) -> Self {
        Self::NetworkError {
            service_name: service_name.into(),
            error: error.into(),
            source: Some(source),
        }
    }

    /// Create a service configuration error
    pub fn service_configuration_error(
        service_name: impl Into<String>,
        details: impl Into<String>,
        config_field: Option<String>,
    ) -> Self {
        Self::ServiceConfigurationError {
            service_name: service_name.into(),
            details: details.into(),
            config_field,
        }
    }

    /// Create a dependency missing error
    pub fn dependency_missing(
        service_name: impl Into<String>,
        dependency: impl Into<String>,
    ) -> Self {
        Self::DependencyMissing {
            service_name: service_name.into(),
            dependency: dependency.into(),
        }
    }

    /// Create a response too large error
    pub fn response_too_large(
        service_name: impl Into<String>,
        size: usize,
        max_size: usize,
    ) -> Self {
        Self::ResponseTooLarge {
            service_name: service_name.into(),
            size,
            max_size,
        }
    }

    /// Create a quota exceeded error
    pub fn quota_exceeded(
        service_name: impl Into<String>,
        quota_type: impl Into<String>,
        current_usage: u64,
        limit: u64,
        reset_time: Option<std::time::SystemTime>,
    ) -> Self {
        Self::QuotaExceeded {
            service_name: service_name.into(),
            quota_type: quota_type.into(),
            current_usage,
            limit,
            reset_time,
        }
    }

    /// Create an invalid request error
    pub fn invalid_request(
        service_name: impl Into<String>,
        field: impl Into<String>,
        value: impl Into<String>,
        reason: impl Into<String>,
    ) -> Self {
        Self::InvalidRequest {
            service_name: service_name.into(),
            field: field.into(),
            value: value.into(),
            reason: reason.into(),
        }
    }

    /// Create a maintenance mode error
    pub fn maintenance_mode(
        service_name: impl Into<String>,
        message: impl Into<String>,
        estimated_downtime: Option<std::time::Duration>,
    ) -> Self {
        Self::MaintenanceMode {
            service_name: service_name.into(),
            message: message.into(),
            estimated_downtime,
        }
    }

    /// Create a version incompatible error
    pub fn version_incompatible(
        service_name: impl Into<String>,
        version: impl Into<String>,
        required_version: impl Into<String>,
    ) -> Self {
        Self::VersionIncompatible {
            service_name: service_name.into(),
            version: version.into(),
            required_version: required_version.into(),
        }
    }

    /// Create a health check failed error
    pub fn health_check_failed(
        service_name: impl Into<String>,
        check_name: impl Into<String>,
        details: impl Into<String>,
    ) -> Self {
        Self::HealthCheckFailed {
            service_name: service_name.into(),
            check_name: check_name.into(),
            details: details.into(),
        }
    }

    /// Create a DNS resolution failed error
    pub fn dns_resolution_failed(
        service_name: impl Into<String>,
        hostname: impl Into<String>,
    ) -> Self {
        Self::DnsResolutionFailed {
            service_name: service_name.into(),
            hostname: hostname.into(),
            source: None,
        }
    }

    /// Create a DNS resolution failed error with source
    pub fn dns_resolution_failed_with_source(
        service_name: impl Into<String>,
        hostname: impl Into<String>,
        source: Box<dyn std::error::Error + Send + Sync>,
    ) -> Self {
        Self::DnsResolutionFailed {
            service_name: service_name.into(),
            hostname: hostname.into(),
            source: Some(source),
        }
    }

    /// Create a connection refused error
    pub fn connection_refused(service_name: impl Into<String>, address: impl Into<String>) -> Self {
        Self::ConnectionRefused {
            service_name: service_name.into(),
            address: address.into(),
            source: None,
        }
    }

    /// Create a connection refused error with source
    pub fn connection_refused_with_source(
        service_name: impl Into<String>,
        address: impl Into<String>,
        source: Box<dyn std::error::Error + Send + Sync>,
    ) -> Self {
        Self::ConnectionRefused {
            service_name: service_name.into(),
            address: address.into(),
            source: Some(source),
        }
    }

    /// Check if this error represents a temporary failure that might be resolved by retrying
    pub fn is_retryable(&self) -> bool {
        match self {
            Self::CircuitOpen { .. } => false, // Circuit breaker prevents immediate retries
            Self::ServiceUnavailable { .. } => true,
            Self::ServiceTimeout { .. } => true,
            Self::RateLimitExceeded { .. } => true,
            Self::NetworkError { .. } => true,
            Self::ConnectionRefused { .. } => true,
            Self::DnsResolutionFailed { .. } => true,
            Self::MaintenanceMode { .. } => false, // Don't retry during maintenance
            Self::AuthenticationFailed { .. } => false, // Retrying won't help
            Self::ServiceConfigurationError { .. } => false,
            Self::DependencyMissing { .. } => false,
            Self::VersionIncompatible { .. } => false,
            Self::QuotaExceeded {
                reset_time: Some(_),
                ..
            } => true, // Retry after reset
            Self::QuotaExceeded {
                reset_time: None, ..
            } => false,
            _ => false,
        }
    }

    /// Get the service name associated with this error
    pub fn service_name(&self) -> Option<&str> {
        match self {
            Self::CircuitOpen { service_name, .. } => Some(service_name),
            Self::ServiceUnavailable { service_name, .. } => Some(service_name),
            Self::ServiceTimeout { service_name, .. } => Some(service_name),
            Self::RateLimitExceeded { service_name, .. } => Some(service_name),
            Self::MalformedResponse { service_name, .. } => Some(service_name),
            Self::AuthenticationFailed { service_name, .. } => Some(service_name),
            Self::UnexpectedStatus { service_name, .. } => Some(service_name),
            Self::NetworkError { service_name, .. } => Some(service_name),
            Self::ServiceConfigurationError { service_name, .. } => Some(service_name),
            Self::DependencyMissing { service_name, .. } => Some(service_name),
            Self::ResponseTooLarge { service_name, .. } => Some(service_name),
            Self::QuotaExceeded { service_name, .. } => Some(service_name),
            Self::InvalidRequest { service_name, .. } => Some(service_name),
            Self::MaintenanceMode { service_name, .. } => Some(service_name),
            Self::VersionIncompatible { service_name, .. } => Some(service_name),
            Self::HealthCheckFailed { service_name, .. } => Some(service_name),
            Self::DnsResolutionFailed { service_name, .. } => Some(service_name),
            Self::ConnectionRefused { service_name, .. } => Some(service_name),
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::time::Duration;

    #[test]
    fn test_circuit_open_error() {
        let last_failure = std::time::SystemTime::now();
        let error = CircuitBreakerError::circuit_open(
            "test-service",
            last_failure,
            5,
            Some(Duration::from_secs(60)),
        );

        assert!(matches!(error, CircuitBreakerError::CircuitOpen { .. }));
        assert_eq!(error.service_name(), Some("test-service"));
        assert!(!error.is_retryable());
    }

    #[test]
    fn test_service_timeout_error() {
        let error = CircuitBreakerError::service_timeout("test-service", 5000, "get_data");

        assert!(matches!(error, CircuitBreakerError::ServiceTimeout { .. }));
        assert_eq!(error.service_name(), Some("test-service"));
        assert!(error.is_retryable());
    }

    #[test]
    fn test_authentication_failed_error() {
        let error = CircuitBreakerError::authentication_failed(
            "test-service",
            "Bearer token",
            "Token expired",
        );

        assert!(matches!(
            error,
            CircuitBreakerError::AuthenticationFailed { .. }
        ));
        assert_eq!(error.service_name(), Some("test-service"));
        assert!(!error.is_retryable());
    }

    #[test]
    fn test_rate_limit_exceeded_error() {
        let error = CircuitBreakerError::rate_limit_exceeded(
            "test-service",
            "Too many requests",
            Some(Duration::from_secs(300)),
            100,
            Duration::from_secs(60),
        );

        assert!(matches!(
            error,
            CircuitBreakerError::RateLimitExceeded { .. }
        ));
        assert_eq!(error.service_name(), Some("test-service"));
        assert!(error.is_retryable());
    }

    #[test]
    fn test_maintenance_mode_error() {
        let error = CircuitBreakerError::maintenance_mode(
            "test-service",
            "Scheduled maintenance",
            Some(Duration::from_secs(3600)),
        );

        assert!(matches!(error, CircuitBreakerError::MaintenanceMode { .. }));
        assert_eq!(error.service_name(), Some("test-service"));
        assert!(!error.is_retryable());
    }
}
