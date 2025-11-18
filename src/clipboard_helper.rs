//! Clipboard detection and paste support

use crate::errors::{ApplicationError, Result};
use arboard::Clipboard;
use std::io::Write;

/// Check if clipboard has substantial content (>= 10 chars) and prompt user
pub fn maybe_use_clipboard() -> Result<Option<String>> {
    let mut clipboard = Clipboard::new().map_err(|e| {
        ApplicationError::Generic(anyhow::anyhow!("Failed to access clipboard: {}", e))
    })?;

    match clipboard.get_text() {
        Ok(text) if is_substantial_content(&text) => prompt_to_use_clipboard(&text),
        Ok(_) => Ok(None),
        Err(_) => Ok(None),
    }
}

fn is_substantial_content(text: &str) -> bool {
    let trimmed = text.trim();

    if trimmed.len() < 10 {
        return false;
    }

    if trimmed.len() < 30 && !trimmed.contains('\n') {
        return false;
    }

    true
}

fn prompt_to_use_clipboard(text: &str) -> Result<Option<String>> {
    let char_count = text.len();
    let line_count = text.lines().count();

    println!("\n📋 Clipboard detected:");
    println!(
        "   {} characters, {} line{}",
        char_count,
        line_count,
        if line_count == 1 { "" } else { "s" }
    );

    let preview = get_preview(text, 100, 3);
    println!("   Preview: {}", preview);

    print!("\nUse clipboard content? [Y/n]: ");
    std::io::stdout().flush().map_err(ApplicationError::Io)?;

    let mut response = String::new();
    std::io::stdin()
        .read_line(&mut response)
        .map_err(ApplicationError::Io)?;

    let response = response.trim().to_lowercase();

    if response.is_empty() || response == "y" || response == "yes" {
        Ok(Some(text.to_string()))
    } else {
        Ok(None)
    }
}

fn get_preview(text: &str, max_chars: usize, max_lines: usize) -> String {
    let lines: Vec<&str> = text.lines().take(max_lines).collect();
    let preview = lines.join(" ");

    if preview.len() > max_chars {
        format!("{}...", &preview[..max_chars])
    } else if text.lines().count() > max_lines {
        format!("{}...", preview)
    } else {
        preview
    }
}

pub fn get_clipboard_content() -> Result<String> {
    let mut clipboard = Clipboard::new().map_err(|e| {
        ApplicationError::Generic(anyhow::anyhow!("Failed to access clipboard: {}", e))
    })?;

    clipboard
        .get_text()
        .map_err(|e| ApplicationError::Generic(anyhow::anyhow!("Failed to read clipboard: {}", e)))
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_substantial_content() {
        assert!(!is_substantial_content("short"));
        assert!(!is_substantial_content("https://example.com"));
        assert!(is_substantial_content(
            "This is a longer piece of text that should qualify"
        ));
        assert!(is_substantial_content("Multi\nline\ntext\nis\nsubstantial"));
    }

    #[test]
    fn test_preview() {
        let text = "Line 1\nLine 2\nLine 3\nLine 4\nLine 5";
        let preview = get_preview(text, 100, 3);
        assert!(preview.contains("Line 1"));
        assert!(preview.contains("Line 2"));
        assert!(preview.contains("Line 3"));
        assert!(!preview.contains("Line 4"));
    }

    #[test]
    fn test_preview_truncate_chars() {
        let long_text = "a".repeat(200);
        let preview = get_preview(&long_text, 50, 10);
        assert_eq!(preview.len(), 53);
    }
}
