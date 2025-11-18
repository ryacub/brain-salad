use crate::errors::Result;
use colored::*;

use crate::database_simple::Database;

pub async fn handle_prune(auto: bool, dry_run: bool, db: &Database) -> Result<()> {
    println!("🔍 {}", "Analyzing ideas for pruning...".bright_blue());
    println!();

    let candidates = db.get_pruning_candidates().await?;

    if candidates.is_empty() {
        println!("✅ {} No ideas need pruning.", "All Clear".green());
        println!("   Your idea bank is well-maintained and focused!");
        return Ok(());
    }

    // Analyze candidates
    let mut delete_candidates = Vec::new();
    let mut archive_candidates = Vec::new();
    let mut review_candidates = Vec::new();

    for idea in &candidates {
        let age_days = idea
            .created_at
            .signed_duration_since(chrono::Utc::now())
            .num_days()
            .abs();
        let action = determine_pruning_action(idea, age_days);

        match action {
            PruningAction::Delete => delete_candidates.push((idea, age_days)),
            PruningAction::Archive => archive_candidates.push((idea, age_days)),
            PruningAction::Review => review_candidates.push((idea, age_days)),
        }
    }

    // Display analysis results
    display_pruning_analysis(&delete_candidates, &archive_candidates, &review_candidates);

    if dry_run {
        println!();
        println!(
            "🔍 {} This was a dry run. No ideas were modified.",
            "Dry Run Complete".bright_blue()
        );
        println!("   Run without --dry-run to actually prune ideas.");
        return Ok(());
    }

    if auto {
        return execute_auto_pruning(delete_candidates, archive_candidates, db).await;
    }

    // Interactive pruning
    execute_interactive_pruning(delete_candidates, archive_candidates, review_candidates, db).await
}

#[derive(Debug, Clone)]
enum PruningAction {
    Delete,
    Archive,
    Review,
}

fn determine_pruning_action(idea: &crate::database::StoredIdea, age_days: i64) -> PruningAction {
    let score = idea.final_score.unwrap_or(0.0);

    // Pruning rules from Phase 2 design
    if score < 3.0 && age_days > 7 {
        PruningAction::Delete
    } else if score < 6.0 && age_days > 14 {
        PruningAction::Archive
    } else if score >= 8.0 {
        PruningAction::Review // High-value items always need manual review
    } else {
        PruningAction::Review // Borderline cases need manual review
    }
}

fn display_pruning_analysis(
    delete_candidates: &[(&crate::database::StoredIdea, i64)],
    archive_candidates: &[(&crate::database::StoredIdea, i64)],
    review_candidates: &[(&crate::database::StoredIdea, i64)],
) {
    println!("📊 {} Pruning Analysis:", "Results".bright_blue().bold());
    println!();

    // Delete candidates
    if !delete_candidates.is_empty() {
        println!(
            "🗑️  {} {} ideas marked for deletion:",
            "Delete".red().bold(),
            delete_candidates.len()
        );
        for (idea, age_days) in delete_candidates.iter().take(5) {
            let preview = truncate_idea(&idea.content, 60);
            println!(
                "   • {} ({:.1}/10, {} days old)",
                preview,
                idea.final_score.unwrap_or(0.0),
                age_days
            );
        }
        if delete_candidates.len() > 5 {
            println!("   • ... and {} more", delete_candidates.len() - 5);
        }
        println!();
    }

    // Archive candidates
    if !archive_candidates.is_empty() {
        println!(
            "📦 {} {} ideas marked for archiving:",
            "Archive".yellow().bold(),
            archive_candidates.len()
        );
        for (idea, age_days) in archive_candidates.iter().take(5) {
            let preview = truncate_idea(&idea.content, 60);
            println!(
                "   • {} ({:.1}/10, {} days old)",
                preview,
                idea.final_score.unwrap_or(0.0),
                age_days
            );
        }
        if archive_candidates.len() > 5 {
            println!("   • ... and {} more", archive_candidates.len() - 5);
        }
        println!();
    }

    // Review candidates
    if !review_candidates.is_empty() {
        println!(
            "🤔 {} {} ideas need manual review:",
            "Review".bright_blue().bold(),
            review_candidates.len()
        );
        for (idea, age_days) in review_candidates.iter().take(5) {
            let preview = truncate_idea(&idea.content, 60);
            println!(
                "   • {} ({:.1}/10, {} days old)",
                preview,
                idea.final_score.unwrap_or(0.0),
                age_days
            );
        }
        if review_candidates.len() > 5 {
            println!("   • ... and {} more", review_candidates.len() - 5);
        }
        println!();
    }

    // Summary
    let total_ideas = delete_candidates.len() + archive_candidates.len() + review_candidates.len();
    println!("📈 Summary:");
    println!(
        "   Total candidates: {}",
        total_ideas.to_string().bright_white()
    );
    println!(
        "   Recommended deletion: {}",
        delete_candidates.len().to_string().red()
    );
    println!(
        "   Recommended archiving: {}",
        archive_candidates.len().to_string().yellow()
    );
    println!(
        "   Manual review needed: {}",
        review_candidates.len().to_string().bright_blue()
    );
}

async fn execute_auto_pruning(
    delete_candidates: Vec<(&crate::database_simple::StoredIdea, i64)>,
    archive_candidates: Vec<(&crate::database_simple::StoredIdea, i64)>,
    db: &Database,
) -> Result<()> {
    println!(
        "🤖 {} Auto-pruning based on rules...",
        "Auto Mode".bright_blue()
    );
    println!();

    let mut deleted_count = 0;
    let mut archived_count = 0;

    // Delete candidates
    for (idea, _) in delete_candidates {
        db.delete_idea(&idea.id).await?;
        deleted_count += 1;
        println!("   🗑️  Deleted: {}", truncate_idea(&idea.content, 50));
    }

    // Archive candidates
    for (idea, _) in archive_candidates {
        db.archive_idea(&idea.id).await?;
        archived_count += 1;
        println!("   📦 Archived: {}", truncate_idea(&idea.content, 50));
    }

    // Summary
    println!();
    println!("✅ {} Auto-pruning complete!", "Done".green().bold());
    println!("   🗑️  Deleted: {} ideas", deleted_count.to_string().red());
    println!(
        "   📦 Archived: {} ideas",
        archived_count.to_string().yellow()
    );
    println!(
        "   💾 Ideas remaining in active bank: {}",
        (db.get_idea_count().await? - deleted_count - archived_count)
            .to_string()
            .green()
    );

    Ok(())
}

async fn execute_interactive_pruning(
    delete_candidates: Vec<(&crate::database_simple::StoredIdea, i64)>,
    archive_candidates: Vec<(&crate::database_simple::StoredIdea, i64)>,
    review_candidates: Vec<(&crate::database_simple::StoredIdea, i64)>,
    db: &Database,
) -> Result<()> {
    use dialoguer::{Confirm, MultiSelect};

    println!(
        "🤔 {} Interactive pruning mode",
        "Interactive Mode".bright_blue().bold()
    );
    println!();

    let mut deleted_count = 0;
    let mut archived_count = 0;
    let mut kept_count = 0;

    // Process delete candidates
    if !delete_candidates.is_empty() {
        println!(
            "🗑️  {} Review candidates for deletion:",
            "Deletion Candidates".red().bold()
        );

        let items: Vec<String> = delete_candidates
            .iter()
            .map(|(idea, age_days)| {
                format!(
                    "{} ({:.1}/10, {} days old)",
                    truncate_idea(&idea.content, 70),
                    idea.final_score.unwrap_or(0.0),
                    age_days
                )
            })
            .collect();

        if Confirm::new()
            .with_prompt(format!(
                "Review {} items for deletion?",
                delete_candidates.len()
            ))
            .default(true)
            .interact()?
        {
            let selections = MultiSelect::new()
                .with_prompt("Select ideas to delete (space to toggle, enter to confirm)")
                .items(&items)
                .interact()?;

            for &index in &selections {
                if let Some((idea, _)) = delete_candidates.get(index) {
                    db.delete_idea(&idea.id).await?;
                    deleted_count += 1;
                    println!("   🗑️  Deleted: {}", truncate_idea(&idea.content, 50));
                }
            }

            kept_count += delete_candidates.len() - selections.len();
        }
    }

    // Process archive candidates
    if !archive_candidates.is_empty() {
        println!();
        println!(
            "📦 {} Review candidates for archiving:",
            "Archive Candidates".yellow().bold()
        );

        let items: Vec<String> = archive_candidates
            .iter()
            .map(|(idea, age_days)| {
                format!(
                    "{} ({:.1}/10, {} days old)",
                    truncate_idea(&idea.content, 70),
                    idea.final_score.unwrap_or(0.0),
                    age_days
                )
            })
            .collect();

        if Confirm::new()
            .with_prompt(format!(
                "Review {} items for archiving?",
                archive_candidates.len()
            ))
            .default(true)
            .interact()?
        {
            let selections = MultiSelect::new()
                .with_prompt("Select ideas to archive (space to toggle, enter to confirm)")
                .items(&items)
                .interact()?;

            for &index in &selections {
                if let Some((idea, _)) = archive_candidates.get(index) {
                    db.archive_idea(&idea.id).await?;
                    archived_count += 1;
                    println!("   📦 Archived: {}", truncate_idea(&idea.content, 50));
                }
            }

            kept_count += archive_candidates.len() - selections.len();
        }
    }

    // Handle review candidates
    if !review_candidates.is_empty() {
        println!();
        println!(
            "🤔 {} Manual review needed for:",
            "Manual Review".bright_blue().bold()
        );
        for (idea, age_days) in &review_candidates {
            println!(
                "   • {} ({:.1}/10, {} days old)",
                truncate_idea(&idea.content, 70),
                idea.final_score.unwrap_or(0.0),
                age_days
            );
        }
        println!("   → Run `telos-matrix review --pruning` for detailed manual review");
        kept_count += review_candidates.len();
    }

    // Final summary
    println!();
    println!("✅ {} Interactive pruning complete!", "Done".green().bold());
    println!("   🗑️  Deleted: {} ideas", deleted_count.to_string().red());
    println!(
        "   📦 Archived: {} ideas",
        archived_count.to_string().yellow()
    );
    println!(
        "   ✅ Kept active: {} ideas",
        kept_count.to_string().green()
    );
    println!(
        "   💾 Ideas remaining in active bank: {}",
        (db.get_idea_count().await? - deleted_count - archived_count)
            .to_string()
            .bright_blue()
    );

    Ok(())
}

fn truncate_idea(content: &str, max_len: usize) -> String {
    if content.len() <= max_len {
        content.to_string()
    } else {
        format!("{}...", &content[..max_len.saturating_sub(3)])
    }
}
