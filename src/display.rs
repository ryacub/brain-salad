use colored::*;
use std::borrow::Cow;

use crate::patterns_simple::PatternMatch;
use crate::scoring::{Recommendation, Score};

pub fn display_analysis_result(
    idea: &str,
    score: &Score,
    patterns: &[PatternMatch],
    idea_id: Option<&str>,
) {
    println!();
    println!("🎯 {}", score.recommendation.text().bright_blue().bold());
    println!("────────────────────────────────────────");
    println!("{}", idea.chars().take(60).collect::<String>());
    if idea.len() > 60 {
        println!("     ...");
    }

    if let Some(id) = idea_id {
        println!("🆔 {}", id.dimmed());
    }

    // Show where the idea works
    println!("\n✅ WHERE THIS IDEA WORKS:");
    if score.mission.ai_alignment > 0.5 {
        println!(
            "   • AI Alignment: {:.1}/1.5 - {}",
            score.mission.ai_alignment,
            get_ai_explanation(score.mission.ai_alignment)
        );
    }
    if score.mission.domain_expertise > 0.5 {
        println!(
            "   • Domain Expertise: {:.1}/1.2 - {}",
            score.mission.domain_expertise,
            get_domain_explanation(score.mission.domain_expertise)
        );
    }
    if score.anti_challenge.rapid_prototyping > 0.5 {
        println!(
            "   • Prototyping: {:.1}/1.0 - {}",
            score.anti_challenge.rapid_prototyping,
            get_perfectionism_explanation(score.anti_challenge.rapid_prototyping)
        );
    }
    if score.mission.revenue_potential > 0.3 {
        println!(
            "   • Revenue: {:.1}/0.5 - {}",
            score.mission.revenue_potential,
            get_income_explanation(score.mission.revenue_potential)
        );
    }

    // Show critical issues
    println!("\n❌ WHERE THIS IDEA FALLS SHORT:");
    if score.anti_challenge.context_switching < 0.5 {
        println!(
            "   • Context Switching: {:.1}/1.2 - {}",
            score.anti_challenge.context_switching,
            get_context_explanation(score.anti_challenge.context_switching)
        );
    }
    if score.strategic.stack_compatibility < 0.5 {
        println!(
            "   • Stack Fit: {:.1}/1.0 - {}",
            score.strategic.stack_compatibility,
            get_stack_explanation(score.strategic.stack_compatibility)
        );
    }
    if score.anti_challenge.accountability < 0.4 {
        println!(
            "   • Accountability: {:.1}/0.8 - {}",
            score.anti_challenge.accountability,
            get_accountability_explanation(score.anti_challenge.accountability)
        );
    }

    // Pattern alerts
    let critical_patterns: Vec<_> = patterns
        .iter()
        .filter(|p| matches!(p.severity, crate::patterns_simple::Severity::Critical))
        .collect();
    let high_patterns: Vec<_> = patterns
        .iter()
        .filter(|p| matches!(p.severity, crate::patterns_simple::Severity::High))
        .collect();
    if !critical_patterns.is_empty() || !high_patterns.is_empty() {
        println!("\n⚠️  CRITICAL PATTERNS DETECTED:");
        for pattern in &critical_patterns {
            println!(
                "   {} {}: {}",
                "🔴".red(),
                pattern.pattern_type.title(),
                pattern.message
            );
            if let Some(suggestion) = &pattern.suggestion {
                println!("      💡 {}", suggestion.bright_blue());
            }
        }
        for pattern in &high_patterns {
            println!(
                "   {} {}: {}",
                "🟠".bright_red(),
                pattern.pattern_type.title(),
                pattern.message
            );
            if let Some(suggestion) = &pattern.suggestion {
                println!("      💡 {}", suggestion.bright_blue());
            }
        }
    }

    // Total score
    println!("\n📊 Overall Score: {:.1}/10.0", score.final_score);

    // Recommendation
    print_recommendation_section(&score.recommendation);
    println!("────────────────────────────────────────\n");
}

pub fn display_quick_save(idea_id: &str) {
    println!("✅ Idea saved (ID: {})", idea_id.green());
}

pub fn display_idea_list(ideas: &[crate::database_simple::StoredIdea]) {
    if ideas.is_empty() {
        println!("📭 {}", "No ideas found.".dimmed());
        return;
    }

    println!("📋 {}", "Your Ideas:".bright_blue());
    println!();

    for (i, idea) in ideas.iter().enumerate() {
        println!(
            "{}. {}",
            (i + 1).to_string().bright_white(),
            idea.content.bright_white()
        );

        if let Some(score) = idea.final_score {
            let colored_score = match score {
                s if s >= 8.0 => format!("{:.1}", s).green(),
                s if s >= 6.0 => format!("{:.1}", s).yellow(),
                s if s >= 4.0 => format!("{:.1}", s).yellow(),
                s => format!("{:.1}", s).red(),
            };
            println!("   📊 Score: {}/10", colored_score);
        }

        if let Some(recommendation) = &idea.recommendation {
            println!("   🎯 {}", recommendation);
        }

        println!(
            "   📅 Created: {}",
            idea.created_at.format("%Y-%m-%d %H:%M")
        );

        if let Some(patterns) = &idea.patterns {
            if !patterns.is_empty() {
                println!("   ⚠️  Patterns: {}", patterns.join(", ").yellow());
            }
        }

        println!("   🆔 ID: {}", idea.id.dimmed());
        println!();
    }
}

fn print_recommendation_section(recommendation: &Recommendation) {
    let (title, color) = match recommendation {
        Recommendation::Priority => ("🔥 PRIORITY NOW - Do this immediately!", "green"),
        Recommendation::Good => ("✅ GOOD ALIGNMENT - Strong candidate for action", "blue"),
        Recommendation::Consider => ("⚠️ CONSIDER LATER - Not bad timing", "yellow"),
        Recommendation::Avoid => ("🚫 AVOID FOR NOW - Wrong timing/context", "red"),
    };

    // Parse color with fallback to white if parsing fails
    let parsed_color = color
        .parse::<colored::Color>()
        .unwrap_or(colored::Color::White);
    println!("{}", title.bright_white().on_color(parsed_color));
    println!();
}

// Helper functions for explanations
fn get_ai_explanation(score: f64) -> &'static str {
    match score {
        s if s >= 1.2 => "Direct AI implementation",
        s if s >= 0.8 => "AI component present",
        s if s >= 0.4 => "AI mentioned",
        _ => "No clear AI component",
    }
}

fn get_income_explanation(score: f64) -> &'static str {
    match score {
        s if s >= 0.4 => "Clear income potential",
        s if s >= 0.2 => "Some business value",
        _ => "No clear income path",
    }
}

fn get_domain_explanation(score: f64) -> &'static str {
    match score {
        s if s >= 1.0 => "Leverages domain expertise",
        s if s >= 0.6 => "Related to tech domain",
        _ => "No domain leverage",
    }
}

fn get_context_explanation(score: f64) -> &'static str {
    match score {
        s if s >= 1.0 => "Perfect stack compliance",
        s if s >= 0.6 => "Mostly compliant",
        _ => "Stack violation risk",
    }
}

fn get_perfectionism_explanation(score: f64) -> &'static str {
    match score {
        s if s >= 0.8 => "Healthy scoping",
        s if s >= 0.4 => "Some scope control",
        _ => "Perfectionism risk",
    }
}

fn get_accountability_explanation(score: f64) -> &'static str {
    match score {
        s if s >= 0.6 => "External accountability",
        s if s >= 0.3 => "Some external factor",
        _ => "Solo project",
    }
}

fn get_stack_explanation(score: f64) -> &'static str {
    match score {
        s if s >= 0.8 => "Uses current stack",
        _ => "Stack mismatch",
    }
}
