//! Simplified Quality Metrics Tracking System
//!
//! This module provides basic quality metrics tracking without complex SQLx query macros.

use crate::commands::analyze_llm::LlmProvider;
use crate::errors::Result;
use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use std::sync::{Arc, RwLock};

/// Simplified quality metrics for LLM analysis
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct SimpleQualityMetrics {
    /// The original idea text
    pub idea: String,
    /// Timestamp when analysis was performed
    pub timestamp: DateTime<Utc>,
    /// Provider that performed the analysis
    pub provider: LlmProvider,
    /// Template used for the analysis
    pub template_id: String,
    /// Idea type classification
    pub idea_type: String,
    /// Whether the response came from cache
    pub from_cache: bool,
    /// If cached, the similarity score
    pub cache_similarity: Option<f64>,

    // Response Quality Metrics
    /// Overall quality score (0.0 - 1.0)
    pub quality_score: f64,
    /// Confidence level of the analysis
    pub confidence_level: String,
    /// Whether fallback scoring was used
    pub fallback_used: bool,

    // Performance Metrics
    /// Total time taken for analysis in milliseconds
    pub response_time_ms: u64,
    /// Final score from LLM analysis
    pub final_score: f64,
    /// Recommendation from LLM
    pub recommendation: String,
}

/// Simplified quality metrics tracker using in-memory storage
pub struct SimpleQualityTracker {
    metrics: Arc<RwLock<Vec<SimpleQualityMetrics>>>,
    template_performance: Arc<RwLock<HashMap<String, TemplateStats>>>,
    provider_performance: Arc<RwLock<HashMap<String, ProviderStats>>>,
}

/// Performance statistics for templates
#[derive(Debug, Clone, Default)]
pub struct TemplateStats {
    pub usage_count: u64,
    pub avg_quality_score: f64,
    pub avg_response_time_ms: f64,
    pub success_rate: f64,
}

/// Performance statistics for providers
#[derive(Debug, Clone, Default)]
pub struct ProviderStats {
    pub analysis_count: u64,
    pub avg_quality_score: f64,
    pub avg_response_time_ms: f64,
    pub success_rate: f64,
    pub cost_effectiveness: f64,
}

impl SimpleQualityTracker {
    /// Create a new simple quality tracker
    pub fn new() -> Self {
        Self {
            metrics: Arc::new(RwLock::new(Vec::new())),
            template_performance: Arc::new(RwLock::new(HashMap::new())),
            provider_performance: Arc::new(RwLock::new(HashMap::new())),
        }
    }

    /// Record quality metrics for an analysis
    pub async fn record_analysis(&self, metrics: SimpleQualityMetrics) -> Result<()> {
        // Store the metrics
        {
            let mut metrics_store = self.metrics.write().unwrap();
            metrics_store.push(metrics.clone());
        }

        // Update template performance
        self.update_template_performance(&metrics).await;

        // Update provider performance
        self.update_provider_performance(&metrics).await;

        Ok(())
    }

    /// Update template performance metrics
    async fn update_template_performance(&self, metrics: &SimpleQualityMetrics) {
        let mut template_perf = self.template_performance.write().unwrap();
        let stats = template_perf
            .entry(metrics.template_id.clone())
            .or_default();

        let new_count = stats.usage_count + 1;
        stats.avg_quality_score = (stats.avg_quality_score * stats.usage_count as f64
            + metrics.quality_score)
            / new_count as f64;
        stats.avg_response_time_ms = (stats.avg_response_time_ms * stats.usage_count as f64
            + metrics.response_time_ms as f64)
            / new_count as f64;
        stats.success_rate = (stats.success_rate * stats.usage_count as f64
            + if !metrics.fallback_used { 1.0 } else { 0.0 })
            / new_count as f64;
        stats.usage_count = new_count;
    }

    /// Update provider performance metrics
    async fn update_provider_performance(&self, metrics: &SimpleQualityMetrics) {
        let mut provider_perf = self.provider_performance.write().unwrap();
        let provider_key = format!("{}-{}", metrics.provider.provider_type(), metrics.idea_type);
        let stats = provider_perf.entry(provider_key).or_default();

        let new_count = stats.analysis_count + 1;
        stats.avg_quality_score = (stats.avg_quality_score * stats.analysis_count as f64
            + metrics.quality_score)
            / new_count as f64;
        stats.avg_response_time_ms = (stats.avg_response_time_ms * stats.analysis_count as f64
            + metrics.response_time_ms as f64)
            / new_count as f64;
        stats.success_rate = (stats.success_rate * stats.analysis_count as f64
            + if !metrics.fallback_used { 1.0 } else { 0.0 })
            / new_count as f64;
        stats.cost_effectiveness = if stats.avg_response_time_ms > 0.0 {
            stats.avg_quality_score / (stats.avg_response_time_ms / 1000.0)
        } else {
            0.0
        };
        stats.analysis_count = new_count;
    }

    /// Get template performance ranking
    pub async fn get_template_rankings(
        &self,
        idea_type: Option<String>,
    ) -> Vec<(String, f64, u64)> {
        let template_perf = self.template_performance.read().unwrap();
        let mut rankings: Vec<_> = template_perf
            .iter()
            .map(|(template_id, stats)| {
                (
                    template_id.clone(),
                    stats.avg_quality_score,
                    stats.usage_count,
                )
            })
            .collect();

        rankings.sort_by(|a, b| b.1.partial_cmp(&a.1).unwrap_or(std::cmp::Ordering::Equal));
        rankings
    }

    /// Get provider performance comparison
    pub async fn get_provider_comparison(&self) -> Vec<ProviderStats> {
        let provider_perf = self.provider_performance.read().unwrap();
        provider_perf.values().cloned().collect()
    }

    /// Get basic quality metrics summary
    pub async fn get_quality_summary(&self, days: u32) -> QualitySummary {
        let metrics_store = self.metrics.read().unwrap();
        let cutoff_time = Utc::now() - chrono::Duration::days(days as i64);

        let recent_metrics: Vec<_> = metrics_store
            .iter()
            .filter(|m| m.timestamp > cutoff_time)
            .collect();

        if recent_metrics.is_empty() {
            return QualitySummary::default();
        }

        let total_analyses = recent_metrics.len() as u64;
        let avg_quality_score =
            recent_metrics.iter().map(|m| m.quality_score).sum::<f64>() / total_analyses as f64;
        let success_rate = recent_metrics.iter().filter(|m| !m.fallback_used).count() as f64
            / total_analyses as f64;
        let cache_hit_rate =
            recent_metrics.iter().filter(|m| m.from_cache).count() as f64 / total_analyses as f64;
        let avg_response_time_ms = recent_metrics
            .iter()
            .map(|m| m.response_time_ms)
            .sum::<u64>()
            / total_analyses;

        // Confidence distribution
        let high_confidence = recent_metrics
            .iter()
            .filter(|m| m.confidence_level == "High")
            .count() as f64
            / total_analyses as f64;
        let medium_confidence = recent_metrics
            .iter()
            .filter(|m| m.confidence_level == "Medium")
            .count() as f64
            / total_analyses as f64;
        let low_confidence = recent_metrics
            .iter()
            .filter(|m| m.confidence_level == "Low")
            .count() as f64
            / total_analyses as f64;

        QualitySummary {
            total_analyses,
            avg_quality_score,
            success_rate,
            cache_hit_rate,
            avg_response_time_ms,
            high_confidence_rate: high_confidence,
            medium_confidence_rate: medium_confidence,
            low_confidence_rate: low_confidence,
        }
    }

    /// Generate simple quality report
    pub async fn generate_quality_report(&self, days: u32) -> String {
        let summary = self.get_quality_summary(days).await;
        let template_rankings = self.get_template_rankings(None).await;
        let provider_comparison = self.get_provider_comparison().await;

        format!(
            r#"
# Quality Metrics Report (Last {} days)

## Overall Performance
- Total Analyses: {}
- Average Quality Score: {:.2}
- Success Rate: {:.1}%
- Cache Hit Rate: {:.1}%
- Average Response Time: {:.0}ms

## Confidence Distribution
- High Confidence: {:.1}%
- Medium Confidence: {:.1}%
- Low Confidence: {:.1}%

## Template Performance
{}
## Provider Performance
{}
"#,
            days,
            summary.total_analyses,
            summary.avg_quality_score,
            summary.success_rate * 100.0,
            summary.cache_hit_rate * 100.0,
            summary.avg_response_time_ms,
            summary.high_confidence_rate * 100.0,
            summary.medium_confidence_rate * 100.0,
            summary.low_confidence_rate * 100.0,
            self.format_template_rankings(&template_rankings),
            self.format_provider_comparison(&provider_comparison)
        )
    }

    fn format_template_rankings(&self, rankings: &[(String, f64, u64)]) -> String {
        if rankings.is_empty() {
            "No template data available.\n".to_string()
        } else {
            let mut result = String::new();
            for (i, (template_id, avg_score, usage_count)) in rankings.iter().take(5).enumerate() {
                result.push_str(&format!(
                    "{}. {} (Quality: {:.2}, Usage: {})\n",
                    i + 1,
                    template_id,
                    avg_score,
                    usage_count
                ));
            }
            result
        }
    }

    fn format_provider_comparison(&self, providers: &[ProviderStats]) -> String {
        if providers.is_empty() {
            "No provider data available.\n".to_string()
        } else {
            let mut result = String::new();
            for provider_perf in providers {
                result.push_str(&format!(
                    "- Provider: Quality {:.2}, Success {:.1}%, Response {:.0}ms, Cost-Effectiveness {:.2}\n",
                    provider_perf.avg_quality_score,
                    provider_perf.success_rate * 100.0,
                    provider_perf.avg_response_time_ms,
                    provider_perf.cost_effectiveness
                ));
            }
            result
        }
    }
}

/// Basic quality summary
#[derive(Debug, Clone, Default)]
pub struct QualitySummary {
    pub total_analyses: u64,
    pub avg_quality_score: f64,
    pub success_rate: f64,
    pub cache_hit_rate: f64,
    pub avg_response_time_ms: u64,
    pub high_confidence_rate: f64,
    pub medium_confidence_rate: f64,
    pub low_confidence_rate: f64,
}

impl Default for SimpleQualityTracker {
    fn default() -> Self {
        Self::new()
    }
}

/// Global quality tracker instance
static QUALITY_TRACKER: std::sync::OnceLock<SimpleQualityTracker> = std::sync::OnceLock::new();

/// Get the global quality tracker instance
pub fn get_quality_tracker() -> &'static SimpleQualityTracker {
    QUALITY_TRACKER.get_or_init(SimpleQualityTracker::new)
}

#[cfg(test)]
mod tests {
    use super::*;
    use chrono::Utc;

    #[test]
    fn test_simple_quality_metrics() {
        let metrics = SimpleQualityMetrics {
            idea: "Test idea".to_string(),
            timestamp: Utc::now(),
            provider: LlmProvider::Ollama,
            template_id: "default_v1".to_string(),
            idea_type: "technical".to_string(),
            from_cache: false,
            cache_similarity: None,
            quality_score: 0.8,
            confidence_level: "High".to_string(),
            fallback_used: false,
            response_time_ms: 500,
            final_score: 7.5,
            recommendation: "Good".to_string(),
        };

        assert_eq!(metrics.idea, "Test idea");
        assert_eq!(metrics.quality_score, 0.8);
    }

    #[tokio::test]
    async fn test_quality_tracker() {
        let tracker = SimpleQualityTracker::new();

        let metrics = SimpleQualityMetrics {
            idea: "Test idea".to_string(),
            timestamp: Utc::now(),
            provider: LlmProvider::Ollama,
            template_id: "technical_v1".to_string(),
            idea_type: "technical".to_string(),
            from_cache: false,
            cache_similarity: None,
            quality_score: 0.8,
            confidence_level: "High".to_string(),
            fallback_used: false,
            response_time_ms: 500,
            final_score: 7.5,
            recommendation: "Good".to_string(),
        };

        tracker.record_analysis(metrics).await.unwrap();

        let summary = tracker.get_quality_summary(7).await;
        assert_eq!(summary.total_analyses, 1);
        assert_eq!(summary.avg_quality_score, 0.8);
    }

    #[tokio::test]
    async fn test_template_performance_tracking() {
        let tracker = SimpleQualityTracker::new();

        let metrics1 = SimpleQualityMetrics {
            idea: "Test idea 1".to_string(),
            timestamp: Utc::now(),
            provider: LlmProvider::Ollama,
            template_id: "technical_v1".to_string(),
            idea_type: "technical".to_string(),
            from_cache: false,
            cache_similarity: None,
            quality_score: 0.8,
            confidence_level: "High".to_string(),
            fallback_used: false,
            response_time_ms: 500,
            final_score: 7.5,
            recommendation: "Good".to_string(),
        };

        let metrics2 = SimpleQualityMetrics {
            idea: "Test idea 2".to_string(),
            timestamp: Utc::now(),
            provider: LlmProvider::Claude,
            template_id: "technical_v1".to_string(),
            idea_type: "technical".to_string(),
            from_cache: false,
            cache_similarity: None,
            quality_score: 0.9,
            confidence_level: "High".to_string(),
            fallback_used: false,
            response_time_ms: 300,
            final_score: 8.5,
            recommendation: "Priority".to_string(),
        };

        tracker.record_analysis(metrics1).await.unwrap();
        tracker.record_analysis(metrics2).await.unwrap();

        let rankings = tracker.get_template_rankings(None).await;
        assert_eq!(rankings.len(), 1);
        assert_eq!(rankings[0].0, "technical_v1");
        assert_eq!(rankings[0].1, 0.85); // Average of 0.8 and 0.9
        assert_eq!(rankings[0].2, 2);
    }
}
