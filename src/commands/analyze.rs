use crate::errors::Result;

use crate::database_simple as database;
use crate::display::display_analysis_result;
use crate::patterns_simple::PatternDetector;
use crate::scoring::ScoringEngine;

pub async fn handle_analyze(
    idea: Option<String>,
    last: bool,
    db: &database::Database,
    scoring_engine: &ScoringEngine,
    pattern_detector: &PatternDetector,
    use_ai: bool,
) -> Result<()> {
    let content = match idea {
        Some(idea) => {
            if idea.trim().is_empty() {
                return Err(crate::errors::ApplicationError::validation(
                    "Idea content cannot be empty",
                ));
            }
            idea
        }
        None => {
            if last {
                match db.get_last_idea().await? {
                    Some(idea) => {
                        println!("🔍 Analyzing last captured idea...");
                        idea.content
                    }
                    None => {
                        return Err(crate::errors::ApplicationError::validation(
                            "No ideas found to analyze",
                        ));
                    }
                }
            } else {
                return Err(crate::errors::ApplicationError::validation(
                    "Please provide an idea to analyze or use --last flag",
                ));
            }
        }
    };

    // Perform detailed analysis concurrently
    let content_clone1 = content.clone();
    let content_clone2 = content.clone();
    let scoring_engine_clone = scoring_engine.clone();
    let pattern_detector_clone = pattern_detector.clone();

    let (score, patterns) = tokio::join!(
        async move { scoring_engine_clone.calculate_score(&content_clone1) },
        async move { pattern_detector_clone.detect_patterns(&content_clone2) }
    );

    let score = score?;

    // Display detailed analysis
    display_analysis_result(&content, &score, &patterns, None);

    // If using AI, enhance analysis (future feature)
    if use_ai {
        // TODO: Add AI enhancement when implemented
        println!("\n🤖 AI enhancement would be added here in future version");
    }

    Ok(())
}
