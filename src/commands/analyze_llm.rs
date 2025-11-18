use crate::database_simple as database;
use crate::errors::Result;
use crate::llm_cache::LlmResponseCache;
use crate::patterns_simple::PatternDetector;
use crate::prompt_templates::{classify_idea_type, get_prompt_builder};
use crate::quality_metrics_simple::{get_quality_tracker, SimpleQualityMetrics};
use crate::response_processing::EnhancedLlmAnalysisResult;
use crate::response_processing::{
    process_llm_response, process_llm_response_with_quality, FallbackScorer,
};
use crate::scoring::ScoringEngine;
use chrono;
use reqwest;
use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use std::time::Duration;
use std::time::Instant;
use tokio::time::timeout;

/// Structure representing the analysis result from LLM
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct LlmAnalysisResult {
    pub scores: LlmScores,
    pub weighted_totals: LlmWeightedTotals,
    pub final_score: f64,
    pub recommendation: String,
    pub explanations: HashMap<String, String>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct LlmScores {
    #[serde(rename = "Mission Alignment")]
    pub mission_alignment: MissionAlignmentScores,
    #[serde(rename = "Anti-Challenge Patterns")]
    pub anti_challenge_patterns: AntiChallengePatternsScores,
    #[serde(rename = "Strategic Fit")]
    pub strategic_fit: StrategicFitScores,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct MissionAlignmentScores {
    #[serde(rename = "Domain Expertise")]
    pub domain_expertise: f64,
    #[serde(rename = "AI Alignment")]
    pub ai_alignment: f64,
    #[serde(rename = "Execution Support")]
    pub execution_support: f64,
    #[serde(rename = "Revenue Potential")]
    pub revenue_potential: f64,
    pub category_total: f64,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct AntiChallengePatternsScores {
    #[serde(rename = "Avoid Context-Switching")]
    pub avoid_context_switching: f64,
    #[serde(rename = "Rapid Prototyping")]
    pub rapid_prototyping: f64,
    #[serde(rename = "Accountability")]
    pub accountability: f64,
    #[serde(rename = "Income Anxiety")]
    pub income_anxiety: f64,
    pub category_total: f64,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct StrategicFitScores {
    #[serde(rename = "Stack Compatibility")]
    pub stack_compatibility: f64,
    #[serde(rename = "Shipping Habit")]
    pub shipping_habit: f64,
    #[serde(rename = "Public Accountability")]
    pub public_accountability: f64,
    #[serde(rename = "Revenue Testing")]
    pub revenue_testing: f64,
    pub category_total: f64,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct LlmWeightedTotals {
    #[serde(rename = "Mission Alignment")]
    pub mission_alignment: f64,
    #[serde(rename = "Anti-Challenge Patterns")]
    pub anti_challenge_patterns: f64,
    #[serde(rename = "Strategic Fit")]
    pub strategic_fit: f64,
}

/// Enum for different LLM providers
#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum LlmProvider {
    OpenAi,
    Claude,
    Ollama,
    Custom,
}

impl LlmProvider {
    /// Get the provider type as a string
    pub fn provider_type(&self) -> String {
        match self {
            LlmProvider::OpenAi => "OpenAI".to_string(),
            LlmProvider::Claude => "Claude".to_string(),
            LlmProvider::Ollama => "Ollama".to_string(),
            LlmProvider::Custom => "Custom".to_string(),
        }
    }
}

/// Configuration for LLM analysis
#[derive(Debug, Clone)]
pub struct LlmConfig {
    pub provider: LlmProvider,
    pub model: String,
    pub api_key: Option<String>,
    pub base_url: Option<String>,
    pub temperature: f32,
    pub max_tokens: u32,
    pub timeout_seconds: u64,
}

impl Default for LlmConfig {
    fn default() -> Self {
        Self {
            provider: LlmProvider::Ollama,
            model: "mistral".to_string(),
            api_key: None,
            base_url: None,
            temperature: 0.3,
            max_tokens: 4096,
            timeout_seconds: 60,
        }
    }
}

/// Main handler for the analyze-llm command
pub async fn handle_analyze_llm(
    idea: String,
    config: LlmConfig,
    db: &database::Database,
    _scoring_engine: &ScoringEngine,
    _pattern_detector: &PatternDetector,
    save_to_db: bool,
) -> Result<()> {
    if idea.trim().is_empty() {
        return Err(crate::errors::ApplicationError::validation(
            "Idea content cannot be empty",
        ));
    }

    let start_time = Instant::now();

    println!(
        "🤖 Enhanced LLM Analysis: {}",
        idea.chars().take(50).collect::<String>()
    );
    if idea.len() > 50 {
        println!("... ({} more characters)", idea.len() - 50);
    }

    // Initialize our enhanced systems
    let prompt_builder = get_prompt_builder();
    let cache = LlmResponseCache::new();
    let quality_tracker = get_quality_tracker();

    // Classify the idea type for dynamic prompt selection
    let idea_type = classify_idea_type(&idea);
    println!(
        "📋 Idea type: {} | Provider: {}",
        idea_type,
        config.provider.provider_type()
    );

    // Check cache first for existing analysis
    println!("🔍 Checking cache for similar ideas...");
    let cache_key = format!("{}_v1.0", idea_type); // Using template version as key
    if let Some((cached_result, similarity_score)) = cache
        .get_response(&idea, &config.provider, &cache_key)
        .await?
    {
        let response_time_ms = start_time.elapsed().as_millis() as u64;
        println!(
            "✅ Cache hit! Similarity: {:.2}, Response time: {}ms",
            similarity_score, response_time_ms
        );

        // Record cache hit in quality metrics
        let metrics = SimpleQualityMetrics {
            idea: idea.clone(),
            timestamp: chrono::Utc::now(),
            provider: config.provider.clone(),
            template_id: format!("{}_v1.0", idea_type),
            idea_type: idea_type.clone(),
            from_cache: true,
            cache_similarity: Some(similarity_score),
            quality_score: 0.9, // High confidence for cache hits
            confidence_level: "High".to_string(),
            fallback_used: false,
            response_time_ms,
            final_score: cached_result.final_score,
            recommendation: cached_result.recommendation.clone(),
        };

        quality_tracker.record_analysis(metrics).await?;

        // Display cached result
        display_enhanced_llm_analysis_result(&idea, &cached_result, Some(similarity_score), true);

        if save_to_db {
            let raw_analysis_json = serde_json::to_string(&cached_result)?;
            db.save_idea(
                &idea,
                Some(cached_result.final_score),
                Some(cached_result.final_score),
                None,
                Some(cached_result.recommendation.clone()),
                Some(raw_analysis_json),
            )
            .await?;
            println!("✅ Cached analysis saved to database");
        }

        return Ok(());
    }

    println!("⚡ No cache hit, performing fresh LLM analysis...");

    // Build dynamic prompt with examples and chain-of-thought
    let include_examples = true; // Enable few-shot learning
    let include_cot = true; // Enable chain-of-thought reasoning
    let dynamic_prompt = prompt_builder.build_prompt(
        &idea,
        &idea_type,
        &config.provider,
        include_examples,
        include_cot,
    );

    println!(
        "📝 Using dynamic prompt with {} examples and {} reasoning steps",
        if include_examples { "few-shot" } else { "no" },
        if include_cot {
            "chain-of-thought"
        } else {
            "no"
        }
    );
    println!(
        "📋 Dynamic prompt preview (first 200 chars): {}",
        &dynamic_prompt[..std::cmp::min(200, dynamic_prompt.len())]
    );

    // Perform enhanced LLM analysis
    let enhanced_result = perform_enhanced_llm_analysis(&dynamic_prompt, &idea, &config).await?;
    let response_time_ms = start_time.elapsed().as_millis() as u64;

    // Cache the new analysis result
    println!("💾 Caching new analysis result...");
    cache
        .cache_response(
            &idea,
            &enhanced_result.base_result,
            &config.provider,
            &cache_key,
            &format!("{:?}", enhanced_result.confidence_level),
            enhanced_result.quality_metrics.quality_score,
        )
        .await?;

    // Record quality metrics
    let metrics = SimpleQualityMetrics {
        idea: idea.clone(),
        timestamp: chrono::Utc::now(),
        provider: config.provider.clone(),
        template_id: format!("{}_v1.0", idea_type),
        idea_type: idea_type.clone(),
        from_cache: false,
        cache_similarity: None,
        quality_score: enhanced_result.quality_metrics.quality_score,
        confidence_level: format!("{:?}", enhanced_result.confidence_level),
        fallback_used: enhanced_result.quality_metrics.fallback_used,
        response_time_ms,
        final_score: enhanced_result.base_result.final_score,
        recommendation: enhanced_result.base_result.recommendation.clone(),
    };

    quality_tracker.record_analysis(metrics).await?;

    // Display the enhanced analysis result
    display_enhanced_llm_analysis_result(&idea, &enhanced_result.base_result, None, false);

    // Optionally save to database
    if save_to_db {
        let raw_analysis_json = serde_json::to_string(&enhanced_result.base_result)?;

        db.save_idea(
            &idea,
            Some(enhanced_result.base_result.final_score),
            Some(enhanced_result.base_result.final_score),
            None, // patterns - we'll set this to None for LLM analysis
            Some(enhanced_result.base_result.recommendation.clone()),
            Some(raw_analysis_json),
        )
        .await?;
        println!("✅ Enhanced analysis saved to database");
    }

    // Display quality summary
    println!("\n📊 Quality Summary:");
    println!("  Response Time: {}ms", response_time_ms);
    println!(
        "  Confidence: {} ({})",
        enhanced_result.confidence_level.emoji(),
        enhanced_result.confidence_level.score()
    );
    println!(
        "  Quality Score: {:.1}%",
        enhanced_result.quality_metrics.quality_score * 100.0
    );
    println!("  Cache Similarity: N/A (fresh analysis)");

    Ok(())
}

/// Substitute the idea text into the prompt template
fn substitute_idea_into_prompt(prompt_template: &str, idea: &str) -> String {
    // The prompt template from IDEA_ANALYSIS_PROMPT.md should already have
    // instructions for analyzing an idea, so we just need to include the idea text
    format!(
        "{}\n\nIDEA TO ANALYZE: {}\n\nPlease provide your analysis in the specified JSON format:",
        prompt_template, idea
    )
}

/// Perform enhanced LLM analysis with quality metrics
async fn perform_enhanced_llm_analysis(
    prompt: &str,
    idea: &str,
    config: &LlmConfig,
) -> Result<EnhancedLlmAnalysisResult> {
    let base_result = match &config.provider {
        LlmProvider::OpenAi => call_openai_api_with_idea(prompt, idea, config).await?,
        LlmProvider::Claude => call_claude_api_with_idea(prompt, idea, config).await?,
        LlmProvider::Ollama => call_ollama_api_with_idea(prompt, idea, config).await?,
        LlmProvider::Custom => call_custom_api_with_idea(prompt, idea, config).await?,
    };

    // Process with enhanced quality analysis
    let response_text = serde_json::to_string(&base_result)?;
    let fallback_scorer = FallbackScorer::new();
    process_llm_response_with_quality(&response_text, idea, &fallback_scorer).await
}

/// Perform the actual LLM analysis (legacy)
async fn perform_llm_analysis(
    prompt: &str,
    idea: &str,
    config: &LlmConfig,
) -> Result<LlmAnalysisResult> {
    match &config.provider {
        LlmProvider::OpenAi => call_openai_api_with_idea(prompt, idea, config).await,
        LlmProvider::Claude => call_claude_api_with_idea(prompt, idea, config).await,
        LlmProvider::Ollama => call_ollama_api_with_idea(prompt, idea, config).await,
        LlmProvider::Custom => call_custom_api_with_idea(prompt, idea, config).await,
    }
}

/// Call OpenAI API with proper error handling and fallback notification
async fn call_openai_api_with_idea(
    prompt: &str,
    idea: &str,
    config: &LlmConfig,
) -> Result<LlmAnalysisResult> {
    println!("🌐 Calling OpenAI API with model: {}", config.model);

    // Check if OpenAI API key is available
    let api_key = match config.api_key.as_ref() {
        Some(key) => {
            if key.is_empty() {
                println!("⚠️  OpenAI API key is empty, using fallback analysis");
                return generate_fallback_analysis(idea).await;
            }
            key
        }
        None => {
            println!("⚠️  OpenAI API key not configured, using fallback analysis");
            return generate_fallback_analysis(idea).await;
        }
    };

    println!("📝 Sending prompt ({} chars) to OpenAI...", prompt.len());
    println!(
        "🔑 Using API key: {}...",
        &api_key[..std::cmp::min(8, api_key.len())]
    );

    let client = reqwest::Client::new();

    let request_body = serde_json::json!({
        "model": config.model,
        "messages": [
            {"role": "user", "content": prompt}
        ],
        "temperature": config.temperature,
        "max_tokens": config.max_tokens
    });

    println!("📤 Making request to OpenAI...");

    let response_result = timeout(
        Duration::from_secs(config.timeout_seconds),
        client
            .post("https://api.openai.com/v1/chat/completions")
            .header("Authorization", format!("Bearer {}", api_key))
            .header("Content-Type", "application/json")
            .json(&request_body)
            .send(),
    )
    .await;

    let response = match response_result {
        Ok(resp) => resp,
        Err(_) => {
            println!(
                "⚠️  OpenAI API request timed out after {} seconds",
                config.timeout_seconds
            );
            println!("🔄 Falling back to rule-based analysis");
            return generate_fallback_analysis(idea).await;
        }
    };

    let response = match response {
        Ok(resp) => resp,
        Err(e) => {
            println!("⚠️  OpenAI API request failed: {}", e);
            println!("🔄 Falling back to rule-based analysis");
            return generate_fallback_analysis(idea).await;
        }
    };

    println!("📥 Received response from OpenAI");

    let response_text = match response.text().await {
        Ok(text) => {
            println!("📄 Response received ({} chars)", text.len());
            text
        }
        Err(e) => {
            println!("⚠️  Failed to read response text: {}", e);
            println!("🔄 Falling back to rule-based analysis");
            return generate_fallback_analysis(idea).await;
        }
    };

    println!("🔍 Processing OpenAI response...");

    // Process the response with validation and fallback
    match process_llm_response(&response_text, idea).await {
        Ok(analysis_result) => {
            println!("✅ Successfully processed OpenAI response");
            Ok(analysis_result)
        }
        Err(e) => {
            println!("⚠️  Failed to process OpenAI response: {}", e);
            println!("🔄 Falling back to rule-based analysis");
            generate_fallback_analysis(idea).await
        }
    }
}

/// Generate fallback analysis using the scoring engine
async fn generate_fallback_analysis(idea: &str) -> Result<LlmAnalysisResult> {
    println!("⚙️  Generating rule-based fallback analysis...");

    // Create a simple fallback analysis based on idea characteristics
    let (final_score, recommendation, explanations) = analyze_idea_with_rules(idea);

    // Create a basic scoring structure
    let analysis_result = LlmAnalysisResult {
        scores: LlmScores {
            mission_alignment: MissionAlignmentScores {
                domain_expertise: final_score * 0.4,
                ai_alignment: final_score * 0.35,
                execution_support: final_score * 0.15,
                revenue_potential: final_score * 0.10,
                category_total: final_score * 0.4,
            },
            anti_challenge_patterns: AntiChallengePatternsScores {
                avoid_context_switching: final_score * 0.35,
                rapid_prototyping: final_score * 0.25,
                accountability: final_score * 0.25,
                income_anxiety: final_score * 0.15,
                category_total: final_score * 0.35,
            },
            strategic_fit: StrategicFitScores {
                stack_compatibility: final_score * 0.4,
                shipping_habit: final_score * 0.3,
                public_accountability: final_score * 0.2,
                revenue_testing: final_score * 0.1,
                category_total: final_score * 0.25,
            },
        },
        weighted_totals: LlmWeightedTotals {
            mission_alignment: final_score * 0.16,        // 40% of 40%
            anti_challenge_patterns: final_score * 0.123, // 35% of 35%
            strategic_fit: final_score * 0.063,           // 25% of 25%
        },
        final_score,
        recommendation: recommendation.to_string(),
        explanations,
    };

    println!(
        "✅ Fallback analysis generated (Score: {:.2}, Recommendation: {})",
        final_score, recommendation
    );
    Ok(analysis_result)
}

/// Simple rule-based idea analysis
fn analyze_idea_with_rules(idea: &str) -> (f64, &'static str, HashMap<String, String>) {
    let idea_lower = idea.to_lowercase();
    let mut explanations = HashMap::new();

    // Basic scoring based on idea characteristics
    let mut score: f64 = 5.0; // Base score

    // Positive indicators
    if idea_lower.contains("build") || idea_lower.contains("create") || idea_lower.contains("make")
    {
        score += 1.0;
        explanations.insert(
            "Domain Expertise".to_string(),
            "Active building language suggests tangible project".to_string(),
        );
    }

    if idea_lower.contains("learn") || idea_lower.contains("study") || idea_lower.contains("master")
    {
        score += 0.5;
        explanations.insert(
            "AI Alignment".to_string(),
            "Learning projects align with skill development".to_string(),
        );
    }

    if idea_lower.contains("app") || idea_lower.contains("tool") || idea_lower.contains("software")
    {
        score += 1.0;
        explanations.insert(
            "Stack Compatibility".to_string(),
            "Software projects leverage existing technical skills".to_string(),
        );
    }

    if idea_lower.contains("quick") || idea_lower.contains("simple") || idea_lower.contains("small")
    {
        score += 0.5;
        explanations.insert(
            "Rapid Prototyping".to_string(),
            "Projects with scope modifiers suggest manageable implementation".to_string(),
        );
    }

    // Negative indicators
    if idea_lower.contains("huge")
        || idea_lower.contains("complex")
        || idea_lower.contains("massive")
    {
        score -= 1.0;
        explanations.insert(
            "Execution Support".to_string(),
            "Complex scope may hinder execution".to_string(),
        );
    }

    if idea.len() > 100 {
        score -= 0.5;
        explanations.insert(
            "Context Switching".to_string(),
            "Longer descriptions may indicate overthinking".to_string(),
        );
    }

    // Clamp score
    score = score.max(0.0).min(10.0);

    // Determine recommendation
    let recommendation = if score >= 8.0 {
        "Priority"
    } else if score >= 6.5 {
        "Good"
    } else if score >= 5.0 {
        "Consider"
    } else {
        "Avoid"
    };

    if explanations.is_empty() {
        explanations.insert(
            "General Assessment".to_string(),
            "Basic rule-based evaluation applied".to_string(),
        );
    }

    (score, recommendation, explanations)
}

/// Call Claude API
async fn call_claude_api_with_idea(
    prompt: &str,
    idea: &str,
    config: &LlmConfig,
) -> Result<LlmAnalysisResult> {
    // Check if Claude API key is available
    let api_key = config.api_key.as_ref().ok_or_else(|| {
        crate::errors::ApplicationError::Configuration("Claude API key not provided".to_string())
    })?;

    let client = reqwest::Client::new();

    let request_body = serde_json::json!({
        "model": config.model,
        "messages": [
            {"role": "user", "content": prompt}
        ],
        "temperature": config.temperature,
        "max_tokens": config.max_tokens as usize
    });

    let response = timeout(
        Duration::from_secs(config.timeout_seconds),
        client
            .post("https://api.anthropic.com/v1/messages")
            .header("x-api-key", api_key)
            .header("Content-Type", "application/json")
            .header("anthropic-version", "2023-06-01")
            .json(&request_body)
            .send(),
    )
    .await
    .map_err(|_| crate::errors::ApplicationError::OperationTimeout {
        timeout_ms: config.timeout_seconds * 1000,
        context: "Claude API call".to_string(),
    })?
    .map_err(|e| crate::errors::ApplicationError::Generic(e.into()))?;

    let response_text = response
        .text()
        .await
        .map_err(|e| crate::errors::ApplicationError::Generic(e.into()))?;

    // Process the response with validation and fallback
    let analysis_result = process_llm_response(&response_text, idea)
        .await
        .map_err(|e| crate::errors::ApplicationError::Generic(e.into()))?;

    Ok(analysis_result)
}

/// Call Ollama API with verbose logging and error handling
async fn call_ollama_api_with_idea(
    prompt: &str,
    idea: &str,
    config: &LlmConfig,
) -> Result<LlmAnalysisResult> {
    println!("🦙 Calling Ollama API with model: {}", config.model);
    println!("📝 Sending prompt ({} chars) to Ollama...", prompt.len());

    use ollama_rs::{generation::completion::request::GenerationRequest, Ollama};

    let ollama = Ollama::default();

    let request = GenerationRequest::new(config.model.clone(), prompt.to_string()).options(
        ollama_rs::generation::options::GenerationOptions::default()
            .temperature(config.temperature)
            .num_predict(config.max_tokens as i32),
    );

    println!("📤 Making request to Ollama...");

    let response_result = timeout(
        Duration::from_secs(config.timeout_seconds),
        ollama.generate(request),
    )
    .await;

    let response = match response_result {
        Ok(resp) => {
            println!("📥 Received response from Ollama");
            resp
        }
        Err(_) => {
            println!(
                "⚠️  Ollama API request timed out after {} seconds",
                config.timeout_seconds
            );
            println!("🔄 Falling back to rule-based analysis");
            return generate_fallback_analysis(idea).await;
        }
    };

    let response = match response {
        Ok(resp) => resp,
        Err(e) => {
            println!("⚠️  Ollama API request failed: {}", e);
            println!("🔄 Falling back to rule-based analysis");
            return generate_fallback_analysis(idea).await;
        }
    };

    let response_text = response.response;
    println!("📄 Response received ({} chars)", response_text.len());

    println!("🔍 Processing Ollama response...");

    // Process the response with validation and fallback
    match process_llm_response(&response_text, idea).await {
        Ok(analysis_result) => {
            println!("✅ Successfully processed Ollama response");
            Ok(analysis_result)
        }
        Err(e) => {
            println!("⚠️  Failed to process Ollama response: {}", e);
            println!("🔄 Falling back to rule-based analysis");
            generate_fallback_analysis(idea).await
        }
    }
}

/// Call custom API endpoint
async fn call_custom_api_with_idea(
    prompt: &str,
    idea: &str,
    config: &LlmConfig,
) -> Result<LlmAnalysisResult> {
    let base_url = config.base_url.as_ref().ok_or_else(|| {
        crate::errors::ApplicationError::Configuration(
            "Custom API base URL not provided".to_string(),
        )
    })?;

    let client = reqwest::Client::new();

    let request_body = serde_json::json!({
        "model": config.model,
        "prompt": prompt,
        "temperature": config.temperature,
        "max_tokens": config.max_tokens
    });

    let response = timeout(
        Duration::from_secs(config.timeout_seconds),
        client
            .post(base_url)
            .header("Content-Type", "application/json")
            .json(&request_body)
            .send(),
    )
    .await
    .map_err(|_| crate::errors::ApplicationError::OperationTimeout {
        timeout_ms: config.timeout_seconds * 1000,
        context: "Custom API call".to_string(),
    })?
    .map_err(|e| crate::errors::ApplicationError::Generic(e.into()))?;

    let response_text = response
        .text()
        .await
        .map_err(|e| crate::errors::ApplicationError::Generic(e.into()))?;

    // Process the response with validation and fallback
    let analysis_result = process_llm_response(&response_text, idea)
        .await
        .map_err(|e| crate::errors::ApplicationError::Generic(e.into()))?;

    Ok(analysis_result)
}

// The extract_json_from_response function is now in the response_processing module
// This is a simplified version that just extracts JSON - the full validation and processing
// is now handled by the response_processing module
fn extract_json_from_response(response: &str) -> Result<String> {
    // First, look for JSON within triple backticks
    if let Some(start) = response.find("```json") {
        if let Some(end) = response[start..].find("```") {
            let json_str = &response[start + 7..start + end]; // Skip "```json"
            return Ok(json_str.trim().to_string());
        }
    }

    // If not found in backticks, look for JSON object directly
    if let Some(start) = response.find('{') {
        let mut brace_count = 0;
        let mut end_pos = start;

        for (i, ch) in response[start..].char_indices() {
            if ch == '{' {
                brace_count += 1;
            } else if ch == '}' {
                brace_count -= 1;
                if brace_count == 0 {
                    end_pos = start + i + 1;
                    break;
                }
            }
        }

        if brace_count == 0 {
            return Ok(response[start..end_pos].to_string());
        }
    }

    // If we couldn't extract JSON, return an error
    Err(crate::errors::ApplicationError::Generic(anyhow::anyhow!(
        "Could not extract JSON from LLM response: {}",
        response
    )))
}

/// Display the LLM analysis result
fn display_llm_analysis_result(idea: &str, analysis: &LlmAnalysisResult) {
    println!("\n📊 LLM Analysis Result:");
    println!("────────────────────────────────────────");
    println!("Idea: {}", idea.chars().take(60).collect::<String>());
    if idea.len() > 60 {
        println!("      ... (truncated)");
    }

    println!("\n🎯 Final Score: {:.2}/10.00", analysis.final_score);
    println!("📋 Recommendation: {}", analysis.recommendation);

    println!("\n📈 Detailed Scores:");
    println!(
        "  Mission Alignment: {:.2}/4.00",
        analysis.scores.mission_alignment.category_total
    );
    println!(
        "  Anti-Challenge Patterns: {:.2}/3.50",
        analysis.scores.anti_challenge_patterns.category_total
    );
    println!(
        "  Strategic Fit: {:.2}/2.50",
        analysis.scores.strategic_fit.category_total
    );

    println!("\n💡 Weighted Totals:");
    println!(
        "  Mission Alignment: {:.2}",
        analysis.weighted_totals.mission_alignment
    );
    println!(
        "  Anti-Challenge Patterns: {:.2}",
        analysis.weighted_totals.anti_challenge_patterns
    );
    println!(
        "  Strategic Fit: {:.2}",
        analysis.weighted_totals.strategic_fit
    );

    println!("\n📝 Explanations:");
    for (key, explanation) in &analysis.explanations {
        println!("  {}: {}", key, explanation);
    }

    println!("────────────────────────────────────────\n");
}

/// Display enhanced LLM analysis result with quality indicators
fn display_enhanced_llm_analysis_result(
    idea: &str,
    analysis: &LlmAnalysisResult,
    cache_similarity: Option<f64>,
    from_cache: bool,
) {
    let cache_indicator = if from_cache {
        format!("⚡ Cache ({:.1}%)", cache_similarity.unwrap_or(0.0) * 100.0)
    } else {
        "🆕 Fresh".to_string()
    };

    println!("\n🎯 Recommendation: {}", analysis.recommendation);
    println!("────────────────────────────────────────");

    // Show idea summary
    println!("💡 {}", idea.chars().take(60).collect::<String>());
    if idea.len() > 60 {
        println!("     ...");
    }

    // Focus on key reasons (strengths and gaps)
    println!("\n✅ WHERE THIS IDEA WORKS:");
    let key_strengths = [
        (
            "AI Alignment",
            analysis.scores.mission_alignment.ai_alignment,
        ),
        (
            "Domain Expertise",
            analysis.scores.mission_alignment.domain_expertise,
        ),
        (
            "Rapid Prototyping",
            analysis.scores.anti_challenge_patterns.rapid_prototyping,
        ),
        (
            "Revenue Potential",
            analysis.scores.mission_alignment.revenue_potential,
        ),
    ];

    for (label, score) in &key_strengths {
        if *score > 0.3 {
            // Show if it's reasonably strong
            if let Some(explanation) = analysis.explanations.get(*label) {
                println!("  • {}: {}", label, explanation);
            } else {
                println!(
                    "  • {}: {:.1}/max - {}",
                    label,
                    score,
                    get_brief_explanation(label, *score)
                );
            }
        }
    }

    println!("\n❌ WHERE THIS IDEA FALLS SHORT:");
    let key_weaknesses = [
        (
            "Context-Switching",
            analysis
                .scores
                .anti_challenge_patterns
                .avoid_context_switching,
        ),
        (
            "Stack Compatibility",
            analysis.scores.strategic_fit.stack_compatibility,
        ),
        (
            "Accountability",
            analysis.scores.anti_challenge_patterns.accountability,
        ),
        (
            "Revenue Testing",
            analysis.scores.strategic_fit.revenue_testing,
        ),
    ];

    for (label, score) in &key_weaknesses {
        if *score < 0.5 {
            // Show if it's weak
            if let Some(explanation) = analysis.explanations.get(*label) {
                println!("  • {}: {}", label, explanation);
            } else {
                println!(
                    "  • {}: {:.1}/max - {}",
                    label,
                    score,
                    get_brief_explanation(label, *score)
                );
            }
        }
    }

    // Overall score and rationale
    println!(
        "\n📊 Overall Score: {:.1}/10 - {}",
        analysis.final_score,
        get_rationale_for_score(analysis.final_score, &analysis.recommendation)
    );

    if from_cache {
        if let Some(similarity) = cache_similarity {
            println!(
                "\n⚡ Previously analyzed ({:.1}% match)",
                similarity * 100.0
            );
        }
    } else {
        println!("\n🔍 Fresh analysis completed");
    }

    println!("────────────────────────────────────────\n");
}

/// Helper function to get brief explanations for scores
fn get_brief_explanation(label: &str, score: f64) -> String {
    match label {
        "AI Alignment" => {
            if score >= 1.0 {
                "Strong AI focus".to_string()
            } else if score >= 0.5 {
                "Some AI component".to_string()
            } else {
                "Little AI relevance".to_string()
            }
        }
        "Domain Expertise" => {
            if score >= 1.0 {
                "Leverages your expertise".to_string()
            } else if score >= 0.5 {
                "Partial domain match".to_string()
            } else {
                "No domain leverage".to_string()
            }
        }
        "Rapid Prototyping" => {
            if score >= 0.8 {
                "Quick iteration possible".to_string()
            } else if score >= 0.4 {
                "Some prototyping speed".to_string()
            } else {
                "Slow prototyping risk".to_string()
            }
        }
        "Revenue Potential" => {
            if score >= 0.4 {
                "Clear income path".to_string()
            } else if score >= 0.2 {
                "Potential income".to_string()
            } else {
                "Unclear revenue".to_string()
            }
        }
        "Context-Switching" => {
            if score >= 1.0 {
                "Uses your stack well".to_string()
            } else if score >= 0.5 {
                "Partial stack match".to_string()
            } else {
                "Risk of context switching".to_string()
            }
        }
        "Stack Compatibility" => {
            if score >= 0.8 {
                "Perfect stack fit".to_string()
            } else if score >= 0.4 {
                "Some stack overlap".to_string()
            } else {
                "Stack mismatch risk".to_string()
            }
        }
        "Accountability" => {
            if score >= 0.6 {
                "External accountability".to_string()
            } else if score >= 0.3 {
                "Some accountability".to_string()
            } else {
                "Solo execution risk".to_string()
            }
        }
        "Revenue Testing" => {
            if score >= 0.2 {
                "Revenue validation possible".to_string()
            } else {
                "No revenue testing".to_string()
            }
        }
        _ => format!("Score: {:.1}/max", score),
    }
}

/// Helper to explain score rationale
fn get_rationale_for_score(score: f64, recommendation: &str) -> String {
    match recommendation {
        "Priority" => "Execute immediately - high strategic value".to_string(),
        "Good" => "Strong potential worth pursuing".to_string(),
        "Consider" => "Has merit but needs refinement".to_string(),
        "Avoid" => "Doesn't align with current priorities".to_string(),
        _ => format!(
            "Score {:.1}/10 indicates {}",
            score,
            recommendation.to_lowercase()
        ),
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_extract_json_from_response_with_backticks() {
        let response = r#"Here's the analysis:
```json
{
    "scores": {
        "Mission Alignment": {
            "Domain Expertise": 1.00,
            "AI Alignment": 1.20,
            "Execution Support": 0.70,
            "Revenue Potential": 0.40,
            "category_total": 3.30
        },
        "Anti-Challenge Patterns": {
            "Avoid Context-Switching": 1.00,
            "Rapid Prototyping": 0.80,
            "Accountability": 0.60,
            "Income Anxiety": 0.40,
            "category_total": 2.80
        },
        "Strategic Fit": {
            "Stack Compatibility": 0.80,
            "Shipping Habit": 0.60,
            "Public Accountability": 0.30,
            "Revenue Testing": 0.20,
            "category_total": 1.90
        }
    },
    "weighted_totals": {
        "Mission Alignment": 3.30,
        "Anti-Challenge Patterns": 2.80,
        "Strategic Fit": 1.90
    },
    "final_score": 8.00,
    "recommendation": "Priority",
    "explanations": {
        "Domain Expertise": "Directly uses existing n8n and Python skills"
    }
}
```
This was a comprehensive analysis."#;

        let result = extract_json_from_response(response).unwrap();
        let parsed: LlmAnalysisResult = serde_json::from_str(&result).unwrap();

        assert_eq!(parsed.final_score, 8.00);
        assert_eq!(parsed.recommendation, "Priority");
        assert_eq!(parsed.scores.mission_alignment.domain_expertise, 1.00);
    }

    #[test]
    fn test_extract_json_from_response_without_backticks() {
        let response = r#"{
    "scores": {
        "Mission Alignment": {
            "Domain Expertise": 0.50,
            "AI Alignment": 0.75,
            "Execution Support": 0.40,
            "Revenue Potential": 0.20,
            "category_total": 1.85
        },
        "Anti-Challenge Patterns": {
            "Avoid Context-Switching": 0.80,
            "Rapid Prototyping": 0.60,
            "Accountability": 0.40,
            "Income Anxiety": 0.30,
            "category_total": 2.10
        },
        "Strategic Fit": {
            "Stack Compatibility": 0.60,
            "Shipping Habit": 0.40,
            "Public Accountability": 0.20,
            "Revenue Testing": 0.10,
            "category_total": 1.30
        }
    },
    "weighted_totals": {
        "Mission Alignment": 1.85,
        "Anti-Challenge Patterns": 2.10,
        "Strategic Fit": 1.30
    },
    "final_score": 5.25,
    "recommendation": "Consider",
    "explanations": {
        "Domain Expertise": "Requires some learning of new skills"
    }
}"#;

        let result = extract_json_from_response(response).unwrap();
        let parsed: LlmAnalysisResult = serde_json::from_str(&result).unwrap();

        assert_eq!(parsed.final_score, 5.25);
        assert_eq!(parsed.recommendation, "Consider");
        assert_eq!(parsed.scores.mission_alignment.domain_expertise, 0.50);
    }

    #[test]
    fn test_extract_json_from_response_with_regular_backticks() {
        let response = r#"Here's the analysis:
```
{
    "scores": {
        "Mission Alignment": {
            "Domain Expertise": 1.00,
            "AI Alignment": 1.20,
            "Execution Support": 0.70,
            "Revenue Potential": 0.40,
            "category_total": 3.30
        },
        "Anti-Challenge Patterns": {
            "Avoid Context-Switching": 1.00,
            "Rapid Prototyping": 0.80,
            "Accountability": 0.60,
            "Income Anxiety": 0.40,
            "category_total": 2.80
        },
        "Strategic Fit": {
            "Stack Compatibility": 0.80,
            "Shipping Habit": 0.60,
            "Public Accountability": 0.30,
            "Revenue Testing": 0.20,
            "category_total": 1.90
        }
    },
    "weighted_totals": {
        "Mission Alignment": 1.32,
        "Anti-Challenge Patterns": 0.98,
        "Strategic Fit": 0.48
    },
    "final_score": 2.78,
    "recommendation": "Consider",
    "explanations": {
        "Domain Expertise": "Valid explanation"
    }
}
```
This was a comprehensive analysis."#;

        let result = extract_json_from_response(response).unwrap();
        let parsed: LlmAnalysisResult = serde_json::from_str(&result).unwrap();

        assert_eq!(parsed.final_score, 2.78);
        assert_eq!(parsed.recommendation, "Consider");
        assert_eq!(parsed.scores.mission_alignment.domain_expertise, 1.00);
    }
}
