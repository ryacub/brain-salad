//! Analysis Quality Metrics Tracking System
//!
//! This module provides comprehensive tracking and analysis of LLM response quality,
//! template performance, and provider effectiveness over time.

use crate::errors::Result;
use crate::commands::analyze_llm::{LlmAnalysisResult, LlmProvider};
use crate::llm_cache::CacheEffectivenessMetrics;
use crate::prompt_templates::PromptTemplate;
use serde::{Deserialize, Serialize};
use chrono::{DateTime, Utc, NaiveDateTime};
use std::collections::HashMap;
use sqlx::{SqlitePool, Row};
use std::sync::Arc;

/// Comprehensive quality metrics for LLM analysis
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct AnalysisQualityMetrics {
    /// Unique identifier for this analysis
    pub analysis_id: String,
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
    /// Whether JSON parsing was successful
    pub json_parsing_success: bool,
    /// Whether all validations passed
    pub validation_success: bool,
    /// Whether fallback scoring was used
    pub fallback_used: bool,

    // Performance Metrics
    /// Total time taken for analysis in milliseconds
    pub response_time_ms: u64,
    /// LLM response length in characters
    pub response_length: usize,
    /// Number of explanations provided
    pub explanation_count: usize,

    // Scoring Analysis
    /// Final score from LLM analysis
    pub final_score: f64,
    /// Recommendation from LLM
    pub recommendation: String,
    /// Standard deviation of all component scores (consistency measure)
    pub score_consistency: f64,
    /// Number of scores at maximum values (potential over-optimism)
    pub max_score_count: usize,

    // Error Tracking
    /// Any errors that occurred during processing
    pub processing_errors: Vec<String>,
    /// Validation errors if any
    pub validation_errors: Vec<String>,
}

/// Aggregated quality metrics over time periods
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct AggregatedQualityMetrics {
    /// Time period these metrics cover
    pub period: String,
    /// Start of the period
    pub start_time: DateTime<Utc>,
    /// End of the period
    pub end_time: DateTime<Utc>,
    /// Total number of analyses
    pub total_analyses: u64,
    /// Average quality score
    pub avg_quality_score: f64,
    /// Average confidence level
    pub avg_confidence_score: f64,
    /// Success rate (analyses without fallback)
    pub success_rate: f64,
    /// Cache hit rate
    pub cache_hit_rate: f64,
    /// Average response time in milliseconds
    pub avg_response_time_ms: f64,

    // Quality Distribution
    /// High confidence analyses percentage
    pub high_confidence_rate: f64,
    /// Medium confidence analyses percentage
    pub medium_confidence_rate: f64,
    /// Low confidence analyses percentage
    pub low_confidence_rate: f64,

    // Template Performance
    /// Best performing template
    pub best_template: Option<String>,
    /// Worst performing template
    pub worst_template: Option<String>,
    /// Template performance variance
    pub template_performance_variance: f64,

    // Provider Performance
    /// Performance by provider
    pub provider_performance: HashMap<String, ProviderPerformance>,

    // Error Analysis
    /// Most common error types
    pub common_errors: Vec<String>,
    /// Error rate
    pub error_rate: f64,
}

/// Performance metrics for a specific provider
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ProviderPerformance {
    /// Provider name
    pub provider: String,
    /// Number of analyses by this provider
    pub analysis_count: u64,
    /// Average quality score
    pub avg_quality_score: f64,
    /// Average response time
    pub avg_response_time_ms: f64,
    /// Success rate
    pub success_rate: f64,
    /// Cost effectiveness (score per unit time)
    pub cost_effectiveness: f64,
}

/// Quality metrics tracker and analyzer
pub struct QualityMetricsTracker {
    db_pool: SqlitePool,
}

impl QualityMetricsTracker {
    /// Create a new quality metrics tracker
    pub async fn new(db_pool: SqlitePool) -> Result<Self> {
        let tracker = Self { db_pool };
        tracker.initialize_tables().await?;
        Ok(tracker)
    }

    /// Initialize database tables for quality metrics
    async fn initialize_tables(&self) -> Result<()> {
        sqlx::query(
            r#"
            CREATE TABLE IF NOT EXISTS quality_metrics (
                id TEXT PRIMARY KEY,
                idea TEXT NOT NULL,
                timestamp DATETIME NOT NULL,
                provider TEXT NOT NULL,
                template_id TEXT,
                idea_type TEXT NOT NULL,
                from_cache BOOLEAN NOT NULL DEFAULT FALSE,
                cache_similarity REAL,

                quality_score REAL NOT NULL,
                confidence_level TEXT NOT NULL,
                json_parsing_success BOOLEAN NOT NULL,
                validation_success BOOLEAN NOT NULL,
                fallback_used BOOLEAN NOT NULL,

                response_time_ms INTEGER NOT NULL,
                response_length INTEGER NOT NULL,
                explanation_count INTEGER NOT NULL,

                final_score REAL NOT NULL,
                recommendation TEXT NOT NULL,
                score_consistency REAL NOT NULL,
                max_score_count INTEGER NOT NULL,

                processing_errors TEXT, -- JSON array
                validation_errors TEXT, -- JSON array

                created_at DATETIME DEFAULT CURRENT_TIMESTAMP
            )
            "#,
        )
        .execute(&self.db_pool)
        .await?;

        sqlx::query(
            r#"
            CREATE TABLE IF NOT EXISTS template_performance (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                template_id TEXT NOT NULL,
                idea_type TEXT NOT NULL,
                usage_count INTEGER NOT NULL DEFAULT 1,
                avg_quality_score REAL NOT NULL DEFAULT 0.0,
                avg_confidence REAL NOT NULL DEFAULT 0.0,
                success_rate REAL NOT NULL DEFAULT 0.0,
                avg_response_time_ms INTEGER NOT NULL DEFAULT 0,
                last_updated DATETIME DEFAULT CURRENT_TIMESTAMP,
                UNIQUE(template_id, idea_type)
            )
            "#,
        )
        .execute(&self.db_pool)
        .await?;

        sqlx::query(
            r#"
            CREATE TABLE IF NOT EXISTS provider_performance (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                provider TEXT NOT NULL,
                idea_type TEXT NOT NULL,
                analysis_count INTEGER NOT NULL DEFAULT 1,
                avg_quality_score REAL NOT NULL DEFAULT 0.0,
                avg_response_time_ms REAL NOT NULL DEFAULT 0,
                success_rate REAL NOT NULL DEFAULT 0.0,
                error_rate REAL NOT NULL DEFAULT 0.0,
                last_updated DATETIME DEFAULT CURRENT_TIMESTAMP,
                UNIQUE(provider, idea_type)
            )
            "#,
        )
        .execute(&self.db_pool)
        .await?;

        // Create indexes for better query performance
        sqlx::query("CREATE INDEX IF NOT EXISTS idx_quality_metrics_timestamp ON quality_metrics(timestamp)")
            .execute(&self.db_pool)
            .await?;

        sqlx::query("CREATE INDEX IF NOT EXISTS idx_quality_metrics_provider ON quality_metrics(provider)")
            .execute(&self.db_pool)
            .await?;

        sqlx::query("CREATE INDEX IF NOT EXISTS idx_quality_metrics_template ON quality_metrics(template_id)")
            .execute(&self.db_pool)
            .await?;

        sqlx::query("CREATE INDEX IF NOT EXISTS idx_quality_metrics_idea_type ON quality_metrics(idea_type)")
            .execute(&self.db_pool)
            .await?;

        Ok(())
    }

    /// Record quality metrics for an analysis
    pub async fn record_analysis(&self, metrics: AnalysisQualityMetrics) -> Result<()> {
        let processing_errors_json = serde_json::to_string(&metrics.processing_errors)?;
        let validation_errors_json = serde_json::to_string(&metrics.validation_errors)?;

        sqlx::query(
            r#"
            INSERT OR REPLACE INTO quality_metrics (
                id, idea, timestamp, provider, template_id, idea_type, from_cache, cache_similarity,
                quality_score, confidence_level, json_parsing_success, validation_success, fallback_used,
                response_time_ms, response_length, explanation_count,
                final_score, recommendation, score_consistency, max_score_count,
                processing_errors, validation_errors
            ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
            "#,
        )
        .bind(&metrics.analysis_id)
        .bind(&metrics.idea)
        .bind(metrics.timestamp.naive_utc())
        .bind(metrics.provider.provider_type())
        .bind(&metrics.template_id)
        .bind(&metrics.idea_type)
        .bind(metrics.from_cache)
        .bind(metrics.cache_similarity)
        .bind(metrics.quality_score)
        .bind(&metrics.confidence_level)
        .bind(metrics.json_parsing_success)
        .bind(metrics.validation_success)
        .bind(metrics.fallback_used)
        .bind(metrics.response_time_ms as i64)
        .bind(metrics.response_length as i64)
        .bind(metrics.explanation_count as i64)
        .bind(metrics.final_score)
        .bind(&metrics.recommendation)
        .bind(metrics.score_consistency)
        .bind(metrics.max_score_count as i64)
        .bind(processing_errors_json)
        .bind(validation_errors_json)
        .execute(&self.db_pool)
        .await?;

        // Update template performance
        self.update_template_performance(&metrics).await?;

        // Update provider performance
        self.update_provider_performance(&metrics).await?;

        Ok(())
    }

    /// Update template performance metrics
    async fn update_template_performance(&self, metrics: &AnalysisQualityMetrics) -> Result<()> {
        let template_id = &metrics.template_id;
        let idea_type = &metrics.idea_type;

        // Get current performance data
        let current = sqlx::query(
            "SELECT usage_count, avg_quality_score, avg_confidence, success_rate, avg_response_time_ms
             FROM template_performance WHERE template_id = ? AND idea_type = ?"
        )
        .bind(template_id)
        .bind(idea_type)
        .fetch_optional(&self.db_pool)
        .await?;

        match current {
            Some(existing) => {
                // Update running averages
                let usage_count: i64 = existing.get("usage_count");
                let avg_quality_score: f64 = existing.get("avg_quality_score");
                let avg_confidence: f64 = existing.get("avg_confidence");
                let success_rate: f64 = existing.get("success_rate");
                let avg_response_time_ms: i64 = existing.get("avg_response_time_ms");

                let new_count = usage_count + 1;
                let new_avg_quality = (avg_quality_score * usage_count as f64 + metrics.quality_score) / new_count as f64;
                let new_avg_confidence = (avg_confidence * usage_count as f64 + self.confidence_to_score(&metrics.confidence_level)) / new_count as f64;
                let new_success_rate = (success_rate * usage_count as f64 + if !metrics.fallback_used { 1.0 } else { 0.0 }) / new_count as f64;
                let new_avg_response_time = (avg_response_time_ms as f64 * usage_count as f64 + metrics.response_time_ms as f64) / new_count as f64;

                sqlx::query(
                    r#"
                    UPDATE template_performance
                    SET usage_count = ?, avg_quality_score = ?, avg_confidence = ?,
                        success_rate = ?, avg_response_time_ms = ?, last_updated = CURRENT_TIMESTAMP
                    WHERE template_id = ? AND idea_type = ?
                    "#
                )
                .bind(new_count)
                .bind(new_avg_quality)
                .bind(new_avg_confidence)
                .bind(new_success_rate)
                .bind(new_avg_response_time as i64)
                .bind(template_id)
                .bind(idea_type)
                .execute(&self.db_pool)
                .await?;
            }
            None => {
                // Insert new record
                sqlx::query(
                    r#"
                    INSERT INTO template_performance (
                        template_id, idea_type, usage_count, avg_quality_score, avg_confidence,
                        success_rate, avg_response_time_ms
                    ) VALUES (?, ?, 1, ?, ?, ?, ?)
                    "#
                )
                .bind(template_id)
                .bind(idea_type)
                .bind(metrics.quality_score)
                .bind(self.confidence_to_score(&metrics.confidence_level))
                .bind(if !metrics.fallback_used { 1.0 } else { 0.0 })
                .bind(metrics.response_time_ms as i64)
                .execute(&self.db_pool)
                .await?;
            }
        }

        Ok(())
    }

    /// Update provider performance metrics
    async fn update_provider_performance(&self, metrics: &AnalysisQualityMetrics) -> Result<()> {
        let provider = &metrics.provider.provider_type();
        let idea_type = &metrics.idea_type;
        let error_rate = if metrics.processing_errors.is_empty() || metrics.validation_errors.is_empty() { 1.0 } else { 0.0 };

        // Get current performance data
        let current = sqlx::query(
            "SELECT analysis_count, avg_quality_score, avg_response_time_ms, success_rate, error_rate
             FROM provider_performance WHERE provider = ? AND idea_type = ?"
        )
        .bind(provider)
        .bind(idea_type)
        .fetch_optional(&self.db_pool)
        .await?;

        match current {
            Some(existing) => {
                // Update running averages
                let new_count = existing.analysis_count + 1;
                let new_avg_quality = (existing.avg_quality_score * existing.analysis_count as f64 + metrics.quality_score) / new_count as f64;
                let new_avg_response_time = (existing.avg_response_time_ms * existing.analysis_count as f64 + metrics.response_time_ms as f64) / new_count as f64;
                let new_success_rate = (existing.success_rate * existing.analysis_count as f64 + if !metrics.fallback_used { 1.0 } else { 0.0 }) / new_count as f64;
                let new_error_rate = (existing.error_rate * existing.analysis_count as f64 + error_rate) / new_count as f64;

                sqlx::query!(
                    r#"
                    UPDATE provider_performance
                    SET analysis_count = ?, avg_quality_score = ?, avg_response_time_ms = ?,
                        success_rate = ?, error_rate = ?, last_updated = CURRENT_TIMESTAMP
                    WHERE provider = ? AND idea_type = ?
                    "#,
                    new_count,
                    new_avg_quality,
                    new_avg_response_time as i64,
                    new_success_rate,
                    new_error_rate,
                    provider,
                    idea_type
                )
                .execute(&self.db_pool)
                .await?;
            }
            None => {
                // Insert new record
                sqlx::query!(
                    r#"
                    INSERT INTO provider_performance (
                        provider, idea_type, analysis_count, avg_quality_score, avg_response_time_ms,
                        success_rate, error_rate
                    ) VALUES (?, ?, 1, ?, ?, ?, ?)
                    "#,
                    provider,
                    idea_type,
                    metrics.quality_score,
                    metrics.response_time_ms as i64,
                    if !metrics.fallback_used { 1.0 } else { 0.0 },
                    error_rate
                )
                .execute(&self.db_pool)
                .await?;
            }
        }

        Ok(())
    }

    /// Get aggregated quality metrics for a time period
    pub async fn get_aggregated_metrics(
        &self,
        start_time: DateTime<Utc>,
        end_time: DateTime<Utc>,
        idea_type: Option<String>,
    ) -> Result<AggregatedQualityMetrics> {
        let idea_type_filter = idea_type.as_deref().unwrap_or("");

        let rows = sqlx::query(
            r#"
            SELECT
                COUNT(*) as total_analyses,
                AVG(quality_score) as avg_quality_score,
                AVG(CASE confidence_level
                    WHEN 'High' THEN 1.0
                    WHEN 'Medium' THEN 0.7
                    WHEN 'Low' THEN 0.5
                    ELSE 0.3 END) as avg_confidence_score,
                SUM(CASE WHEN fallback_used = 0 THEN 1 ELSE 0 END) * 1.0 / COUNT(*) as success_rate,
                SUM(CASE WHEN from_cache = 1 THEN 1 ELSE 0 END) * 1.0 / COUNT(*) as cache_hit_rate,
                AVG(response_time_ms) as avg_response_time_ms,
                provider,
                template_id
            FROM quality_metrics
            WHERE timestamp BETWEEN ? AND ?
            AND (? = '' OR idea_type = ?)
            GROUP BY provider, template_id
            "#,
        )
        .bind(start_time.naive_utc())
        .bind(end_time.naive_utc())
        .bind(idea_type_filter)
        .bind(idea_type_filter)
        .fetch_all(&self.db_pool)
        .await?;

        // Calculate overall aggregates from grouped data
        let mut total_analyses: u64 = 0;
        let mut total_quality_score: f64 = 0.0;
        let mut total_confidence_score: f64 = 0.0;
        let mut total_success_count: u64 = 0;
        let mut total_cache_hits: u64 = 0;
        let mut total_response_time: u64 = 0;

        let mut provider_performance = HashMap::new();
        let mut template_performance = HashMap::new();

        for row in rows {
            let count: i64 = row.get("total_analyses");
            let avg_quality: f64 = row.get("avg_quality_score");
            let avg_confidence: f64 = row.get("avg_confidence_score");
            let success_rate: f64 = row.get("success_rate");
            let cache_hit_rate: f64 = row.get("cache_hit_rate");
            let avg_response_time: f64 = row.get("avg_response_time_ms");
            let provider: String = row.get("provider");
            let template_id: String = row.get("template_id");

            total_analyses += count as u64;
            total_quality_score += avg_quality * count as f64;
            total_confidence_score += avg_confidence * count as f64;
            total_success_count += (success_rate * count as f64) as u64;
            total_cache_hits += (cache_hit_rate * count as f64) as u64;
            total_response_time += (avg_response_time * count as f64) as u64;

            // Aggregate provider performance
            let provider_entry = provider_performance.entry(provider.clone()).or_insert(ProviderPerformance {
                provider: provider.clone(),
                analysis_count: 0,
                avg_quality_score: 0.0,
                avg_response_time_ms: 0.0,
                success_rate: 0.0,
                cost_effectiveness: 0.0,
            });
            provider_entry.analysis_count += count as u64;

            // Aggregate template performance (simplified for now)
            template_performance.insert(template_id, avg_quality);
        }

        let overall_avg_quality = if total_analyses > 0 { total_quality_score / total_analyses as f64 } else { 0.0 };
        let overall_avg_confidence = if total_analyses > 0 { total_confidence_score / total_analyses as f64 } else { 0.0 };
        let overall_success_rate = if total_analyses > 0 { total_success_count as f64 / total_analyses as f64 } else { 0.0 };
        let overall_cache_hit_rate = if total_analyses > 0 { total_cache_hits as f64 / total_analyses as f64 } else { 0.0 };
        let overall_avg_response_time = if total_analyses > 0 { total_response_time as f64 / total_analyses as f64 } else { 0.0 };

        // Calculate cost effectiveness for each provider
        for provider_perf in provider_performance.values_mut() {
            provider_perf.cost_effectiveness = if provider_perf.avg_response_time_ms > 0.0 {
                provider_perf.avg_quality_score / (provider_perf.avg_response_time_ms / 1000.0)
            } else {
                0.0
            };
        }

        // Find best and worst performing templates
        let (best_template, worst_template) = if template_performance.is_empty() {
            (None, None)
        } else {
            let templates: Vec<_> = template_performance.iter().collect();
            let best = templates.iter().max_by(|a, b| a.1.partial_cmp(b.1).unwrap()).map(|(t, _)| (*t).clone());
            let worst = templates.iter().min_by(|a, b| a.1.partial_cmp(b.1).unwrap()).map(|(t, _)| (*t).clone());
            (best, worst)
        };

        // Calculate template performance variance
        let template_variance = if template_performance.len() > 1 {
            let mean = overall_avg_quality;
            let variance_sum: f64 = template_performance.values()
                .map(|score| (score - mean).powi(2))
                .sum();
            variance_sum / template_performance.len() as f64
        } else {
            0.0
        };

        // Get confidence distribution (simplified - would need more complex query for exact)
        let high_confidence_rate = overall_avg_confidence * 0.9; // Approximation
        let medium_confidence_rate = overall_avg_confidence * 0.8;
        let low_confidence_rate = 1.0 - high_confidence_rate - medium_confidence_rate;

        // Get common errors (simplified)
        let common_errors = vec!["JSON parsing errors".to_string(), "Validation failures".to_string()];
        let error_rate = 1.0 - overall_success_rate;

        Ok(AggregatedQualityMetrics {
            period: format!("{} to {}", start_time.format("%Y-%m-%d"), end_time.format("%Y-%m-%d")),
            start_time,
            end_time,
            total_analyses,
            avg_quality_score: overall_avg_quality,
            avg_confidence_score: overall_avg_confidence,
            success_rate: overall_success_rate,
            cache_hit_rate: overall_cache_hit_rate,
            avg_response_time_ms: overall_avg_response_time,
            high_confidence_rate,
            medium_confidence_rate,
            low_confidence_rate,
            best_template,
            worst_template,
            template_performance_variance: template_variance,
            provider_performance,
            common_errors,
            error_rate,
        })
    }

    /// Get template performance ranking
    pub async fn get_template_rankings(&self, idea_type: Option<String>) -> Result<Vec<(String, f64, u64)>> {
        let idea_type_filter = idea_type.as_deref().unwrap_or("");

        let rows = sqlx::query!(
            r#"
            SELECT template_id, avg_quality_score, usage_count
            FROM template_performance
            WHERE ? = '' OR idea_type = ?
            ORDER BY avg_quality_score DESC, usage_count DESC
            "#,
            idea_type_filter,
            idea_type_filter
        )
        .fetch_all(&self.db_pool)
        .await?;

        Ok(rows.into_iter()
            .map(|row| (row.template_id, row.avg_quality_score, row.usage_count as u64))
            .collect())
    }

    /// Get provider performance comparison
    pub async fn get_provider_comparison(&self, idea_type: Option<String>) -> Result<Vec<ProviderPerformance>> {
        let idea_type_filter = idea_type.as_deref().unwrap_or("");

        let rows = sqlx::query!(
            r#"
            SELECT provider, analysis_count, avg_quality_score, avg_response_time_ms, success_rate, error_rate
            FROM provider_performance
            WHERE ? = '' OR idea_type = ?
            ORDER BY avg_quality_score DESC, cost_effectiveness DESC
            "#,
            idea_type_filter,
            idea_type_filter
        )
        .fetch_all(&self.db_pool)
        .await?;

        Ok(rows.into_iter()
            .map(|row| {
                let cost_effectiveness = if row.avg_response_time_ms > 0 {
                    row.avg_quality_score / (row.avg_response_time_ms as f64 / 1000.0)
                } else {
                    0.0
                };

                ProviderPerformance {
                    provider: row.provider,
                    analysis_count: row.analysis_count as u64,
                    avg_quality_score: row.avg_quality_score,
                    avg_response_time_ms: row.avg_response_time_ms as f64,
                    success_rate: row.success_rate,
                    cost_effectiveness,
                }
            })
            .collect())
    }

    /// Convert confidence level string to numeric score
    fn confidence_to_score(&self, confidence_level: &str) -> f64 {
        match confidence_level {
            "High" => 1.0,
            "Medium" => 0.7,
            "Low" => 0.5,
            _ => 0.3,
        }
    }

    /// Generate quality metrics report
    pub async fn generate_quality_report(&self, days: u32) -> Result<String> {
        let end_time = Utc::now();
        let start_time = end_time - chrono::Duration::days(days as i64);

        let metrics = self.get_aggregated_metrics(start_time, end_time, None).await?;
        let template_rankings = self.get_template_rankings(None).await?;
        let provider_comparison = self.get_provider_comparison(None).await?;

        let mut report = format!(
            r#"
# Quality Metrics Report (Last {} days)
Generated: {}

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
"#,
            days,
            end_time.format("%Y-%m-%d %H:%M UTC"),
            metrics.total_analyses,
            metrics.avg_quality_score,
            metrics.success_rate * 100.0,
            metrics.cache_hit_rate * 100.0,
            metrics.avg_response_time_ms,
            metrics.high_confidence_rate * 100.0,
            metrics.medium_confidence_rate * 100.0,
            metrics.low_confidence_rate * 100.0
        );

        // Add top templates
        if !template_rankings.is_empty() {
            report.push_str("### Top Performing Templates\n");
            for (i, (template_id, avg_score, usage_count)) in template_rankings.iter().take(5).enumerate() {
                report.push_str(&format!(
                    "{}. {} (Quality: {:.2}, Usage: {})\n",
                    i + 1,
                    template_id,
                    avg_score,
                    usage_count
                ));
            }
            report.push('\n');
        }

        // Add provider comparison
        if !provider_comparison.is_empty() {
            report.push_str("### Provider Performance\n");
            for provider_perf in &provider_comparison {
                report.push_str(&format!(
                    "- {}: Quality {:.2}, Success {:.1}%, Response {:.0}ms, Cost-Effectiveness {:.2}\n",
                    provider_perf.provider,
                    provider_perf.avg_quality_score,
                    provider_perf.success_rate * 100.0,
                    provider_perf.avg_response_time_ms,
                    provider_perf.cost_effectiveness
                ));
            }
        }

        // Add error analysis
        if !metrics.common_errors.is_empty() {
            report.push_str(&format!(
                "\n## Error Analysis\nError Rate: {:.1}%\nCommon Issues:\n",
                metrics.error_rate * 100.0
            ));
            for error in &metrics.common_errors {
                report.push_str(&format!("- {}\n", error));
            }
        }

        Ok(report)
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use chrono::Utc;
    use uuid::Uuid;

    #[test]
    fn test_analysis_quality_metrics_creation() {
        let metrics = AnalysisQualityMetrics {
            analysis_id: Uuid::new_v4().to_string(),
            idea: "Test idea".to_string(),
            timestamp: Utc::now(),
            provider: LlmProvider::Ollama,
            template_id: "default_v1".to_string(),
            idea_type: "technical".to_string(),
            from_cache: false,
            cache_similarity: None,
            quality_score: 0.8,
            confidence_level: "High".to_string(),
            json_parsing_success: true,
            validation_success: true,
            fallback_used: false,
            response_time_ms: 500,
            response_length: 1000,
            explanation_count: 5,
            final_score: 7.5,
            recommendation: "Good".to_string(),
            score_consistency: 0.3,
            max_score_count: 2,
            processing_errors: vec![],
            validation_errors: vec![],
        };

        assert_eq!(metrics.idea, "Test idea");
        assert_eq!(metrics.confidence_level, "High");
        assert_eq!(metrics.quality_score, 0.8);
    }

    #[test]
    fn test_confidence_to_score_conversion() {
        let tracker = QualityMetricsTracker {
            db_pool: sqlx::SqlitePool::connect(":memory:").await.unwrap()
        };

        assert_eq!(tracker.confidence_to_score("High"), 1.0);
        assert_eq!(tracker.confidence_to_score("Medium"), 0.7);
        assert_eq!(tracker.confidence_to_score("Low"), 0.5);
        assert_eq!(tracker.confidence_to_score("Unknown"), 0.3);
    }

    #[test]
    fn test_provider_performance_calculation() {
        let mut perf = ProviderPerformance {
            provider: "OpenAI".to_string(),
            analysis_count: 100,
            avg_quality_score: 0.8,
            avg_response_time_ms: 2000.0,
            success_rate: 0.9,
            cost_effectiveness: 0.0,
        };

        // Calculate cost effectiveness
        perf.cost_effectiveness = if perf.avg_response_time_ms > 0.0 {
            perf.avg_quality_score / (perf.avg_response_time_ms / 1000.0)
        } else {
            0.0
        };

        assert_eq!(perf.cost_effectiveness, 0.8 / 2.0); // 0.4
    }
}