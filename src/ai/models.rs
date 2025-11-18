// AI model configurations and data structures
// This file will contain specific model configurations for different LLMs

pub struct ModelConfig {
    pub name: String,
    pub temperature: f32,
    pub max_tokens: usize,
    pub system_prompt: Option<String>,
}

impl ModelConfig {
    pub fn llama_3_1_8b() -> Self {
        Self {
            name: "llama3.1:8b".to_string(),
            temperature: 0.3,
            max_tokens: 2000,
            system_prompt: Some(
                "You are an AI assistant helping with personal productivity and decision-making."
                    .to_string(),
            ),
        }
    }

    pub fn llama_3_1_70b() -> Self {
        Self {
            name: "llama3.1:70b".to_string(),
            temperature: 0.2,
            max_tokens: 3000,
            system_prompt: Some("You are an AI assistant specializing in personal productivity and behavioral pattern analysis.".to_string()),
        }
    }

    pub fn qwen_2_7b() -> Self {
        Self {
            name: "qwen2:7b".to_string(),
            temperature: 0.4,
            max_tokens: 1500,
            system_prompt: Some(
                "You are a helpful AI assistant for task management and decision support."
                    .to_string(),
            ),
        }
    }
}
