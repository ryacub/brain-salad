use clap::{Parser, Subcommand};
use tokio::signal;
use tokio::sync::oneshot;

use crate::background_tasks::TaskManager;
use crate::commands::{analyze, analyze_llm, dump, prune, review, score};
use crate::database_simple as database;
use crate::errors::{ApplicationError, Result};
use crate::patterns_simple::PatternDetector;
use crate::scoring::ScoringEngine;

mod ai;
mod background_tasks;
mod clipboard_helper;
mod commands;
mod config;
mod database_simple;
mod display;
mod errors;
mod health;
mod implementations;
mod llm_cache;
mod llm_fallback;
mod logging;
mod metrics;
mod patterns_simple;
mod prompt_manager;
mod prompt_templates;
mod quality_metrics_simple;
mod response_processing;
mod scoring;
mod telos;
mod traits;
mod types;
mod validation;

#[derive(Parser)]
#[command(name = "tm")]
#[command(about = "Idea capture + Telos-aligned analysis")]
#[command(version = "0.1.0")]
#[command(author = "Ray Yacub")]
struct Cli {
    #[command(subcommand)]
    command: Commands,

    /// Disable AI analysis, use rule-based only
    #[arg(long, global = true)]
    #[arg(help = "Use rule-based analysis only (no AI)")]
    no_ai: bool,
}

#[derive(Debug, Subcommand)]
enum Commands {
    /// Capture an idea and get immediate analysis
    Dump {
        /// The idea text (omit for interactive input)
        idea: Option<String>,

        /// Open editor for multi-line input
        #[arg(short, long)]
        interactive: bool,

        /// Save without analysis
        #[arg(short, long)]
        quick: bool,

        /// Force use of Claude CLI instead of Ollama
        #[arg(long)]
        claude: bool,
    },

    /// Analyze an idea (from text or last captured)
    Analyze {
        /// Idea text to analyze
        idea: Option<String>,

        /// Analyze the most recently captured idea
        #[arg(long)]
        last: bool,
    },

    /// Analyze an idea using LLM with complete prompt template
    AnalyzeLlm {
        /// Idea text to analyze with LLM
        idea: String,

        /// LLM provider (openai, claude, ollama)
        #[arg(long, default_value = "ollama")]
        provider: String,

        /// Model name to use
        #[arg(long, default_value = "mistral")]
        model: String,

        /// API key for the LLM provider (if required)
        #[arg(long)]
        api_key: Option<String>,

        /// Base URL for custom LLM API
        #[arg(long)]
        base_url: Option<String>,

        /// Save the analysis result to the database
        #[arg(long)]
        save: bool,

        /// Temperature for the LLM (0.0 to 1.0)
        #[arg(long, default_value = "0.3")]
        temperature: f32,

        /// Maximum tokens for the response
        #[arg(long, default_value = "4096")]
        max_tokens: u32,
    },

    /// Quick score an idea without saving
    Score {
        /// Idea text to score
        idea: String,
    },

    /// Review and browse captured ideas
    Review {
        /// Limit number of ideas to show
        #[arg(short, long, default_value = "10")]
        limit: usize,

        /// Filter by minimum score
        #[arg(long, default_value = "0.0")]
        min_score: f64,

        /// Show ideas needing pruning review
        #[arg(long)]
        pruning: bool,
    },

    /// Manage old ideas (archive/delete)
    Prune {
        /// Auto-prune without confirmation
        #[arg(long)]
        auto: bool,

        /// Show what would be pruned (dry run)
        #[arg(long)]
        dry_run: bool,
    },

    /// Bulk operations on multiple ideas
    Bulk {
        #[command(subcommand)]
        command: crate::commands::bulk::BulkCommands,
    },

    /// Analytics and reporting
    Analytics {
        #[command(subcommand)]
        command: crate::commands::analytics::AnalyticsCommands,
    },

    /// Link and manage idea relationships
    Link {
        #[command(subcommand)]
        command: crate::commands::link::LinkCommands,
    },

    /// Health check and monitoring
    Health {
        /// Output format (json or text)
        #[arg(long, default_value = "text", value_parser = ["json", "text"])]
        format: String,
    },

    /// Manage Ollama LLM service
    Llm {
        #[command(subcommand)]
        action: LlmAction,
    },
}

#[derive(Debug, Subcommand)]
enum LlmAction {
    /// Show Ollama status and available models
    Status,
    /// Start Ollama service
    Start,
    /// Stop Ollama service
    Stop,
}

#[tokio::main(flavor = "multi_thread", worker_threads = 4)]
async fn main() -> Result<()> {
    // Initialize structured logging first
    let log_config = logging::LoggingConfig {
        level: tracing::Level::INFO,
        json_format: std::env::var("TELOS_LOG_JSON").is_ok(),
        include_timestamps: true,
        display_target: false,
        log_directory: std::env::var("TELOS_LOG_DIR").ok().map(|s| s.into()),
        max_log_file_size: Some(10 * 1024 * 1024), // 10MB
        log_file_retention: Some(5),
    };

    if let Err(e) = logging::init_logging(log_config) {
        eprintln!("Failed to initialize logging: {}", e);
        return Err(ApplicationError::Configuration(format!(
            "Logging initialization failed: {}",
            e
        )));
    }

    let cli = Cli::parse();

    // Initialize the prompt manager
    crate::prompt_manager::initialize_prompt_manager("./IDEA_ANALYSIS_PROMPT.md")
        .await
        .map_err(|e| {
            tracing::error!(error = %e, "Failed to initialize prompt manager");
            crate::errors::ApplicationError::Configuration(format!(
                "Prompt manager initialization failed: {}",
                e
            ))
        })?;

    // Log application startup
    tracing::info!(
        app_name = "telos-matrix",
        version = env!("CARGO_PKG_VERSION"),
        command = ?cli.command,
        no_ai = cli.no_ai,
        "Application starting"
    );

    // Set up graceful shutdown handling
    let (_shutdown_tx, mut shutdown_rx) = oneshot::channel::<()>();

    // Spawn signal handler for graceful shutdown
    let signal_handler = tokio::spawn(async move {
        #[cfg(unix)]
        {
            let mut sigterm = signal::unix::signal(signal::unix::SignalKind::terminate())
                .expect("Failed to setup SIGTERM handler");
            let mut sigint = signal::unix::signal(signal::unix::SignalKind::interrupt())
                .expect("Failed to setup SIGINT handler");

            tokio::select! {
                _ = sigterm.recv() => {
                    println!("\n🛑 Received SIGTERM, initiating graceful shutdown...");
                }
                _ = sigint.recv() => {
                    println!("\n🛑 Received Ctrl+C, initiating graceful shutdown...");
                }
            }
        }

        #[cfg(not(unix))]
        {
            let mut sigint = signal::windows::ctrl_c().expect("Failed to setup Ctrl+C handler");

            let _ = sigint.recv().await;
            println!("\n🛑 Received Ctrl+C, initiating graceful shutdown...");
        }
    });

    // Initialize database with health check
    let db_timer = logging::OperationTimer::new("database_initialization");
    let db = database::Database::new().await.map_err(|e| {
        tracing::error!(error = %e, "Database initialization failed");
        e
    })?;

    db.health_check().await.map_err(|e| {
        tracing::error!(error = %e, "Database health check failed");
        ApplicationError::Database(crate::errors::DatabaseError::Connection {
            source: sqlx::Error::Protocol("Health check failed".to_string()),
            database_path: None,
        })
    })?;
    db_timer.complete();

    // Load configuration
    let config_timer = logging::OperationTimer::new("configuration_loading");
    let config = crate::config::ConfigPaths::load().map_err(|e| {
        tracing::error!(error = %e, "Configuration loading failed");
        crate::errors::ApplicationError::Configuration(format!(
            "Configuration loading failed: {}",
            e
        ))
    })?;
    config.ensure_directories_exist().map_err(|e| {
        tracing::error!(error = %e, "Failed to create configuration directories");
        crate::errors::ApplicationError::Configuration(format!("Directory creation failed: {}", e))
    })?;
    config_timer.complete();

    // Initialize analysis components
    let scoring_timer = logging::OperationTimer::new("scoring_engine_initialization");
    let scoring_engine = ScoringEngine::with_config(&config).await.map_err(|e| {
        tracing::error!(error = %e, "Scoring engine initialization failed");
        e
    })?;
    let pattern_detector = PatternDetector::new();
    scoring_timer.complete();

    // Initialize task manager for background operations
    let mut task_manager = TaskManager::new();

    // Initialize health checks
    {
        let mut health_monitor = health::get_health_monitor_mut().await;
        health_monitor.add_check(Box::new(health::MemoryHealthChecker));
        health_monitor.add_check(Box::new(health::DiskSpaceHealthChecker));
    }

    // Resources used directly without Arc wrapping for single-threaded execution

    // Handle commands with cancellation support
    let correlation_id = logging::generate_correlation_id();
    let request_span =
        logging::create_request_span("command_execution", Some(&correlation_id), None);

    let command_timer = logging::OperationTimer::new_with_fields(
        "command_execution",
        &[("command", &format!("{:?}", cli.command))],
    );

    let command_result = tokio::select! {
        result = async {
            match cli.command {
                Commands::Dump { idea, interactive, quick, claude } => {
                    // Always use LLM unless --no-ai flag is set
                    let use_llm = !cli.no_ai;

                    tracing::info!(
                        command = "dump",
                        idea_provided = idea.is_some(),
                        interactive = interactive,
                        quick = quick,
                        claude = claude,
                        use_llm = use_llm,
                        "Executing dump command"
                    );
                    dump::handle_dump(
                        idea,
                        interactive,
                        quick,
                        claude,
                        &db,
                        &scoring_engine,
                        &pattern_detector,
                        use_llm,
                    ).await
                }
                Commands::Analyze { idea, last } => {
                    tracing::info!(
                        command = "analyze",
                        idea_provided = idea.is_some(),
                        analyze_last = last,
                        "Executing analyze command"
                    );
                    analyze::handle_analyze(
                        idea,
                        last,
                        &db,
                        &scoring_engine,
                        &pattern_detector,
                        !cli.no_ai
                    ).await
                }
                Commands::AnalyzeLlm { idea, provider, model, api_key, base_url, save, temperature, max_tokens } => {
                    tracing::info!(
                        command = "analyze-llm",
                        provider = provider,
                        model = model,
                        save_to_db = save,
                        "Executing analyze-llm command"
                    );

                    // Create LLM config from command arguments
                    let llm_provider = match provider.as_str() {
                        "openai" => analyze_llm::LlmProvider::OpenAi,
                        "claude" => analyze_llm::LlmProvider::Claude,
                        "ollama" => analyze_llm::LlmProvider::Ollama,
                        _ => analyze_llm::LlmProvider::Custom,
                    };

                    let config = analyze_llm::LlmConfig {
                        provider: llm_provider,
                        model,
                        api_key,
                        base_url,
                        temperature,
                        max_tokens,
                        timeout_seconds: 60, // Default timeout
                    };

                    analyze_llm::handle_analyze_llm(
                        idea,
                        config,
                        &db,
                        &scoring_engine,
                        &pattern_detector,
                        save
                    ).await
                }
                Commands::Score { idea } => {
                    tracing::info!(
                        command = "score",
                        idea_length = idea.len(),
                        "Executing score command"
                    );
                    score::handle_score(&idea, &scoring_engine, &pattern_detector).await
                }
                Commands::Review { limit, min_score, pruning } => {
                    tracing::info!(
                        command = "review",
                        limit = limit,
                        min_score = min_score,
                        pruning = pruning,
                        "Executing review command"
                    );
                    review::handle_review(limit, min_score, pruning, &db).await
                }
                Commands::Prune { auto, dry_run } => {
                    tracing::info!(
                        command = "prune",
                        auto = auto,
                        dry_run = dry_run,
                        "Executing prune command"
                    );
                    prune::handle_prune(auto, dry_run, &db).await
                }
                Commands::Health { format } => {
                    tracing::info!(
                        command = "health",
                        format = format,
                        "Executing health check command"
                    );
                    health::handle_health_check(&format).await
                }
                Commands::Bulk { command } => {
                    tracing::info!(
                        command = "bulk",
                        subcommand = ?command,
                        "Executing bulk operation command"
                    );
                    commands::bulk::handle_bulk(command, &db).await
                }
                Commands::Analytics { command } => {
                    tracing::info!(
                        command = "analytics",
                        subcommand = ?command,
                        "Executing analytics command"
                    );
                    commands::analytics::handle_analytics(command, &db).await
                }
                Commands::Link { command } => {
                    tracing::info!(
                        command = "link",
                        subcommand = ?command,
                        "Executing link operation command"
                    );
                    commands::link::handle_link(command, &db).await
                }
                Commands::Llm { action } => {
                    tracing::info!(
                        command = "llm",
                        subcommand = ?action,
                        "Executing LLM management command"
                    );
                    match action {
                        LlmAction::Status => {
                            commands::llm::handle_llm_status().await?;
                        }
                        LlmAction::Start => {
                            commands::llm::handle_llm_start().await?;
                        }
                        LlmAction::Stop => {
                            commands::llm::handle_llm_stop().await?;
                        }
                    }
                    Ok(())
                }
            }
        } => result,
        _ = &mut shutdown_rx => {
            tracing::warn!(
                correlation_id = %correlation_id,
                "Operation cancelled due to shutdown signal"
            );
            println!("⚠️  Operation cancelled due to shutdown signal");
            Err(ApplicationError::operation_cancelled("Command execution"))
        }
        _ = signal_handler => {
            tracing::warn!(
                correlation_id = %correlation_id,
                "Operation cancelled due to signal"
            );
            println!("⚠️  Operation cancelled due to signal");
            Err(ApplicationError::operation_cancelled("Signal received"))
        }
    };

    match &command_result {
        Ok(_) => {
            command_timer.complete();
            tracing::info!(
                correlation_id = %correlation_id,
                "Command completed successfully"
            );
        }
        Err(e) => {
            command_timer.error(e);
            tracing::error!(
                correlation_id = %correlation_id,
                error = %e,
                "Command failed"
            );
        }
    }

    drop(request_span);

    // Perform graceful shutdown
    println!("🔄 Shutting down gracefully...");
    tracing::info!("Starting graceful shutdown");

    let shutdown_timer = logging::OperationTimer::new("graceful_shutdown");

    // Shutdown background tasks
    if let Err(e) = task_manager.shutdown().await {
        eprintln!("⚠️  Error during task manager shutdown: {}", e);
        tracing::warn!(error = %e, "Task manager shutdown encountered error");
    } else {
        tracing::info!("Task manager shutdown completed successfully");
    }

    // Close database connections
    if let Err(e) = db.close().await {
        eprintln!("⚠️  Error during database shutdown: {}", e);
        tracing::warn!(error = %e, "Database shutdown encountered error");
    } else {
        tracing::info!("Database shutdown completed successfully");
    }

    // Return the command result or shutdown status
    match command_result {
        Ok(_) => {
            shutdown_timer.complete();
            println!("✅ Shutdown completed successfully");
            tracing::info!("Application shutdown completed successfully");
            Ok(())
        }
        Err(e) => {
            // Check if it's a cancellation error (expected during shutdown)
            if matches!(e, ApplicationError::OperationCancelled { .. }) {
                shutdown_timer.complete();
                println!("✅ Graceful shutdown completed");
                tracing::info!("Graceful shutdown completed (cancelled operation)");
                Ok(())
            } else {
                shutdown_timer.error(&e);
                eprintln!("❌ Command failed: {}", e);
                tracing::error!(
                    error = %e,
                    "Application shutdown with errors"
                );
                Err(e)
            }
        }
    }
}
