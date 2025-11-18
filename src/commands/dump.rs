use crate::errors::Result;
use std::io::Write;
use tokio::io::{AsyncBufReadExt, AsyncReadExt, BufReader};

use crate::clipboard_helper;
use crate::database_simple as database;
use crate::display::display_analysis_result;
use crate::patterns_simple::PatternDetector;
use crate::scoring::ScoringEngine;

pub async fn handle_dump(
    idea: Option<String>,
    interactive: bool,
    quick: bool,
    force_claude: bool,
    db: &database::Database,
    scoring_engine: &ScoringEngine,
    pattern_detector: &PatternDetector,
    use_ai: bool,
) -> Result<()> {
    // If an idea is provided directly (not through interactive mode), process it normally
    if idea.is_some() {
        let content = {
            let idea = idea.unwrap();
            if idea.trim().is_empty() {
                return Err(crate::errors::ApplicationError::validation(
                    "Idea content cannot be empty",
                ));
            }
            idea
        };

        if quick {
            let idea_id = db.save_idea(&content, None, None, None, None, None).await?;
            println!("✅ Idea saved (ID: {})", idea_id);
            return Ok(());
        }

        let (score, patterns) = if use_ai {
            if force_claude {
                println!("✨ Using Claude CLI (forced)...");
                try_claude_cli_only(&content).await?
            } else {
                println!("🔍 Starting LLM analysis...");
                try_ollama_with_autostart(&content).await?
            }
        } else {
            println!("📊 Using rule-based analysis (--no-ai)");
            perform_rule_based_analysis(&content, scoring_engine, pattern_detector).await?
        };

        // Save with analysis and timeout
        let db_timeout = tokio::time::Duration::from_secs(10);

        let idea_id = tokio::time::timeout(
            db_timeout,
            db.save_idea(
                &content,
                Some(score.raw_score),
                Some(score.final_score),
                Some(
                    patterns
                        .iter()
                        .map(|p| format!("{}: {}", p.pattern_type.emoji(), p.pattern_type.title()))
                        .collect(),
                ),
                Some(format!(
                    "{} {}",
                    score.recommendation.emoji(),
                    score.recommendation.text()
                )),
                Some(serde_json::to_string(&score).map_err(|e| {
                    crate::errors::ApplicationError::Scoring(
                        crate::errors::ScoringError::ScoreSerialization { source: e },
                    )
                })?),
            ),
        )
        .await
        .map_err(|_| {
            crate::errors::ApplicationError::operation_timeout(
                db_timeout.as_millis() as u64,
                "Database save operation",
            )
        })??;

        // Display results
        display_analysis_result(&content, &score, &patterns, Some(&idea_id));

        // Prompt for next actions
        prompt_next_actions(&idea_id, db).await?;

        Ok(())
    } else {
        // If no idea provided, run in interactive loop mode
        run_interactive_loop(
            interactive,
            quick,
            force_claude,
            db,
            scoring_engine,
            pattern_detector,
            use_ai,
        )
        .await
    }
}

async fn get_interactive_input() -> Result<String> {
    // Check clipboard first
    if let Ok(Some(clipboard_text)) = clipboard_helper::maybe_use_clipboard() {
        println!("✅ Using clipboard content");
        return Ok(clipboard_text);
    }

    // Normal multi-line input
    println!("📝 Enter your idea (Ctrl+D on Unix, Ctrl+Z on Windows to finish):");
    println!();

    let mut input = String::new();
    println!("> ");
    std::io::stdout().flush()?;

    let mut stdin = BufReader::new(tokio::io::stdin());

    stdin
        .read_to_string(&mut input)
        .await
        .map_err(crate::errors::ApplicationError::Io)?;

    let trimmed = input.trim().to_string();

    if trimmed.is_empty() {
        return Err(crate::errors::ApplicationError::validation(
            "No input provided",
        ));
    }

    Ok(trimmed)
}

async fn prompt_single_line() -> Result<String> {
    // Check clipboard first
    if let Ok(Some(clipboard_text)) = clipboard_helper::maybe_use_clipboard() {
        println!("✅ Using clipboard content");
        return Ok(clipboard_text);
    }

    // Normal single-line prompt
    print!("💡 Enter your idea: ");
    std::io::stdout().flush()?;

    let mut input = String::new();
    let mut stdin = BufReader::new(tokio::io::stdin());

    stdin
        .read_line(&mut input)
        .await
        .map_err(crate::errors::ApplicationError::Io)?;

    let trimmed = input.trim().to_string();

    if trimmed.is_empty() {
        return Err(crate::errors::ApplicationError::validation(
            "No input provided",
        ));
    }

    Ok(trimmed)
}

async fn run_interactive_loop(
    interactive: bool,
    quick: bool,
    force_claude: bool,
    db: &database::Database,
    scoring_engine: &ScoringEngine,
    pattern_detector: &PatternDetector,
    use_ai: bool,
) -> Result<()> {
    loop {
        let content = if interactive {
            match get_interactive_input().await {
                Ok(content) => content,
                Err(e) => {
                    println!("Error getting input: {}", e);
                    break;
                }
            }
        } else {
            match prompt_single_line().await {
                Ok(content) => content,
                Err(e) => {
                    println!("Error getting input: {}", e);
                    break;
                }
            }
        };

        if quick {
            let idea_id = db.save_idea(&content, None, None, None, None, None).await?;
            println!("✅ Idea saved (ID: {})", idea_id);
        } else {
            let (score, patterns) = if use_ai {
                if force_claude {
                    println!("✨ Using Claude CLI (forced)...");
                    match try_claude_cli_only(&content).await {
                        Ok(result) => result,
                        Err(e) => {
                            println!("Claude failed, using rule-based: {}", e);
                            perform_rule_based_analysis(&content, scoring_engine, pattern_detector)
                                .await?
                        }
                    }
                } else {
                    println!("🔍 Starting LLM analysis...");
                    match try_ollama_with_autostart(&content).await {
                        Ok(result) => result,
                        Err(e) => {
                            println!("LLM failed, using rule-based: {}", e);
                            perform_rule_based_analysis(&content, scoring_engine, pattern_detector)
                                .await?
                        }
                    }
                }
            } else {
                println!("📊 Using rule-based analysis (--no-ai)");
                perform_rule_based_analysis(&content, scoring_engine, pattern_detector).await?
            };

            // Save with analysis and timeout
            let db_timeout = tokio::time::Duration::from_secs(10);

            let idea_id = tokio::time::timeout(
                db_timeout,
                db.save_idea(
                    &content,
                    Some(score.raw_score),
                    Some(score.final_score),
                    Some(
                        patterns
                            .iter()
                            .map(|p| {
                                format!("{}: {}", p.pattern_type.emoji(), p.pattern_type.title())
                            })
                            .collect(),
                    ),
                    Some(format!(
                        "{} {}",
                        score.recommendation.emoji(),
                        score.recommendation.text()
                    )),
                    Some(serde_json::to_string(&score).map_err(|e| {
                        crate::errors::ApplicationError::Scoring(
                            crate::errors::ScoringError::ScoreSerialization { source: e },
                        )
                    })?),
                ),
            )
            .await
            .map_err(|_| {
                crate::errors::ApplicationError::operation_timeout(
                    db_timeout.as_millis() as u64,
                    "Database save operation",
                )
            })??;

            // Display results
            display_analysis_result(&content, &score, &patterns, Some(&idea_id));
        }

        // Prompt for next actions
        let action = prompt_next_actions_interactive(db).await?;

        match action {
            NextAction::AddAnother => {
                // Continue the loop to add another idea
                continue;
            }
            NextAction::Review => {
                super::review::handle_review(10, 0.0, false, db).await?;
            }
            NextAction::DetailedAnalysis => match db.get_last_idea().await? {
                Some(idea) => {
                    super::analyze::handle_analyze(
                        Some(idea.content.clone()),
                        false,
                        db,
                        &crate::scoring::ScoringEngine::new().await?,
                        &PatternDetector::new(),
                        false,
                    )
                    .await?;
                }
                None => {
                    println!("❌ No idea found to analyze");
                }
            },
            NextAction::SetPriority => {
                println!("✅ Marked as priority (Note: priority tagging coming soon)");
            }
            NextAction::Exit => {
                println!("👋 Happy idea prioritizing!");
                break;
            }
        }
    }

    Ok(())
}

enum NextAction {
    AddAnother,
    Review,
    DetailedAnalysis,
    SetPriority,
    Exit,
}

async fn prompt_next_actions_interactive(_db: &database::Database) -> Result<NextAction> {
    use dialoguer::Select;

    // Check if stdin is connected to a terminal
    if atty::is(atty::Stream::Stdin) {
        println!();
        println!("🤔 What would you like to do next?");

        let options = vec![
            "Add another idea",
            "Review all ideas",
            "Get detailed analysis of this idea",
            "Set this idea as priority",
            "Exit",
        ];

        let selection = Select::new()
            .with_prompt("Choose an action")
            .items(&options)
            .interact()
            .map_err(|e| {
                crate::errors::ApplicationError::Generic(anyhow::anyhow!("Dialog error: {}", e))
            })?;

        match selection {
            0 => Ok(NextAction::AddAnother),
            1 => Ok(NextAction::Review),
            2 => Ok(NextAction::DetailedAnalysis),
            3 => Ok(NextAction::SetPriority),
            4 => Ok(NextAction::Exit),
            _ => unreachable!(),
        }
    } else {
        // If not in interactive mode, default to Exit to avoid hanging
        println!();
        println!("💡 Input was piped - exiting automatically");
        Ok(NextAction::Exit)
    }
}

async fn prompt_next_actions(_idea_id: &str, db: &database::Database) -> Result<()> {
    // Check if stdin is connected to a terminal
    if atty::is(atty::Stream::Stdin) {
        use dialoguer::Select;

        println!();
        println!("🤔 What would you like to do next?");

        let options = vec![
            "Add another idea",
            "Review all ideas",
            "Get detailed analysis of this idea",
            "Set this idea as priority",
            "Exit",
        ];

        let selection = Select::new()
            .with_prompt("Choose an action")
            .items(&options)
            .interact()
            .map_err(|e| {
                crate::errors::ApplicationError::Generic(anyhow::anyhow!("Dialog error: {}", e))
            })?;

        match selection {
            0 => {
                // Add another idea - just return to main prompt
                println!();
                println!("Ready for next idea...");
            }
            1 => {
                // Review all ideas
                println!();
                super::review::handle_review(10, 0.0, false, db).await?;
            }
            2 => {
                // Detailed analysis
                println!();
                match db.get_last_idea().await? {
                    Some(idea) => {
                        super::analyze::handle_analyze(
                            Some(idea.content.clone()),
                            false,
                            db,
                            &crate::scoring::ScoringEngine::new().await?,
                            &PatternDetector::new(),
                            false,
                        )
                        .await?;
                    }
                    None => {
                        println!("❌ No idea found to analyze");
                    }
                }
            }
            3 => {
                // Set as priority
                println!("✅ Marked as priority (Note: priority tagging coming soon)");
            }
            4 => {
                // Exit
                println!("👋 Goodbye!");
            }
            _ => unreachable!(),
        }
    } else {
        // If not in interactive mode, just notify and continue
        println!("💡 Input was piped - completing analysis automatically");
    }

    Ok(())
}

async fn try_llm_analysis_with_fallback(
    content: &str,
) -> Result<(
    crate::scoring::Score,
    Vec<crate::patterns_simple::PatternMatch>,
)> {
    if crate::llm_fallback::is_ollama_running().await {
        println!("Using Ollama (local)...");

        match try_ollama_analysis(content).await {
            Ok(result) => {
                println!("Ollama analysis complete");
                return Ok(result);
            }
            Err(e) => {
                println!("Ollama failed: {}", e);
                println!("Trying Claude CLI...");
            }
        }
    } else {
        println!("Ollama not running");
    }

    if crate::llm_fallback::is_claude_cli_available().await {
        println!("Using Claude CLI...");

        match try_claude_cli_analysis(content).await {
            Ok(result) => {
                println!("Claude CLI analysis complete");
                return Ok(result);
            }
            Err(e) => {
                println!("Claude CLI failed: {}", e);
                println!("Falling back to rule-based...");
            }
        }
    } else {
        println!("Claude CLI not available");
    }

    println!("Using rule-based analysis (no LLM available)");
    let scoring_engine = crate::scoring::ScoringEngine::new().await?;
    let pattern_detector = crate::patterns_simple::PatternDetector::new();
    perform_rule_based_analysis(content, &scoring_engine, &pattern_detector).await
}

async fn try_ollama_analysis(
    content: &str,
) -> Result<(
    crate::scoring::Score,
    Vec<crate::patterns_simple::PatternMatch>,
)> {
    use crate::ai::AiAnalyzer;
    use crate::telos::TelosParser;

    let ai_analyzer = AiAnalyzer::new();
    let scoring_engine = crate::scoring::ScoringEngine::new().await?;
    let base_score = scoring_engine.calculate_score(content)?;

    // Parse telos context for AI enhancement
    let telos_parser = match TelosParser::new() {
        Ok(parser) => parser,
        Err(_) => {
            // If config loading failed, create a parser with a dummy path to avoid breaking
            TelosParser::with_path("dummy.md")
        }
    };
    let telos_context = telos_parser.parse().await?;

    let _enhancement = ai_analyzer
        .enhance_analysis(content, &base_score, &telos_context)
        .await?;

    let pattern_detector = crate::patterns_simple::PatternDetector::new();
    let patterns = pattern_detector.detect_patterns(content);

    Ok((base_score, patterns))
}

async fn try_claude_cli_analysis(
    content: &str,
) -> Result<(
    crate::scoring::Score,
    Vec<crate::patterns_simple::PatternMatch>,
)> {
    use crate::llm_fallback;
    use crate::prompt_manager::get_prompt_manager;

    let prompt_manager = get_prompt_manager().await.map_err(|e| {
        crate::errors::ApplicationError::Configuration(format!("Failed to get prompt: {}", e))
    })?;

    let prompt_template = prompt_manager.get_analysis_prompt_str();

    let response = llm_fallback::analyze_with_claude_cli(content, prompt_template).await?;

    if let Some(json) = llm_fallback::extract_json(&response) {
        return parse_llm_json_response(&json);
    }

    println!("Claude returned plain text, attempting to parse...");
    parse_llm_plain_text_response(&response)
}

fn parse_llm_json_response(
    json: &serde_json::Value,
) -> Result<(
    crate::scoring::Score,
    Vec<crate::patterns_simple::PatternMatch>,
)> {
    use crate::patterns_simple::{PatternMatch, PatternType, Severity};
    use crate::scoring::{
        AntiChallengeScores, MissionScores, Recommendation, Score, StrategicScores,
    };
    use std::collections::HashMap;

    let final_score = json["final_score"].as_f64().ok_or_else(|| {
        crate::errors::ApplicationError::Generic(anyhow::anyhow!("Missing final_score in JSON"))
    })?;

    let rec_str = json["recommendation"].as_str().unwrap_or("Consider");
    let recommendation = match rec_str {
        "Priority" => Recommendation::Priority,
        "Good" => Recommendation::Good,
        "Consider" => Recommendation::Consider,
        "Avoid" => Recommendation::Avoid,
        _ => Recommendation::Consider,
    };

    let reasoning = if let Some(explanations) = json["explanations"].as_object() {
        explanations
            .iter()
            .map(|(k, v)| format!("{}: {}", k, v.as_str().unwrap_or("")))
            .collect::<Vec<_>>()
            .join("\n")
    } else {
        "LLM analysis completed".to_string()
    };

    // Create a Score struct with proper structure
    let score = Score {
        mission: MissionScores {
            domain_expertise: 0.0,
            ai_alignment: 0.0,
            execution_support: 0.0,
            revenue_potential: 0.0,
            total: final_score * 0.4,
        },
        anti_challenge: AntiChallengeScores {
            context_switching: 0.0,
            rapid_prototyping: 0.0,
            accountability: 0.0,
            income_anxiety: 0.0,
            total: final_score * 0.35,
        },
        strategic: StrategicScores {
            stack_compatibility: 0.0,
            shipping_habit: 0.0,
            public_accountability: 0.0,
            revenue_testing: 0.0,
            total: final_score * 0.25,
        },
        raw_score: final_score,
        final_score,
        recommendation,
        scoring_details: vec![format!("LLM Score: {:.1}/10", final_score)],
        explanations: HashMap::new(),
    };

    let mut patterns = Vec::new();

    if let Some(scores_obj) = json["scores"].as_object() {
        if let Some(anti) = scores_obj
            .get("Anti-Challenge Patterns")
            .and_then(|v| v.as_object())
        {
            if let Some(ctx_switch) = anti.get("Avoid Context-Switching").and_then(|v| v.as_f64()) {
                if ctx_switch < 0.6 {
                    patterns.push(PatternMatch {
                        pattern_type: PatternType::ContextSwitching,
                        severity: if ctx_switch < 0.3 {
                            Severity::High
                        } else {
                            Severity::Medium
                        },
                        matches: vec!["Low stack continuity score".to_string()],
                        message: "Context-switching risk detected".to_string(),
                        suggestion: Some("Consider using current tech stack".to_string()),
                    });
                }
            }

            if let Some(proto) = anti.get("Rapid Prototyping").and_then(|v| v.as_f64()) {
                if proto < 0.5 {
                    patterns.push(PatternMatch {
                        pattern_type: PatternType::Perfectionism,
                        severity: Severity::Medium,
                        matches: vec!["Low rapid prototyping score".to_string()],
                        message: "Perfectionism risk detected".to_string(),
                        suggestion: Some("Focus on MVP first".to_string()),
                    });
                }
            }
        }
    }

    Ok((score, patterns))
}

fn parse_llm_plain_text_response(
    text: &str,
) -> Result<(
    crate::scoring::Score,
    Vec<crate::patterns_simple::PatternMatch>,
)> {
    use crate::llm_fallback;
    use crate::scoring::{
        AntiChallengeScores, MissionScores, Recommendation, Score, StrategicScores,
    };
    use std::collections::HashMap;

    let (final_score, rec_str) =
        llm_fallback::parse_plain_text_response(text).unwrap_or((5.0, "Consider".to_string()));

    let recommendation = match rec_str.as_str() {
        "Priority" => Recommendation::Priority,
        "Good" => Recommendation::Good,
        "Consider" => Recommendation::Consider,
        "Avoid" => Recommendation::Avoid,
        _ => Recommendation::Consider,
    };

    let score = Score {
        mission: MissionScores {
            domain_expertise: 0.0,
            ai_alignment: 0.0,
            execution_support: 0.0,
            revenue_potential: 0.0,
            total: final_score * 0.4,
        },
        anti_challenge: AntiChallengeScores {
            context_switching: 0.0,
            rapid_prototyping: 0.0,
            accountability: 0.0,
            income_anxiety: 0.0,
            total: final_score * 0.35,
        },
        strategic: StrategicScores {
            stack_compatibility: 0.0,
            shipping_habit: 0.0,
            public_accountability: 0.0,
            revenue_testing: 0.0,
            total: final_score * 0.25,
        },
        raw_score: final_score,
        final_score,
        recommendation,
        scoring_details: vec![format!("LLM Score: {:.1}/10 (from text)", final_score)],
        explanations: HashMap::new(),
    };

    let patterns = Vec::new();

    Ok((score, patterns))
}

async fn perform_rule_based_analysis(
    content: &str,
    scoring_engine: &crate::scoring::ScoringEngine,
    pattern_detector: &crate::patterns_simple::PatternDetector,
) -> Result<(
    crate::scoring::Score,
    Vec<crate::patterns_simple::PatternMatch>,
)> {
    let (score_result, pattern_result) = tokio::join!(
        async {
            let scoring_engine_clone = scoring_engine.clone();
            let content_clone = content.to_string();
            tokio::task::spawn_blocking(move || {
                scoring_engine_clone.calculate_score(&content_clone)
            })
            .await
            .map_err(|e| {
                crate::errors::ApplicationError::Generic(anyhow::anyhow!(
                    "Scoring task failed: {}",
                    e
                ))
            })?
        },
        async {
            let pattern_detector_clone = pattern_detector.clone();
            let content_clone = content.to_string();
            tokio::task::spawn_blocking(move || {
                pattern_detector_clone.detect_patterns(&content_clone)
            })
            .await
            .map_err(|e| {
                crate::errors::ApplicationError::Generic(anyhow::anyhow!(
                    "Pattern detection task failed: {}",
                    e
                ))
            })
        }
    );

    Ok((score_result?, pattern_result?))
}

/// Try Ollama with auto-start if not running
async fn try_ollama_with_autostart(
    content: &str,
) -> Result<(
    crate::scoring::Score,
    Vec<crate::patterns_simple::PatternMatch>,
)> {
    use crate::llm_fallback;

    // Check if Ollama is running
    if !llm_fallback::is_ollama_running().await {
        println!("ℹ️  Ollama not running, starting it...");

        // Try to auto-start
        match llm_fallback::start_ollama().await {
            Ok(_) => {
                println!("🤖 Using Ollama...");
            }
            Err(e) => {
                println!("⚠️  Failed to start Ollama: {}", e);
                println!("   Falling back to rule-based analysis");
                let scoring_engine = crate::scoring::ScoringEngine::new().await?;
                let pattern_detector = crate::patterns_simple::PatternDetector::new();
                return perform_rule_based_analysis(content, &scoring_engine, &pattern_detector)
                    .await;
            }
        }
    } else {
        println!("🤖 Using Ollama...");
    }

    // Use Ollama for analysis
    match try_ollama_analysis(content).await {
        Ok(result) => {
            println!("✅ Analysis complete");
            Ok(result)
        }
        Err(e) => {
            println!("⚠️  Ollama analysis failed: {}", e);
            println!("   Falling back to rule-based");
            let scoring_engine = crate::scoring::ScoringEngine::new().await?;
            let pattern_detector = crate::patterns_simple::PatternDetector::new();
            perform_rule_based_analysis(content, &scoring_engine, &pattern_detector).await
        }
    }
}

/// Force Claude CLI only (skip Ollama entirely)
async fn try_claude_cli_only(
    content: &str,
) -> Result<(
    crate::scoring::Score,
    Vec<crate::patterns_simple::PatternMatch>,
)> {
    use crate::llm_fallback;

    // Check if Claude CLI is available
    if !llm_fallback::is_claude_cli_available().await {
        println!("❌ Claude CLI not found");
        println!("   Falling back to rule-based");
        let scoring_engine = crate::scoring::ScoringEngine::new().await?;
        let pattern_detector = crate::patterns_simple::PatternDetector::new();
        return perform_rule_based_analysis(content, &scoring_engine, &pattern_detector).await;
    }

    // Use Claude CLI
    match try_claude_cli_analysis(content).await {
        Ok(result) => {
            println!("✅ Claude analysis complete");
            Ok(result)
        }
        Err(e) => {
            println!("⚠️  Claude CLI failed: {}", e);
            println!("   Falling back to rule-based");
            let scoring_engine = crate::scoring::ScoringEngine::new().await?;
            let pattern_detector = crate::patterns_simple::PatternDetector::new();
            perform_rule_based_analysis(content, &scoring_engine, &pattern_detector).await
        }
    }
}
