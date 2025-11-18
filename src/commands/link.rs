//! Commands for linking and managing idea relationships

use crate::database_simple as database;
use crate::errors::Result;
use colored::*;
use dialoguer::Confirm;

#[derive(clap::Subcommand, Debug)]
pub enum LinkCommands {
    /// Link two ideas together with a relationship
    Create {
        /// Source idea ID to link from
        #[arg(required = true)]
        source_id: String,

        /// Target idea ID to link to
        #[arg(required = true)]
        target_id: String,

        /// Type of relationship
        #[arg(required=true, value_parser = ["depends_on", "related_to", "part_of", "parent", "child", "duplicate", "blocks", "blocked_by", "similar_to"])]
        relationship_type: String,

        /// Confirm before creating the relationship
        #[arg(short, long)]
        confirm: bool,
    },

    /// List all relationships for an idea
    List {
        /// Idea ID to show relationships for
        #[arg(required = true)]
        idea_id: String,
    },

    /// Show related ideas for an idea
    Show {
        /// Idea ID to show related ideas for
        #[arg(required = true)]
        idea_id: String,

        /// Filter by relationship type
        #[arg(long)]
        relationship_type: Option<String>,
    },

    /// Remove a relationship between ideas
    Remove {
        /// Relationship ID to remove
        #[arg(required = true)]
        relationship_id: String,

        /// Confirm before removing
        #[arg(short, long)]
        confirm: bool,
    },

    /// Find dependency paths between ideas
    Path {
        /// Starting idea ID
        #[arg(required = true)]
        from: String,

        /// Target idea ID
        #[arg(required = true)]
        to: String,
    },
}

pub async fn handle_link(cmd: LinkCommands, db: &database::Database) -> Result<()> {
    match cmd {
        LinkCommands::Create {
            source_id,
            target_id,
            relationship_type,
            confirm,
        } => handle_create_link(source_id, target_id, relationship_type, confirm, db).await,
        LinkCommands::List { idea_id } => handle_list_links(idea_id, db).await,
        LinkCommands::Show {
            idea_id,
            relationship_type,
        } => handle_show_related(idea_id, relationship_type, db).await,
        LinkCommands::Remove {
            relationship_id,
            confirm,
        } => handle_remove_link(relationship_id, confirm, db).await,
        LinkCommands::Path { from, to } => handle_find_path(from, to, db).await,
    }
}

async fn handle_create_link(
    source_id: String,
    target_id: String,
    relationship_type: String,
    confirm: bool,
    db: &database::Database,
) -> Result<()> {
    // Verify both ideas exist
    let source_idea = db.get_by_id(&source_id).await?;
    let source_idea = match source_idea {
        Some(idea) => idea,
        None => {
            println!("❌ Source idea not found: {}", source_id.red());
            return Ok(());
        }
    };

    let target_idea = db.get_by_id(&target_id).await?;
    let target_idea = match target_idea {
        Some(idea) => idea,
        None => {
            println!("❌ Target idea not found: {}", target_id.red());
            return Ok(());
        }
    };

    let rel_type_enum = match relationship_type.parse::<database::RelationshipType>() {
        Ok(rt) => rt,
        Err(_) => {
            println!(
                "❌ Invalid relationship type: '{}'",
                relationship_type.red()
            );
            println!("Valid types: depends_on, related_to, part_of, parent, child, duplicate, blocks, blocked_by, similar_to");
            return Ok(());
        }
    };

    if confirm {
        println!("🔗 Creating relationship:");
        println!(
            "  Source: {} ({})",
            source_id.blue(),
            source_idea.content.chars().take(50).collect::<String>()
        );
        println!(
            "  Target: {} ({})",
            target_id.blue(),
            target_idea.content.chars().take(50).collect::<String>()
        );
        println!("  Type: {}", relationship_type.cyan());

        let confirmed = Confirm::new()
            .with_prompt("Create this relationship?")
            .interact()?;

        if !confirmed {
            println!("❌ Operation cancelled.");
            return Ok(());
        }
    }

    match db
        .create_relationship(&source_id, &target_id, rel_type_enum)
        .await
    {
        Ok(relationship_id) => {
            println!("✅ Relationship created successfully!");
            println!("   ID: {}", relationship_id.green());
        }
        Err(e) => {
            println!("❌ Failed to create relationship: {}", e.to_string().red());
        }
    }

    Ok(())
}

async fn handle_list_links(idea_id: String, db: &database::Database) -> Result<()> {
    // Verify idea exists
    let idea = db.get_by_id(&idea_id).await?;
    let idea = match idea {
        Some(idea) => idea,
        None => {
            println!("❌ Idea not found: {}", idea_id.red());
            return Ok(());
        }
    };

    let relationships = db.get_relationships_for_idea(&idea_id).await?;

    if relationships.is_empty() {
        println!("📭 No relationships found for idea: {}", idea_id.yellow());
        return Ok(());
    }

    println!(
        "🔗 Relationships for idea: {} ({})",
        idea_id.blue(),
        idea.content.chars().take(50).collect::<String>()
    );
    println!();

    for (i, rel) in relationships.iter().enumerate() {
        // Get the related idea details
        let related_idea_id = if rel.source_idea_id == idea_id {
            &rel.target_idea_id
        } else {
            &rel.source_idea_id
        };

        let related_idea = db.get_by_id(related_idea_id).await?;

        println!(
            "{}. {} {} {}",
            (i + 1).to_string().bright_blue(),
            rel.relationship_type.to_string().cyan(),
            "→".bright_black(),
            related_idea
                .map(|idea| idea.content.chars().take(60).collect::<String>())
                .unwrap_or_else(|| format!("(idea {} not found)", related_idea_id))
                .white()
        );
        println!(
            "   ID: {} ↔ {}",
            rel.source_idea_id.dimmed(),
            rel.target_idea_id.dimmed()
        );
        println!(
            "   Created: {}",
            rel.created_at.format("%Y-%m-%d %H:%M").to_string().dimmed()
        );
        println!();
    }

    Ok(())
}

async fn handle_show_related(
    idea_id: String,
    relationship_type: Option<String>,
    db: &database::Database,
) -> Result<()> {
    // Verify idea exists
    let idea = db.get_by_id(&idea_id).await?;
    let idea = match idea {
        Some(idea) => idea,
        None => {
            println!("❌ Idea not found: {}", idea_id.red());
            return Ok(());
        }
    };

    let rel_type_enum = if let Some(rel_type) = relationship_type {
        match rel_type.parse::<database::RelationshipType>() {
            Ok(rt) => Some(rt),
            Err(_) => {
                println!("❌ Invalid relationship type: '{}'", rel_type.red());
                println!("Valid types: depends_on, related_to, part_of, parent, child, duplicate, blocks, blocked_by, similar_to");
                return Ok(());
            }
        }
    } else {
        None
    };

    let related_ideas = db
        .get_related_ideas(&idea_id, rel_type_enum.clone())
        .await?;

    if related_ideas.is_empty() {
        println!("📭 No related ideas found for: {}", idea_id.yellow());
        return Ok(());
    }

    let filter_text = if let Some(rel_type) = &rel_type_enum {
        format!(" (filtered by: {})", rel_type.to_string().cyan())
    } else {
        "".to_string()
    };

    println!(
        "🔗 Related ideas for: {}{} ({})",
        idea_id.blue(),
        filter_text,
        idea.content.chars().take(50).collect::<String>()
    );
    println!();

    for (i, (related_idea, rel_type)) in related_ideas.iter().enumerate() {
        let score_display = if let Some(score) = related_idea.final_score {
            format!(" 📊 {:.1}/10", score)
        } else {
            "".to_string()
        };

        println!(
            "{}. {} {}",
            (i + 1).to_string().bright_blue(),
            rel_type.to_string().cyan(),
            related_idea.content.white()
        );
        println!(
            "   ID: {}{}",
            related_idea.id.dimmed(),
            score_display.dimmed()
        );
        println!(
            "   Created: {}",
            related_idea
                .created_at
                .format("%Y-%m-%d %H:%M")
                .to_string()
                .dimmed()
        );
        println!();
    }

    Ok(())
}

async fn handle_remove_link(
    relationship_id: String,
    confirm: bool,
    db: &database::Database,
) -> Result<()> {
    // Get details about this relationship
    // Since we don't have a direct getter, we'll just remove by ID

    if confirm {
        println!(
            ".unlink Removing relationship: {}",
            relationship_id.yellow()
        );

        let confirmed = Confirm::new()
            .with_prompt("Permanently remove this relationship?")
            .interact()?;

        if !confirmed {
            println!("❌ Removal cancelled.");
            return Ok(());
        }
    }

    match db.delete_relationship(&relationship_id).await {
        Ok(()) => {
            println!(
                "✅ Relationship removed successfully: {}",
                relationship_id.green()
            );
        }
        Err(e) => {
            println!("❌ Failed to remove relationship: {}", e.to_string().red());
        }
    }

    Ok(())
}

async fn handle_find_path(from: String, to: String, db: &database::Database) -> Result<()> {
    // For now, just a simple implementation that finds direct connections
    // A full pathfinding algorithm would be more complex

    println!("🔍 Finding path from {} to {}...", from.cyan(), to.cyan());

    // Check if there's a direct connection
    let from_idea = db.get_by_id(&from).await?;
    let to_idea = db.get_by_id(&to).await?;

    if from_idea.is_none() {
        println!("❌ Starting idea not found: {}", from.red());
        return Ok(());
    }

    if to_idea.is_none() {
        println!("❌ Target idea not found: {}", to.red());
        return Ok(());
    }

    let from_relationships = db.get_relationships_for_idea(&from).await?;

    // Check if 'to' is a direct relationship of 'from'
    let direct_connection = from_relationships
        .iter()
        .find(|rel| rel.source_idea_id == from || rel.target_idea_id == from);

    if let Some(rel) = direct_connection {
        if rel.source_idea_id == to || rel.target_idea_id == to {
            println!("🎯 Direct connection found!");
            println!(
                "  {} {} {}",
                from.blue(),
                rel.relationship_type.to_string().cyan(),
                to.blue()
            );
            return Ok(());
        }
    }

    // Find indirect paths (simple 2-step search for now)
    let mut paths_found = Vec::new();

    for rel in &from_relationships {
        let connected_idea_id = if rel.source_idea_id == from {
            rel.target_idea_id.clone()
        } else {
            rel.source_idea_id.clone()
        };

        // Check if this connected idea has a relationship with the target
        let connected_relationships = db.get_relationships_for_idea(&connected_idea_id).await?;
        for connected_rel in &connected_relationships {
            let check_id = if connected_rel.source_idea_id == connected_idea_id {
                connected_rel.target_idea_id.clone()
            } else {
                connected_rel.source_idea_id.clone()
            };

            if check_id == to {
                paths_found.push((
                    format!("{} -> {} -> {}", from, connected_idea_id, to),
                    format!(
                        "({} -> {} -> {})",
                        rel.relationship_type, connected_rel.relationship_type, "?"
                    ),
                ));
            }
        }
    }

    if !paths_found.is_empty() {
        println!("🔗 Indirect paths found:");
        for (path, details) in paths_found {
            println!("  {}", path.green());
            println!("  {}", details.dimmed());
        }
    } else {
        println!(
            "❌ No direct or 2-step path found between {} and {}",
            from.red(),
            to.red()
        );
        println!("💡 Try linking ideas that might connect these two concepts");
    }

    Ok(())
}
