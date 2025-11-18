use thiserror::Error;

/// Database-specific errors
#[derive(Error, Debug)]
pub enum DatabaseError {
    #[error("Database connection failed: {source}")]
    Connection {
        #[source]
        source: sqlx::Error,
        database_path: Option<String>,
    },

    #[error("Query failed: {query}")]
    QueryFailed {
        query: String,
        #[source]
        source: sqlx::Error,
    },

    #[error("Database migration failed: {source}")]
    Migration {
        #[source]
        source: sqlx::migrate::MigrateError,
    },

    #[error("Invalid database path: {path}")]
    InvalidPath {
        path: String,
        #[source]
        source: Option<std::io::Error>,
    },

    #[error("Database file not found: {path}")]
    DatabaseNotFound { path: String },

    #[error("Serialization error: {source}")]
    Serialization {
        source: serde_json::Error,
        field: Option<String>,
    },

    #[error("Deserialization error: {source}")]
    Deserialization {
        source: serde_json::Error,
        field: Option<String>,
        data: Option<String>,
    },

    #[error("Row conversion failed: {source}")]
    RowConversion {
        #[source]
        source: sqlx::Error,
        column: Option<String>,
    },

    #[error("Database constraints violated: {constraint}")]
    ConstraintViolation { constraint: String },

    #[error("Database is locked: {source}")]
    DatabaseLocked {
        #[source]
        source: sqlx::Error,
    },

    #[error("Backup operation failed: {operation}")]
    BackupFailed {
        operation: String,
        #[source]
        source: std::io::Error,
    },

    #[error("Database integrity error: {details}")]
    IntegrityError { details: String },

    #[error("Retry operation failed after {attempts} attempts: {operation}")]
    RetryExhausted {
        operation: String,
        attempts: u32,
        total_duration_ms: u64,
        #[source]
        last_error: sqlx::Error,
    },

    #[error("Database operation timeout after {timeout_ms}ms: {operation}")]
    OperationTimeout {
        operation: String,
        timeout_ms: u64,
        #[source]
        source: sqlx::Error,
    },

    #[error("Connection pool exhausted: {details}")]
    PoolExhausted { details: String },

    #[error("Deadlock detected: {query}")]
    DeadlockDetected {
        query: String,
        #[source]
        source: sqlx::Error,
    },
}

impl DatabaseError {
    /// Create a connection error with database path context
    pub fn connection(source: sqlx::Error, database_path: Option<&str>) -> Self {
        Self::Connection {
            source,
            database_path: database_path.map(|s| s.to_string()),
        }
    }

    /// Create a query failed error with the problematic query
    pub fn query_failed(query: impl Into<String>, source: sqlx::Error) -> Self {
        Self::QueryFailed {
            query: query.into(),
            source,
        }
    }

    /// Create an invalid path error
    pub fn invalid_path(path: impl Into<String>) -> Self {
        Self::InvalidPath {
            path: path.into(),
            source: None,
        }
    }

    /// Create a connection timeout error
    pub fn connection_timeout(timeout_ms: u64, _operation: impl Into<String>) -> Self {
        Self::Connection {
            source: sqlx::Error::Protocol(format!("Connection timeout after {}ms", timeout_ms)),
            database_path: None,
        }
    }

    /// Create an invalid path error with underlying IO error
    pub fn invalid_path_with_source(path: impl Into<String>, source: std::io::Error) -> Self {
        Self::InvalidPath {
            path: path.into(),
            source: Some(source),
        }
    }

    /// Create a serialization error with field context
    pub fn serialization_field(source: serde_json::Error, field: impl Into<String>) -> Self {
        Self::Serialization {
            source,
            field: Some(field.into()),
        }
    }

    /// Create a deserialization error with field and data context
    pub fn deserialization_with_data(
        source: serde_json::Error,
        field: impl Into<String>,
        data: impl Into<String>,
    ) -> Self {
        Self::Deserialization {
            source,
            field: Some(field.into()),
            data: Some(data.into()),
        }
    }

    /// Create a row conversion error with column context
    pub fn row_conversion(source: sqlx::Error, column: Option<&str>) -> Self {
        Self::RowConversion {
            source,
            column: column.map(|s| s.to_string()),
        }
    }

    /// Create a backup failed error
    pub fn backup_failed(operation: impl Into<String>, source: std::io::Error) -> Self {
        Self::BackupFailed {
            operation: operation.into(),
            source,
        }
    }

    /// Create a retry exhausted error
    pub fn retry_exhausted(
        operation: impl Into<String>,
        attempts: u32,
        total_duration_ms: u64,
        last_error: sqlx::Error,
    ) -> Self {
        Self::RetryExhausted {
            operation: operation.into(),
            attempts,
            total_duration_ms,
            last_error,
        }
    }

    /// Create an operation timeout error
    pub fn operation_timeout(
        operation: impl Into<String>,
        timeout_ms: u64,
        source: sqlx::Error,
    ) -> Self {
        Self::OperationTimeout {
            operation: operation.into(),
            timeout_ms,
            source,
        }
    }

    /// Create a pool exhausted error
    pub fn pool_exhausted(details: impl Into<String>) -> Self {
        Self::PoolExhausted {
            details: details.into(),
        }
    }

    /// Create a deadlock detected error
    pub fn deadlock_detected(query: impl Into<String>, source: sqlx::Error) -> Self {
        Self::DeadlockDetected {
            query: query.into(),
            source,
        }
    }

    /// Create a duplicate entry error
    pub fn duplicate_entry(
        entity_type: impl Into<String>,
        field: impl Into<String>,
        value: impl Into<String>,
    ) -> Self {
        Self::ConstraintViolation {
            constraint: format!(
                "Duplicate entry for {} field {}: {}",
                entity_type.into(),
                field.into(),
                value.into()
            ),
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use sqlx::Error as SqlxError;

    #[test]
    fn test_database_error_creation() {
        let error = DatabaseError::invalid_path("/invalid/path");
        assert!(matches!(error, DatabaseError::InvalidPath { .. }));

        let error_str = error.to_string();
        assert!(error_str.contains("Invalid database path"));
        assert!(error_str.contains("/invalid/path"));
    }

    #[test]
    fn test_query_failed_error() {
        let query = "SELECT * FROM ideas";
        let sqlx_error = SqlxError::RowNotFound;
        let error = DatabaseError::query_failed(query, sqlx_error);

        assert!(matches!(error, DatabaseError::QueryFailed { .. }));

        let error_str = error.to_string();
        assert!(error_str.contains("Query failed"));
        assert!(error_str.contains("SELECT * FROM ideas"));
    }

    #[test]
    fn test_connection_error_with_path() {
        let sqlx_error = SqlxError::RowNotFound;
        let error = DatabaseError::connection(sqlx_error, Some("/path/to/db.sqlite"));

        if let DatabaseError::Connection { database_path, .. } = error {
            assert_eq!(database_path.as_deref(), Some("/path/to/db.sqlite"));
        } else {
            panic!("Expected Connection error");
        }
    }

    #[test]
    fn test_serialization_field_error() {
        // Create a serialization error by attempting to deserialize invalid JSON
        let json_error = serde_json::from_str::<serde_json::Value>("{ invalid json }").unwrap_err();
        let error = DatabaseError::serialization_field(json_error, "score");

        if let DatabaseError::Serialization { field, .. } = error {
            assert_eq!(field.as_deref(), Some("score"));
        } else {
            panic!("Expected Serialization error");
        }
    }
}
