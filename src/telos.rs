use crate::errors::Result;
use serde::{Deserialize, Serialize};
use std::borrow::Cow;
use std::path::Path;
use tokio::fs;

use crate::scoring::TelosConfig;

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ParsedTelos {
    pub problems: Vec<Problem>,
    pub missions: Vec<Mission>,
    pub goals: Vec<Goal>,
    pub challenges: Vec<Challenge>,
    pub strategies: Vec<Strategy>,
    pub current_stack: Vec<String>,
    pub domain_keywords: Vec<String>,
    pub last_updated: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Problem {
    pub id: String,
    pub title: String,
    pub description: String,
    pub why_it_matters: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Mission {
    pub id: String,
    pub title: String,
    pub description: String,
    pub actions: Vec<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Goal {
    pub id: String,
    pub title: String,
    pub description: String,
    pub metric: String,
    pub deadline: String,
    pub why: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Challenge {
    pub id: String,
    pub title: String,
    pub description: String,
    pub evidence: Option<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Strategy {
    pub id: String,
    pub title: String,
    pub description: String,
    pub implementation: Vec<String>,
    pub why_it_works: String,
}

pub struct TelosParser {
    telos_path: std::path::PathBuf,
}

impl TelosParser {
    pub fn new() -> Result<Self> {
        // Try to load config first, fallback to error as before
        match crate::config::ConfigPaths::load() {
            Ok(config) => Ok(Self::from_config(&config)),
            Err(e) => {
                eprintln!("Warning: Failed to load configuration, TelosParser::new() will not work properly: {}", e);
                // Return a parser with an empty path that will fail on parse - this matches previous behavior
                Err(crate::errors::ApplicationError::Configuration(
                    "TelosParser::new() requires configuration. Use TelosParser::with_path() or TelosParser::from_config() instead.".to_string()
                ))
            }
        }
    }

    pub fn with_path<P: AsRef<Path>>(telos_path: P) -> Self {
        Self {
            telos_path: telos_path.as_ref().to_path_buf(),
        }
    }

    pub fn from_config(config: &crate::config::ConfigPaths) -> Self {
        Self {
            telos_path: config.telos_file.clone(),
        }
    }

    pub async fn parse(&self) -> Result<ParsedTelos> {
        let content = fs::read_to_string(&self.telos_path)
            .await
            .map_err(crate::errors::ApplicationError::Io)?;

        // Extract sections using regex patterns
        let problems = self.extract_problems(&content)?;
        let missions = self.extract_missions(&content)?;
        let goals = self.extract_goals(&content)?;
        let challenges = self.extract_challenges(&content)?;
        let strategies = self.extract_strategies(&content)?;

        // Extract current stack and domain keywords
        let current_stack = self.extract_current_stack(&content)?;
        let domain_keywords = self.extract_domain_keywords(&content)?;

        // Get last updated date - use Cow to avoid allocation when possible
        let last_updated = self
            .extract_last_updated(&content)
            .unwrap_or(Cow::Borrowed("Unknown"));

        Ok(ParsedTelos {
            problems,
            missions,
            goals,
            challenges,
            strategies,
            current_stack,
            domain_keywords,
            last_updated: last_updated.into_owned(),
        })
    }

    pub async fn parse_for_scoring(&self) -> Result<TelosConfig> {
        let parsed = self.parse().await?;

        Ok(TelosConfig {
            current_stack: parsed.current_stack,
            domain_keywords: parsed.domain_keywords,
            income_deadline: parsed
                .goals
                .iter()
                .find(|g| g.id.contains("G1") || g.title.contains("income"))
                .map(|g| g.deadline.clone())
                .unwrap_or_else(|| "2026-01-15".to_string()),
            active_goals: parsed.goals.into_iter().map(|g| g.title).collect(),
            active_strategies: parsed.strategies.into_iter().map(|s| s.id).collect(),
            challenges: parsed.challenges.into_iter().map(|c| c.title).collect(),
        })
    }

    fn extract_problems(&self, content: &str) -> Result<Vec<Problem>> {
        let problems_section = self.extract_section(content, "PROBLEMS")?;

        let mut problems = Vec::new();
        let problem_blocks = self.extract_subsections(&problems_section, r"^### P\d+:");

        for (i, block) in problem_blocks.iter().enumerate() {
            let title = self
                .extract_heading(block)
                .unwrap_or(format!("Problem {}", i + 1));
            let description = self.extract_paragraphs_after_heading(block).join("\n");
            let why_it_matters = self
                .extract_section_content(block, "Why this matters:")
                .unwrap_or_default();

            problems.push(Problem {
                id: format!("P{}", i + 1),
                title,
                description,
                why_it_matters,
            });
        }

        Ok(problems)
    }

    fn extract_missions(&self, content: &str) -> Result<Vec<Mission>> {
        let missions_section = self.extract_section(content, "MISSIONS")?;

        let mut missions = Vec::new();
        let mission_blocks = self.extract_subsections(&missions_section, r"^### M\d+:");

        for (i, block) in mission_blocks.iter().enumerate() {
            let title = self
                .extract_heading(block)
                .unwrap_or(format!("Mission {}", i + 1));
            let description = self.extract_paragraphs_after_heading(block).join("\n");
            let actions = self.extract_list_items(block, "**Specific actions:**");

            missions.push(Mission {
                id: format!("M{}", i + 1),
                title,
                description,
                actions,
            });
        }

        Ok(missions)
    }

    fn extract_goals(&self, content: &str) -> Result<Vec<Goal>> {
        let goals_section = self.extract_section(content, "GOALS")?;

        let mut goals = Vec::new();
        let goal_blocks = self.extract_subsections(&goals_section, r"^### G\d+:");

        for (i, block) in goal_blocks.iter().enumerate() {
            let title = self
                .extract_heading(block)
                .unwrap_or(format!("Goal {}", i + 1));
            let description = self
                .extract_section_content(block, "- **What:**")
                .unwrap_or_default();
            let metric = self
                .extract_section_content(block, "- **Metric:**")
                .unwrap_or_default();
            let deadline = self
                .extract_section_content(block, "- **Deadline:**")
                .unwrap_or_default();
            let why = self
                .extract_section_content(block, "- **Why:**")
                .unwrap_or_default();

            goals.push(Goal {
                id: format!("G{}", i + 1),
                title,
                description,
                metric,
                deadline,
                why,
            });
        }

        Ok(goals)
    }

    fn extract_challenges(&self, content: &str) -> Result<Vec<Challenge>> {
        let challenges_section = self.extract_section(content, "CHALLENGES")?;

        let mut challenges = Vec::new();
        let challenge_blocks = self.extract_subsections(&challenges_section, r"^### C\d+:");

        for (i, block) in challenge_blocks.iter().enumerate() {
            let title = self
                .extract_heading(block)
                .unwrap_or(format!("Challenge {}", i + 1));
            let description = self.extract_paragraphs_after_heading(block).join("\n");
            let evidence = self.extract_section_content(block, "**Evidence:**");

            challenges.push(Challenge {
                id: format!("C{}", i + 1),
                title,
                description,
                evidence,
            });
        }

        Ok(challenges)
    }

    fn extract_strategies(&self, content: &str) -> Result<Vec<Strategy>> {
        let strategies_section = self.extract_section(content, "STRATEGIES")?;

        let mut strategies = Vec::new();
        let strategy_blocks = self.extract_subsections(&strategies_section, r"^### S\d+:");

        for (i, block) in strategy_blocks.iter().enumerate() {
            let title = self
                .extract_heading(block)
                .unwrap_or(format!("Strategy {}", i + 1));
            let description = self
                .extract_section_content(block, "**The Rule:**")
                .unwrap_or_default();
            let implementation = self.extract_list_items(block, "**Implementation:**");
            let why_it_works = self
                .extract_section_content(block, "**Why this works:**")
                .unwrap_or_default();

            strategies.push(Strategy {
                id: format!("S{}", i + 1),
                title,
                description,
                implementation,
                why_it_works,
            });
        }

        Ok(strategies)
    }

    fn extract_current_stack(&self, content: &str) -> Result<Vec<String>> {
        // Look for current stack in strategy S1
        let s1_content = content
            .find("S1: The \"One Stack, One Month\" Rule")
            .and_then(|pos| content[pos..].find("Implementation:"))
            .map(|pos| {
                let start = pos + content[pos..].find("Implementation:").unwrap();
                &content[start..start + 500] // Next 500 chars should contain stack info
            });

        if let Some(s1_text) = s1_content {
            // Extract the stack from November implementation
            if let Some(november_line) =
                s1_text.lines().find(|line| line.contains("November 2025:"))
            {
                let stack_part = november_line.split(':').nth(1).unwrap_or("");
                let stack_items: Vec<String> = stack_part
                    .split('+')
                    .map(|s| s.trim().to_lowercase())
                    .collect();

                if !stack_items.is_empty() {
                    return Ok(stack_items);
                }
            }
        }

        // Fallback to hardcoded current stack based on Ray's Telos
        Ok(vec![
            "python".to_string(),
            "langchain".to_string(),
            "openai".to_string(),
            "gpt".to_string(),
            "api".to_string(),
            "streamlit".to_string(),
            "web app".to_string(),
        ])
    }

    fn extract_domain_keywords(&self, content: &str) -> Result<Vec<String>> {
        let mut keywords = Vec::new();

        // Look for domain expertise in G1 description
        if content.contains("hotel")
            || content.contains("hospitality")
            || content.contains("Hilton")
        {
            keywords.extend_from_slice(&[
                "hotel".to_string(),
                "hospitality".to_string(),
                "hilton".to_string(),
            ]);
        }

        if content.contains("mobile") || content.contains("Android") || content.contains("app") {
            keywords.extend_from_slice(&[
                "mobile".to_string(),
                "android".to_string(),
                "app".to_string(),
                "application".to_string(),
            ]);
        }

        if content.contains("software")
            || content.contains("development")
            || content.contains("programming")
        {
            keywords.extend_from_slice(&[
                "software".to_string(),
                "development".to_string(),
                "programming".to_string(),
                "tech".to_string(),
            ]);
        }

        Ok(keywords)
    }

    fn extract_last_updated<'a>(&self, content: &'a str) -> Option<Cow<'a, str>> {
        // Look for "Last Updated:" in the header
        content
            .lines()
            .find(|line| line.contains("Last Updated:"))
            .and_then(|line| line.split(':').nth(1))
            .map(|date| {
                let trimmed = date.trim();
                if trimmed.is_empty() {
                    Cow::Borrowed("Unknown")
                } else {
                    Cow::Borrowed(trimmed)
                }
            })
    }

    // Helper methods
    fn extract_section(&self, content: &str, section_name: &str) -> Result<String> {
        let start_pattern = format!("## {}\n", section_name);
        let start = content
            .find(&start_pattern)
            .ok_or_else(|| anyhow::anyhow!("Section '{}' not found", section_name))?
            + start_pattern.len();

        let end = content[start..]
            .find("\n## ")
            .map(|pos| start + pos)
            .unwrap_or(content.len());

        Ok(content[start..end].trim().to_string())
    }

    fn extract_subsections(&self, section_content: &str, pattern: &str) -> Vec<String> {
        let regex = regex::Regex::new(pattern).unwrap();
        let mut subsections = Vec::new();

        let matches: Vec<_> = regex.find_iter(section_content).collect();
        for (i, mat) in matches.iter().enumerate() {
            let start = mat.start();
            let end = if i + 1 < matches.len() {
                matches[i + 1].start()
            } else {
                section_content.len()
            };

            subsections.push(section_content[start..end].trim().to_string());
        }

        subsections
    }

    fn extract_heading(&self, content: &str) -> Option<String> {
        content.lines().next().map(|line| line.trim().to_string())
    }

    fn extract_paragraphs_after_heading(&self, content: &str) -> Vec<String> {
        content
            .lines()
            .skip(1) // Skip heading
            .take_while(|line| {
                !line.starts_with("###") && !line.starts_with("##") && !line.starts_with("###")
            })
            .filter(|line| !line.trim().is_empty())
            .map(|line| line.trim().to_string())
            .collect()
    }

    fn extract_section_content(&self, content: &str, section_marker: &str) -> Option<String> {
        content
            .lines()
            .find(|line| line.contains(section_marker))
            .and_then(|line| {
                let parts: Vec<&str> = line.split(section_marker).collect();
                if parts.len() > 1 {
                    Some(parts[1].trim().to_string())
                } else {
                    None
                }
            })
    }

    fn extract_list_items(&self, content: &str, list_marker: &str) -> Vec<String> {
        if let Some(list_start) = content.find(list_marker) {
            let list_content = &content[list_start + list_marker.len()..];

            list_content
                .lines()
                .take_while(|line| line.trim().starts_with('-'))
                .map(|line| line.trim_start_matches('-').trim().to_string())
                .collect()
        } else {
            Vec::new()
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_extract_current_stack() {
        // Create a parser with a dummy path for testing
        let telos_parser = TelosParser::with_path("dummy.md");
        // This would require actual telos.md content to test properly
        // For now, we just verify the fallback works
        let stack = telos_parser.extract_current_stack("").unwrap();
        assert!(!stack.is_empty());
        assert!(stack.contains(&"python".to_string()));
    }

    #[test]
    fn test_extract_domain_keywords() {
        // Create a parser with a dummy path for testing
        let telos_parser = TelosParser::with_path("dummy.md");
        let content = "Worked at Hilton on mobile app development";
        let keywords = telos_parser.extract_domain_keywords(content).unwrap();

        assert!(keywords.contains(&"hotel".to_string()));
        assert!(keywords.contains(&"hilton".to_string()));
        assert!(keywords.contains(&"mobile".to_string()));
    }
}
