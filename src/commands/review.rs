use crate::errors::Result;
use colored::*;
use dialoguer::Select;

use crate::database_simple as database;
use crate::display::display_idea_list;

pub async fn handle_review(
    limit: usize,
    min_score: f64,
    pruning: bool,
    db: &database::Database,
) -> Result<()> {
    if pruning {
        return handle_pruning_review(db).await;
    }

    let ideas = db.get_ideas_with_filters(limit, min_score).await?;

    if ideas.is_empty() {
        println!(
            "📭 {} No ideas found matching your criteria.",
            "No Results:".bright_yellow()
        );

        if min_score > 0.0 {
            println!("💡 Try with a lower minimum score: `telos-matrix review --min-score 0.0`");
        }
        return Ok(());
    }

    // Display header with statistics
    display_review_header(&ideas, min_score, limit);

    // Show ideas
    display_idea_list(&ideas);

    // Interactive actions
    if !ideas.is_empty() {
        prompt_review_actions(&ideas, db).await?;
    }

    Ok(())
}

async fn handle_pruning_review(db: &database::Database) -> Result<()> {
    let candidates = db.get_pruning_candidates().await?;

    if candidates.is_empty() {
        println!("✅ {} No ideas need pruning review.", "All Clear:".green());
        println!("   Your idea bank is well-maintained!");
        return Ok(());
    }

    println!(
        "🔍 {} {} ideas need pruning review",
        "Pruning Review:".bright_yellow(),
        candidates.len()
    );
    println!();

    let mut delete_count = 0;
    let mut archive_count = 0;
    let mut keep_count = 0;

    for (i, idea) in candidates.iter().enumerate() {
        let age_days = idea
            .created_at
            .signed_duration_since(chrono::Utc::now())
            .num_days()
            .abs();
        let (action, _action_color) = get_pruning_action(idea, age_days);

        println!(
            "{}. {}",
            (i + 1).to_string().bright_white(),
            idea.content.bright_white()
        );

        if let Some(score) = idea.final_score {
            let colored_score = match score {
                s if s >= 8.0 => format!("{:.1}", s).green(),
                s if s >= 6.0 => format!("{:.1}", s).yellow(),
                s => format!("{:.1}", s).red(),
            };
            println!("   📊 Score: {}/10", colored_score);
        }

        println!("   📅 Age: {} days", age_days);
        match action {
            "DELETE" => println!("   🗑️ {}: {}", "red".red(), action),
            "ARCHIVE" => println!("   📦 {}: {}", "yellow".yellow(), action),
            "KEEP" => println!("   ✅ {}: {}", "green".green(), action),
            _ => println!("   🤔 {}: {}", "blue".blue(), action),
        }

        if let Some(suggestion) = get_pruning_suggestion(idea, age_days) {
            println!("   💡 {}", suggestion.bright_blue());
        }

        // Interactive action
        let options = vec![
            "Delete permanently",
            "Archive (keep but hide)",
            "Keep active",
            "Skip to next",
        ];

        let selection = Select::new()
            .with_prompt("Choose action")
            .items(&options)
            .default(3) // Default to skip
            .interact()?;

        match selection {
            0 => {
                db.delete_idea(&idea.id).await?;
                delete_count += 1;
                println!("   ✅ {}", "Deleted".red());
            }
            1 => {
                db.archive_idea(&idea.id).await?;
                archive_count += 1;
                println!("   ✅ {}", "Archived".yellow());
            }
            2 => {
                keep_count += 1;
                println!("   ✅ {}", "Kept active".green());
            }
            3 => {
                println!("   ⏭️ {}", "Skipped".dimmed());
            }
            _ => unreachable!(),
        }

        println!();
    }

    // Summary
    println!("📊 {} Pruning Summary:", "Complete".bright_blue().bold());
    println!("   🗑️  Deleted: {} ideas", delete_count.to_string().red());
    println!(
        "   📦 Archived: {} ideas",
        archive_count.to_string().yellow()
    );
    println!(
        "   ✅ Kept active: {} ideas",
        keep_count.to_string().green()
    );

    if delete_count + archive_count > 0 {
        println!();
        println!("🎉 Your idea bank is now cleaner and more focused!");
    }

    Ok(())
}

fn display_review_header(ideas: &[crate::database::StoredIdea], min_score: f64, limit: usize) {
    let total_count = ideas.len();

    println!(
        "📋 {} {} ideas found",
        "Idea Review".bright_blue().bold(),
        total_count
    );

    if min_score > 0.0 {
        println!("   📊 Filter: score ≥ {:.1}", min_score);
    }

    if limit > 0 && total_count >= limit {
        println!("   📄 Limit: showing first {} ideas", limit);
    }

    // Score distribution
    if !ideas.is_empty() {
        let high_priority = ideas
            .iter()
            .filter(|i| i.final_score.unwrap_or(0.0) >= 8.0)
            .count();
        let good = ideas
            .iter()
            .filter(|i| {
                let score = i.final_score.unwrap_or(0.0);
                (6.0..8.0).contains(&score)
            })
            .count();
        let consider = ideas
            .iter()
            .filter(|i| {
                let score = i.final_score.unwrap_or(0.0);
                (4.0..6.0).contains(&score)
            })
            .count();
        let avoid = ideas
            .iter()
            .filter(|i| i.final_score.unwrap_or(0.0) < 4.0)
            .count();

        println!(
            "   📈 Distribution: {} 🔥 priority, {} ✅ good, {} ⚠️ consider, {} 🚫 avoid",
            high_priority.to_string().green(),
            good.to_string().blue(),
            consider.to_string().yellow(),
            avoid.to_string().red()
        );
    }

    println!();
}

async fn prompt_review_actions(
    ideas: &[crate::database_simple::StoredIdea],
    db: &database::Database,
) -> Result<()> {
    use dialoguer::Confirm;

    println!(
        "🤔 {} What would you like to do?",
        "Next Actions".bright_blue().bold()
    );

    let options = vec![
        "Score a new idea",
        "Analyze one of these ideas in detail",
        "Review lower-scoring ideas",
        "Run pruning review",
        "Export ideas to file",
        "Exit",
    ];

    let selection = Select::new()
        .with_prompt("Choose an action")
        .items(&options)
        .interact()?;

    match selection {
        0 => {
            // Score a new idea
            println!("\n💡 Enter your idea to score:");
            println!("   (or press Ctrl+C to cancel)");
            println!();

            // In a real implementation, you'd prompt for input here
            println!("Use: `telos-matrix score \"your idea here\"`");
        }
        1 => {
            // Analyze one idea in detail
            if ideas.is_empty() {
                println!("No ideas available for detailed analysis.");
                return Ok(());
            }

            let idea_options: Vec<String> = ideas
                .iter()
                .take(10) // Limit to first 10 for selection
                .map(|idea| {
                    let preview = if idea.content.len() > 50 {
                        format!("{}...", &idea.content[..47])
                    } else {
                        idea.content.clone()
                    };
                    format!(
                        "{} (Score: {:.1})",
                        preview,
                        idea.final_score.unwrap_or(0.0)
                    )
                })
                .collect();

            let selection = Select::new()
                .with_prompt("Select an idea to analyze in detail")
                .items(&idea_options)
                .interact()?;

            if let Some(idea) = ideas.get(selection) {
                println!("\n🔍 Detailed analysis for:");
                println!("\"{}\"", idea.content);
                println!("\nUse: `telos-matrix analyze --last` to see full analysis");
            }
        }
        2 => {
            // Review lower-scoring ideas
            println!("\n📉 Review lower-scoring ideas:");
            println!("Use: `telos-matrix review --min-score 4.0`");
        }
        3 => {
            // Run pruning review
            if Confirm::new()
                .with_prompt("Review ideas for pruning?")
                .default(true)
                .interact()?
            {
                handle_pruning_review(db).await?;
            }
        }
        4 => {
            // Export ideas
            println!("\n📤 Export functionality coming soon!");
            println!("For now, ideas are stored in: ~/.local/share/telos-matrix/ideas.db");
        }
        5 => {
            // Exit
            println!("\n👋 Happy idea prioritizing!");
        }
        _ => unreachable!(),
    }

    Ok(())
}

fn get_pruning_action(
    idea: &crate::database::StoredIdea,
    age_days: i64,
) -> (&'static str, colored::Color) {
    let score = idea.final_score.unwrap_or(0.0);

    if score < 3.0 && age_days > 7 {
        ("DELETE", colored::Color::Red)
    } else if score < 6.0 && age_days > 14 {
        ("ARCHIVE", colored::Color::Yellow)
    } else if score >= 8.0 {
        ("KEEP (High Value)", colored::Color::Green)
    } else {
        ("CONSIDER", colored::Color::Blue)
    }
}

fn get_pruning_suggestion(idea: &crate::database::StoredIdea, age_days: i64) -> Option<String> {
    let score = idea.final_score.unwrap_or(0.0);

    if score < 3.0 && age_days > 7 {
        Some("Low value and old - safe to delete".to_string())
    } else if score < 6.0 && age_days > 14 {
        Some("Medium value but old - archive for potential future use".to_string())
    } else if score >= 8.0 {
        Some("High priority idea - keep active".to_string())
    } else {
        None
    }
}
