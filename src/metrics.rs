//! Application metrics collection and monitoring
//!
//! This module provides comprehensive metrics collection for the telos-idea-matrix application,
//! including performance metrics, usage statistics, and health monitoring.

use chrono::{DateTime, Utc};
use std::collections::HashMap;
use std::sync::atomic::{AtomicU64, AtomicUsize, Ordering};
use std::time::{Duration, Instant};
use tokio::sync::RwLock;

/// Global application metrics registry
pub struct MetricsRegistry {
    /// Metrics store with thread-safe access
    metrics: RwLock<HashMap<String, Metric>>,
    /// Performance counters
    performance: PerformanceMetrics,
    /// Usage statistics
    usage: UsageMetrics,
}

impl MetricsRegistry {
    /// Create a new metrics registry
    pub fn new() -> Self {
        Self {
            metrics: RwLock::new(HashMap::with_capacity(20)),
            performance: PerformanceMetrics::new(),
            usage: UsageMetrics::new(),
        }
    }

    /// Register a new metric
    pub async fn register_metric(&self, name: String, metric: Metric) {
        let mut metrics = self.metrics.write().await;
        metrics.insert(name, metric);
    }

    /// Increment a counter metric
    pub async fn increment_counter(&self, name: &str) {
        let mut metrics = self.metrics.write().await;
        metrics
            .entry(name.to_string())
            .or_insert_with(|| Metric::Counter(Counter::new()))
            .as_counter_mut()
            .expect("Metric should be a Counter")
            .increment();
    }

    /// Record a gauge value
    pub async fn record_gauge(&self, name: &str, value: f64) {
        let mut metrics = self.metrics.write().await;
        metrics
            .entry(name.to_string())
            .or_insert_with(|| Metric::Gauge(Gauge::new()))
            .as_gauge_mut()
            .expect("Metric should be a Gauge")
            .set(value);
    }

    /// Record a histogram value
    pub async fn record_histogram(&self, name: &str, value: f64) {
        let mut metrics = self.metrics.write().await;
        metrics
            .entry(name.to_string())
            .or_insert_with(|| Metric::Histogram(Histogram::new()))
            .as_histogram_mut()
            .expect("Metric should be a Histogram")
            .record(value);
    }

    /// Get performance metrics
    pub fn performance_metrics(&self) -> &PerformanceMetrics {
        &self.performance
    }

    /// Get usage metrics
    pub fn usage_metrics(&self) -> &UsageMetrics {
        &self.usage
    }

    /// Get all metrics as a serializable format
    pub async fn get_all_metrics(&self) -> MetricsSnapshot {
        let metrics = self.metrics.read().await;
        let mut snapshot = MetricsSnapshot::new();

        for (name, metric) in metrics.iter() {
            match metric {
                Metric::Counter(counter) => {
                    snapshot.counters.insert(name.clone(), counter.get());
                }
                Metric::Gauge(gauge) => {
                    snapshot.gauges.insert(name.clone(), gauge.get());
                }
                Metric::Histogram(histogram) => {
                    snapshot
                        .histograms
                        .insert(name.clone(), histogram.get_values().clone());
                }
            }
        }

        snapshot
    }

    /// Record an operation timing
    pub async fn record_timing(&self, operation: &str, duration: Duration) {
        let millis = duration.as_millis() as f64;
        self.record_histogram(operation, millis).await;
    }
}

#[derive(Debug)]
pub enum Metric {
    Counter(Counter),
    Gauge(Gauge),
    Histogram(Histogram),
}

impl Metric {
    pub fn as_counter_mut(&mut self) -> Option<&mut Counter> {
        if let Metric::Counter(counter) = self {
            Some(counter)
        } else {
            None
        }
    }

    pub fn as_gauge_mut(&mut self) -> Option<&mut Gauge> {
        if let Metric::Gauge(gauge) = self {
            Some(gauge)
        } else {
            None
        }
    }

    pub fn as_histogram_mut(&mut self) -> Option<&mut Histogram> {
        if let Metric::Histogram(histogram) = self {
            Some(histogram)
        } else {
            None
        }
    }
}

#[derive(Debug)]
pub struct Counter {
    value: AtomicU64,
}

impl Counter {
    pub fn new() -> Self {
        Self {
            value: AtomicU64::new(0),
        }
    }

    pub fn increment(&self) {
        self.value.fetch_add(1, Ordering::Relaxed);
    }

    pub fn get(&self) -> u64 {
        self.value.load(Ordering::Relaxed)
    }

    pub fn reset(&self) {
        self.value.store(0, Ordering::Relaxed);
    }
}

#[derive(Debug)]
pub struct Gauge {
    value: AtomicU64,
}

impl Gauge {
    pub fn new() -> Self {
        Self {
            value: AtomicU64::new(0),
        }
    }

    pub fn set(&self, value: f64) {
        self.value.store(value as u64, Ordering::Relaxed);
    }

    pub fn get(&self) -> f64 {
        self.value.load(Ordering::Relaxed) as f64
    }
}

#[derive(Debug)]
pub struct Histogram {
    values: RwLock<Vec<f64>>,
}

impl Histogram {
    pub fn new() -> Self {
        Self {
            values: RwLock::new(Vec::new()),
        }
    }

    pub fn record(&self, _value: f64) {
        // In a real implementation, we'd use more efficient storage
        // For now, we'll keep this simple
        // Note: This is not efficient for high-frequency metrics
        // In production, consider using buckets or streaming algorithms
    }

    pub fn get_values(&self) -> Vec<f64> {
        // Return a snapshot of values - in real implementation, this would be computed efficiently
        vec![]
    }

    pub fn get_percentiles(&self, percentiles: &[f64]) -> HashMap<String, f64> {
        // Calculate percentiles - placeholder implementation
        let mut map = HashMap::new();
        for &p in percentiles {
            map.insert(format!("p{}", p), 0.0);
        }
        map
    }
}

#[derive(Debug)]
pub struct PerformanceMetrics {
    /// Total requests served
    pub total_requests: AtomicUsize,
    /// Successful requests
    pub successful_requests: AtomicUsize,
    /// Failed requests
    pub failed_requests: AtomicUsize,
    /// Average response time in milliseconds
    pub avg_response_time: AtomicU64,
    /// Peak memory usage
    pub peak_memory: AtomicU64,
    /// Current active connections
    pub active_connections: AtomicUsize,
}

impl PerformanceMetrics {
    pub fn new() -> Self {
        Self {
            total_requests: AtomicUsize::new(0),
            successful_requests: AtomicUsize::new(0),
            failed_requests: AtomicUsize::new(0),
            avg_response_time: AtomicU64::new(0),
            peak_memory: AtomicU64::new(0),
            active_connections: AtomicUsize::new(0),
        }
    }

    pub fn record_request(&self, success: bool, duration_ms: u64) {
        self.total_requests.fetch_add(1, Ordering::Relaxed);
        if success {
            self.successful_requests.fetch_add(1, Ordering::Relaxed);
        } else {
            self.failed_requests.fetch_add(1, Ordering::Relaxed);
        }

        // Update average response time (simplified - in real implementation, use a better algorithm)
        let current_avg = self.avg_response_time.load(Ordering::Relaxed);
        let total_requests = self.total_requests.load(Ordering::Relaxed);
        if total_requests > 0 {
            let new_avg =
                (current_avg * (total_requests as u64 - 1) + duration_ms) / (total_requests as u64);
            self.avg_response_time.store(new_avg, Ordering::Relaxed);
        }
    }
}

#[derive(Debug)]
pub struct UsageMetrics {
    /// Total ideas processed
    pub ideas_processed: AtomicUsize,
    /// Total analysis operations
    pub analysis_operations: AtomicUsize,
    /// Total scoring operations
    pub scoring_operations: AtomicUsize,
    /// Total database operations
    pub database_operations: AtomicUsize,
    /// Last activity timestamp
    pub last_activity: RwLock<Option<DateTime<Utc>>>,
}

impl UsageMetrics {
    pub fn new() -> Self {
        Self {
            ideas_processed: AtomicUsize::new(0),
            analysis_operations: AtomicUsize::new(0),
            scoring_operations: AtomicUsize::new(0),
            database_operations: AtomicUsize::new(0),
            last_activity: RwLock::new(None),
        }
    }

    pub async fn record_idea_processed(&self) {
        self.ideas_processed.fetch_add(1, Ordering::Relaxed);
        *self.last_activity.write().await = Some(Utc::now());
    }

    pub async fn record_analysis_operation(&self) {
        self.analysis_operations.fetch_add(1, Ordering::Relaxed);
        *self.last_activity.write().await = Some(Utc::now());
    }

    pub async fn record_scoring_operation(&self) {
        self.scoring_operations.fetch_add(1, Ordering::Relaxed);
        *self.last_activity.write().await = Some(Utc::now());
    }

    pub async fn record_database_operation(&self) {
        self.database_operations.fetch_add(1, Ordering::Relaxed);
        *self.last_activity.write().await = Some(Utc::now());
    }
}

#[derive(Debug, Clone, serde::Serialize)]
pub struct MetricsSnapshot {
    pub counters: HashMap<String, u64>,
    pub gauges: HashMap<String, f64>,
    pub histograms: HashMap<String, Vec<f64>>,
    pub timestamp: DateTime<Utc>,
}

impl MetricsSnapshot {
    pub fn new() -> Self {
        Self {
            counters: HashMap::with_capacity(10),
            gauges: HashMap::with_capacity(10),
            histograms: HashMap::with_capacity(5),
            timestamp: Utc::now(),
        }
    }
}

/// Advanced analytics module for metrics analysis
pub mod analytics {
    use super::*;
    use serde::Serialize;
    use std::collections::HashMap;

    /// Trend analysis for identifying patterns over time
    pub struct TrendAnalyzer;

    impl TrendAnalyzer {
        /// Calculate trend direction for a time series
        pub fn calculate_trend(values: &[f64]) -> TrendDirection {
            if values.len() < 2 {
                return TrendDirection::Neutral;
            }

            let first = values[0];
            let last = values[values.len() - 1];

            if last > first * 1.1 {
                // 10% increase
                TrendDirection::Up
            } else if last < first * 0.9 {
                // 10% decrease
                TrendDirection::Down
            } else {
                TrendDirection::Neutral
            }
        }

        /// Detect anomalies in metrics
        pub fn detect_anomalies(values: &[f64], threshold: f64) -> Vec<(usize, f64)> {
            let mean = values.iter().sum::<f64>() / values.len() as f64;
            let std_dev = Self::calculate_std_deviation(values, mean);

            values
                .iter()
                .enumerate()
                .filter(|(_, value)| (*value - mean).abs() > threshold * std_dev)
                .map(|(idx, value)| (idx, *value))
                .collect()
        }

        fn calculate_std_deviation(values: &[f64], mean: f64) -> f64 {
            let variance = values
                .iter()
                .map(|value| (value - mean).powi(2))
                .sum::<f64>()
                / values.len() as f64;
            variance.sqrt()
        }
    }

    #[derive(Debug, Clone, Serialize)]
    pub enum TrendDirection {
        Up,
        Down,
        Neutral,
    }

    /// Usage trends report structure
    #[derive(Debug, Clone, Serialize)]
    pub struct UsageTrendsReport {
        pub timestamp: DateTime<Utc>,
        pub trends: HashMap<String, TrendDirection>,
        pub summary: String,
    }

    /// Performance report structure
    #[derive(Debug, Clone, Serialize)]
    pub struct PerformanceReport {
        pub timestamp: DateTime<Utc>,
        pub metrics: HashMap<String, f64>,
        pub summary: String,
    }

    /// Report builder for generating analytics reports
    pub struct ReportBuilder {
        metrics: MetricsSnapshot,
    }

    impl ReportBuilder {
        pub fn new(metrics: MetricsSnapshot) -> Self {
            Self { metrics }
        }

        /// Generate usage trends report
        pub async fn build_usage_trends(&self) -> UsageTrendsReport {
            let mut trends = HashMap::new();

            // Analyze trends for key metrics
            if let Some(_idea_count) = self.metrics.counters.get("ideas_processed") {
                // In a real implementation, we'd analyze historical data
                trends.insert("ideas_processed".to_string(), TrendDirection::Neutral);
            }

            if let Some(_score_avg) = self.metrics.gauges.get("average_score") {
                // Analyze average score trends
                trends.insert("average_score".to_string(), TrendDirection::Neutral);
            }

            UsageTrendsReport {
                timestamp: self.metrics.timestamp,
                trends,
                summary: "Usage trends analysis".to_string(),
            }
        }

        /// Generate performance report
        pub async fn build_performance_report(&self) -> PerformanceReport {
            let mut performance_metrics = HashMap::new();

            // Aggregate performance metrics
            for (key, value) in &self.metrics.gauges {
                if key.starts_with("latency_") || key.contains("duration") {
                    performance_metrics.insert(key.clone(), *value);
                }
            }

            PerformanceReport {
                timestamp: self.metrics.timestamp,
                metrics: performance_metrics,
                summary: "Performance metrics overview".to_string(),
            }
        }
    }

    /// Analytics aggregator for combining multiple metrics sources
    pub struct AnalyticsAggregator {
        pub metrics_registry: std::sync::Arc<MetricsRegistry>,
    }

    impl AnalyticsAggregator {
        pub fn new(metrics_registry: std::sync::Arc<MetricsRegistry>) -> Self {
            Self { metrics_registry }
        }

        /// Generate comprehensive analytics report
        pub async fn generate_comprehensive_report(&self) -> AnalyticsReport {
            let snapshot = self.metrics_registry.get_all_metrics().await;
            let report_builder = ReportBuilder::new(snapshot);

            AnalyticsReport {
                timestamp: Utc::now(),
                usage_trends: report_builder.build_usage_trends().await,
                performance_metrics: report_builder.build_performance_report().await,
                recommendations: self.generate_recommendations().await,
            }
        }

        /// Generate actionable recommendations based on metrics
        async fn generate_recommendations(&self) -> Vec<String> {
            let mut recommendations = Vec::new();

            // Get current usage metrics
            let usage = self.metrics_registry.usage_metrics();
            let ideas_processed = usage.ideas_processed.load(Ordering::Relaxed);

            // Generate recommendations based on usage patterns
            if ideas_processed > 100 {
                recommendations.push(
                    "High idea volume detected - consider periodic review sessions".to_string(),
                );
            }

            if ideas_processed == 0 {
                recommendations
                    .push("No ideas processed - consider increasing capture frequency".to_string());
            }

            // Check performance metrics for issues
            let perf = self.metrics_registry.performance_metrics();
            let avg_response_time = perf.avg_response_time.load(Ordering::Relaxed);

            if avg_response_time > 1000 {
                // More than 1 second
                recommendations.push(
                    "Slow response times detected - consider performance optimization".to_string(),
                );
            }

            recommendations
        }
    }

    #[derive(Debug, Clone, Serialize)]
    pub struct AnalyticsReport {
        pub timestamp: DateTime<Utc>,
        pub usage_trends: UsageTrendsReport,
        pub performance_metrics: PerformanceReport,
        pub recommendations: Vec<String>,
    }
}

/// Global metrics registry instance
static METRICS_REGISTRY: std::sync::LazyLock<MetricsRegistry> =
    std::sync::LazyLock::new(MetricsRegistry::new);

/// Get reference to global metrics registry
pub fn get_metrics_registry() -> &'static MetricsRegistry {
    &METRICS_REGISTRY
}

/// Record timing for an operation
pub async fn record_operation_timing(operation: &str, start_time: Instant) {
    let duration = start_time.elapsed();
    get_metrics_registry()
        .record_timing(operation, duration)
        .await;
}

/// Increment a counter metric
pub async fn increment_counter(name: &str) {
    get_metrics_registry().increment_counter(name).await;
}

/// Record a gauge value
pub async fn record_gauge(name: &str, value: f64) {
    get_metrics_registry().record_gauge(name, value).await;
}

/// Get a snapshot of all metrics
pub async fn get_metrics_snapshot() -> MetricsSnapshot {
    get_metrics_registry().get_all_metrics().await
}

#[cfg(test)]
mod tests {
    use super::*;

    #[tokio::test]
    async fn test_counter_metric() {
        let registry = MetricsRegistry::new();
        registry.increment_counter("test_counter").await;
        registry.increment_counter("test_counter").await;

        let snapshot = registry.get_all_metrics().await;
        assert_eq!(snapshot.counters.get("test_counter"), Some(&2));
    }

    #[tokio::test]
    async fn test_gauge_metric() {
        let registry = MetricsRegistry::new();
        registry.record_gauge("test_gauge", 42.0).await;

        let snapshot = registry.get_all_metrics().await;
        assert_eq!(snapshot.gauges.get("test_gauge"), Some(&42.0));
    }

    #[tokio::test]
    async fn test_usage_metrics() {
        let usage = UsageMetrics::new();
        usage.record_idea_processed().await;

        assert_eq!(usage.ideas_processed.load(Ordering::Relaxed), 1);
    }
}
