use crate::errors::Result;
use crate::llm_fallback;

/// Handle `tm llm status` command
pub async fn handle_llm_status() -> Result<()> {
    println!("🔍 Checking Ollama status...\n");

    let status = llm_fallback::get_ollama_status().await;

    if status.running {
        println!("✅ Ollama is running");
        println!("   Endpoint: http://localhost:11434");

        if !status.models.is_empty() {
            println!("\n📦 Available models:");
            for model in &status.models {
                println!("   • {}", model);
            }
        } else {
            println!("\n⚠️  No models found. Pull a model with: ollama pull mistral");
        }
    } else {
        println!("❌ Ollama is not running");
        println!("\n💡 Start Ollama with: tm llm start");
    }

    Ok(())
}

/// Handle `tm llm start` command
pub async fn handle_llm_start() -> Result<()> {
    // Check if already running
    if llm_fallback::is_ollama_running().await {
        println!("ℹ️  Ollama is already running");
        return Ok(());
    }

    // Start Ollama
    llm_fallback::start_ollama().await?;

    // Show status
    let status = llm_fallback::get_ollama_status().await;
    if !status.models.is_empty() {
        println!("\n✅ Ready to use with {} model(s)", status.models.len());
    } else {
        println!("\n⚠️  No models installed. Pull one with:");
        println!("   ollama pull mistral");
    }

    Ok(())
}

/// Handle `tm llm stop` command
pub async fn handle_llm_stop() -> Result<()> {
    llm_fallback::stop_ollama().await
}
