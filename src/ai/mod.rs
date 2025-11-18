use crate::errors::CircuitBreakerError;
use ollama_rs::generation::completion::request::GenerationRequest;
use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use std::sync::Arc;
use std::time::{Duration, Instant};
use tokio::sync::RwLock;

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct AiEnhancement {
    pub mission_alignment_insights: Vec<String>,
    pub pattern_analysis: HashMap<String, PatternInsight>,
    pub contextual_factors: Vec<String>,
    pub suggested_actions: Vec<SuggestedAction>,
    pub confidence_score: f64,
    pub reasoning: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct PatternInsight {
    pub detected: bool,
    pub severity: String, // "low", "medium", "high", "critical"
    pub explanation: String,
    pub suggestions: Vec<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct SuggestedAction {
    pub action: String,
    pub priority: String, // "immediate", "this_week", "later"
    pub reasoning: String,
}

pub struct AiAnalyzer {
    ollama_client: Option<ollama_rs::Ollama>,
    model_name: String,
    prompts: PromptTemplates,
    circuit_breaker: CircuitBreaker,
    retry_config: RetryConfig,
}

/// Configuration for retry logic
#[derive(Debug, Clone, Copy)]
pub struct RetryConfig {
    /// Maximum number of retry attempts
    pub max_attempts: u32,
    /// Initial backoff delay
    pub initial_delay_ms: u64,
    /// Maximum backoff delay
    pub max_delay_ms: u64,
    /// Backoff multiplier
    pub multiplier: f64,
    /// Whether to add jitter to prevent thundering herd
    pub jitter: bool,
}

impl RetryConfig {
    pub fn initial_delay(&self) -> Duration {
        Duration::from_millis(self.initial_delay_ms)
    }

    pub fn max_delay(&self) -> Duration {
        Duration::from_millis(self.max_delay_ms)
    }
}

impl Default for RetryConfig {
    fn default() -> Self {
        Self {
            max_attempts: 3,
            initial_delay_ms: 500,
            max_delay_ms: 30_000, // 30 seconds
            multiplier: 2.0,
            jitter: true,
        }
    }
}

/// Circuit breaker state for AI service protection
#[derive(Debug, Clone)]
pub enum CircuitBreakerState {
    Closed,   // Normal operation
    Open,     // Failing, reject requests
    HalfOpen, // Testing if service has recovered
}

/// Circuit breaker configuration
#[derive(Debug, Clone, Copy)]
pub struct CircuitBreakerConfig {
    /// Number of failures before opening the circuit
    pub failure_threshold: u32,
    /// How long to wait before transitioning from Open to HalfOpen
    pub recovery_timeout_ms: u64,
    /// Number of successful requests before closing circuit (in HalfOpen state)
    pub success_threshold: u32,
    /// Request timeout
    pub request_timeout_ms: u64,
    /// Rate limit requests per second
    pub rate_limit: u32,
}

impl CircuitBreakerConfig {
    pub fn recovery_timeout(&self) -> Duration {
        Duration::from_millis(self.recovery_timeout_ms)
    }

    pub fn request_timeout(&self) -> Duration {
        Duration::from_millis(self.request_timeout_ms)
    }
}

impl Default for CircuitBreakerConfig {
    fn default() -> Self {
        Self {
            failure_threshold: 5,
            recovery_timeout_ms: 60_000, // 60 seconds
            success_threshold: 3,
            request_timeout_ms: 30_000, // 30 seconds
            rate_limit: 10,
        }
    }
}

/// Circuit breaker for AI service protection
#[derive(Debug)]
pub struct CircuitBreaker {
    config: CircuitBreakerConfig,
    state: Arc<RwLock<CircuitBreakerState>>,
    failure_count: Arc<RwLock<u32>>,
    success_count: Arc<RwLock<u32>>,
    last_failure_time: Arc<RwLock<Option<Instant>>>,
    service_name: String,
}

impl CircuitBreaker {
    pub fn new(service_name: String, config: CircuitBreakerConfig) -> Self {
        Self {
            config,
            state: Arc::new(RwLock::new(CircuitBreakerState::Closed)),
            failure_count: Arc::new(RwLock::new(0)),
            success_count: Arc::new(RwLock::new(0)),
            last_failure_time: Arc::new(RwLock::new(None)),
            service_name,
        }
    }

    pub async fn execute<F, T, E>(&self, operation: F) -> Result<T, CircuitBreakerError>
    where
        F: std::future::Future<Output = Result<T, E>>,
        E: std::error::Error + Send + Sync + 'static,
    {
        // Check circuit state before executing
        self.check_circuit_state().await?;

        let start_time = Instant::now();

        // Execute with timeout
        let result = tokio::time::timeout(self.config.request_timeout(), operation).await;

        match result {
            Ok(Ok(value)) => {
                // Success - record and potentially close circuit
                self.record_success().await;

                let duration = start_time.elapsed();
                tracing::info!(
                    service_name = %self.service_name,
                    duration_ms = duration.as_millis(),
                    "AI request completed successfully"
                );

                Ok(value)
            }
            Ok(Err(e)) => {
                // Failure - record and potentially open circuit
                self.record_failure().await;

                let duration = start_time.elapsed();
                let error_string = e.to_string();

                tracing::warn!(
                    service_name = %self.service_name,
                    error = %error_string,
                    duration_ms = duration.as_millis(),
                    "AI request failed"
                );

                Err(CircuitBreakerError::service_unavailable_with_source(
                    &self.service_name,
                    &error_string,
                    Box::new(e) as Box<dyn std::error::Error + Send + Sync>,
                ))
            }
            Err(_) => {
                // Timeout - record as failure
                self.record_failure().await;

                let duration = start_time.elapsed();

                tracing::warn!(
                    service_name = %self.service_name,
                    timeout_ms = self.config.request_timeout().as_millis(),
                    actual_duration_ms = duration.as_millis(),
                    "AI request timed out"
                );

                Err(CircuitBreakerError::service_timeout(
                    &self.service_name,
                    self.config.request_timeout_ms,
                    "AI request",
                ))
            }
        }
    }

    async fn check_circuit_state(&self) -> Result<(), CircuitBreakerError> {
        let mut state = self.state.write().await;
        let last_failure = self.last_failure_time.read().await;

        match *state {
            CircuitBreakerState::Closed => Ok(()),
            CircuitBreakerState::Open => {
                if let Some(last_failure_time) = *last_failure {
                    if last_failure_time.elapsed() > self.config.recovery_timeout() {
                        // Try to recover
                        *state = CircuitBreakerState::HalfOpen;
                        *self.success_count.write().await = 0;
                        tracing::info!(
                            service_name = %self.service_name,
                            "Circuit breaker transitioning to half-open"
                        );
                        Ok(())
                    } else {
                        // Still in open state
                        let retry_after =
                            self.config.recovery_timeout() - last_failure_time.elapsed();
                        // Convert Instant to SystemTime (approximate)
                        let system_time =
                            std::time::SystemTime::now() - last_failure_time.elapsed();
                        Err(CircuitBreakerError::circuit_open(
                            &self.service_name,
                            system_time,
                            *self.failure_count.read().await,
                            Some(retry_after),
                        ))
                    }
                } else {
                    // Shouldn't happen, but handle gracefully
                    *state = CircuitBreakerState::Closed;
                    Ok(())
                }
            }
            CircuitBreakerState::HalfOpen => Ok(()),
        }
    }

    async fn record_success(&self) {
        let state = self.state.read().await;

        match *state {
            CircuitBreakerState::HalfOpen => {
                let mut success_count = self.success_count.write().await;
                *success_count += 1;

                if *success_count >= self.config.success_threshold {
                    // Close the circuit
                    drop(state); // Release read lock
                    let mut state = self.state.write().await;
                    *state = CircuitBreakerState::Closed;
                    *self.failure_count.write().await = 0;

                    tracing::info!(
                        service_name = %self.service_name,
                        "Circuit breaker closed after successful recovery"
                    );
                }
            }
            CircuitBreakerState::Closed => {
                // Reset failure count on success
                *self.failure_count.write().await = 0;
            }
            _ => {}
        }
    }

    async fn record_failure(&self) {
        let mut failure_count = self.failure_count.write().await;
        *failure_count += 1;

        let current_failures = *failure_count;
        let mut state = self.state.write().await;

        match *state {
            CircuitBreakerState::Closed => {
                if current_failures >= self.config.failure_threshold {
                    *state = CircuitBreakerState::Open;
                    *self.last_failure_time.write().await = Some(Instant::now());

                    tracing::warn!(
                        service_name = %self.service_name,
                        failure_count = current_failures,
                        threshold = self.config.failure_threshold,
                        "Circuit breaker opened due to failures"
                    );
                }
            }
            CircuitBreakerState::HalfOpen => {
                *state = CircuitBreakerState::Open;
                *self.last_failure_time.write().await = Some(Instant::now());

                tracing::warn!(
                    service_name = %self.service_name,
                    "Circuit breaker reopened after failure in half-open state"
                );
            }
            CircuitBreakerState::Open => {
                // Already open, just update last failure time
                *self.last_failure_time.write().await = Some(Instant::now());
            }
        }
    }

    pub async fn get_state(&self) -> CircuitBreakerState {
        self.state.read().await.clone()
    }

    pub async fn get_metrics(&self) -> CircuitBreakerMetrics {
        CircuitBreakerMetrics {
            state: self.get_state().await,
            failure_count: *self.failure_count.read().await,
            success_count: *self.success_count.read().await,
            last_failure_time: *self.last_failure_time.read().await,
        }
    }
}

/// Metrics for circuit breaker monitoring
#[derive(Debug, Clone)]
pub struct CircuitBreakerMetrics {
    pub state: CircuitBreakerState,
    pub failure_count: u32,
    pub success_count: u32,
    pub last_failure_time: Option<Instant>,
}

#[derive(Debug, Clone)]
pub struct PromptTemplates {
    pub analysis_template: String,
    pub pattern_detection_template: String,
    pub recommendation_template: String,
}

impl AiAnalyzer {
    pub fn new() -> Self {
        Self::with_configs(
            "mistral-dolphin",
            CircuitBreakerConfig::default(),
            RetryConfig::default(),
        )
    }

    pub fn with_model(model_name: &str) -> Self {
        Self::with_configs(
            model_name,
            CircuitBreakerConfig::default(),
            RetryConfig::default(),
        )
    }

    pub fn with_configs(
        model_name: &str,
        circuit_config: CircuitBreakerConfig,
        retry_config: RetryConfig,
    ) -> Self {
        let analyzer = Self {
            ollama_client: None, // AI integration temporarily disabled for MVP
            model_name: model_name.to_string(),
            prompts: PromptTemplates::default(),
            circuit_breaker: CircuitBreaker::new(format!("ollama-{}", model_name), circuit_config),
            retry_config,
        };

        tracing::info!(
            model = %model_name,
            circuit_failure_threshold = circuit_config.failure_threshold,
            circuit_recovery_timeout_ms = circuit_config.recovery_timeout_ms,
            retry_max_attempts = retry_config.max_attempts,
            "AI Analyzer initialized"
        );

        analyzer
    }

    pub async fn enhance_analysis(
        &self,
        idea: &str,
        base_score: &crate::scoring::Score,
        telos_context: &crate::telos::ParsedTelos,
    ) -> Result<Option<AiEnhancement>, anyhow::Error> {
        // If AI is not available, return None gracefully
        if self.ollama_client.is_none() {
            tracing::debug!("AI client not available, skipping enhancement");
            return Ok(None);
        }

        // Log the request
        tracing::info!(
            model = %self.model_name,
            idea_length = idea.len(),
            base_score = base_score.final_score,
            "Starting AI analysis enhancement"
        );

        // Create context from Telos
        let context = self.create_telos_context(telos_context);

        // Create prompt
        let prompt = self
            .prompts
            .analysis_template
            .replace("{CONTEXT}", &context)
            .replace("{IDEA}", idea)
            .replace("{BASE_SCORE}", &base_score.final_score.to_string())
            .replace("{BASE_RECOMMENDATION}", base_score.recommendation.text());

        // Execute AI call with circuit breaker protection
        match self.call_ai_service(&prompt).await {
            Ok(enhancement) => {
                tracing::info!(
                    model = %self.model_name,
                    confidence_score = enhancement.confidence_score,
                    "AI analysis enhancement completed successfully"
                );
                Ok(Some(enhancement))
            }
            Err(e) => {
                tracing::warn!(
                    model = %self.model_name,
                    error = %e,
                    "AI analysis enhancement failed, continuing without enhancement"
                );
                // Continue without AI enhancement rather than failing the entire operation
                Ok(None)
            }
        }
    }

    pub async fn enhance_pattern_detection(
        &self,
        idea: &str,
        base_patterns: &[crate::patterns_simple::PatternMatch],
        telos_context: &crate::telos::ParsedTelos,
    ) -> Result<Option<HashMap<String, PatternInsight>>, anyhow::Error> {
        if self.ollama_client.is_none() {
            return Ok(None);
        }

        let _client = self.ollama_client.as_ref().unwrap();

        // Create context
        let context = self.create_pattern_context(telos_context);
        let existing_patterns = base_patterns
            .iter()
            .map(|p| format!("{}: {}", p.pattern_type.title(), p.message))
            .collect::<Vec<_>>()
            .join(", ");

        let _prompt = self
            .prompts
            .pattern_detection_template
            .replace("{CONTEXT}", &context)
            .replace("{IDEA}", idea)
            .replace("{EXISTING_PATTERNS}", &existing_patterns);

        // AI integration disabled for MVP
        Ok(None)
    }

    fn create_telos_context(&self, telos: &crate::telos::ParsedTelos) -> String {
        format!(
            "CURRENT TELOS CONTEXT:\n\
            Active Goals: {}\n\
            Current Strategy: {}\n\
            Main Challenges: {}\n\
            Current Tech Stack: {}\n\
            Domain Expertise: {}",
            telos
                .goals
                .iter()
                .map(|g| g.title.as_str())
                .collect::<Vec<_>>()
                .join(", "),
            telos
                .strategies
                .first()
                .map(|s| s.title.as_str())
                .unwrap_or("Unknown"),
            telos
                .challenges
                .iter()
                .take(2)
                .map(|c| c.title.as_str())
                .collect::<Vec<_>>()
                .join(", "),
            telos.current_stack.join(" + "),
            telos.domain_keywords.join(" + ")
        )
    }

    fn create_pattern_context(&self, telos: &crate::telos::ParsedTelos) -> String {
        format!(
            "USER'S DOCUMENTED PATTERNS:\n\
            Main Challenges: {}\n\
            Current Strategy Focus: {}\n\
            Known Traps: Context-switching to new tech stacks, Perfectionism leading to scope creep, Procrastination through consumption",
            telos.challenges.iter().map(|c| c.title.as_str()).collect::<Vec<_>>().join(", "),
            telos.strategies.first().map(|s| s.title.as_str()).unwrap_or("Unknown")
        )
    }

    fn parse_ai_response(&self, response: &str) -> anyhow::Result<AiEnhancement> {
        // Try to parse as JSON first
        if response.trim().starts_with('{') {
            return Ok(serde_json::from_str(response)?);
        }

        // Fallback: extract insights from text
        Ok(self.parse_text_insights(response))
    }

    fn parse_pattern_response(
        &self,
        response: &str,
    ) -> anyhow::Result<HashMap<String, PatternInsight>> {
        // Try to parse as JSON
        if response.trim().starts_with('{') {
            return Ok(serde_json::from_str(response)?);
        }

        // Fallback: create simple insights
        let mut insights = HashMap::new();

        insights.insert(
            "context-switching".to_string(),
            PatternInsight {
                detected: response.to_lowercase().contains("context")
                    || response.to_lowercase().contains("stack"),
                severity: "medium".to_string(),
                explanation: "AI detected potential context-switching indicators".to_string(),
                suggestions: vec!["Stay focused on current stack".to_string()],
            },
        );

        Ok(insights)
    }

    fn parse_text_insights(&self, response: &str) -> AiEnhancement {
        let lines: Vec<&str> = response.lines().collect();
        let mut insights = Vec::new();
        let mut suggested_actions = Vec::new();

        for line in lines {
            if line.to_lowercase().contains("insight")
                || line.to_lowercase().contains("observation")
            {
                insights.push(line.trim().to_string());
            }

            if line.to_lowercase().contains("suggest") || line.to_lowercase().contains("recommend")
            {
                suggested_actions.push(SuggestedAction {
                    action: line.trim().to_string(),
                    priority: "this_week".to_string(),
                    reasoning: "AI-generated suggestion".to_string(),
                });
            }
        }

        AiEnhancement {
            mission_alignment_insights: insights,
            pattern_analysis: HashMap::new(),
            contextual_factors: Vec::new(),
            suggested_actions,
            confidence_score: 0.7, // Default confidence for text parsing
            reasoning: "Parsed from AI text response".to_string(),
        }
    }

    /// Execute an operation with exponential backoff retry logic
    async fn execute_with_retry<F, T>(
        &self,
        operation_name: &str,
        mut operation: F,
    ) -> Result<T, CircuitBreakerError>
    where
        F: FnMut() -> std::pin::Pin<
            Box<dyn std::future::Future<Output = Result<T, CircuitBreakerError>> + Send>,
        >,
        T: Send + 'static,
    {
        let mut attempt = 0;
        let mut delay = self.retry_config.initial_delay();

        loop {
            attempt += 1;

            tracing::debug!(
                operation = operation_name,
                attempt = attempt,
                max_attempts = self.retry_config.max_attempts,
                "Executing AI operation"
            );

            let result = operation().await;

            match result {
                Ok(value) => {
                    if attempt > 1 {
                        tracing::info!(
                            operation = operation_name,
                            attempt = attempt,
                            "AI operation succeeded after retries"
                        );
                    }
                    return Ok(value);
                }
                Err(e) => {
                    if attempt >= self.retry_config.max_attempts {
                        tracing::error!(
                            operation = operation_name,
                            attempt = attempt,
                            error = %e,
                            "AI operation failed after all retries"
                        );
                        return Err(e);
                    }

                    if !e.is_retryable() {
                        tracing::warn!(
                            operation = operation_name,
                            attempt = attempt,
                            error = %e,
                            "AI operation error is not retryable"
                        );
                        return Err(e);
                    }

                    // Calculate delay with jitter if enabled
                    let actual_delay = if self.retry_config.jitter {
                        let jitter_factor = 0.8 + (rand::random::<f64>() * 0.4); // 80% to 120%
                        Duration::from_millis((delay.as_millis() as f64 * jitter_factor) as u64)
                    } else {
                        delay
                    };

                    tracing::info!(
                        operation = operation_name,
                        attempt = attempt,
                        error = %e,
                        delay_ms = actual_delay.as_millis(),
                        "AI operation failed, retrying after delay"
                    );

                    tokio::time::sleep(actual_delay).await;

                    // Exponential backoff
                    delay = std::cmp::min(
                        Duration::from_millis(
                            (delay.as_millis() as f64 * self.retry_config.multiplier) as u64,
                        ),
                        self.retry_config.max_delay(),
                    );
                }
            }
        }
    }

    /// Real AI service call using Ollama API with circuit breaker protection
    async fn call_ai_service(&self, prompt: &str) -> Result<AiEnhancement, CircuitBreakerError> {
        let client = self.ollama_client.as_ref().ok_or_else(|| {
            CircuitBreakerError::service_unavailable("ai_service", "Ollama client not initialized")
        })?;

        tracing::debug!("Making AI service call to model: {}", self.model_name);

        // Use circuit breaker to protect the AI service call
        self.circuit_breaker
            .execute(async {
                // Create the Ollama request - simplified approach
                let request = GenerationRequest::new(self.model_name.clone(), prompt.to_string());

                // Make the actual API call
                let response = client.generate(request).await.map_err(|e| {
                    CircuitBreakerError::service_unavailable(
                        "ai_service",
                        format!("Ollama API call failed: {}", e),
                    )
                })?;

                // Parse the response
                let ai_response = response.response;
                tracing::debug!("Received AI response: {} chars", ai_response.len());

                // Convert to AiEnhancement
                self.parse_ai_response(&ai_response).map_err(|e| {
                    CircuitBreakerError::malformed_response(
                        "ai_service",
                        format!("Failed to parse AI response: {}", e),
                    )
                })
            })
            .await
    }

    pub fn is_available(&self) -> bool {
        self.ollama_client.is_some()
    }

    pub async fn test_connection(&self) -> Result<bool, CircuitBreakerError> {
        if self.ollama_client.is_none() {
            return Ok(false);
        }

        // Test with circuit breaker
        #[derive(Debug)]
        struct TestError;

        impl std::fmt::Display for TestError {
            fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                write!(f, "Test connection error")
            }
        }

        impl std::error::Error for TestError {}

        self.circuit_breaker
            .execute(async {
                // Simulate connection test
                tokio::time::sleep(Duration::from_millis(100)).await;
                Ok::<bool, TestError>(true)
            })
            .await
    }

    /// Get circuit breaker metrics for monitoring
    pub async fn get_circuit_breaker_metrics(&self) -> CircuitBreakerMetrics {
        self.circuit_breaker.get_metrics().await
    }
}

impl Default for PromptTemplates {
    fn default() -> Self {
        Self {
            analysis_template: r#"You are analyzing ideas for Ray Yacub's Telos framework.

CONTEXT: {CONTEXT}

IDEA TO ANALYZE: {IDEA}

BASE ANALYSIS: Score {BASE_SCORE}/10, Recommendation: {BASE_RECOMMENDATION}

Please provide additional insights in JSON format:
{{
    "mission_alignment_insights": [
        "Specific insight about how this aligns with Ray's missions",
        "Another insight about mission alignment"
    ],
    "pattern_analysis": {{
        "context-switching": {{
            "detected": true/false,
            "severity": "low/medium/high/critical",
            "explanation": "Why this pattern was detected",
            "suggestions": ["Specific suggestion 1", "Specific suggestion 2"]
        }},
        "perfectionism": {{ ... }},
        "procrastination": {{ ... }}
    }},
    "contextual_factors": [
        "Current deadline pressure consideration",
        "Energy level consideration",
        "Resource availability consideration"
    ],
    "suggested_actions": [
        {{
            "action": "Specific actionable step",
            "priority": "immediate/this_week/later",
            "reasoning": "Why this action makes sense"
        }}
    ],
    "confidence_score": 0.85,
    "reasoning": "Brief explanation of your analysis approach"
}}"#.to_string(),

            pattern_detection_template: r#"Analyze this idea for behavioral patterns based on Ray's Telos.

CONTEXT: {CONTEXT}
IDEA: {IDEA}
EXISTING RULE-BASED PATTERNS: {EXISTING_PATTERNS}

Provide pattern analysis in JSON format:
{{
    "pattern_name": {{
        "detected": true/false,
        "severity": "low/medium/high/critical",
        "explanation": "Why this pattern is present",
        "suggestions": ["Specific suggestion 1", "Specific suggestion 2"]
    }}
}}"#.to_string(),

            recommendation_template: r#"Based on the analysis of this idea against Ray's Telos, provide a recommendation in JSON format:
{{
    "recommendation": "PRIORITIZE/GOOD/CONSIDER/AVOID",
    "confidence": 0.8,
    "reasoning": "Detailed explanation",
    "next_steps": ["Step 1", "Step 2"]
}}"#.to_string(),
        }
    }
}

// Empty module files for future expansion
pub mod models;
pub mod prompts;
