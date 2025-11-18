//! Structured logging utilities for the telos-idea-matrix application.
//!
//! This module provides comprehensive logging functionality with:
//! - Request/response tracing with correlation IDs
//! - Performance metrics and timing
//! - Error tracking with context
//! - Security event logging
//! - Structured JSON output for analysis

use std::time::{Duration, Instant};
use tracing::{error, info, span, warn, Level, Span};
use tracing_subscriber::{layer::SubscriberExt, util::SubscriberInitExt, EnvFilter, Layer};
use uuid::Uuid;

/// Application-wide logging configuration
#[derive(Debug, Clone)]
pub struct LoggingConfig {
    /// Minimum log level to output
    pub level: Level,
    /// Whether to output in JSON format (useful for production)
    pub json_format: bool,
    /// Whether to include timestamps in console output
    pub include_timestamps: bool,
    /// Whether to display target module in console output
    pub display_target: bool,
    /// Directory for log files (optional)
    pub log_directory: Option<std::path::PathBuf>,
    /// Maximum log file size in bytes before rotation
    pub max_log_file_size: Option<u64>,
    /// Number of log files to retain
    pub log_file_retention: Option<usize>,
}

impl Default for LoggingConfig {
    fn default() -> Self {
        Self {
            level: Level::INFO,
            json_format: false,
            include_timestamps: true,
            display_target: false,
            log_directory: None,
            max_log_file_size: Some(10 * 1024 * 1024), // 10MB
            log_file_retention: Some(5),
        }
    }
}

/// Initialize the global tracing subscriber with the given configuration
pub fn init_logging(config: LoggingConfig) -> Result<(), Box<dyn std::error::Error>> {
    let env_filter = EnvFilter::try_from_default_env()
        .unwrap_or_else(|_| EnvFilter::new(format!("telos_idea_matrix={}", config.level)));

    // Create a subscriber with only error logs going to console for clean user experience
    let console_subscriber = tracing_subscriber::registry().with(
        tracing_subscriber::fmt::layer()
            .with_target(config.display_target)
            .with_timer(tracing_subscriber::fmt::time::UtcTime::rfc_3339())
            .with_level(true)
            .with_ansi(true)
            .with_filter(tracing_subscriber::filter::LevelFilter::ERROR),
    );

    // Determine log directory
    let log_dir = if let Some(configured_dir) = &config.log_directory {
        configured_dir.clone()
    } else {
        dirs::data_dir()
            .unwrap_or_else(|| std::path::PathBuf::from("/tmp"))
            .join("telos-idea-matrix")
            .join("logs")
    };

    std::fs::create_dir_all(&log_dir)?;

    // Use daily rotation for detailed logs
    let file_appender = tracing_appender::rolling::daily(&log_dir, "telos-matrix.log");

    let file_layer = tracing_subscriber::fmt::layer()
        .with_writer(file_appender)
        .with_ansi(false)
        .json()
        .with_filter(env_filter); // Apply the original filter to file logs

    // Combine console (error only) and file (full logs) subscribers
    console_subscriber.with(file_layer).init();

    info!(
        app_name = "telos-idea-matrix",
        version = env!("CARGO_PKG_VERSION"),
        log_level = %config.level,
        json_format = config.json_format,
        log_directory = %log_dir.display(),
        "Logging initialized"
    );

    Ok(())
}

/// A utility for measuring operation duration and logging performance metrics
#[derive(Debug)]
pub struct OperationTimer {
    start_time: Instant,
    operation_name: String,
    span: Span,
}

impl OperationTimer {
    /// Start timing a new operation
    pub fn new(operation_name: impl Into<String>) -> Self {
        let name = operation_name.into();
        let span = span!(Level::INFO, "operation", name = %name);

        tracing::trace!(
            operation_name = %name,
            "Starting operation"
        );

        Self {
            start_time: Instant::now(),
            operation_name: name,
            span,
        }
    }

    /// Start timing a new operation with additional context
    pub fn new_with_fields(operation_name: impl Into<String>, fields: &[(&str, &str)]) -> Self {
        let name = operation_name.into();
        let span = span!(Level::INFO, "operation", name = %name);

        // Add custom fields to the span
        for (key, value) in fields {
            span.record(*key, value);
        }

        tracing::trace!(
            operation_name = %name,
            fields = ?fields,
            "Starting operation with context"
        );

        Self {
            start_time: Instant::now(),
            operation_name: name,
            span,
        }
    }

    /// Complete the operation and log its duration
    pub fn complete(self) {
        let duration = self.start_time.elapsed();

        info!(
            operation_name = %self.operation_name,
            duration_ms = duration.as_millis(),
            "Operation completed successfully"
        );

        drop(self.span);
    }

    /// Complete the operation with additional context
    pub fn complete_with_context(self, additional_fields: &[(&str, &str)]) {
        let duration = self.start_time.elapsed();

        info!(
            operation_name = %self.operation_name,
            duration_ms = duration.as_millis(),
            additional_context = ?additional_fields,
            "Operation completed successfully"
        );

        drop(self.span);
    }

    /// Record an error for this operation
    pub fn error(self, error: &dyn std::error::Error) {
        let duration = self.start_time.elapsed();

        error!(
            operation_name = %self.operation_name,
            duration_ms = duration.as_millis(),
            error = %error,
            error_type = %std::any::type_name_of_val(error),
            "Operation failed"
        );

        drop(self.span);
    }

    /// Get the elapsed duration without completing the operation
    pub fn elapsed(&self) -> Duration {
        self.start_time.elapsed()
    }
}

/// Generate a unique correlation ID for request tracking
pub fn generate_correlation_id() -> String {
    Uuid::new_v4().to_string()
}

/// Log security-related events with standardized fields
pub fn log_security_event(
    event_type: &str,
    severity: SecurityEventSeverity,
    details: &str,
    user_context: Option<&str>,
    ip_address: Option<&str>,
) {
    let correlation_id = generate_correlation_id();

    match severity {
        SecurityEventSeverity::Low => {
            tracing::info!(
                event_type = event_type,
                severity = %severity,
                details = details,
                user_context = user_context,
                ip_address = ip_address,
                correlation_id = %correlation_id,
                "Security event"
            );
        }
        SecurityEventSeverity::Medium => {
            tracing::warn!(
                event_type = event_type,
                severity = %severity,
                details = details,
                user_context = user_context,
                ip_address = ip_address,
                correlation_id = %correlation_id,
                "Security event"
            );
        }
        SecurityEventSeverity::High | SecurityEventSeverity::Critical => {
            tracing::error!(
                event_type = event_type,
                severity = %severity,
                details = details,
                user_context = user_context,
                ip_address = ip_address,
                correlation_id = %correlation_id,
                "Security event"
            );
        }
    }
}

/// Severity levels for security events
#[derive(Debug, Clone, Copy)]
pub enum SecurityEventSeverity {
    Low,
    Medium,
    High,
    Critical,
}

impl std::fmt::Display for SecurityEventSeverity {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            SecurityEventSeverity::Low => write!(f, "low"),
            SecurityEventSeverity::Medium => write!(f, "medium"),
            SecurityEventSeverity::High => write!(f, "high"),
            SecurityEventSeverity::Critical => write!(f, "critical"),
        }
    }
}

/// Log database operations with timing and context
pub fn log_database_operation(
    operation: &str,
    table: Option<&str>,
    duration: Duration,
    success: bool,
    error: Option<&str>,
) {
    if success {
        info!(
            operation = operation,
            table = table,
            duration_ms = duration.as_millis(),
            "Database operation completed"
        );
    } else {
        error!(
            operation = operation,
            table = table,
            duration_ms = duration.as_millis(),
            error = error,
            "Database operation failed"
        );
    }
}

/// Log AI service interactions
pub fn log_ai_request(
    model: &str,
    request_type: &str,
    duration: Duration,
    success: bool,
    tokens_used: Option<u32>,
    error: Option<&str>,
) {
    if success {
        info!(
            model = model,
            request_type = request_type,
            duration_ms = duration.as_millis(),
            tokens_used = tokens_used,
            "AI request completed"
        );
    } else {
        warn!(
            model = model,
            request_type = request_type,
            duration_ms = duration.as_millis(),
            error = error,
            "AI request failed"
        );
    }
}

/// Log validation errors with structured context
pub fn log_validation_error(
    field: &str,
    value: &str,
    validation_type: &str,
    reason: &str,
    user_context: Option<&str>,
) {
    warn!(
        field = field,
        value = %value.chars().take(100).collect::<String>(), // Limit value length
        validation_type = validation_type,
        reason = reason,
        user_context = user_context,
        "Input validation failed"
    );
}

/// Create a child span for request processing with correlation ID
pub fn create_request_span(
    operation: &str,
    correlation_id: Option<&str>,
    user_id: Option<&str>,
) -> Span {
    let id = if let Some(cid) = correlation_id {
        cid.to_string()
    } else {
        generate_correlation_id()
    };

    let span = span!(
        Level::INFO,
        "request",
        operation = operation,
        correlation_id = %id,
        user_id = user_id
    );

    tracing::info!(
        operation = operation,
        correlation_id = %id,
        user_id = user_id,
        "Request started"
    );

    span
}

/// Log application metrics
pub fn log_metrics(metrics: &AppMetrics) {
    tracing::info!(
        total_requests = metrics.total_requests,
        successful_requests = metrics.successful_requests,
        failed_requests = metrics.failed_requests,
        avg_response_time_ms = metrics.avg_response_time_ms,
        active_connections = metrics.active_connections,
        "Application metrics"
    );
}

/// Application performance metrics
#[derive(Debug, Default)]
pub struct AppMetrics {
    pub total_requests: u64,
    pub successful_requests: u64,
    pub failed_requests: u64,
    pub avg_response_time_ms: f64,
    pub active_connections: u32,
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::thread;
    use std::time::Duration;

    #[test]
    fn test_operation_timer() {
        let timer = OperationTimer::new("test_operation");
        thread::sleep(Duration::from_millis(10));

        let elapsed = timer.elapsed();
        assert!(elapsed.as_millis() >= 10);

        timer.complete();
    }

    #[test]
    fn test_correlation_id_generation() {
        let id1 = generate_correlation_id();
        let id2 = generate_correlation_id();

        assert_ne!(id1, id2);
        assert_eq!(id1.len(), 36); // UUID string length
    }

    #[test]
    fn test_security_event_severity_display() {
        assert_eq!(SecurityEventSeverity::Low.to_string(), "low");
        assert_eq!(SecurityEventSeverity::Critical.to_string(), "critical");
    }

    #[test]
    fn test_logging_config_default() {
        let config = LoggingConfig::default();
        assert_eq!(config.level, Level::INFO);
        assert!(!config.json_format);
        assert!(config.include_timestamps);
    }
}
