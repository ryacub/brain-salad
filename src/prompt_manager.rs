use std::sync::Arc;
use tokio::fs;
use tokio::sync::OnceCell;

static PROMPT_MANAGER: OnceCell<PromptManager> = OnceCell::const_new();

/// Manages the idea analysis prompt used for both internal scoring and LLM analysis
pub struct PromptManager {
    analysis_prompt: Arc<String>,
}

impl PromptManager {
    /// Creates a new PromptManager instance by loading the analysis prompt
    pub async fn new(prompt_file_path: &str) -> Result<Self, Box<dyn std::error::Error>> {
        let prompt_content = fs::read_to_string(prompt_file_path)
            .await
            .map_err(|e| format!("Failed to read prompt file '{}': {}", prompt_file_path, e))?;

        Ok(Self {
            analysis_prompt: Arc::new(prompt_content),
        })
    }

    /// Gets the analysis prompt content
    pub fn get_analysis_prompt(&self) -> Arc<String> {
        Arc::clone(&self.analysis_prompt)
    }

    /// Gets the analysis prompt as a string slice
    pub fn get_analysis_prompt_str(&self) -> &str {
        &self.analysis_prompt
    }
}

/// Returns a reference to the global PromptManager instance
pub async fn get_prompt_manager() -> Result<&'static PromptManager, Box<dyn std::error::Error>> {
    match PROMPT_MANAGER
        .get_or_init(|| async {
            match PromptManager::new("./IDEA_ANALYSIS_PROMPT.md").await {
                Ok(manager) => manager,
                Err(e) => {
                    // In case of error during initialization, we create a manager with an error message
                    // However, in a real scenario, we'd want to handle this more gracefully
                    PromptManager {
                        analysis_prompt: Arc::new(format!("Error loading prompt: {}", e)),
                    }
                }
            }
        })
        .await
    {
        manager if manager.analysis_prompt.starts_with("Error loading prompt:") => Err(format!(
            "Prompt manager failed to initialize: {}",
            manager.analysis_prompt
        )
        .into()),
        manager => Ok(manager),
    }
}

/// Initializes the global PromptManager instance with the actual prompt file content
pub async fn initialize_prompt_manager(
    prompt_file_path: &str,
) -> Result<(), Box<dyn std::error::Error>> {
    let prompt_manager = PromptManager::new(prompt_file_path).await?;
    PROMPT_MANAGER
        .set(prompt_manager)
        .map_err(|_| std::io::Error::other("PromptManager already initialized"))?;
    Ok(())
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::io::Write;
    use tempfile::NamedTempFile;

    #[tokio::test]
    async fn test_prompt_manager_creation() {
        // Create a temporary file with prompt content
        let mut temp_file = NamedTempFile::new().unwrap();
        let prompt_content = "Test analysis prompt content";
        temp_file.write_all(prompt_content.as_bytes()).unwrap();
        let temp_path = temp_file.path().to_string_lossy().to_string();

        // Test creation
        let manager = PromptManager::new(&temp_path).await.unwrap();

        assert_eq!(manager.get_analysis_prompt_str(), prompt_content);
    }

    #[tokio::test]
    async fn test_prompt_manager_file_not_found() {
        let result = PromptManager::new("non_existent_file.md").await;
        assert!(result.is_err());
    }
}
