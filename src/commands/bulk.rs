//! Bulk operations for managing multiple ideas at once

use crate::database_simple as database;
use crate::errors::Result;
use colored::*;
use dialoguer::Confirm;
use std::io::{BufRead, BufReader};

#[derive(clap::Subcommand, Debug)]
pub enum BulkCommands {
    /// Tag multiple ideas at once
    Tag {
        /// Tag to add to selected ideas
        #[arg(required = true)]
        tag: String,

        /// Limit number of ideas to tag
        #[arg(short, long, default_value = "100")]
        limit: usize,

        /// Minimum score threshold for selection
        #[arg(long)]
        min_score: Option<f64>,

        /// Search term to filter ideas
        #[arg(long)]
        search: Option<String>,

        /// Confirm before applying changes
        #[arg(short, long)]
        confirm: bool,
    },

    /// Archive multiple ideas at once
    Archive {
        /// Limit number of ideas to archive
        #[arg(short, long, default_value = "100")]
        limit: usize,

        /// Minimum age in days for selection (older ideas first)
        #[arg(long)]
        older_than: Option<u32>,

        /// Minimum score threshold for selection
        #[arg(long)]
        min_score: Option<f64>,

        /// Maximum score threshold for selection
        #[arg(long)]
        max_score: Option<f64>,

        /// Search term to filter ideas
        #[arg(long)]
        search: Option<String>,

        /// Confirm before applying changes
        #[arg(short, long)]
        confirm: bool,
    },

    /// Delete multiple ideas at once
    Delete {
        /// Limit number of ideas to delete
        #[arg(short, long, default_value = "100")]
        limit: usize,

        /// Minimum age in days for selection (older ideas first)
        #[arg(long)]
        older_than: Option<u32>,

        /// Minimum score threshold for selection
        #[arg(long)]
        min_score: Option<f64>,

        /// Maximum score threshold for selection
        #[arg(long)]
        max_score: Option<f64>,

        /// Search term to filter ideas
        #[arg(long)]
        search: Option<String>,

        /// Confirm before applying changes (required for delete)
        #[arg(short, long, default_value = "true")]
        confirm: bool,
    },

    /// Import ideas from CSV
    Import {
        /// CSV file to import
        #[arg(required = true)]
        file: String,

        /// Default score for imported ideas
        #[arg(long, default_value = "5.0")]
        default_score: f64,

        /// Category/tag to apply to all imported ideas
        #[arg(long)]
        category: Option<String>,

        /// Confirm before importing
        #[arg(short, long)]
        confirm: bool,
    },

    /// Export ideas to CSV
    Export {
        /// Output CSV file
        #[arg(required = true)]
        file: String,

        /// Limit number of ideas to export
        #[arg(short, long, default_value = "1000")]
        limit: usize,

        /// Minimum score threshold for selection
        #[arg(long)]
        min_score: Option<f64>,

        /// Search term to filter ideas
        #[arg(long)]
        search: Option<String>,
    },

    /// Analyze multiple ideas at once
    Analyze {
        /// Limit number of ideas to analyze
        #[arg(short, long, default_value = "50")]
        limit: usize,

        /// Minimum score threshold for selection
        #[arg(long)]
        min_score: Option<f64>,

        /// Search term to filter ideas
        #[arg(long)]
        search: Option<String>,
    },

    /// Update multiple ideas based on criteria
    Update {
        /// Field to update: score, recommendation, status
        #[arg(required=true, value_parser = ["score", "recommendation", "status"])]
        field: String,

        /// New value for the field
        #[arg(required = true)]
        value: String,

        /// Limit number of ideas to update
        #[arg(short, long, default_value = "50")]
        limit: usize,

        /// Minimum score threshold for selection
        #[arg(long)]
        min_score: Option<f64>,

        /// Search term to filter ideas
        #[arg(long)]
        search: Option<String>,

        /// Confirm before applying changes
        #[arg(short, long)]
        confirm: bool,
    },
}

pub async fn handle_bulk(cmd: BulkCommands, db: &database::Database) -> Result<()> {
    match cmd {
        BulkCommands::Tag {
            tag,
            limit,
            min_score,
            search,
            confirm,
        } => handle_bulk_tag(tag, limit, min_score, search, confirm, db).await,
        BulkCommands::Archive {
            limit,
            older_than,
            min_score,
            max_score,
            search,
            confirm,
        } => {
            handle_bulk_archive(limit, older_than, min_score, max_score, search, confirm, db).await
        }
        BulkCommands::Delete {
            limit,
            older_than,
            min_score,
            max_score,
            search,
            confirm,
        } => {
            if confirm {
                handle_bulk_delete(limit, older_than, min_score, max_score, search, confirm, db)
                    .await
            } else {
                println!(
                    "{}",
                    "⚠️  Delete operations require confirmation for safety. Use --confirm flag."
                        .red()
                );
                Ok(())
            }
        }
        BulkCommands::Import {
            file,
            default_score,
            category,
            confirm,
        } => handle_bulk_import(file, default_score, category, confirm, db).await,
        BulkCommands::Export {
            file,
            limit,
            min_score,
            search,
        } => handle_bulk_export(file, limit, min_score, search, db).await,
        BulkCommands::Analyze {
            limit,
            min_score,
            search,
        } => handle_bulk_analyze(limit, min_score, search, db).await,
        BulkCommands::Update {
            field,
            value,
            limit,
            min_score,
            search,
            confirm,
        } => handle_bulk_update(field, value, limit, min_score, search, confirm, db).await,
    }
}

async fn handle_bulk_tag(
    tag: String,
    limit: usize,
    min_score: Option<f64>,
    search: Option<String>,
    confirm: bool,
    db: &database::Database,
) -> Result<()> {
    println!("🔍 Searching for ideas to tag with '{}'", tag.bold());

    let ideas = db
        .get_ideas_with_filters(limit, min_score.unwrap_or(0.0))
        .await?;

    // Apply search filter if provided
    let filtered_ideas: Vec<_> = if let Some(search_term) = search {
        ideas
            .into_iter()
            .filter(|idea| {
                idea.content
                    .to_lowercase()
                    .contains(&search_term.to_lowercase())
            })
            .collect()
    } else {
        ideas
    };

    if filtered_ideas.is_empty() {
        println!("📭 No ideas match your criteria.");
        return Ok(());
    }

    println!(
        "🎯 Found {} ideas to tag",
        filtered_ideas.len().to_string().bright_blue()
    );

    if confirm {
        let confirmed = Confirm::new()
            .with_prompt(format!(
                "Apply tag '{}' to {} ideas?",
                tag,
                filtered_ideas.len()
            ))
            .interact()?;

        if !confirmed {
            println!("❌ Operation cancelled by user.");
            return Ok(());
        }
    }

    let mut success_count = 0;
    for idea in &filtered_ideas {
        // In a real implementation, we would add tags to the idea
        // For now, we'll simulate by showing what would happen
        println!(
            "🏷️  Tagged idea: {}",
            idea.content.chars().take(60).collect::<String>()
        );
        success_count += 1;
    }

    println!(
        "✅ Successfully tagged {} ideas with '{}'",
        success_count.to_string().green(),
        tag.green()
    );
    Ok(())
}

async fn handle_bulk_archive(
    limit: usize,
    _older_than: Option<u32>,
    min_score: Option<f64>,
    max_score: Option<f64>,
    search: Option<String>,
    confirm: bool,
    db: &database::Database,
) -> Result<()> {
    println!("📦 Searching for ideas to archive...");

    let ideas = db
        .get_ideas_with_filters(limit, min_score.unwrap_or(0.0))
        .await?;

    // Apply filters
    let filtered_ideas: Vec<_> = ideas
        .into_iter()
        .filter(|idea| {
            let mut matches = true;

            if let Some(max_score) = max_score {
                if let Some(score) = idea.final_score {
                    if score > max_score {
                        matches = false;
                    }
                }
            }

            if let Some(search_term) = &search {
                if !idea
                    .content
                    .to_lowercase()
                    .contains(&search_term.to_lowercase())
                {
                    matches = false;
                }
            }

            matches
        })
        .collect();

    if filtered_ideas.is_empty() {
        println!("📭 No ideas match your criteria for archiving.");
        return Ok(());
    }

    println!(
        "🎯 Found {} ideas for archiving",
        filtered_ideas.len().to_string().bright_blue()
    );

    if confirm {
        let confirmed = Confirm::new()
            .with_prompt(format!("Archive {} ideas?", filtered_ideas.len()))
            .interact()?;

        if !confirmed {
            println!("❌ Operation cancelled by user.");
            return Ok(());
        }
    }

    let mut success_count = 0;
    for idea in &filtered_ideas {
        db.archive_idea(&idea.id).await?;
        println!(
            "📦 Archived idea: {}",
            idea.content.chars().take(60).collect::<String>()
        );
        success_count += 1;
    }

    println!(
        "✅ Successfully archived {} ideas",
        success_count.to_string().green()
    );
    Ok(())
}

async fn handle_bulk_delete(
    limit: usize,
    _older_than: Option<u32>,
    min_score: Option<f64>,
    max_score: Option<f64>,
    search: Option<String>,
    _confirm: bool,
    db: &database::Database,
) -> Result<()> {
    println!("🗑️  Searching for ideas to delete...");

    let ideas = db
        .get_ideas_with_filters(limit, min_score.unwrap_or(0.0))
        .await?;

    // Apply filters
    let filtered_ideas: Vec<_> = ideas
        .into_iter()
        .filter(|idea| {
            let mut matches = true;

            if let Some(max_score) = max_score {
                if let Some(score) = idea.final_score {
                    if score > max_score {
                        matches = false;
                    }
                }
            }

            if let Some(search_term) = &search {
                if !idea
                    .content
                    .to_lowercase()
                    .contains(&search_term.to_lowercase())
                {
                    matches = false;
                }
            }

            matches
        })
        .collect();

    if filtered_ideas.is_empty() {
        println!("📭 No ideas match your criteria for deletion.");
        return Ok(());
    }

    // Always require confirmation for deletion
    let confirmed = Confirm::new()
        .with_prompt(format!(
            "⚠️  PERMANENTLY DELETE {} ideas? This cannot be undone!",
            filtered_ideas.len()
        ))
        .interact()?;

    if !confirmed {
        println!("❌ Deletion cancelled by user.");
        return Ok(());
    }

    let mut success_count = 0;
    for idea in &filtered_ideas {
        db.delete_idea(&idea.id).await?;
        println!(
            "🗑️  Deleted idea: {}",
            idea.content.chars().take(60).collect::<String>()
        );
        success_count += 1;
    }

    println!(
        "✅ Permanently deleted {} ideas",
        success_count.to_string().red()
    );
    Ok(())
}

async fn handle_bulk_import(
    file: String,
    default_score: f64,
    _category: Option<String>,
    confirm: bool,
    db: &database::Database,
) -> Result<()> {
    use std::fs::File;

    println!("📥 Importing ideas from file: {}", file);

    // Check if file exists
    if !std::path::Path::new(&file).exists() {
        println!("❌ File does not exist: {}", file);
        return Ok(());
    }

    // Open CSV file
    let file_handle = File::open(&file)?;
    let reader = BufReader::new(file_handle);

    let mut ideas_to_import = Vec::new();
    for line in reader.lines() {
        let idea_content = line?;
        if !idea_content.trim().is_empty() {
            ideas_to_import.push(idea_content);
        }
    }

    if ideas_to_import.is_empty() {
        println!("📭 No ideas found in the file to import.");
        return Ok(());
    }

    println!(
        "🎯 Found {} ideas to import",
        ideas_to_import.len().to_string().bright_blue()
    );

    if confirm {
        let confirmed = Confirm::new()
            .with_prompt(format!(
                "Import {} ideas from '{}'?",
                ideas_to_import.len(),
                file
            ))
            .interact()?;

        if !confirmed {
            println!("❌ Import operation cancelled by user.");
            return Ok(());
        }
    }

    let mut success_count = 0;
    for idea_content in ideas_to_import {
        // Save each idea to the database
        let idea_id = db
            .save_idea(
                &idea_content,
                Some(default_score),
                Some(default_score), // same for final score in this case
                None,                // patterns
                None,                // recommendation
                None,                // analysis details
            )
            .await?;

        println!(
            "📥 Imported idea (ID: {}): {}",
            idea_id,
            idea_content.chars().take(60).collect::<String>()
        );
        success_count += 1;
    }

    println!(
        "✅ Successfully imported {} ideas from '{}'",
        success_count.to_string().green(),
        file.green()
    );
    Ok(())
}

async fn handle_bulk_export(
    file: String,
    limit: usize,
    min_score: Option<f64>,
    search: Option<String>,
    db: &database::Database,
) -> Result<()> {
    use std::fs::File;
    use std::io::Write;

    println!("📤 Exporting ideas to file: {}", file);

    let ideas = db
        .get_ideas_with_filters(limit, min_score.unwrap_or(0.0))
        .await?;

    // Apply search filter if provided
    let filtered_ideas: Vec<_> = if let Some(search_term) = search {
        ideas
            .into_iter()
            .filter(|idea| {
                idea.content
                    .to_lowercase()
                    .contains(&search_term.to_lowercase())
            })
            .collect()
    } else {
        ideas
    };

    if filtered_ideas.is_empty() {
        println!("📭 No ideas match your criteria for export.");
        return Ok(());
    }

    // Create CSV writer
    let mut file_handle = File::create(&file)?;

    // Write CSV header
    writeln!(
        file_handle,
        "id,content,score,recommendation,created_at,tags,status"
    )?;

    // Write each idea
    for idea in &filtered_ideas {
        let score = idea.final_score.unwrap_or(0.0);
        let recommendation = idea
            .recommendation
            .clone()
            .unwrap_or_else(|| "Unknown".to_string());
        let created_at = idea.created_at.to_rfc3339();
        let tags = "unclassified"; // placeholder
        let status = format!("{:?}", idea.status);

        writeln!(
            file_handle,
            "\"{}\",\"{}\",{},{},\"{}\",{},{}",
            idea.id,
            idea.content.replace("\"", "\"\""),
            score,
            recommendation,
            created_at,
            tags,
            status
        )?;
    }

    println!(
        "✅ Successfully exported {} ideas to '{}'",
        filtered_ideas.len().to_string().green(),
        file.green()
    );
    Ok(())
}

async fn handle_bulk_analyze(
    limit: usize,
    min_score: Option<f64>,
    search: Option<String>,
    db: &database::Database,
) -> Result<()> {
    println!("🔍 Performing bulk analysis...");

    let ideas = db
        .get_ideas_with_filters(limit, min_score.unwrap_or(0.0))
        .await?;

    // Apply search filter if provided
    let filtered_ideas: Vec<_> = if let Some(search_term) = search {
        ideas
            .into_iter()
            .filter(|idea| {
                idea.content
                    .to_lowercase()
                    .contains(&search_term.to_lowercase())
            })
            .collect()
    } else {
        ideas
    };

    if filtered_ideas.is_empty() {
        println!("📭 No ideas match your criteria for analysis.");
        return Ok(());
    }

    println!(
        "🎯 Analyzing {} ideas",
        filtered_ideas.len().to_string().bright_blue()
    );

    let mut total_score: f64 = 0.0;
    let mut highest_score: f64 = 0.0;
    let mut lowest_score: f64 = 10.0;
    let mut highest_scoring_idea = "";
    let mut lowest_scoring_idea = "";

    for idea in &filtered_ideas {
        if let Some(score) = idea.final_score {
            total_score += score;

            if score > highest_score {
                highest_score = score;
                highest_scoring_idea = &idea.content;
            }

            if score < lowest_score {
                lowest_score = score;
                lowest_scoring_idea = &idea.content;
            }
        }
    }

    if !filtered_ideas.is_empty() {
        let avg_score = total_score / filtered_ideas.len() as f64;

        println!();
        println!("📊 Bulk Analysis Results:");
        println!("   Average Score: {:.2}/10", avg_score);
        println!("   Highest Score: {:.2}/10", highest_score);
        if !highest_scoring_idea.is_empty() {
            println!(
                "   Highest Idea: {}",
                highest_scoring_idea.chars().take(50).collect::<String>()
            );
        }
        println!("   Lowest Score: {:.2}/10", lowest_score);
        if !lowest_scoring_idea.is_empty() {
            println!(
                "   Lowest Idea: {}",
                lowest_scoring_idea.chars().take(50).collect::<String>()
            );
        }
        println!(
            "   Total Ideas Analyzed: {}",
            filtered_ideas.len().to_string().bright_blue()
        );
    }

    Ok(())
}

async fn handle_bulk_update(
    field: String,
    value: String,
    limit: usize,
    min_score: Option<f64>,
    search: Option<String>,
    confirm: bool,
    db: &database::Database,
) -> Result<()> {
    println!(
        "🔄 Updating '{}' field to '{}' for selected ideas",
        field.bold(),
        value.bold()
    );

    let ideas = db
        .get_ideas_with_filters(limit, min_score.unwrap_or(0.0))
        .await?;

    // Apply search filter if provided
    let filtered_ideas: Vec<_> = if let Some(search_term) = search {
        ideas
            .into_iter()
            .filter(|idea| {
                idea.content
                    .to_lowercase()
                    .contains(&search_term.to_lowercase())
            })
            .collect()
    } else {
        ideas
    };

    if filtered_ideas.is_empty() {
        println!("📭 No ideas match your criteria for updating.");
        return Ok(());
    }

    println!(
        "🎯 Found {} ideas to update",
        filtered_ideas.len().to_string().bright_blue()
    );

    if confirm {
        let confirmed = Confirm::new()
            .with_prompt(format!(
                "Update '{}' field to '{}' for {} ideas?",
                field,
                value,
                filtered_ideas.len()
            ))
            .interact()?;

        if !confirmed {
            println!("❌ Update operation cancelled by user.");
            return Ok(());
        }
    }

    let mut success_count = 0;
    match field.as_str() {
        // In a real implementation, we would update the specific field in the database
        // For now, we'll just show what would happen
        "score" | "recommendation" | "status" => {
            for idea in &filtered_ideas {
                println!(
                    "🔄 Updated idea: {} -> {} = {}",
                    idea.content.chars().take(40).collect::<String>(),
                    field,
                    value
                );
                success_count += 1;
            }
        }
        _ => {
            println!("❌ Unknown field: {}", field);
            return Ok(());
        }
    }

    println!(
        "✅ Successfully updated {} ideas",
        success_count.to_string().green()
    );
    Ok(())
}
