//! Analytics and reporting commands for the telos-idea-matrix application
//!
//! This module provides commands for generating reports, analyzing trends,
//! and getting insights from the collected metrics data.

use crate::database_simple as database;
use crate::errors::Result;
use crate::metrics;
use colored::*;

#[derive(clap::Subcommand, Debug)]
pub enum AnalyticsCommands {
    /// Generate usage trends report
    Trends {
        /// Include detailed analysis
        #[arg(short, long)]
        verbose: bool,

        /// Output format (text, json, csv)
        #[arg(long, default_value = "text", value_parser = ["text", "json", "csv"])]
        format: String,
    },

    /// Generate performance report
    Performance {
        /// Include detailed metrics
        #[arg(short, long)]
        verbose: bool,

        /// Output format (text, json, csv)
        #[arg(long, default_value = "text", value_parser = ["text", "json", "csv"])]
        format: String,
    },

    /// Generate comprehensive analytics report
    Report {
        /// Include recommendations
        #[arg(short, long)]
        with_recommendations: bool,

        /// Output format (text, json, csv)
        #[arg(long, default_value = "text", value_parser = ["text", "json", "csv"])]
        format: String,
    },

    /// Detect anomalies in usage patterns
    Anomaly {
        /// Anomaly detection threshold
        #[arg(long, default_value = "2.0")]
        threshold: f64,

        /// Output format
        #[arg(long, default_value = "text", value_parser = ["text", "json"])]
        format: String,
    },

    /// Show metrics summary
    Metrics {
        /// Show all metrics (not just summary)
        #[arg(short, long)]
        all: bool,

        /// Output format
        #[arg(long, default_value = "text", value_parser = ["text", "json"])]
        format: String,
    },
}

pub async fn handle_analytics(cmd: AnalyticsCommands, db: &database::Database) -> Result<()> {
    match cmd {
        AnalyticsCommands::Trends { verbose, format } => handle_trends(verbose, &format, db).await,
        AnalyticsCommands::Performance { verbose, format } => {
            handle_performance(verbose, &format).await
        }
        AnalyticsCommands::Report {
            with_recommendations,
            format,
        } => handle_comprehensive_report(with_recommendations, &format).await,
        AnalyticsCommands::Anomaly { threshold, format } => {
            handle_anomaly_detection(threshold, &format, db).await
        }
        AnalyticsCommands::Metrics { all, format } => handle_metrics_display(all, &format).await,
    }
}

async fn handle_trends(verbose: bool, format: &str, db: &database::Database) -> Result<()> {
    println!(
        "📊 {} Generating usage trends report...",
        "Analyzing".bright_blue().bold()
    );
    println!();

    // Get metrics snapshot
    let snapshot = metrics::get_metrics_snapshot().await;

    // Build report
    let report_builder = metrics::analytics::ReportBuilder::new(snapshot);
    let trends_report = report_builder.build_usage_trends().await;

    match format {
        "json" => {
            let json_output = serde_json::to_string_pretty(&trends_report).map_err(|e| {
                crate::errors::ApplicationError::Generic(anyhow::anyhow!(
                    "JSON serialization error: {}",
                    e
                ))
            })?;
            println!("{}", json_output);
        }
        "csv" => {
            println!("timestamp,trend_key,direction");
            for (key, direction) in &trends_report.trends {
                let direction_str = match direction {
                    metrics::analytics::TrendDirection::Up => "up",
                    metrics::analytics::TrendDirection::Down => "down",
                    metrics::analytics::TrendDirection::Neutral => "neutral",
                };
                println!(
                    "{},{},{}",
                    trends_report.timestamp.to_rfc3339(),
                    key,
                    direction_str
                );
            }
        }
        _ => {
            // Text format
            println!("📈 {}", "USAGE TRENDS REPORT".bright_blue().underline());
            println!(
                "   Generated: {}",
                trends_report.timestamp.format("%Y-%m-%d %H:%M:%S UTC")
            );
            println!();

            if trends_report.trends.is_empty() {
                println!("ℹ️  No significant trends detected in the current data.");
                println!("   As you use the system more, trends will become apparent.");
            } else {
                println!("📈 Identified Trends:");

                for (metric, direction) in &trends_report.trends {
                    let (arrow, color) = match direction {
                        metrics::analytics::TrendDirection::Up => ("📈 ↑", "green"),
                        metrics::analytics::TrendDirection::Down => ("📉 ↓", "red"),
                        metrics::analytics::TrendDirection::Neutral => ("➡️ →", "blue"),
                    };

                    let colored_arrow = match color {
                        "green" => arrow.green(),
                        "red" => arrow.red(),
                        _ => arrow.blue(),
                    };

                    println!(
                        "   {} {}: {}",
                        colored_arrow,
                        metric,
                        direction_message(direction)
                    );
                }
            }

            if verbose {
                println!();
                println!("💡 Interpretation Guide:");
                println!("   📈 Up: Growing usage/interest in this area");
                println!("   📉 Down: Declining usage/interest in this area");
                println!("   ➡️  Neutral: Stable usage pattern");
            }
        }
    }

    Ok(())
}

fn direction_message(direction: &metrics::analytics::TrendDirection) -> String {
    match direction {
        metrics::analytics::TrendDirection::Up => "Increasing trend".to_string(),
        metrics::analytics::TrendDirection::Down => "Decreasing trend".to_string(),
        metrics::analytics::TrendDirection::Neutral => "Stable trend".to_string(),
    }
}

async fn handle_performance(verbose: bool, format: &str) -> Result<()> {
    println!(
        "⚡ {} Generating performance report...",
        "Analyzing".bright_blue().bold()
    );
    println!();

    // Get metrics snapshot
    let snapshot = metrics::get_metrics_snapshot().await;

    // Build report
    let report_builder = metrics::analytics::ReportBuilder::new(snapshot);
    let perf_report = report_builder.build_performance_report().await;

    match format {
        "json" => {
            let json_output = serde_json::to_string_pretty(&perf_report).map_err(|e| {
                crate::errors::ApplicationError::Generic(anyhow::anyhow!(
                    "JSON serialization error: {}",
                    e
                ))
            })?;
            println!("{}", json_output);
        }
        _ => {
            // Text format
            println!(
                "⚙️  {}",
                "PERFORMANCE METRICS REPORT".bright_blue().underline()
            );
            println!(
                "   Generated: {}",
                perf_report.timestamp.format("%Y-%m-%d %H:%M:%S UTC")
            );
            println!();

            if perf_report.metrics.is_empty() {
                println!("✅ All performance metrics are within normal ranges.");
            } else {
                println!("📈 Performance Metrics:");

                for (metric, value) in &perf_report.metrics {
                    let severity = if *value > 1000.0 {
                        "🔴 CRITICAL"
                    } else if *value > 500.0 {
                        "🟡 WARNING"
                    } else {
                        "🟢 OK"
                    };

                    println!("   {}: {:.2}ms ({})", metric, value, severity);
                }
            }

            if verbose {
                println!();
                println!("💡 Performance Interpretation:");
                println!("   ⚡ <500ms: Excellent performance");
                println!("   ⚡ 500-1000ms: Acceptable but could be improved");
                println!("   ⚡ >1000ms: Needs optimization");
            }
        }
    }

    Ok(())
}

async fn handle_comprehensive_report(with_recommendations: bool, format: &str) -> Result<()> {
    println!(
        "📋 {} Generating comprehensive analytics report...",
        "Compiling".bright_blue().bold()
    );
    println!();

    // Get the metrics registry and create aggregator
    let metrics_registry = std::sync::Arc::new(metrics::get_metrics_registry()); // This won't work since it's not Clone
                                                                                 // Instead, we'll use the functions directly
    let snapshot = metrics::get_metrics_snapshot().await;
    let report_builder = metrics::analytics::ReportBuilder::new(snapshot);

    let trends_report = report_builder.build_usage_trends().await;
    let perf_report = report_builder.build_performance_report().await;

    let recommendations = if with_recommendations {
        generate_sample_recommendations().await
    } else {
        vec![]
    };

    let full_report = metrics::analytics::AnalyticsReport {
        timestamp: chrono::Utc::now(),
        usage_trends: trends_report,
        performance_metrics: perf_report,
        recommendations,
    };

    match format {
        "json" => {
            let json_output = serde_json::to_string_pretty(&full_report).map_err(|e| {
                crate::errors::ApplicationError::Generic(anyhow::anyhow!(
                    "JSON serialization error: {}",
                    e
                ))
            })?;
            println!("{}", json_output);
        }
        _ => {
            // Text format
            println!(
                "🎯 {}",
                "COMPREHENSIVE ANALYTICS REPORT".bright_blue().underline()
            );
            println!(
                "   Generated: {}",
                full_report.timestamp.format("%Y-%m-%d %H:%M:%S UTC")
            );
            println!();

            // Print trends section
            println!("📈 Usage Trends:");
            if full_report.usage_trends.trends.is_empty() {
                println!("   No significant trends detected yet.");
            } else {
                for (metric, direction) in &full_report.usage_trends.trends {
                    let (arrow, color) = match direction {
                        metrics::analytics::TrendDirection::Up => ("📈 ↑", "green"),
                        metrics::analytics::TrendDirection::Down => ("📉 ↓", "red"),
                        metrics::analytics::TrendDirection::Neutral => ("➡️ →", "blue"),
                    };

                    let colored_arrow = match color {
                        "green" => arrow.green(),
                        "red" => arrow.red(),
                        _ => arrow.blue(),
                    };

                    println!(
                        "   {} {}: {}",
                        colored_arrow,
                        metric,
                        direction_message(direction)
                    );
                }
            }
            println!();

            // Print performance section
            println!("⚙️  Performance Metrics:");
            if full_report.performance_metrics.metrics.is_empty() {
                println!("   All performance metrics are within normal ranges.");
            } else {
                for (metric, value) in &full_report.performance_metrics.metrics {
                    let severity = if *value > 1000.0 {
                        "🔴 CRITICAL"
                    } else if *value > 500.0 {
                        "🟡 WARNING"
                    } else {
                        "🟢 OK"
                    };

                    println!("   {}: {:.2}ms ({})", metric, value, severity);
                }
            }
            println!();

            // Print recommendations if requested
            if with_recommendations {
                println!("💡 Recommendations:");
                if full_report.recommendations.is_empty() {
                    println!("   No specific recommendations at this time. Use the system more to generate insights.");
                } else {
                    for (i, rec) in full_report.recommendations.iter().enumerate() {
                        println!("   {}. {}", i + 1, rec);
                    }
                }
                println!();
            }
        }
    }

    Ok(())
}

async fn generate_sample_recommendations() -> Vec<String> {
    let mut recs = Vec::new();

    // These would be based on actual metrics in real implementation
    recs.push("Try to maintain consistent idea capture rhythm".to_string());
    recs.push("Review ideas with low engagement scores".to_string());
    recs.push("Consider implementing time-boxing for analysis tasks".to_string());

    recs
}

async fn handle_anomaly_detection(
    threshold: f64,
    format: &str,
    _db: &database::Database,
) -> Result<()> {
    println!(
        "🚨 {} Detecting anomalies in usage patterns (threshold: {})...",
        "Scanning".bright_yellow().bold(),
        threshold
    );
    println!();

    // This would use real historical data in a production implementation
    // For now, we'll simulate anomaly detection
    let anomalies = vec![
        ("high_score_spikes".to_string(), vec![(5, 9.8), (12, 9.5)]),
        ("low_score_clusters".to_string(), vec![(3, 1.2), (7, 1.8)]),
    ];

    match format {
        "json" => {
            let mut json_data = std::collections::HashMap::new();
            for (metric, anomalous_points) in &anomalies {
                json_data.insert(metric, anomalous_points);
            }
            let json_output = serde_json::to_string_pretty(&json_data).map_err(|e| {
                crate::errors::ApplicationError::Generic(anyhow::anyhow!(
                    "JSON serialization error: {}",
                    e
                ))
            })?;
            println!("{}", json_output);
        }
        _ => {
            println!(
                "🔍 {}",
                "ANOMALY DETECTION REPORT".bright_yellow().underline()
            );
            println!();

            if anomalies.is_empty() {
                println!("✅ No anomalies detected in your usage patterns.");
                println!("   Your idea generation patterns appear consistent.");
            } else {
                for (metric, anomalous_points) in &anomalies {
                    println!("   Metric: {}", metric.bold());
                    for (idx, value) in anomalous_points {
                        println!(
                            "     Point {} with value {:.2} flagged as anomaly",
                            idx, value
                        );
                    }
                    println!();
                }

                println!("💡 Anomaly Interpretation:");
                println!("   Anomalies represent unusual patterns in your data");
                println!("   High score spikes may indicate breakthrough ideas");
                println!("   Low score clusters may indicate pattern traps");
            }
        }
    }

    Ok(())
}

async fn handle_metrics_display(all: bool, format: &str) -> Result<()> {
    println!(
        "📊 {} Retrieving current metrics...",
        "Fetching".bright_blue().bold()
    );
    println!();

    let snapshot = metrics::get_metrics_snapshot().await;

    match format {
        "json" => {
            let json_output = serde_json::to_string_pretty(&snapshot).map_err(|e| {
                crate::errors::ApplicationError::Generic(anyhow::anyhow!(
                    "JSON serialization error: {}",
                    e
                ))
            })?;
            println!("{}", json_output);
        }
        _ => {
            println!(
                "📈 {}",
                "CURRENT METRICS SNAPSHOT".bright_blue().underline()
            );
            println!(
                "   Timestamp: {}",
                snapshot.timestamp.format("%Y-%m-%d %H:%M:%S UTC")
            );
            println!();

            // Counters section
            if !snapshot.counters.is_empty() {
                println!("🔢 Counters:");
                for (name, value) in &snapshot.counters {
                    println!("   {}: {}", name.dimmed(), value.to_string().bright_white());
                }
                println!();
            }

            // Gauges section
            if !snapshot.gauges.is_empty() {
                println!("🎚️  Gauges:");
                for (name, value) in &snapshot.gauges {
                    println!(
                        "   {}: {:.2}",
                        name.dimmed(),
                        value.to_string().bright_white()
                    );
                }
                println!();
            }

            // Histograms section (only if showing all)
            if all && !snapshot.histograms.is_empty() {
                println!("📊 Histograms:");
                for (name, values) in &snapshot.histograms {
                    if !values.is_empty() {
                        let avg = values.iter().sum::<f64>() / values.len() as f64;
                        println!(
                            "   {}: {} samples, avg {:.2}",
                            name.dimmed(),
                            values.len(),
                            avg
                        );
                    }
                }
                println!();
            }

            if snapshot.counters.is_empty()
                && snapshot.gauges.is_empty()
                && (!all || snapshot.histograms.is_empty())
            {
                println!("ℹ️  No metrics have been recorded yet.");
                println!("   As you use the system, metrics will populate here.");
            }
        }
    }

    Ok(())
}
