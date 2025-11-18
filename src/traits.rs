//! Trait abstractions for common operations to improve code reuse and testability

use crate::types::{PatternType, QueryLimit};
use async_trait::async_trait;
use std::collections::HashMap;

/// Generic repository pattern for data storage operations
#[async_trait]
pub trait Repository<T, ID> {
    /// Error type for this repository
    type Error;

    /// Save an entity to the repository
    async fn save(&self, entity: T) -> std::result::Result<ID, Self::Error>;

    /// Find an entity by its ID
    async fn find_by_id(&self, id: ID) -> std::result::Result<Option<T>, Self::Error>;

    /// Find all entities with optional filtering
    async fn find_all(
        &self,
        filter: Option<&dyn Filter<T>>,
    ) -> std::result::Result<Vec<T>, Self::Error>;

    /// Update an existing entity
    async fn update(&self, id: ID, entity: T) -> std::result::Result<(), Self::Error>;

    /// Delete an entity by its ID
    async fn delete(&self, id: ID) -> std::result::Result<(), Self::Error>;

    /// Count entities with optional filtering
    async fn count(
        &self,
        filter: Option<&dyn Filter<T>>,
    ) -> std::result::Result<usize, Self::Error>;
}

/// Generic filtering trait for repository queries
pub trait Filter<T> {
    /// Check if an entity matches this filter
    fn matches(&self, entity: &T) -> bool;

    /// Get filter description for debugging
    fn description(&self) -> String;
}

/// Generic query builder pattern
pub trait QueryBuilder<T> {
    /// Filter entities by a predicate
    fn filter<F>(self, predicate: F) -> Self
    where
        F: Fn(&T) -> bool + 'static;

    /// Limit the number of results
    fn limit(self, limit: QueryLimit) -> Self;

    /// Order the results
    fn order_by<F>(self, compare: F) -> Self
    where
        F: Fn(&T, &T) -> std::cmp::Ordering + 'static;

    /// Execute the query and return results
    async fn execute(self)
        -> std::result::Result<Vec<T>, Box<dyn std::error::Error + Send + Sync>>;
}

/// Generic scoring trait for idea evaluation
#[async_trait]
pub trait Scorer<I: Send + Sync> {
    /// The score type produced by this scorer
    type Score: Send;

    /// Configuration for this scorer
    type Config;

    /// Create a new scorer with the given configuration
    fn new(config: Self::Config) -> Self;

    /// Score an idea asynchronously
    async fn score(
        &self,
        input: &I,
    ) -> std::result::Result<Self::Score, Box<dyn std::error::Error + Send + Sync>>;

    /// Score multiple ideas in batch
    async fn score_batch(
        &self,
        inputs: &[I],
    ) -> std::result::Result<Vec<Self::Score>, Box<dyn std::error::Error + Send + Sync>> {
        let mut results = Vec::with_capacity(inputs.len());
        for input in inputs {
            results.push(self.score(input).await?);
        }
        Ok(results)
    }

    /// Validate the input before scoring
    fn validate_input(
        &self,
        input: &I,
    ) -> std::result::Result<(), Box<dyn std::error::Error + Send + Sync>>;
}

/// Generic pattern detection trait
#[async_trait]
pub trait PatternDetector<I: Send + Sync> {
    /// Pattern type produced by this detector
    type Pattern: Send;

    /// Configuration for this detector
    type Config;

    /// Create a new detector with the given configuration
    fn new(config: Self::Config) -> Self;

    /// Detect patterns in the input asynchronously
    async fn detect_patterns(
        &self,
        input: &I,
    ) -> std::result::Result<Vec<Self::Pattern>, Box<dyn std::error::Error + Send + Sync>>;

    /// Detect a single pattern type
    async fn detect_pattern_type(
        &self,
        input: &I,
        pattern_type: &PatternType,
    ) -> std::result::Result<Option<Self::Pattern>, Box<dyn std::error::Error + Send + Sync>>;

    /// Get all supported pattern types
    fn supported_patterns(&self) -> Vec<PatternType>;

    /// Validate input before pattern detection
    fn validate_input(
        &self,
        input: &I,
    ) -> std::result::Result<(), Box<dyn std::error::Error + Send + Sync>>;
}

/// Generic AI enhancement trait for intelligent analysis
#[async_trait]
pub trait AiEnhancer<I: Send + Sync, E: Send> {
    /// Configuration for the AI enhancer
    type Config;

    /// Create a new enhancer with the given configuration
    fn new(config: Self::Config) -> Self;

    /// Enhance analysis with AI
    async fn enhance(
        &self,
        input: &I,
    ) -> std::result::Result<E, Box<dyn std::error::Error + Send + Sync>>;

    /// Check if AI enhancement is available
    fn is_available(&self) -> bool;

    /// Get model information
    fn model_info(&self) -> ModelInfo;
}

/// Information about an AI model
#[derive(Debug, Clone)]
pub struct ModelInfo {
    pub name: String,
    pub version: String,
    pub capabilities: Vec<String>,
    pub max_tokens: Option<usize>,
    pub temperature: Option<f32>,
}

/// Generic caching trait for performance optimization
#[async_trait]
pub trait Cache<K: Send + Sync, V: Send> {
    /// Error type for cache operations
    type Error: std::error::Error + Send + Sync + 'static;

    /// Get a value from the cache
    async fn get(&self, key: &K) -> std::result::Result<Option<V>, Self::Error>;

    /// Set a value in the cache with optional TTL
    async fn set(
        &self,
        key: K,
        value: V,
        ttl: Option<std::time::Duration>,
    ) -> std::result::Result<(), Self::Error>;

    /// Remove a value from the cache
    async fn remove(&self, key: &K) -> std::result::Result<bool, Self::Error>;

    /// Clear all values from the cache
    async fn clear(&self) -> std::result::Result<(), Self::Error>;

    /// Get cache statistics
    async fn stats(&self) -> CacheStats;
}

/// Cache statistics
#[derive(Debug, Clone, Default)]
pub struct CacheStats {
    pub hits: u64,
    pub misses: u64,
    pub size: usize,
    pub hit_rate: f64,
}

impl CacheStats {
    pub fn new(hits: u64, misses: u64, size: usize) -> Self {
        let total = hits + misses;
        let hit_rate = if total > 0 {
            hits as f64 / total as f64
        } else {
            0.0
        };
        Self {
            hits,
            misses,
            size,
            hit_rate,
        }
    }
}

/// Generic validation trait
pub trait Validator<T> {
    /// Error type for validation
    type Error: std::error::Error;

    /// Validate the input
    fn validate(&self, input: &T) -> std::result::Result<(), Self::Error>;

    /// Validate and return a reference to the input if valid
    fn validated<'a>(&self, input: &'a T) -> std::result::Result<&'a T, Self::Error> {
        self.validate(input)?;
        Ok(input)
    }

    /// Check if input is valid without returning errors
    fn is_valid(&self, input: &T) -> bool {
        self.validate(input).is_ok()
    }
}

/// Generic transformer trait for data processing
pub trait Transformer<I, O> {
    /// Error type for transformation
    type Error: std::error::Error;

    /// Transform input to output
    fn transform(&self, input: I) -> std::result::Result<O, Self::Error>;

    /// Transform multiple items
    fn transform_batch(&self, inputs: Vec<I>) -> std::result::Result<Vec<O>, Self::Error> {
        let mut outputs = Vec::with_capacity(inputs.len());
        for input in inputs {
            outputs.push(self.transform(input)?);
        }
        Ok(outputs)
    }
}

/// Async transformer trait for heavy operations
#[async_trait]
pub trait AsyncTransformer<I: Send + 'static, O: Send> {
    /// Error type for transformation
    type Error: std::error::Error + Send + Sync + 'static;

    /// Transform input to output asynchronously
    async fn transform(&self, input: I) -> std::result::Result<O, Self::Error>;

    /// Transform multiple items in parallel
    async fn transform_batch(&self, inputs: Vec<I>) -> std::result::Result<Vec<O>, Self::Error> {
        let futures: Vec<_> = inputs
            .into_iter()
            .map(|input| self.transform(input))
            .collect();

        futures::future::try_join_all(futures).await
    }
}

/// Generic service trait with health checking
#[async_trait]
pub trait Service {
    /// Service status
    type Status;

    /// Service configuration
    type Config;

    /// Create a new service with configuration
    fn new(config: Self::Config) -> Self;

    /// Check if the service is healthy
    async fn health_check(
        &self,
    ) -> std::result::Result<Self::Status, Box<dyn std::error::Error + Send + Sync>>;

    /// Start the service
    async fn start(&mut self) -> std::result::Result<(), Box<dyn std::error::Error + Send + Sync>>;

    /// Stop the service gracefully
    async fn stop(&mut self) -> std::result::Result<(), Box<dyn std::error::Error + Send + Sync>>;

    /// Get service name
    fn name(&self) -> &str;

    /// Get service version
    fn version(&self) -> &str;
}

/// Generic metrics trait for performance monitoring
#[async_trait]
pub trait Metrics {
    /// Error type for metrics operations
    type Error: std::error::Error + Send + Sync + 'static;

    /// Record a counter metric
    async fn increment_counter(
        &self,
        name: &str,
        tags: Option<&HashMap<String, String>>,
    ) -> std::result::Result<(), Self::Error>;

    /// Record a gauge metric
    async fn set_gauge(
        &self,
        name: &str,
        value: f64,
        tags: Option<&HashMap<String, String>>,
    ) -> std::result::Result<(), Self::Error>;

    /// Record a histogram/timer metric
    async fn record_timer(
        &self,
        name: &str,
        duration: std::time::Duration,
        tags: Option<&HashMap<String, String>>,
    ) -> std::result::Result<(), Self::Error>;

    /// Get all metric values
    async fn get_metrics(&self) -> std::result::Result<HashMap<String, f64>, Self::Error>;
}

/// Generic configuration trait
pub trait Configuration {
    /// Error type for configuration operations
    type Error: std::error::Error;

    /// Load configuration from a source
    fn load() -> std::result::Result<Self, Self::Error>
    where
        Self: Sized;

    /// Validate the configuration
    fn validate(&self) -> std::result::Result<(), Self::Error>;

    /// Get a configuration value by key
    fn get<T>(&self, key: &str) -> std::result::Result<T, Self::Error>
    where
        T: std::str::FromStr,
        T::Err: std::fmt::Display;

    /// Set a configuration value by key
    fn set<T>(&mut self, key: &str, value: T) -> std::result::Result<(), Self::Error>
    where
        T: std::fmt::Display;

    /// Check if a configuration key exists
    fn contains_key(&self, key: &str) -> bool;
}

/// Middleware trait for request/response processing
pub trait Middleware<Request, Response> {
    /// Error type for middleware operations
    type Error: std::error::Error;

    /// Process the request and optionally modify it
    fn process_request(&self, request: Request) -> std::result::Result<Request, Self::Error>;

    /// Process the response and optionally modify it
    fn process_response(&self, response: Response) -> std::result::Result<Response, Self::Error>;

    /// Get middleware name for debugging
    fn name(&self) -> &str;
}

/// Generic event trait for pub/sub patterns
#[async_trait]
pub trait Event {
    /// Event type identifier
    fn event_type(&self) -> &str;

    /// Event timestamp
    fn timestamp(&self) -> chrono::DateTime<chrono::Utc>;

    /// Serialize event to JSON
    fn to_json(&self) -> Result<String, serde_json::Error>;

    /// Get event ID if available
    fn id(&self) -> Option<&str> {
        None
    }
}

/// Generic event publisher trait
#[async_trait]
pub trait EventPublisher<E: Event + Send> {
    /// Error type for publishing operations
    type Error: std::error::Error + Send + Sync + 'static;

    /// Publish an event
    async fn publish(&self, event: E) -> std::result::Result<(), Self::Error>;

    /// Publish multiple events
    async fn publish_batch(&self, events: Vec<E>) -> std::result::Result<(), Self::Error>;

    /// Get publisher name
    fn name(&self) -> &str;
}

/// Generic event subscriber trait
#[async_trait]
pub trait EventSubscriber<E: Event + Send + Sync> {
    /// Error type for subscription operations
    type Error: std::error::Error + Send + Sync + 'static;

    /// Handle an incoming event
    async fn handle(&self, event: &E) -> std::result::Result<(), Self::Error>;

    /// Get subscriber name
    fn name(&self) -> &str;

    /// Get event types this subscriber handles
    fn handles_events(&self) -> Vec<&'static str>;
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::collections::HashMap;

    // Example implementation for testing
    #[derive(Debug, Clone)]
    struct TestEntity {
        id: String,
        name: String,
        value: i32,
    }

    struct TestFilter {
        min_value: i32,
    }

    impl Filter<TestEntity> for TestFilter {
        fn matches(&self, entity: &TestEntity) -> bool {
            entity.value >= self.min_value
        }

        fn description(&self) -> String {
            format!("value >= {}", self.min_value)
        }
    }

    #[test]
    fn test_filter_trait() {
        let entity = TestEntity {
            id: "1".to_string(),
            name: "test".to_string(),
            value: 10,
        };

        let filter = TestFilter { min_value: 5 };
        assert!(filter.matches(&entity));

        let filter = TestFilter { min_value: 15 };
        assert!(!filter.matches(&entity));
    }

    #[test]
    fn test_cache_stats() {
        let stats = CacheStats::new(80, 20, 100);
        assert_eq!(stats.hits, 80);
        assert_eq!(stats.misses, 20);
        assert_eq!(stats.size, 100);
        assert!((stats.hit_rate - 0.8).abs() < f64::EPSILON);
    }

    // Simple validator example
    struct TestValidator;

    #[derive(Debug)]
    struct TestValidationError(String);

    impl std::fmt::Display for TestValidationError {
        fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
            write!(f, "{}", self.0)
        }
    }

    impl std::error::Error for TestValidationError {}

    impl Validator<String> for TestValidator {
        type Error = TestValidationError;

        fn validate(&self, input: &String) -> std::result::Result<(), Self::Error> {
            if input.is_empty() {
                Err(TestValidationError("Input cannot be empty".to_string()))
            } else {
                Ok(())
            }
        }
    }

    #[test]
    fn test_validator_trait() {
        let validator = TestValidator;

        assert!(validator.is_valid(&"hello".to_string()));
        assert!(!validator.is_valid(&"".to_string()));

        assert!(validator.validated(&"hello".to_string()).is_ok());
        assert!(validator.validated(&"".to_string()).is_err());
    }

    // Simple transformer example
    struct UpperCaseTransformer;

    impl Transformer<String, String> for UpperCaseTransformer {
        type Error = std::convert::Infallible;

        fn transform(&self, input: String) -> std::result::Result<String, Self::Error> {
            Ok(input.to_uppercase())
        }
    }

    #[test]
    fn test_transformer_trait() {
        let transformer = UpperCaseTransformer;
        let result = transformer.transform("hello".to_string()).unwrap();
        assert_eq!(result, "HELLO");

        let batch = transformer
            .transform_batch(vec!["hello".to_string(), "world".to_string()])
            .unwrap();
        assert_eq!(batch, vec!["HELLO", "WORLD"]);
    }
}
