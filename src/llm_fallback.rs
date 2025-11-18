//! Simple LLM provider fallback using shell commands

use crate::errors::{ApplicationError, Result};
use std::time::Duration;
use tokio::process::Command;

pub async fn is_ollama_running() -> bool {
    let client = match reqwest::Client::builder()
        .timeout(Duration::from_millis(500))
        .build()
    {
        Ok(c) => c,
        Err(_) => return false,
    };

    match client.get("http://localhost:11434/api/tags").send().await {
        Ok(resp) => resp.status().is_success(),
        Err(_) => false,
    }
}

pub async fn is_claude_cli_available() -> bool {
    let which_cmd = if cfg!(target_os = "windows") {
        "where"
    } else {
        "which"
    };

    match Command::new(which_cmd).arg("claude").output().await {
        Ok(output) => output.status.success(),
        Err(_) => false,
    }
}

pub async fn analyze_with_claude_cli(idea: &str, prompt_template: &str) -> Result<String> {
    let full_prompt = format!(
        "{}\n\n---\n\nIdea to analyze:\n{}\n\nReturn your analysis. Use the JSON format specified above if possible, otherwise provide structured plain text.",
        prompt_template,
        idea
    );

    let mut child = Command::new("claude")
        .stdin(std::process::Stdio::piped())
        .stdout(std::process::Stdio::piped())
        .stderr(std::process::Stdio::piped())
        .spawn()
        .map_err(|e| ApplicationError::Generic(anyhow::anyhow!("Failed to spawn claude: {}", e)))?;

    if let Some(mut stdin) = child.stdin.take() {
        use tokio::io::AsyncWriteExt;
        stdin
            .write_all(full_prompt.as_bytes())
            .await
            .map_err(ApplicationError::Io)?;
        drop(stdin);
    }

    let output = tokio::time::timeout(Duration::from_secs(120), child.wait_with_output()).await;

    match output {
        Ok(Ok(result)) if result.status.success() => {
            let stdout = String::from_utf8_lossy(&result.stdout).to_string();
            Ok(stdout)
        }
        Ok(Ok(result)) => {
            let stderr = String::from_utf8_lossy(&result.stderr);
            Err(ApplicationError::Generic(anyhow::anyhow!(
                "Claude CLI failed: {}",
                stderr
            )))
        }
        Ok(Err(e)) => Err(ApplicationError::Generic(anyhow::anyhow!(
            "Failed to execute claude: {}",
            e
        ))),
        Err(_) => Err(ApplicationError::operation_timeout(120000, "Claude CLI")),
    }
}

pub fn extract_json(text: &str) -> Option<serde_json::Value> {
    if let Ok(json) = serde_json::from_str(text.trim()) {
        return Some(json);
    }

    if let Some(start_idx) = text.find("```json") {
        let after_start = &text[start_idx + 7..];
        if let Some(end_idx) = after_start.find("```") {
            let json_str = after_start[..end_idx].trim();
            if let Ok(json) = serde_json::from_str(json_str) {
                return Some(json);
            }
        }
    }

    if let Some(start_idx) = text.find("```") {
        let after_start = &text[start_idx + 3..];
        if let Some(end_idx) = after_start.find("```") {
            let json_str = after_start[..end_idx].trim();
            if let Some(newline) = json_str.find('\n') {
                let potential_json = &json_str[newline..].trim();
                if let Ok(json) = serde_json::from_str(potential_json) {
                    return Some(json);
                }
            }
            if let Ok(json) = serde_json::from_str(json_str) {
                return Some(json);
            }
        }
    }

    if let Some(start) = text.find('{') {
        if let Some(end) = text.rfind('}') {
            if start < end {
                let json_str = &text[start..=end];
                if let Ok(json) = serde_json::from_str(json_str) {
                    return Some(json);
                }
            }
        }
    }

    None
}

pub fn parse_plain_text_response(text: &str) -> Option<(f64, String)> {
    let score_patterns = [
        r"(?i)final[_\s]?score[:\s]+(\d+\.?\d*)",
        r"(?i)score[:\s]+(\d+\.?\d*)",
        r"(?i)rating[:\s]+(\d+\.?\d*)",
    ];

    let mut score = None;
    for pattern in &score_patterns {
        if let Ok(re) = regex::Regex::new(pattern) {
            if let Some(captures) = re.captures(text) {
                if let Some(score_match) = captures.get(1) {
                    if let Ok(s) = score_match.as_str().parse::<f64>() {
                        score = Some(s);
                        break;
                    }
                }
            }
        }
    }

    let recommendation = if text.contains("Priority") || text.contains("priority") {
        "Priority"
    } else if text.contains("Good") || text.contains("good fit") {
        "Good"
    } else if text.contains("Consider") || text.contains("consider") {
        "Consider"
    } else if text.contains("Avoid") || text.contains("avoid") {
        "Avoid"
    } else {
        "Consider"
    };

    score.map(|s| (s, recommendation.to_string()))
}

/// Start Ollama in background and wait for it to be ready
pub async fn start_ollama() -> Result<()> {
    println!("🚀 Starting Ollama...");

    // Spawn ollama serve in background
    let child = Command::new("ollama")
        .arg("serve")
        .stdout(std::process::Stdio::null())
        .stderr(std::process::Stdio::null())
        .spawn()
        .map_err(|e| {
            ApplicationError::Generic(anyhow::anyhow!(
                "Failed to start Ollama: {}. Is Ollama installed?",
                e
            ))
        })?;

    // Detach the process so it continues running
    drop(child);

    // Wait for Ollama to be ready (poll with 15s timeout)
    let start_time = std::time::Instant::now();
    let timeout = Duration::from_secs(15);

    loop {
        if start_time.elapsed() > timeout {
            return Err(ApplicationError::operation_timeout(
                15000,
                "Ollama startup - it may still be starting in background",
            ));
        }

        if is_ollama_running().await {
            println!("✅ Ollama is ready");
            return Ok(());
        }

        // Wait 500ms before next check
        tokio::time::sleep(Duration::from_millis(500)).await;
        print!(".");
        use std::io::Write;
        std::io::stdout().flush().ok();
    }
}

/// Stop any running Ollama instance
pub async fn stop_ollama() -> Result<()> {
    println!("🛑 Stopping Ollama...");

    // Kill any ollama process (macOS/Linux)
    let result = if cfg!(target_os = "windows") {
        Command::new("taskkill")
            .args(["/F", "/IM", "ollama.exe"])
            .output()
            .await
    } else {
        Command::new("pkill").arg("ollama").output().await
    };

    match result {
        Ok(output) if output.status.success() => {
            println!("✅ Ollama stopped");
            Ok(())
        }
        Ok(_) => {
            println!("ℹ️  Ollama was not running");
            Ok(())
        }
        Err(e) => Err(ApplicationError::Generic(anyhow::anyhow!(
            "Failed to stop Ollama: {}",
            e
        ))),
    }
}

/// Get Ollama status with details
pub async fn get_ollama_status() -> OllamaStatus {
    if is_ollama_running().await {
        // Try to get model list
        let client = reqwest::Client::builder()
            .timeout(Duration::from_secs(2))
            .build();

        if let Ok(client) = client {
            if let Ok(resp) = client.get("http://localhost:11434/api/tags").send().await {
                if let Ok(json) = resp.json::<serde_json::Value>().await {
                    let models = json["models"]
                        .as_array()
                        .map(|arr| {
                            arr.iter()
                                .filter_map(|m| m["name"].as_str().map(String::from))
                                .collect()
                        })
                        .unwrap_or_default();

                    return OllamaStatus {
                        running: true,
                        models,
                    };
                }
            }
        }

        OllamaStatus {
            running: true,
            models: vec![],
        }
    } else {
        OllamaStatus {
            running: false,
            models: vec![],
        }
    }
}

/// Ollama status information
#[derive(Debug, Clone)]
pub struct OllamaStatus {
    pub running: bool,
    pub models: Vec<String>,
}

#[cfg(test)]
mod tests {
    use super::*;

    #[tokio::test]
    async fn test_ollama_check_fast() {
        let start = std::time::Instant::now();
        let _ = is_ollama_running().await;
        assert!(start.elapsed().as_millis() < 1000);
    }

    #[test]
    fn test_extract_json_from_markdown() {
        let text = r#"Here's the analysis:
```json
{"final_score": 8.5, "recommendation": "Good"}
```
Done!"#;
        let json = extract_json(text);
        assert!(json.is_some());
        assert_eq!(json.unwrap()["final_score"], 8.5);
    }
}
