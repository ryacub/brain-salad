use crate::errors::DatabaseError;
pub type DatabaseResult<T> = std::result::Result<T, DatabaseError>;
use crate::logging;
use crate::metrics;
use chrono::{DateTime, Utc};
use rand::Rng;
use serde::{Deserialize, Serialize};
use sqlx::{query, query_scalar, sqlite::SqlitePoolOptions, Error as SqlxError, Row, SqlitePool};
use std::path::PathBuf;
use std::str::FromStr;
use std::sync::Arc;
use std::time::{Duration, Instant};
use tokio::fs;
use uuid::Uuid;

#[derive(Debug, Clone)]
pub struct PoolStatus {
    pub size: u32,
    pub idle_connections: u32,
}

#[derive(Debug, Clone)]
pub struct DatabaseMetrics {
    pub pool_status: PoolStatus,
    pub db_path: PathBuf,
    pub retry_config: RetryConfig,
    pub last_health_check: DateTime<Utc>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct StoredIdea {
    pub id: String,
    pub content: String,
    pub raw_score: Option<f64>,
    pub final_score: Option<f64>,
    pub patterns: Option<Vec<String>>,
    pub recommendation: Option<String>,
    pub analysis_details: Option<String>,
    pub created_at: DateTime<Utc>,
    pub reviewed_at: Option<DateTime<Utc>>,
    pub status: IdeaStatus,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct IdeaRelationship {
    pub id: String,
    pub source_idea_id: String,
    pub target_idea_id: String,
    pub relationship_type: RelationshipType,
    pub created_at: DateTime<Utc>,
}

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub enum RelationshipType {
    DependsOn,
    RelatedTo,
    PartOf,
    Parent,
    Child,
    Duplicate,
    Blocks,
    BlockedBy,
    SimilarTo,
}

impl std::fmt::Display for RelationshipType {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            RelationshipType::DependsOn => write!(f, "depends_on"),
            RelationshipType::RelatedTo => write!(f, "related_to"),
            RelationshipType::PartOf => write!(f, "part_of"),
            RelationshipType::Parent => write!(f, "parent"),
            RelationshipType::Child => write!(f, "child"),
            RelationshipType::Duplicate => write!(f, "duplicate"),
            RelationshipType::Blocks => write!(f, "blocks"),
            RelationshipType::BlockedBy => write!(f, "blocked_by"),
            RelationshipType::SimilarTo => write!(f, "similar_to"),
        }
    }
}

impl std::str::FromStr for RelationshipType {
    type Err = String;

    fn from_str(s: &str) -> Result<Self, Self::Err> {
        match s {
            "depends_on" => Ok(RelationshipType::DependsOn),
            "related_to" => Ok(RelationshipType::RelatedTo),
            "part_of" => Ok(RelationshipType::PartOf),
            "parent" => Ok(RelationshipType::Parent),
            "child" => Ok(RelationshipType::Child),
            "duplicate" => Ok(RelationshipType::Duplicate),
            "blocks" => Ok(RelationshipType::Blocks), // Note: Fixed typo in the original enum
            "blocked_by" => Ok(RelationshipType::BlockedBy), // Note: Fixed typo in the original enum
            "similar_to" => Ok(RelationshipType::SimilarTo),
            _ => Err(format!("Invalid relationship type: {}", s)),
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize, sqlx::Type)]
#[sqlx(type_name = "text")]
pub enum IdeaStatus {
    Active,
    Archived,
    Deleted,
}

impl std::fmt::Display for IdeaStatus {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            IdeaStatus::Active => write!(f, "active"),
            IdeaStatus::Archived => write!(f, "archived"),
            IdeaStatus::Deleted => write!(f, "deleted"),
        }
    }
}

#[derive(Clone, Debug)]
pub struct RetryConfig {
    pub max_attempts: u32,
    pub base_delay_ms: u64,
    pub max_delay_ms: u64,
    pub backoff_multiplier: f64,
    pub jitter: bool,
    pub retryable_errors: Vec<String>,
}

impl Default for RetryConfig {
    fn default() -> Self {
        Self {
            max_attempts: 3,
            base_delay_ms: 100,
            max_delay_ms: 5000,
            backoff_multiplier: 2.0,
            jitter: true,
            retryable_errors: vec![
                "database is locked".to_string(),
                "database table is locked".to_string(),
                "connection refused".to_string(),
                "connection timeout".to_string(),
                "connection reset".to_string(),
                "temporary failure".to_string(),
                "busy".to_string(),
                "timeout".to_string(),
            ],
        }
    }
}

impl RetryConfig {
    /// Check if an error is retryable based on the configuration
    pub fn is_retryable_error(&self, error: &str) -> bool {
        let error_lower = error.to_lowercase();
        self.retryable_errors
            .iter()
            .any(|retryable| error_lower.contains(&retryable.to_lowercase()))
    }

    /// Calculate delay for the next retry attempt
    pub fn calculate_delay(&self, attempt: u32) -> Duration {
        let base_delay = self.base_delay_ms as f64;
        let delay = base_delay * self.backoff_multiplier.powi(attempt as i32 - 1);
        let delay = delay.min(self.max_delay_ms as f64);

        let mut rng = rand::thread_rng();
        let final_delay = if self.jitter {
            let jitter_factor = 0.8 + (rng.gen::<f64>() * 0.4); // 80% to 120%
            delay * jitter_factor
        } else {
            delay
        };

        Duration::from_millis(final_delay as u64)
    }
}

pub struct Database {
    pool: SqlitePool,
    db_path: Arc<PathBuf>,
    retry_config: RetryConfig,
}

impl Database {
    pub async fn new() -> DatabaseResult<Self> {
        let db_path = PathBuf::from("./data/ideas.db");
        Self::with_config(db_path, RetryConfig::default()).await
    }

    pub async fn with_config(db_path: PathBuf, retry_config: RetryConfig) -> DatabaseResult<Self> {
        // Ensure directory exists
        if let Some(parent) = db_path.parent() {
            fs::create_dir_all(parent).await.map_err(|e| {
                DatabaseError::invalid_path_with_source(parent.to_str().unwrap_or(""), e)
            })?;
        }

        // Enhanced connection pool configuration
        let db_url = format!("sqlite://{}?mode=rwc", db_path.display());

        let pool = SqlitePoolOptions::new()
            .max_connections(2) // Optimized for single-threaded use
            .min_connections(1) // Pre-warm connections
            .idle_timeout(Duration::from_secs(5 * 60)) // 5 minutes
            .acquire_timeout(Duration::from_secs(5)) // Acquire timeout
            .connect(&db_url)
            .await
            .map_err(|e| DatabaseError::connection(e, Some(db_path.to_str().unwrap_or(""))))?;

        // Create table if it doesn't exist
        let create_table_query = r#"
            CREATE TABLE IF NOT EXISTS ideas (
                id TEXT PRIMARY KEY,
                content TEXT NOT NULL,
                raw_score REAL,
                final_score REAL,
                patterns TEXT,
                recommendation TEXT,
                analysis_details TEXT,
                created_at TEXT NOT NULL,
                reviewed_at TEXT,
                status TEXT NOT NULL DEFAULT 'active'
            )
            "#;

        // Create relationships table for idea linking
        let create_relationships_table = r#"
            CREATE TABLE IF NOT EXISTS idea_relationships (
                id TEXT PRIMARY KEY,
                source_idea_id TEXT NOT NULL,
                target_idea_id TEXT NOT NULL,
                relationship_type TEXT NOT NULL,
                created_at TEXT NOT NULL,
                FOREIGN KEY (source_idea_id) REFERENCES ideas (id),
                FOREIGN KEY (target_idea_id) REFERENCES ideas (id)
            )
            "#;

        query(create_table_query)
            .execute(&pool)
            .await
            .map_err(|e| DatabaseError::query_failed("CREATE TABLE ideas", e))?;

        // Create relationships table
        query(create_relationships_table)
            .execute(&pool)
            .await
            .map_err(|e| DatabaseError::query_failed("CREATE TABLE idea_relationships", e))?;

        let create_relationships_indexes = vec![
            "CREATE INDEX IF NOT EXISTS idx_relationships_source ON idea_relationships(source_idea_id)",
            "CREATE INDEX IF NOT EXISTS idx_relationships_target ON idea_relationships(target_idea_id)",
            "CREATE INDEX IF NOT EXISTS idx_relationships_type ON idea_relationships(relationship_type)",
        ];

        // Create indexes for relationships
        for index_query in create_relationships_indexes {
            sqlx::query(index_query)
                .execute(&pool)
                .await
                .map_err(|e| DatabaseError::query_failed(index_query, e))?;
        }

        // Create indexes for better query performance
        let create_indexes = vec![
            "CREATE INDEX IF NOT EXISTS idx_ideas_created_at ON ideas(created_at)",
            "CREATE INDEX IF NOT EXISTS idx_ideas_final_score ON ideas(final_score)",
            "CREATE INDEX IF NOT EXISTS idx_ideas_status ON ideas(status)",
            "CREATE INDEX IF NOT EXISTS idx_ideas_status_score ON ideas(status, final_score)",
        ];

        // Create indexes for relationships table
        let _create_relationships_indexes = ["CREATE INDEX IF NOT EXISTS idx_relationships_source ON idea_relationships(source_idea_id)",
            "CREATE INDEX IF NOT EXISTS idx_relationships_target ON idea_relationships(target_idea_id)",
            "CREATE INDEX IF NOT EXISTS idx_relationships_type ON idea_relationships(relationship_type)"];

        for index_query in create_indexes {
            query(index_query)
                .execute(&pool)
                .await
                .map_err(|e| DatabaseError::query_failed(index_query, e))?;
        }

        Ok(Database {
            pool,
            db_path: Arc::new(db_path),
            retry_config,
        })
    }

    // Execute operation with enhanced retry logic and structured logging
    async fn execute_with_retry<F, Fut, T>(
        &self,
        operation: F,
        operation_name: &str,
    ) -> DatabaseResult<T>
    where
        F: Fn() -> Fut,
        Fut: std::future::Future<Output = DatabaseResult<T>>,
    {
        let start_time = Instant::now();
        let mut attempt = 0;
        let mut _last_error: Option<DatabaseError> = None;

        loop {
            attempt += 1;

            tracing::debug!(
                operation = operation_name,
                attempt = attempt,
                max_attempts = self.retry_config.max_attempts,
                "Executing database operation"
            );

            let operation_start = Instant::now();
            let result = operation().await;
            let operation_duration = operation_start.elapsed();

            match result {
                Ok(value) => {
                    let total_duration = start_time.elapsed();

                    if attempt > 1 {
                        tracing::info!(
                            operation = operation_name,
                            attempt = attempt,
                            operation_duration_ms = operation_duration.as_millis(),
                            total_duration_ms = total_duration.as_millis(),
                            "Database operation succeeded after retries"
                        );
                    } else {
                        tracing::trace!(
                            operation = operation_name,
                            operation_duration_ms = operation_duration.as_millis(),
                            "Database operation completed successfully"
                        );
                    }

                    // Log the successful database operation
                    logging::log_database_operation(
                        operation_name,
                        Some("ideas"),
                        operation_duration,
                        true,
                        None,
                    );

                    return Ok(value);
                }
                Err(e) => {
                    let error_string = e.to_string();
                    _last_error = Some(DatabaseError::query_failed(
                        "execute_with_timeout_and_retry",
                        sqlx::Error::Protocol(format!("Operation failed: {}", error_string)),
                    ));

                    // Check if error is retryable
                    if attempt >= self.retry_config.max_attempts
                        || !self.retry_config.is_retryable_error(&error_string)
                    {
                        let total_duration = start_time.elapsed();

                        tracing::error!(
                            operation = operation_name,
                            attempt = attempt,
                            error = %error_string,
                            operation_duration_ms = operation_duration.as_millis(),
                            total_duration_ms = total_duration.as_millis(),
                            is_retryable = self.retry_config.is_retryable_error(&error_string),
                            max_attempts = self.retry_config.max_attempts,
                            "Database operation failed permanently"
                        );

                        // Log the failed database operation
                        logging::log_database_operation(
                            operation_name,
                            Some("ideas"),
                            operation_duration,
                            false,
                            Some(&error_string),
                        );

                        return Err(DatabaseError::retry_exhausted(
                            operation_name,
                            attempt,
                            total_duration.as_millis() as u64,
                            // Convert to SQLx Error if possible, otherwise create a generic one
                            match e {
                                DatabaseError::Connection { source, .. } => source,
                                DatabaseError::QueryFailed { source, .. } => source,
                                _ => SqlxError::Protocol(error_string),
                            },
                        ));
                    }

                    // Calculate delay and wait
                    let delay = self.retry_config.calculate_delay(attempt);

                    tracing::warn!(
                        operation = operation_name,
                        attempt = attempt,
                        error = %error_string,
                        delay_ms = delay.as_millis(),
                        "Database operation failed, retrying after delay"
                    );

                    tokio::time::sleep(delay).await;
                }
            }
        }
    }

    // Execute operation with timeout and retry logic
    async fn execute_with_timeout_and_retry<F, Fut, T>(
        &self,
        operation: F,
        operation_name: &str,
        timeout: Duration,
    ) -> DatabaseResult<T>
    where
        F: Fn() -> Fut,
        Fut: std::future::Future<Output = DatabaseResult<T>>,
    {
        let retry_operation = || async {
            tokio::time::timeout(timeout, operation())
                .await
                .map_err(|_| {
                    DatabaseError::operation_timeout(
                        operation_name,
                        timeout.as_millis() as u64,
                        SqlxError::Protocol("Operation timed out".to_string()),
                    )
                })?
        };

        self.execute_with_retry(retry_operation, operation_name)
            .await
    }

    // Health check for database connectivity with timeout
    pub async fn health_check(&self) -> DatabaseResult<()> {
        let health_timeout = tokio::time::Duration::from_secs(5);

        tokio::time::timeout(
            health_timeout,
            query("SELECT 1 as health_check").fetch_one(&self.pool),
        )
        .await
        .map_err(|_| {
            DatabaseError::connection_timeout(health_timeout.as_millis() as u64, "health check")
        })?
        .map(|_| ())
        .map_err(|e| DatabaseError::query_failed("health check", e))?;
        Ok(())
    }

    // Get database connection pool status
    pub async fn get_pool_status(&self) -> PoolStatus {
        let size = self.pool.size();
        let idle_connections = self.pool.num_idle() as u32;

        tracing::debug!(
            pool_size = size,
            idle_connections = idle_connections,
            active_connections = size - idle_connections,
            "Database pool status queried"
        );

        PoolStatus {
            size,
            idle_connections,
        }
    }

    /// Get database metrics for monitoring
    pub async fn get_metrics(&self) -> DatabaseMetrics {
        let pool_status = self.get_pool_status().await;

        DatabaseMetrics {
            pool_status,
            db_path: Arc::clone(&self.db_path).as_ref().clone(),
            retry_config: self.retry_config.clone(),
            last_health_check: Utc::now(), // This could be tracked more precisely
        }
    }

    // Close database connection pool gracefully
    pub async fn close(&self) -> DatabaseResult<()> {
        self.pool.close().await;
        Ok(())
    }

    pub async fn save_idea(
        &self,
        content: &str,
        raw_score: Option<f64>,
        final_score: Option<f64>,
        patterns: Option<Vec<String>>,
        recommendation: Option<String>,
        analysis_details: Option<String>,
    ) -> DatabaseResult<String> {
        // Record operation start time for metrics
        let start_time = Instant::now();
        let id = Uuid::new_v4().to_string();
        let now = Utc::now();

        let insert_query = r#"
            INSERT INTO ideas (
                id, content, raw_score, final_score, patterns,
                recommendation, analysis_details, created_at, status
            ) VALUES (?1, ?2, ?3, ?4, ?5, ?6, ?7, ?8, ?9)
            "#;

        let serialized_patterns = patterns
            .as_ref()
            .map(serde_json::to_string)
            .transpose()
            .map_err(|e| DatabaseError::serialization_field(e, "patterns"))?;

        self.execute_with_timeout_and_retry(
            || async {
                query(insert_query)
                    .bind(&id)
                    .bind(content)
                    .bind(raw_score)
                    .bind(final_score)
                    .bind(&serialized_patterns)
                    .bind(&recommendation)
                    .bind(&analysis_details)
                    .bind(now.to_rfc3339())
                    .bind(IdeaStatus::Active.to_string())
                    .execute(&self.pool)
                    .await
                    .map(|_| ())
                    .map_err(|e| DatabaseError::query_failed("INSERT INTO ideas", e))
            },
            "save_idea",
            Duration::from_secs(30),
        )
        .await?;

        tracing::info!(
            idea_id = %id,
            content_length = content.len(),
            "Successfully saved idea to database"
        );

        // Record performance metrics
        tokio::spawn(async move {
            metrics::record_operation_timing("database_save_idea", start_time).await;
            metrics::increment_counter("database_save_operations").await;
        });

        Ok(id)
    }

    pub async fn get_last_idea(&self) -> DatabaseResult<Option<StoredIdea>> {
        let select_query = r#"
            SELECT id, content, raw_score, final_score, patterns,
                   recommendation, analysis_details, created_at, reviewed_at, status
            FROM ideas
            WHERE status = 'active'
            ORDER BY created_at DESC
            LIMIT 1
            "#;

        let row = self
            .execute_with_timeout_and_retry(
                || async {
                    query(select_query)
                        .fetch_optional(&self.pool)
                        .await
                        .map_err(|e| DatabaseError::query_failed("SELECT last idea", e))
                },
                "get_last_idea",
                Duration::from_secs(10),
            )
            .await?;

        if let Some(row) = row {
            let created_at_str: String = row.get("created_at");
            let created_at = DateTime::parse_from_rfc3339(&created_at_str)
                .map_err(|e| DatabaseError::InvalidPath {
                    path: format!("Failed to parse datetime field 'created_at': {}", e),
                    source: None,
                })?
                .with_timezone(&Utc);

            let patterns = row
                .get::<Option<String>, _>("patterns")
                .map(|p| serde_json::from_str::<Vec<String>>(&p))
                .transpose()
                .map_err(|e| {
                    DatabaseError::deserialization_with_data(
                        e,
                        "patterns",
                        row.get::<Option<String>, _>("patterns").unwrap_or_default(),
                    )
                })?;

            let status_str: String = row.get("status");
            let status = match status_str.as_str() {
                "active" => IdeaStatus::Active,
                "archived" => IdeaStatus::Archived,
                "deleted" => IdeaStatus::Deleted,
                _ => IdeaStatus::Active,
            };

            Ok(Some(StoredIdea {
                id: row.get("id"),
                content: row.get("content"),
                raw_score: row.get("raw_score"),
                final_score: row.get("final_score"),
                patterns,
                recommendation: row.get("recommendation"),
                analysis_details: row.get("analysis_details"),
                created_at,
                reviewed_at: row.get::<Option<String>, _>("reviewed_at").and_then(|r| {
                    DateTime::parse_from_rfc3339(&r)
                        .ok()
                        .map(|dt| dt.with_timezone(&Utc))
                }),
                status,
            }))
        } else {
            Ok(None)
        }
    }

    pub async fn get_ideas_with_filters(
        &self,
        limit: usize,
        min_score: f64,
    ) -> DatabaseResult<Vec<StoredIdea>> {
        let filter_query = r#"
            SELECT id, content, raw_score, final_score, patterns,
                   recommendation, analysis_details, created_at, reviewed_at, status
            FROM ideas
            WHERE status = 'active' AND (final_score >= ?1 OR final_score IS NULL)
            ORDER BY created_at DESC
            LIMIT ?2
            "#;

        let rows = query(filter_query)
            .bind(min_score)
            .bind(limit as i64)
            .fetch_all(&self.pool)
            .await
            .map_err(|e| DatabaseError::query_failed("SELECT ideas with filters", e))?;

        let ideas: Result<Vec<_>, DatabaseError> = rows
            .into_iter()
            .map(|row| {
                let created_at_str: String = row.get("created_at");
                let created_at = DateTime::parse_from_rfc3339(&created_at_str)
                    .map_err(|e| DatabaseError::InvalidPath {
                        path: format!("Failed to parse datetime field 'created_at': {}", e),
                        source: None,
                    })?
                    .with_timezone(&Utc);

                let patterns = row
                    .get::<Option<String>, _>("patterns")
                    .and_then(|p| serde_json::from_str::<Vec<String>>(&p).ok());

                let status_str: String = row.get("status");
                let status = match status_str.as_str() {
                    "active" => IdeaStatus::Active,
                    "archived" => IdeaStatus::Archived,
                    "deleted" => IdeaStatus::Deleted,
                    _ => IdeaStatus::Active,
                };

                Ok(StoredIdea {
                    id: row.get("id"),
                    content: row.get("content"),
                    raw_score: row.get("raw_score"),
                    final_score: row.get("final_score"),
                    patterns,
                    recommendation: row.get("recommendation"),
                    analysis_details: row.get("analysis_details"),
                    created_at,
                    reviewed_at: row
                        .get::<Option<String>, _>("reviewed_at")
                        .and_then(|r| DateTime::parse_from_rfc3339(&r).ok())
                        .map(|dt| dt.with_timezone(&Utc)),
                    status,
                })
            })
            .collect();

        ideas
    }

    pub async fn get_pruning_candidates(&self) -> DatabaseResult<Vec<StoredIdea>> {
        let seven_days_ago = Utc::now() - chrono::Duration::days(7);
        let fourteen_days_ago = Utc::now() - chrono::Duration::days(14);

        let pruning_query = r#"
            SELECT id, content, raw_score, final_score, patterns,
                   recommendation, analysis_details, created_at, reviewed_at, status
            FROM ideas
            WHERE status = 'active'
            AND (
                (final_score < 3.0 AND created_at < ?1) OR
                (final_score < 6.0 AND created_at < ?2)
            )
            ORDER BY created_at ASC
            "#;

        let rows = query(pruning_query)
            .bind(seven_days_ago.to_rfc3339())
            .bind(fourteen_days_ago.to_rfc3339())
            .fetch_all(&self.pool)
            .await
            .map_err(|e| DatabaseError::query_failed("SELECT pruning candidates", e))?;

        let ideas: Result<Vec<_>, DatabaseError> = rows
            .into_iter()
            .map(|row| {
                let created_at_str: String = row.get("created_at");
                let created_at = DateTime::parse_from_rfc3339(&created_at_str)
                    .map_err(|e| DatabaseError::InvalidPath {
                        path: format!("Failed to parse datetime field 'created_at': {}", e),
                        source: None,
                    })?
                    .with_timezone(&Utc);

                let patterns = row
                    .get::<Option<String>, _>("patterns")
                    .and_then(|p| serde_json::from_str::<Vec<String>>(&p).ok());

                let status_str: String = row.get("status");
                let status = match status_str.as_str() {
                    "active" => IdeaStatus::Active,
                    "archived" => IdeaStatus::Archived,
                    "deleted" => IdeaStatus::Deleted,
                    _ => IdeaStatus::Active,
                };

                Ok(StoredIdea {
                    id: row.get("id"),
                    content: row.get("content"),
                    raw_score: row.get("raw_score"),
                    final_score: row.get("final_score"),
                    patterns,
                    recommendation: row.get("recommendation"),
                    analysis_details: row.get("analysis_details"),
                    created_at,
                    reviewed_at: row
                        .get::<Option<String>, _>("reviewed_at")
                        .and_then(|r| DateTime::parse_from_rfc3339(&r).ok())
                        .map(|dt| dt.with_timezone(&Utc)),
                    status,
                })
            })
            .collect();

        ideas
    }

    pub async fn archive_idea(&self, id: &str) -> DatabaseResult<()> {
        query("UPDATE ideas SET status = 'archived' WHERE id = ?1")
            .bind(id)
            .execute(&self.pool)
            .await
            .map_err(|e| DatabaseError::query_failed("UPDATE idea to archived", e))?;

        Ok(())
    }

    pub async fn delete_idea(&self, id: &str) -> DatabaseResult<()> {
        query("UPDATE ideas SET status = 'deleted' WHERE id = ?1")
            .bind(id)
            .execute(&self.pool)
            .await
            .map_err(|e| DatabaseError::query_failed("UPDATE idea to deleted", e))?;

        Ok(())
    }

    pub async fn get_idea_count(&self) -> DatabaseResult<i64> {
        let count: i64 =
            query_scalar("SELECT COUNT(*) as count FROM ideas WHERE status = 'active'")
                .fetch_one(&self.pool)
                .await
                .map_err(|e| DatabaseError::query_failed("SELECT COUNT(*)", e))?;

        Ok(count)
    }

    // Relationship management functions
    pub async fn create_relationship(
        &self,
        source_idea_id: &str,
        target_idea_id: &str,
        relationship_type: RelationshipType,
    ) -> Result<String, DatabaseError> {
        let id = uuid::Uuid::new_v4().to_string();
        let now = Utc::now();

        // First check if relationship already exists to avoid duplicates
        let check_query = r#"
            SELECT COUNT(*) as count FROM idea_relationships
            WHERE source_idea_id = ?1 AND target_idea_id = ?2 AND relationship_type = ?3
        "#;

        let count_row = sqlx::query(check_query)
            .bind(source_idea_id)
            .bind(target_idea_id)
            .bind(relationship_type.to_string())
            .fetch_one(&self.pool)
            .await;

        if let Ok(row) = count_row {
            if row.get::<i64, _>("count") > 0 {
                return Err(DatabaseError::duplicate_entry(
                    "IdeaRelationship",
                    format!(
                        "Relationship of type '{}' between {} and {}",
                        relationship_type, source_idea_id, target_idea_id
                    ),
                    format!("{} -> {}", source_idea_id, target_idea_id),
                ));
            }
        }

        let insert_query = r#"
            INSERT INTO idea_relationships (
                id, source_idea_id, target_idea_id, relationship_type, created_at
            ) VALUES (?1, ?2, ?3, ?4, ?5)
        "#;

        query(insert_query)
            .bind(&id)
            .bind(source_idea_id)
            .bind(target_idea_id)
            .bind(relationship_type.to_string())
            .bind(now.to_rfc3339())
            .execute(&self.pool)
            .await
            .map(|_| ())
            .map_err(|e| DatabaseError::query_failed("INSERT INTO idea_relationships", e))?;

        tracing::info!(
            relationship_id = %id,
            source_idea_id = source_idea_id,
            target_idea_id = target_idea_id,
            relationship_type = %relationship_type,
            "Created idea relationship"
        );

        Ok(id)
    }

    pub async fn get_relationships_for_idea(
        &self,
        idea_id: &str,
    ) -> Result<Vec<IdeaRelationship>, DatabaseError> {
        let select_query = r#"
            SELECT ir.id, ir.source_idea_id, ir.target_idea_id, ir.relationship_type, ir.created_at
            FROM idea_relationships ir
            WHERE ir.source_idea_id = ?1 OR ir.target_idea_id = ?1
            ORDER BY ir.created_at DESC
        "#;

        let rows = query(select_query)
            .bind(idea_id)
            .fetch_all(&self.pool)
            .await
            .map_err(|e| DatabaseError::query_failed("SELECT relationships for idea", e))?;

        let relationships: Result<Vec<_>, DatabaseError> = rows
            .into_iter()
            .map(|row| {
                let created_at_str: String = row.get("created_at");
                let created_at = DateTime::parse_from_rfc3339(&created_at_str)
                    .map_err(|e| DatabaseError::InvalidPath {
                        path: format!("Failed to parse datetime field 'created_at': {}", e),
                        source: None,
                    })?
                    .with_timezone(&Utc);

                let relationship_type_str: String = row.get("relationship_type");
                let relationship_type = RelationshipType::from_str(&relationship_type_str)
                    .map_err(|e| DatabaseError::ConstraintViolation {
                        constraint: format!("Invalid relationship type: {}", e),
                    })?;

                Ok(IdeaRelationship {
                    id: row.get("id"),
                    source_idea_id: row.get("source_idea_id"),
                    target_idea_id: row.get("target_idea_id"),
                    relationship_type,
                    created_at,
                })
            })
            .collect();

        relationships
    }

    pub async fn get_related_ideas(
        &self,
        idea_id: &str,
        relationship_type: Option<RelationshipType>,
    ) -> Result<Vec<(StoredIdea, RelationshipType)>, DatabaseError> {
        let base_query = r#"
            SELECT
                ir.relationship_type,
                i.id, i.content, i.raw_score, i.final_score, i.patterns,
                i.recommendation, i.analysis_details, i.created_at, i.reviewed_at, i.status
            FROM idea_relationships ir
            JOIN ideas i ON (i.id = CASE
                WHEN ir.source_idea_id = ?1 THEN ir.target_idea_id
                ELSE ir.source_idea_id
            END)
            WHERE (ir.source_idea_id = ?1 OR ir.target_idea_id = ?1)
              AND i.status = 'active'
        "#;

        let final_query = if relationship_type.is_some() {
            format!("{} AND ir.relationship_type = ?2", base_query)
        } else {
            base_query.to_string()
        };

        let rows = if let Some(rel_type) = relationship_type {
            sqlx::query(&final_query)
                .bind(idea_id)
                .bind(rel_type.to_string())
                .fetch_all(&self.pool)
                .await
        } else {
            sqlx::query(&final_query)
                .bind(idea_id)
                .fetch_all(&self.pool)
                .await
        }
        .map_err(|e| DatabaseError::query_failed("SELECT related ideas", e))?;

        let mut related_ideas = Vec::new();

        for row in rows {
            let relationship_type_str: String = row
                .try_get("relationship_type")
                .map_err(|e| DatabaseError::query_failed("GET relationship_type", e))?;
            let relationship_type =
                RelationshipType::from_str(&relationship_type_str).map_err(|e| {
                    DatabaseError::ConstraintViolation {
                        constraint: format!("Invalid relationship type: {}", e),
                    }
                })?;

            let idea_id: String = row
                .try_get("id")
                .map_err(|e| DatabaseError::query_failed("GET id", e))?;
            let content: String = row
                .try_get("content")
                .map_err(|e| DatabaseError::query_failed("GET content", e))?;
            let raw_score: Option<f64> = row
                .try_get("raw_score")
                .map_err(|e| DatabaseError::query_failed("GET raw_score", e))?;
            let final_score: Option<f64> = row
                .try_get("final_score")
                .map_err(|e| DatabaseError::query_failed("GET final_score", e))?;
            let patterns_str: Option<String> = row
                .try_get("patterns")
                .map_err(|e| DatabaseError::query_failed("GET patterns", e))?;
            let recommendation: Option<String> = row
                .try_get("recommendation")
                .map_err(|e| DatabaseError::query_failed("GET recommendation", e))?;
            let analysis_details: Option<String> = row
                .try_get("analysis_details")
                .map_err(|e| DatabaseError::query_failed("GET analysis_details", e))?;
            let created_at_str: String = row
                .try_get("created_at")
                .map_err(|e| DatabaseError::query_failed("GET created_at", e))?;
            let reviewed_at_str: Option<String> = row
                .try_get("reviewed_at")
                .map_err(|e| DatabaseError::query_failed("GET reviewed_at", e))?;
            let status_str: String = row
                .try_get("status")
                .map_err(|e| DatabaseError::query_failed("GET status", e))?;

            let created_at = DateTime::parse_from_rfc3339(&created_at_str)
                .map_err(|e| DatabaseError::InvalidPath {
                    path: format!("Failed to parse datetime field 'created_at': {}", e),
                    source: None,
                })?
                .with_timezone(&Utc);

            let patterns = patterns_str.and_then(|p| serde_json::from_str::<Vec<String>>(&p).ok());

            let status = match status_str.as_str() {
                "active" => IdeaStatus::Active,
                "archived" => IdeaStatus::Archived,
                "deleted" => IdeaStatus::Deleted,
                _ => IdeaStatus::Active,
            };

            let stored_idea = StoredIdea {
                id: idea_id,
                content,
                raw_score,
                final_score,
                patterns,
                recommendation,
                analysis_details,
                created_at,
                reviewed_at: reviewed_at_str
                    .and_then(|r| DateTime::parse_from_rfc3339(&r).ok())
                    .map(|dt| dt.with_timezone(&Utc)),
                status,
            };

            related_ideas.push((stored_idea, relationship_type));
        }

        Ok(related_ideas)
    }

    pub async fn delete_relationship(&self, relationship_id: &str) -> Result<(), DatabaseError> {
        let delete_query = r#"
            DELETE FROM idea_relationships
            WHERE id = ?1
        "#;

        query(delete_query)
            .bind(relationship_id)
            .execute(&self.pool)
            .await
            .map(|_| ())
            .map_err(|e| DatabaseError::query_failed("DELETE FROM idea_relationships", e))?;

        tracing::info!(
            relationship_id = relationship_id,
            "Deleted idea relationship"
        );

        Ok(())
    }

    pub async fn delete_all_relationships_for_idea(
        &self,
        idea_id: &str,
    ) -> Result<(), DatabaseError> {
        let delete_query = r#"
            DELETE FROM idea_relationships
            WHERE source_idea_id = ?1 OR target_idea_id = ?1
        "#;

        query(delete_query)
            .bind(idea_id)
            .execute(&self.pool)
            .await
            .map(|_| ())
            .map_err(|e| DatabaseError::query_failed("DELETE ALL relationships for idea", e))?;

        tracing::info!(idea_id = idea_id, "Deleted all relationships for idea");

        Ok(())
    }

    /// Get idea by ID
    pub async fn get_by_id(&self, idea_id: &str) -> Result<Option<StoredIdea>, DatabaseError> {
        let select_query = r#"
            SELECT
                id, content, raw_score, final_score, patterns, recommendation,
                analysis_details, created_at, reviewed_at, status
            FROM ideas
            WHERE id = ?1
        "#;

        let row = query(select_query)
            .bind(idea_id)
            .fetch_optional(&self.pool)
            .await
            .map_err(|e| DatabaseError::query_failed("SELECT idea by ID", e))?;

        match row {
            Some(row) => {
                let created_at_str: String = row.get("created_at");
                let created_at = DateTime::parse_from_rfc3339(&created_at_str)
                    .map_err(|e| DatabaseError::InvalidPath {
                        path: format!("Error parsing datetime field 'created_at': {}", e),
                        source: None,
                    })?
                    .with_timezone(&Utc);

                let patterns_str: Option<String> = row.get("patterns");
                let patterns =
                    patterns_str.and_then(|p| serde_json::from_str::<Vec<String>>(&p).ok());

                let status_str: String = row.get("status");
                let status = match status_str.as_str() {
                    "active" => IdeaStatus::Active,
                    "archived" => IdeaStatus::Archived,
                    "deleted" => IdeaStatus::Deleted,
                    _ => IdeaStatus::Active,
                };

                let reviewed_at_str: Option<String> = row.get("reviewed_at");
                let reviewed_at = if let Some(reviewed_at_str) = reviewed_at_str {
                    Some(
                        DateTime::parse_from_rfc3339(&reviewed_at_str)
                            .map_err(|e| DatabaseError::InvalidPath {
                                path: format!("Error parsing datetime field 'reviewed_at': {}", e),
                                source: None,
                            })?
                            .with_timezone(&Utc),
                    )
                } else {
                    None
                };

                Ok(Some(StoredIdea {
                    id: row.get("id"),
                    content: row.get("content"),
                    raw_score: row.get("raw_score"),
                    final_score: row.get("final_score"),
                    patterns,
                    recommendation: row.get("recommendation"),
                    analysis_details: row.get("analysis_details"),
                    created_at,
                    reviewed_at,
                    status,
                }))
            }
            None => Ok(None),
        }
    }
}
