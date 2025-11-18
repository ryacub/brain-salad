use crate::errors::Result;
use colored::Colorize;

use crate::display::display_analysis_result;
use crate::patterns_simple::PatternDetector;
use crate::scoring::ScoringEngine;

pub async fn handle_score(
    idea: &str,
    scoring_engine: &ScoringEngine,
    pattern_detector: &PatternDetector,
) -> Result<()> {
    if idea.trim().is_empty() {
        return Err(crate::errors::ApplicationError::validation(
            "Idea content cannot be empty",
        ));
    }

    // Perform quick scoring concurrently (no database save)
    let idea_clone1 = idea.to_string();
    let idea_clone2 = idea.to_string();
    let scoring_engine_clone = scoring_engine.clone();
    let pattern_detector_clone = pattern_detector.clone();

    let (score, patterns) = tokio::join!(
        async move { scoring_engine_clone.calculate_score(&idea_clone1) },
        async move { pattern_detector_clone.detect_patterns(&idea_clone2) }
    );

    let score = score?;

    // Display analysis results
    display_analysis_result(idea, &score, &patterns, None);

    // Also provide quick summary
    display_quick_summary(&score);

    Ok(())
}

fn display_quick_summary(score: &crate::scoring::Score) {
    println!();
    println!("🔍 {} QUICK SUMMARY", "SCORE ANALYSIS".bright_blue().bold());
    println!();

    let (emoji, color) = match score.final_score {
        s if s >= 8.0 => ("🔥", "green"),
        s if s >= 6.0 => ("✅", "blue"),
        s if s >= 4.0 => ("⚠️", "yellow"),
        _ => ("🚫", "red"),
    };

    let colored_score = match score.final_score {
        s if s >= 8.0 => format!("{:.1}", s).green(),
        s if s >= 6.0 => format!("{:.1}", s).bright_blue(),
        s if s >= 4.0 => format!("{:.1}", s).yellow(),
        s => format!("{:.1}", s).red(),
    };

    println!("{} Overall Score: {}/10", emoji, colored_score);
    let parsed_color = color
        .parse::<colored::Color>()
        .unwrap_or(colored::Color::White);
    println!(
        "{} {}",
        "Recommendation:".bold(),
        format!(
            "{} {}",
            score.recommendation.emoji(),
            score.recommendation.text()
        )
        .color(parsed_color)
    );

    // Top contributing factors
    let mut factors = Vec::new();

    if score.mission.total >= 2.8 {
        factors.push("Strong mission alignment".to_string());
    }
    if score.anti_challenge.total >= 2.5 {
        factors.push("Combats your challenges".to_string());
    }
    if score.strategic.total >= 1.8 {
        factors.push("Fits current strategy".to_string());
    }

    if !factors.is_empty() {
        println!();
        println!("💪 Key Strengths:");
        for factor in factors {
            println!("   ✓ {}", factor);
        }
    }

    // Action recommendations
    println!();
    print_action_recommendations(score);
}

fn print_action_recommendations(score: &crate::scoring::Score) {
    match score.recommendation {
        crate::scoring::Recommendation::Priority => {
            println!(
                "🎯 {} Add to today's work queue",
                "IMMEDIATE ACTION:".bright_green().bold()
            );
            println!("   → This aligns perfectly with your Telos goals and current strategy");
            println!("   → Low risk of context-switching or procrastination");
            println!("   → Consider starting work within the next 24 hours");
        }
        crate::scoring::Recommendation::Good => {
            println!(
                "📋 {} Add to this week's backlog",
                "STRONG CONSIDERATION:".bright_blue().bold()
            );
            println!("   → Good fit for your current goals and timeline");
            println!("   → May need minor scoping or timeline adjustment");
            println!("   → Consider scheduling for specific date/time");
        }
        crate::scoring::Recommendation::Consider => {
            println!(
                "⏰ {} Queue for later review",
                "FUTURE OPPORTUNITY:".yellow().bold()
            );
            println!("   → Some alignment but wrong timing or scope");
            println!("   → May become higher priority with deadline changes");
            println!("   → Reassess in 1-2 weeks with updated context");
        }
        crate::scoring::Recommendation::Avoid => {
            println!(
                "🛑 {} Major risk factors detected",
                "POSTPONE INDEFINITELY:".red().bold()
            );
            println!("   → High context-switching or perfectionism risk");
            println!("   → Violates current strategy or timeline");
            println!("   → Consider adding to 'Never/Maybe' list");
        }
    }
}
