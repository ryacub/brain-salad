//! Response processing and validation for LLM analysis system.
//! This module provides JSON validation, error handling, and fallback scoring for LLM responses.

use crate::commands::analyze_llm::{
    AntiChallengePatternsScores, LlmAnalysisResult, LlmScores, LlmWeightedTotals,
    MissionAlignmentScores, StrategicFitScores,
};
use crate::errors::Result;
use crate::scoring::{Score, ScoringEngine, TelosConfig};
use regex::Regex;
use serde::{Deserialize, Serialize};
use std::collections::HashMap;

/// Confidence level for LLM analysis results
#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub enum ConfidenceLevel {
    High,
    Medium,
    Low,
    VeryLow,
}

impl ConfidenceLevel {
    /// Get numeric confidence score (0.0 - 1.0)
    pub fn score(&self) -> f64 {
        match self {
            ConfidenceLevel::High => 0.9,
            ConfidenceLevel::Medium => 0.7,
            ConfidenceLevel::Low => 0.5,
            ConfidenceLevel::VeryLow => 0.3,
        }
    }

    /// Get display emoji
    pub fn emoji(&self) -> &'static str {
        match self {
            ConfidenceLevel::High => "🟢",
            ConfidenceLevel::Medium => "🟡",
            ConfidenceLevel::Low => "🟠",
            ConfidenceLevel::VeryLow => "🔴",
        }
    }

    /// Get description
    pub fn description(&self) -> &'static str {
        match self {
            ConfidenceLevel::High => "High confidence - analysis appears reliable",
            ConfidenceLevel::Medium => "Medium confidence - analysis has some uncertainties",
            ConfidenceLevel::Low => "Low confidence - analysis may have issues",
            ConfidenceLevel::VeryLow => "Very low confidence - fallback scoring used",
        }
    }
}

/// Quality metrics for LLM response
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ResponseQualityMetrics {
    pub validation_passed: bool,
    pub json_parsing_passed: bool,
    pub scores_in_range: bool,
    pub calculations_correct: bool,
    pub explanations_provided: bool,
    pub confidence_level: ConfidenceLevel,
    pub fallback_used: bool,
    pub validation_errors: Vec<String>,
    pub quality_score: f64, // 0.0 - 1.0
}

impl ResponseQualityMetrics {
    pub fn new() -> Self {
        Self {
            validation_passed: false,
            json_parsing_passed: false,
            scores_in_range: false,
            calculations_correct: false,
            explanations_provided: false,
            confidence_level: ConfidenceLevel::VeryLow,
            fallback_used: false,
            validation_errors: Vec::new(),
            quality_score: 0.0,
        }
    }

    /// Calculate overall quality score
    pub fn calculate_quality_score(&mut self) {
        let mut score = 0.0;
        let mut factors = 0;

        if self.json_parsing_passed {
            score += 0.3;
            factors += 1;
        }
        if self.scores_in_range {
            score += 0.2;
            factors += 1;
        }
        if self.calculations_correct {
            score += 0.2;
            factors += 1;
        }
        if self.explanations_provided {
            score += 0.1;
            factors += 1;
        }
        if !self.fallback_used {
            score += 0.2;
            factors += 1;
        }

        self.quality_score = if factors > 0 { score } else { 0.0 };
    }

    /// Determine confidence level based on metrics
    pub fn determine_confidence_level(&mut self) {
        self.confidence_level = if self.fallback_used {
            ConfidenceLevel::VeryLow
        } else if self.quality_score >= 0.9 {
            ConfidenceLevel::High
        } else if self.quality_score >= 0.7 {
            ConfidenceLevel::Medium
        } else if self.quality_score >= 0.5 {
            ConfidenceLevel::Low
        } else {
            ConfidenceLevel::VeryLow
        };
    }
}

/// Enhanced LLM analysis result with confidence and quality metrics
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct EnhancedLlmAnalysisResult {
    pub base_result: LlmAnalysisResult,
    pub confidence_level: ConfidenceLevel,
    pub quality_metrics: ResponseQualityMetrics,
    pub processing_notes: Vec<String>,
}

/// Configuration for response validation
#[derive(Debug, Clone)]
pub struct ResponseValidationConfig {
    /// Maximum allowed score values for each category
    pub max_mission_score: f64,
    pub max_anti_challenge_score: f64,
    pub max_strategic_score: f64,
    /// Allowed recommendation values
    pub allowed_recommendations: Vec<String>,
    /// Whether to enforce two decimal places precision
    pub enforce_two_decimal_places: bool,
    /// Maximum length for explanation strings
    pub max_explanation_length: usize,
    /// Confidence calculation thresholds
    pub high_confidence_threshold: f64,
    pub medium_confidence_threshold: f64,
    pub low_confidence_threshold: f64,
}

impl Default for ResponseValidationConfig {
    fn default() -> Self {
        Self {
            max_mission_score: 4.0,
            max_anti_challenge_score: 3.5,
            max_strategic_score: 2.5,
            allowed_recommendations: vec![
                "Priority".to_string(),
                "Good".to_string(),
                "Consider".to_string(),
                "Avoid".to_string(),
            ],
            enforce_two_decimal_places: true,
            max_explanation_length: 500,
            high_confidence_threshold: 0.9,
            medium_confidence_threshold: 0.7,
            low_confidence_threshold: 0.5,
        }
    }
}

/// Validator for LLM response processing
#[derive(Debug, Clone)]
pub struct ResponseValidator {
    config: ResponseValidationConfig,
    // Pre-compiled regex for decimal validation
    decimal_pattern: Regex,
}

impl ResponseValidator {
    /// Create a new validator with default configuration
    pub fn new() -> Self {
        Self::with_config(ResponseValidationConfig::default())
    }

    /// Create a new validator with custom configuration
    pub fn with_config(config: ResponseValidationConfig) -> Self {
        Self {
            config,
            decimal_pattern: Regex::new(r"^\d+\.\d{2}$").expect("Invalid decimal regex"),
        }
    }

    /// Validate the complete LLM analysis result
    pub fn validate_analysis_result(&self, result: &LlmAnalysisResult) -> Result<()> {
        // Validate scores structure
        self.validate_scores(&result.scores)?;

        // Validate weighted totals
        self.validate_weighted_totals(&result.weighted_totals, &result.scores)?;

        // Validate final score
        self.validate_final_score(result.final_score, &result.scores)?;

        // Validate recommendation
        self.validate_recommendation(&result.recommendation)?;

        // Validate explanations
        self.validate_explanations(&result.explanations)?;

        Ok(())
    }

    /// Validate the scores structure
    fn validate_scores(&self, scores: &LlmScores) -> Result<()> {
        // Validate Mission Alignment scores
        self.validate_mission_alignment_scores(&scores.mission_alignment)?;

        // Validate Anti-Challenge Patterns scores
        self.validate_anti_challenge_scores(&scores.anti_challenge_patterns)?;

        // Validate Strategic Fit scores
        self.validate_strategic_scores(&scores.strategic_fit)?;

        Ok(())
    }

    /// Validate Mission Alignment scores
    fn validate_mission_alignment_scores(&self, scores: &MissionAlignmentScores) -> Result<()> {
        // Validate individual scores
        self.validate_score_range(scores.domain_expertise, 0.0, 1.2, "Domain Expertise")?;
        self.validate_score_range(scores.ai_alignment, 0.0, 1.5, "AI Alignment")?;
        self.validate_score_range(scores.execution_support, 0.0, 0.8, "Execution Support")?;
        self.validate_score_range(scores.revenue_potential, 0.0, 0.5, "Revenue Potential")?;

        // Validate category total
        let expected_total = scores.domain_expertise
            + scores.ai_alignment
            + scores.execution_support
            + scores.revenue_potential;
        self.validate_score_range(
            scores.category_total,
            0.0,
            self.config.max_mission_score,
            "Mission Alignment category total",
        )?;

        // Check if the total matches the sum of components (with small tolerance for floating point errors)
        if (scores.category_total - expected_total).abs() > 0.01 {
            return Err(crate::errors::ApplicationError::Validation(
                crate::errors::ValidationError::invalid_value(
                    "Mission Alignment category total",
                    format!("{:.2}", scores.category_total),
                    format!("should be {:.2} (sum of components)", expected_total),
                ),
            ));
        }

        // Validate decimal precision
        if self.config.enforce_two_decimal_places {
            self.validate_two_decimal_places(scores.domain_expertise, "Domain Expertise")?;
            self.validate_two_decimal_places(scores.ai_alignment, "AI Alignment")?;
            self.validate_two_decimal_places(scores.execution_support, "Execution Support")?;
            self.validate_two_decimal_places(scores.revenue_potential, "Revenue Potential")?;
            self.validate_two_decimal_places(
                scores.category_total,
                "Mission Alignment category total",
            )?;
        }

        Ok(())
    }

    /// Validate Anti-Challenge Patterns scores
    fn validate_anti_challenge_scores(&self, scores: &AntiChallengePatternsScores) -> Result<()> {
        // Validate individual scores
        self.validate_score_range(
            scores.avoid_context_switching,
            0.0,
            1.2,
            "Avoid Context-Switching",
        )?;
        self.validate_score_range(scores.rapid_prototyping, 0.0, 1.0, "Rapid Prototyping")?;
        self.validate_score_range(scores.accountability, 0.0, 0.8, "Accountability")?;
        self.validate_score_range(scores.income_anxiety, 0.0, 0.5, "Income Anxiety")?;

        // Validate category total
        let expected_total = scores.avoid_context_switching
            + scores.rapid_prototyping
            + scores.accountability
            + scores.income_anxiety;
        self.validate_score_range(
            scores.category_total,
            0.0,
            self.config.max_anti_challenge_score,
            "Anti-Challenge Patterns category total",
        )?;

        // Check if the total matches the sum of components
        if (scores.category_total - expected_total).abs() > 0.01 {
            return Err(crate::errors::ApplicationError::Validation(
                crate::errors::ValidationError::invalid_value(
                    "Anti-Challenge Patterns category total",
                    format!("{:.2}", scores.category_total),
                    format!("should be {:.2} (sum of components)", expected_total),
                ),
            ));
        }

        // Validate decimal precision
        if self.config.enforce_two_decimal_places {
            self.validate_two_decimal_places(
                scores.avoid_context_switching,
                "Avoid Context-Switching",
            )?;
            self.validate_two_decimal_places(scores.rapid_prototyping, "Rapid Prototyping")?;
            self.validate_two_decimal_places(scores.accountability, "Accountability")?;
            self.validate_two_decimal_places(scores.income_anxiety, "Income Anxiety")?;
            self.validate_two_decimal_places(
                scores.category_total,
                "Anti-Challenge Patterns category total",
            )?;
        }

        Ok(())
    }

    /// Validate Strategic Fit scores
    fn validate_strategic_scores(&self, scores: &StrategicFitScores) -> Result<()> {
        // Validate individual scores
        self.validate_score_range(scores.stack_compatibility, 0.0, 1.0, "Stack Compatibility")?;
        self.validate_score_range(scores.shipping_habit, 0.0, 0.8, "Shipping Habit")?;
        self.validate_score_range(
            scores.public_accountability,
            0.0,
            0.4,
            "Public Accountability",
        )?;
        self.validate_score_range(scores.revenue_testing, 0.0, 0.3, "Revenue Testing")?;

        // Validate category total
        let expected_total = scores.stack_compatibility
            + scores.shipping_habit
            + scores.public_accountability
            + scores.revenue_testing;
        self.validate_score_range(
            scores.category_total,
            0.0,
            self.config.max_strategic_score,
            "Strategic Fit category total",
        )?;

        // Check if the total matches the sum of components
        if (scores.category_total - expected_total).abs() > 0.01 {
            return Err(crate::errors::ApplicationError::Validation(
                crate::errors::ValidationError::invalid_value(
                    "Strategic Fit category total",
                    format!("{:.2}", scores.category_total),
                    format!("should be {:.2} (sum of components)", expected_total),
                ),
            ));
        }

        // Validate decimal precision
        if self.config.enforce_two_decimal_places {
            self.validate_two_decimal_places(scores.stack_compatibility, "Stack Compatibility")?;
            self.validate_two_decimal_places(scores.shipping_habit, "Shipping Habit")?;
            self.validate_two_decimal_places(
                scores.public_accountability,
                "Public Accountability",
            )?;
            self.validate_two_decimal_places(scores.revenue_testing, "Revenue Testing")?;
            self.validate_two_decimal_places(
                scores.category_total,
                "Strategic Fit category total",
            )?;
        }

        Ok(())
    }

    /// Validate weighted totals based on scores and weights
    fn validate_weighted_totals(
        &self,
        weighted_totals: &LlmWeightedTotals,
        scores: &LlmScores,
    ) -> Result<()> {
        // Mission Alignment: 40% weight
        let expected_mission = scores.mission_alignment.category_total * 0.4;
        if (weighted_totals.mission_alignment - expected_mission).abs() > 0.01 {
            return Err(crate::errors::ApplicationError::Validation(
                crate::errors::ValidationError::invalid_value(
                    "Mission Alignment weighted total",
                    format!("{:.2}", weighted_totals.mission_alignment),
                    format!(
                        "should be {:.2} (category total {:.2} * 0.4)",
                        expected_mission, scores.mission_alignment.category_total
                    ),
                ),
            ));
        }

        // Anti-Challenge Patterns: 35% weight
        let expected_anti_challenge = scores.anti_challenge_patterns.category_total * 0.35;
        if (weighted_totals.anti_challenge_patterns - expected_anti_challenge).abs() > 0.01 {
            return Err(crate::errors::ApplicationError::Validation(
                crate::errors::ValidationError::invalid_value(
                    "Anti-Challenge Patterns weighted total",
                    format!("{:.2}", weighted_totals.anti_challenge_patterns),
                    format!(
                        "should be {:.2} (category total {:.2} * 0.35)",
                        expected_anti_challenge, scores.anti_challenge_patterns.category_total
                    ),
                ),
            ));
        }

        // Strategic Fit: 25% weight
        let expected_strategic = scores.strategic_fit.category_total * 0.25;
        if (weighted_totals.strategic_fit - expected_strategic).abs() > 0.01 {
            return Err(crate::errors::ApplicationError::Validation(
                crate::errors::ValidationError::invalid_value(
                    "Strategic Fit weighted total",
                    format!("{:.2}", weighted_totals.strategic_fit),
                    format!(
                        "should be {:.2} (category total {:.2} * 0.25)",
                        expected_strategic, scores.strategic_fit.category_total
                    ),
                ),
            ));
        }

        // Validate decimal precision
        if self.config.enforce_two_decimal_places {
            self.validate_two_decimal_places(
                weighted_totals.mission_alignment,
                "Mission Alignment weighted total",
            )?;
            self.validate_two_decimal_places(
                weighted_totals.anti_challenge_patterns,
                "Anti-Challenge Patterns weighted total",
            )?;
            self.validate_two_decimal_places(
                weighted_totals.strategic_fit,
                "Strategic Fit weighted total",
            )?;
        }

        Ok(())
    }

    /// Validate final score based on weighted totals
    fn validate_final_score(&self, final_score: f64, scores: &LlmScores) -> Result<()> {
        let expected_final = scores.mission_alignment.category_total * 0.4
            + scores.anti_challenge_patterns.category_total * 0.35
            + scores.strategic_fit.category_total * 0.25;

        if (final_score - expected_final).abs() > 0.01 {
            return Err(crate::errors::ApplicationError::Validation(
                crate::errors::ValidationError::invalid_value(
                    "Final score",
                    format!("{:.2}", final_score),
                    format!("should be {:.2} (sum of weighted totals)", expected_final),
                ),
            ));
        }

        // Validate range
        self.validate_score_range(final_score, 0.0, 10.0, "Final score")?;

        // Validate decimal precision
        if self.config.enforce_two_decimal_places {
            self.validate_two_decimal_places(final_score, "Final score")?;
        }

        Ok(())
    }

    /// Validate recommendation string
    fn validate_recommendation(&self, recommendation: &str) -> Result<()> {
        if !self
            .config
            .allowed_recommendations
            .contains(&recommendation.to_string())
        {
            return Err(crate::errors::ApplicationError::Validation(
                crate::errors::ValidationError::invalid_value(
                    "Recommendation",
                    recommendation.to_string(),
                    format!(
                        "must be one of: {}",
                        self.config.allowed_recommendations.join(", ")
                    ),
                ),
            ));
        }

        Ok(())
    }

    /// Validate explanations map
    fn validate_explanations(&self, explanations: &HashMap<String, String>) -> Result<()> {
        for (key, explanation) in explanations {
            if explanation.len() > self.config.max_explanation_length {
                return Err(crate::errors::ApplicationError::Validation(
                    crate::errors::ValidationError::invalid_value(
                        format!("Explanation for '{}'", key),
                        format!("{} characters", explanation.len()),
                        format!(
                            "must be {} characters or less",
                            self.config.max_explanation_length
                        ),
                    ),
                ));
            }

            if explanation.trim().is_empty() {
                return Err(crate::errors::ApplicationError::Validation(
                    crate::errors::ValidationError::empty_field(format!(
                        "Explanation for '{}'",
                        key
                    )),
                ));
            }
        }

        Ok(())
    }

    /// Validate score is within the specified range
    fn validate_score_range(&self, score: f64, min: f64, max: f64, field_name: &str) -> Result<()> {
        if score < min || score > max {
            return Err(crate::errors::ApplicationError::Validation(
                crate::errors::ValidationError::invalid_value(
                    field_name,
                    format!("{:.2}", score),
                    format!("must be in range [{:.1}, {:.1}]", min, max),
                ),
            ));
        }

        if !score.is_finite() {
            return Err(crate::errors::ApplicationError::Validation(
                crate::errors::ValidationError::invalid_value(
                    field_name,
                    format!("{}", score),
                    "must be a finite number".to_string(),
                ),
            ));
        }

        Ok(())
    }

    /// Validate that a score has exactly two decimal places
    fn validate_two_decimal_places(&self, score: f64, field_name: &str) -> Result<()> {
        // Convert to string and check if it has exactly 2 decimal places
        let score_str = score.to_string();

        // Check if the string representation has exactly 2 digits after the decimal point
        if let Some(decimal_pos) = score_str.find('.') {
            let decimals = &score_str[decimal_pos + 1..];
            if decimals.len() != 2 {
                return Err(crate::errors::ApplicationError::Validation(
                    crate::errors::ValidationError::invalid_value(
                        field_name,
                        format!("{}", score),
                        "must have exactly two decimal places".to_string(),
                    ),
                ));
            }
        } else {
            // No decimal point means 0 decimal places, which is invalid for our requirement
            return Err(crate::errors::ApplicationError::Validation(
                crate::errors::ValidationError::invalid_value(
                    field_name,
                    format!("{}", score),
                    "must have exactly two decimal places".to_string(),
                ),
            ));
        }

        Ok(())
    }
}

/// Enhanced fallback scoring system when LLM analysis fails
pub struct FallbackScorer {
    scoring_engine: Option<ScoringEngine>,
    telos_config: Option<TelosConfig>,
}

impl FallbackScorer {
    /// Create a new fallback scorer without scoring engine (default behavior)
    pub fn new() -> Self {
        Self {
            scoring_engine: None,
            telos_config: None,
        }
    }

    /// Create a fallback scorer with scoring engine integration
    pub fn with_scoring_engine(scoring_engine: ScoringEngine, telos_config: TelosConfig) -> Self {
        Self {
            scoring_engine: Some(scoring_engine),
            telos_config: Some(telos_config),
        }
    }

    /// Generate fallback analysis when LLM analysis fails
    pub async fn generate_fallback_analysis(&self, idea: &str) -> LlmAnalysisResult {
        // Try to use the real scoring engine if available
        if let (Some(scoring_engine), Some(telos_config)) =
            (&self.scoring_engine, &self.telos_config)
        {
            Self::generate_scoring_engine_fallback(scoring_engine, telos_config, idea).await
        } else {
            Self::generate_default_fallback(idea)
        }
    }

    /// Generate fallback using the actual scoring engine
    async fn generate_scoring_engine_fallback(
        scoring_engine: &ScoringEngine,
        _telos_config: &TelosConfig,
        idea: &str,
    ) -> LlmAnalysisResult {
        match scoring_engine.calculate_score(idea) {
            Ok(score) => Self::convert_rule_based_score_to_llm_format(score),
            Err(_) => {
                // If scoring engine fails, fall back to default
                Self::generate_default_fallback(idea)
            }
        }
    }

    /// Convert rule-based score to LLM analysis format
    fn convert_rule_based_score_to_llm_format(score: Score) -> LlmAnalysisResult {
        LlmAnalysisResult {
            scores: LlmScores {
                mission_alignment: MissionAlignmentScores {
                    domain_expertise: score.mission.domain_expertise,
                    ai_alignment: score.mission.ai_alignment,
                    execution_support: score.mission.execution_support,
                    revenue_potential: score.mission.revenue_potential,
                    category_total: score.mission.total,
                },
                anti_challenge_patterns: AntiChallengePatternsScores {
                    avoid_context_switching: score.anti_challenge.context_switching,
                    rapid_prototyping: score.anti_challenge.rapid_prototyping,
                    accountability: score.anti_challenge.accountability,
                    income_anxiety: score.anti_challenge.income_anxiety,
                    category_total: score.anti_challenge.total,
                },
                strategic_fit: StrategicFitScores {
                    stack_compatibility: score.strategic.stack_compatibility,
                    shipping_habit: score.strategic.shipping_habit,
                    public_accountability: score.strategic.public_accountability,
                    revenue_testing: score.strategic.revenue_testing,
                    category_total: score.strategic.total,
                },
            },
            weighted_totals: LlmWeightedTotals {
                mission_alignment: score.mission.total * 0.4,
                anti_challenge_patterns: score.anti_challenge.total * 0.35,
                strategic_fit: score.strategic.total * 0.25,
            },
            final_score: score.final_score,
            recommendation: match score.recommendation {
                crate::scoring::Recommendation::Priority => "Priority".to_string(),
                crate::scoring::Recommendation::Good => "Good".to_string(),
                crate::scoring::Recommendation::Consider => "Consider".to_string(),
                crate::scoring::Recommendation::Avoid => "Avoid".to_string(),
            },
            explanations: score.explanations,
        }
    }

    /// Generate a default fallback analysis
    fn generate_default_fallback(idea: &str) -> LlmAnalysisResult {
        // Default to a moderate score with explanations
        let default_mission = MissionAlignmentScores {
            domain_expertise: 0.60,
            ai_alignment: 0.75,
            execution_support: 0.50,
            revenue_potential: 0.25,
            category_total: 2.10,
        };

        let default_anti_challenge = AntiChallengePatternsScores {
            avoid_context_switching: 0.80,
            rapid_prototyping: 0.60,
            accountability: 0.40,
            income_anxiety: 0.30,
            category_total: 2.10,
        };

        let default_strategic = StrategicFitScores {
            stack_compatibility: 0.60,
            shipping_habit: 0.40,
            public_accountability: 0.20,
            revenue_testing: 0.15,
            category_total: 1.35,
        };

        let final_score = default_mission.category_total * 0.4
            + default_anti_challenge.category_total * 0.35
            + default_strategic.category_total * 0.25;

        let recommendation = if final_score >= 8.5 {
            "Priority".to_string()
        } else if final_score >= 7.0 {
            "Good".to_string()
        } else if final_score >= 5.0 {
            "Consider".to_string()
        } else {
            "Avoid".to_string()
        };

        let mut explanations = HashMap::new();
        explanations.insert(
            "Default".to_string(),
            format!(
                "Default analysis for idea: {}",
                idea.chars().take(50).collect::<String>()
            ),
        );

        LlmAnalysisResult {
            scores: LlmScores {
                mission_alignment: default_mission.clone(),
                anti_challenge_patterns: default_anti_challenge.clone(),
                strategic_fit: default_strategic.clone(),
            },
            weighted_totals: LlmWeightedTotals {
                mission_alignment: default_mission.category_total * 0.4,
                anti_challenge_patterns: default_anti_challenge.category_total * 0.35,
                strategic_fit: default_strategic.category_total * 0.25,
            },
            final_score,
            recommendation,
            explanations,
        }
    }
}

/// Process and validate LLM response with fallback and quality metrics
pub async fn process_llm_response_with_quality(
    response_text: &str,
    idea: &str,
    fallback_scorer: &FallbackScorer,
) -> Result<EnhancedLlmAnalysisResult> {
    let mut quality_metrics = ResponseQualityMetrics::new();
    let mut processing_notes = Vec::new();

    // First, try to extract and validate JSON from response
    let json_result = extract_json_from_response(response_text);

    match json_result {
        Ok(json_str) => {
            quality_metrics.json_parsing_passed = true;
            processing_notes.push("✅ JSON extraction successful".to_string());

            // Try to parse the JSON
            match serde_json::from_str::<LlmAnalysisResult>(&json_str) {
                Ok(mut analysis_result) => {
                    // Validate the parsed result
                    let validator = ResponseValidator::new();

                    match validator.validate_analysis_result(&analysis_result) {
                        Ok(()) => {
                            quality_metrics.validation_passed = true;
                            quality_metrics.scores_in_range = true;
                            quality_metrics.calculations_correct = true;
                            processing_notes.push("✅ All validation checks passed".to_string());

                            // Check if explanations are provided
                            quality_metrics.explanations_provided =
                                !analysis_result.explanations.is_empty();
                            if quality_metrics.explanations_provided {
                                processing_notes.push("✅ Explanations provided".to_string());
                            } else {
                                processing_notes.push("⚠️ No explanations provided".to_string());
                            }

                            // Calculate quality and confidence
                            quality_metrics.calculate_quality_score();
                            quality_metrics.determine_confidence_level();

                            Ok(EnhancedLlmAnalysisResult {
                                base_result: analysis_result,
                                confidence_level: quality_metrics.confidence_level.clone(),
                                quality_metrics,
                                processing_notes,
                            })
                        }
                        Err(validation_error) => {
                            // Validation failed, log the error and try fixing
                            processing_notes
                                .push(format!("⚠️ Validation failed: {}", validation_error));
                            quality_metrics
                                .validation_errors
                                .push(validation_error.to_string());

                            // Try to fix common issues
                            analysis_result = fix_common_validation_issues(analysis_result);
                            processing_notes
                                .push("🔧 Attempted to fix validation issues".to_string());

                            // Validate the fixed result
                            match validator.validate_analysis_result(&analysis_result) {
                                Ok(()) => {
                                    quality_metrics.validation_passed = true;
                                    quality_metrics.scores_in_range = true;
                                    quality_metrics.calculations_correct = true;
                                    processing_notes
                                        .push("✅ Fixed result passed validation".to_string());

                                    quality_metrics.calculate_quality_score();
                                    quality_metrics.determine_confidence_level();

                                    Ok(EnhancedLlmAnalysisResult {
                                        base_result: analysis_result,
                                        confidence_level: quality_metrics.confidence_level.clone(),
                                        quality_metrics,
                                        processing_notes,
                                    })
                                }
                                Err(_) => {
                                    // Even after fixes, validation failed - use fallback
                                    processing_notes.push(
                                        "❌ Fixed result still invalid, using fallback".to_string(),
                                    );
                                    quality_metrics.fallback_used = true;
                                    quality_metrics.calculate_quality_score();
                                    quality_metrics.determine_confidence_level();

                                    let fallback_result =
                                        fallback_scorer.generate_fallback_analysis(idea).await;
                                    Ok(EnhancedLlmAnalysisResult {
                                        base_result: fallback_result,
                                        confidence_level: ConfidenceLevel::VeryLow,
                                        quality_metrics,
                                        processing_notes,
                                    })
                                }
                            }
                        }
                    }
                }
                Err(parse_error) => {
                    // JSON parsing failed, use fallback
                    processing_notes.push(format!("❌ JSON parsing failed: {}", parse_error));
                    quality_metrics.fallback_used = true;
                    quality_metrics.calculate_quality_score();
                    quality_metrics.determine_confidence_level();

                    let fallback_result = fallback_scorer.generate_fallback_analysis(idea).await;
                    Ok(EnhancedLlmAnalysisResult {
                        base_result: fallback_result,
                        confidence_level: ConfidenceLevel::VeryLow,
                        quality_metrics,
                        processing_notes,
                    })
                }
            }
        }
        Err(extraction_error) => {
            // JSON extraction failed, use fallback
            processing_notes.push(format!("❌ JSON extraction failed: {}", extraction_error));
            quality_metrics.fallback_used = true;
            quality_metrics.calculate_quality_score();
            quality_metrics.determine_confidence_level();

            let fallback_result = fallback_scorer.generate_fallback_analysis(idea).await;
            Ok(EnhancedLlmAnalysisResult {
                base_result: fallback_result,
                confidence_level: ConfidenceLevel::VeryLow,
                quality_metrics,
                processing_notes,
            })
        }
    }
}

/// Process and validate LLM response with fallback (legacy function for backward compatibility)
pub async fn process_llm_response(response_text: &str, idea: &str) -> Result<LlmAnalysisResult> {
    let fallback_scorer = FallbackScorer::new();
    let enhanced_result =
        process_llm_response_with_quality(response_text, idea, &fallback_scorer).await?;
    Ok(enhanced_result.base_result)
}

/// Extract JSON from response text (handles cases where LLM returns more than just JSON)
fn extract_json_from_response(response: &str) -> Result<String> {
    // First, look for JSON within triple backticks
    if let Some(start) = response.find("```json") {
        if let Some(end) = response[start..].find("```") {
            let json_str = &response[start + 7..start + end]; // Skip "```json"
            return Ok(json_str.trim().to_string());
        }
    }

    // Look for regular code blocks as well
    if let Some(start) = response.find("```") {
        if let Some(end) = response[start..].find("```") {
            let potential_json = &response[start + 3..start + end]; // Skip "```"
                                                                    // Check if it looks like JSON
            if potential_json.trim_start().starts_with('{') {
                return Ok(potential_json.trim().to_string());
            }
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

/// Fix common validation issues in the analysis result
fn fix_common_validation_issues(mut result: LlmAnalysisResult) -> LlmAnalysisResult {
    // Fix Mission Alignment scores
    result.scores.mission_alignment.domain_expertise = result
        .scores
        .mission_alignment
        .domain_expertise
        .max(0.0)
        .min(1.2);
    result.scores.mission_alignment.ai_alignment = result
        .scores
        .mission_alignment
        .ai_alignment
        .max(0.0)
        .min(1.5);
    result.scores.mission_alignment.execution_support = result
        .scores
        .mission_alignment
        .execution_support
        .max(0.0)
        .min(0.8);
    result.scores.mission_alignment.revenue_potential = result
        .scores
        .mission_alignment
        .revenue_potential
        .max(0.0)
        .min(0.5);

    // Recalculate category total
    result.scores.mission_alignment.category_total =
        result.scores.mission_alignment.domain_expertise
            + result.scores.mission_alignment.ai_alignment
            + result.scores.mission_alignment.execution_support
            + result.scores.mission_alignment.revenue_potential;

    // Fix Anti-Challenge scores
    result
        .scores
        .anti_challenge_patterns
        .avoid_context_switching = result
        .scores
        .anti_challenge_patterns
        .avoid_context_switching
        .max(0.0)
        .min(1.2);
    result.scores.anti_challenge_patterns.rapid_prototyping = result
        .scores
        .anti_challenge_patterns
        .rapid_prototyping
        .max(0.0)
        .min(1.0);
    result.scores.anti_challenge_patterns.accountability = result
        .scores
        .anti_challenge_patterns
        .accountability
        .max(0.0)
        .min(0.8);
    result.scores.anti_challenge_patterns.income_anxiety = result
        .scores
        .anti_challenge_patterns
        .income_anxiety
        .max(0.0)
        .min(0.5);

    // Recalculate category total
    result.scores.anti_challenge_patterns.category_total = result
        .scores
        .anti_challenge_patterns
        .avoid_context_switching
        + result.scores.anti_challenge_patterns.rapid_prototyping
        + result.scores.anti_challenge_patterns.accountability
        + result.scores.anti_challenge_patterns.income_anxiety;

    // Fix Strategic scores
    result.scores.strategic_fit.stack_compatibility = result
        .scores
        .strategic_fit
        .stack_compatibility
        .max(0.0)
        .min(1.0);
    result.scores.strategic_fit.shipping_habit =
        result.scores.strategic_fit.shipping_habit.max(0.0).min(0.8);
    result.scores.strategic_fit.public_accountability = result
        .scores
        .strategic_fit
        .public_accountability
        .max(0.0)
        .min(0.4);
    result.scores.strategic_fit.revenue_testing = result
        .scores
        .strategic_fit
        .revenue_testing
        .max(0.0)
        .min(0.3);

    // Recalculate category total
    result.scores.strategic_fit.category_total = result.scores.strategic_fit.stack_compatibility
        + result.scores.strategic_fit.shipping_habit
        + result.scores.strategic_fit.public_accountability
        + result.scores.strategic_fit.revenue_testing;

    // Recalculate weighted totals
    result.weighted_totals.mission_alignment = result.scores.mission_alignment.category_total * 0.4;
    result.weighted_totals.anti_challenge_patterns =
        result.scores.anti_challenge_patterns.category_total * 0.35;
    result.weighted_totals.strategic_fit = result.scores.strategic_fit.category_total * 0.25;

    // Recalculate final score
    result.final_score = result.weighted_totals.mission_alignment
        + result.weighted_totals.anti_challenge_patterns
        + result.weighted_totals.strategic_fit;

    // Fix recommendation if invalid
    if !["Priority", "Good", "Consider", "Avoid"].contains(&result.recommendation.as_str()) {
        // Determine recommendation based on final score
        result.recommendation = if result.final_score >= 8.5 {
            "Priority".to_string()
        } else if result.final_score >= 7.0 {
            "Good".to_string()
        } else if result.final_score >= 5.0 {
            "Consider".to_string()
        } else {
            "Avoid".to_string()
        };
    }

    // Round scores to 2 decimal places
    result.scores.mission_alignment.domain_expertise =
        (result.scores.mission_alignment.domain_expertise * 100.0).round() / 100.0;
    result.scores.mission_alignment.ai_alignment =
        (result.scores.mission_alignment.ai_alignment * 100.0).round() / 100.0;
    result.scores.mission_alignment.execution_support =
        (result.scores.mission_alignment.execution_support * 100.0).round() / 100.0;
    result.scores.mission_alignment.revenue_potential =
        (result.scores.mission_alignment.revenue_potential * 100.0).round() / 100.0;
    result.scores.mission_alignment.category_total =
        (result.scores.mission_alignment.category_total * 100.0).round() / 100.0;

    result
        .scores
        .anti_challenge_patterns
        .avoid_context_switching = (result
        .scores
        .anti_challenge_patterns
        .avoid_context_switching
        * 100.0)
        .round()
        / 100.0;
    result.scores.anti_challenge_patterns.rapid_prototyping =
        (result.scores.anti_challenge_patterns.rapid_prototyping * 100.0).round() / 100.0;
    result.scores.anti_challenge_patterns.accountability =
        (result.scores.anti_challenge_patterns.accountability * 100.0).round() / 100.0;
    result.scores.anti_challenge_patterns.income_anxiety =
        (result.scores.anti_challenge_patterns.income_anxiety * 100.0).round() / 100.0;
    result.scores.anti_challenge_patterns.category_total =
        (result.scores.anti_challenge_patterns.category_total * 100.0).round() / 100.0;

    result.scores.strategic_fit.stack_compatibility =
        (result.scores.strategic_fit.stack_compatibility * 100.0).round() / 100.0;
    result.scores.strategic_fit.shipping_habit =
        (result.scores.strategic_fit.shipping_habit * 100.0).round() / 100.0;
    result.scores.strategic_fit.public_accountability =
        (result.scores.strategic_fit.public_accountability * 100.0).round() / 100.0;
    result.scores.strategic_fit.revenue_testing =
        (result.scores.strategic_fit.revenue_testing * 100.0).round() / 100.0;
    result.scores.strategic_fit.category_total =
        (result.scores.strategic_fit.category_total * 100.0).round() / 100.0;

    result.weighted_totals.mission_alignment =
        (result.weighted_totals.mission_alignment * 100.0).round() / 100.0;
    result.weighted_totals.anti_challenge_patterns =
        (result.weighted_totals.anti_challenge_patterns * 100.0).round() / 100.0;
    result.weighted_totals.strategic_fit =
        (result.weighted_totals.strategic_fit * 100.0).round() / 100.0;

    result.final_score = (result.final_score * 100.0).round() / 100.0;

    result
}

/// Display enhanced LLM analysis result with confidence and quality metrics
pub fn display_enhanced_analysis_result(idea: &str, enhanced_result: &EnhancedLlmAnalysisResult) {
    println!("\n🤖 Enhanced LLM Analysis Result:");
    println!("════════════════════════════════════════════════════════");
    println!("Idea: {}", idea.chars().take(60).collect::<String>());
    if idea.len() > 60 {
        println!("      ... (truncated)");
    }

    // Confidence level
    println!(
        "\n{} Confidence Level: {}",
        enhanced_result.confidence_level.emoji(),
        enhanced_result.confidence_level.description()
    );
    println!(
        "Quality Score: {:.1}%",
        enhanced_result.quality_metrics.quality_score * 100.0
    );

    // Core scores
    println!(
        "\n🎯 Final Score: {:.2}/10.00",
        enhanced_result.base_result.final_score
    );
    println!(
        "📋 Recommendation: {}",
        enhanced_result.base_result.recommendation
    );

    println!("\n📈 Detailed Scores:");
    println!(
        "  Mission Alignment: {:.2}/4.00",
        enhanced_result
            .base_result
            .scores
            .mission_alignment
            .category_total
    );
    println!(
        "  Anti-Challenge Patterns: {:.2}/3.50",
        enhanced_result
            .base_result
            .scores
            .anti_challenge_patterns
            .category_total
    );
    println!(
        "  Strategic Fit: {:.2}/2.50",
        enhanced_result
            .base_result
            .scores
            .strategic_fit
            .category_total
    );

    println!("\n💡 Weighted Totals:");
    println!(
        "  Mission Alignment: {:.2}",
        enhanced_result
            .base_result
            .weighted_totals
            .mission_alignment
    );
    println!(
        "  Anti-Challenge Patterns: {:.2}",
        enhanced_result
            .base_result
            .weighted_totals
            .anti_challenge_patterns
    );
    println!(
        "  Strategic Fit: {:.2}",
        enhanced_result.base_result.weighted_totals.strategic_fit
    );

    // Quality metrics
    println!("\n📊 Quality Metrics:");
    println!(
        "  JSON Parsing: {} {}",
        if enhanced_result.quality_metrics.json_parsing_passed {
            "✅"
        } else {
            "❌"
        },
        if enhanced_result.quality_metrics.json_parsing_passed {
            "Passed"
        } else {
            "Failed"
        }
    );
    println!(
        "  Validation: {} {}",
        if enhanced_result.quality_metrics.validation_passed {
            "✅"
        } else {
            "❌"
        },
        if enhanced_result.quality_metrics.validation_passed {
            "Passed"
        } else {
            "Failed"
        }
    );
    println!(
        "  Scores in Range: {} {}",
        if enhanced_result.quality_metrics.scores_in_range {
            "✅"
        } else {
            "❌"
        },
        if enhanced_result.quality_metrics.scores_in_range {
            "Valid"
        } else {
            "Invalid"
        }
    );
    println!(
        "  Calculations: {} {}",
        if enhanced_result.quality_metrics.calculations_correct {
            "✅"
        } else {
            "❌"
        },
        if enhanced_result.quality_metrics.calculations_correct {
            "Correct"
        } else {
            "Incorrect"
        }
    );
    println!(
        "  Explanations: {} {}",
        if enhanced_result.quality_metrics.explanations_provided {
            "✅"
        } else {
            "⚠️"
        },
        if enhanced_result.quality_metrics.explanations_provided {
            "Provided"
        } else {
            "Missing"
        }
    );
    println!(
        "  Fallback Used: {} {}",
        if enhanced_result.quality_metrics.fallback_used {
            "⚠️"
        } else {
            "✅"
        },
        if enhanced_result.quality_metrics.fallback_used {
            "Yes"
        } else {
            "No"
        }
    );

    // Processing notes
    if !enhanced_result.processing_notes.is_empty() {
        println!("\n📝 Processing Notes:");
        for note in &enhanced_result.processing_notes {
            println!("  {}", note);
        }
    }

    // Explanations
    if !enhanced_result.base_result.explanations.is_empty() {
        println!("\n💬 Explanations:");
        for (key, explanation) in &enhanced_result.base_result.explanations {
            println!("  {}: {}", key, explanation);
        }
    }

    // Validation errors (if any)
    if !enhanced_result.quality_metrics.validation_errors.is_empty() {
        println!("\n⚠️ Validation Errors:");
        for error in &enhanced_result.quality_metrics.validation_errors {
            println!("  • {}", error);
        }
    }

    println!("════════════════════════════════════════════════════════\n");
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_response_validator_creation() {
        let validator = ResponseValidator::new();
        assert_eq!(validator.config.max_mission_score, 4.0);
        assert_eq!(validator.config.max_anti_challenge_score, 3.5);
        assert_eq!(validator.config.max_strategic_score, 2.5);
    }

    #[test]
    fn test_valid_mission_alignment_scores() {
        let validator = ResponseValidator::new();
        let scores = MissionAlignmentScores {
            domain_expertise: 1.00,
            ai_alignment: 1.20,
            execution_support: 0.70,
            revenue_potential: 0.40,
            category_total: 3.30,
        };

        assert!(validator.validate_mission_alignment_scores(&scores).is_ok());
    }

    #[test]
    fn test_invalid_mission_alignment_scores() {
        let validator = ResponseValidator::new();
        let scores = MissionAlignmentScores {
            domain_expertise: 2.00, // Too high
            ai_alignment: 1.20,
            execution_support: 0.70,
            revenue_potential: 0.40,
            category_total: 4.30, // Doesn't match sum
        };

        assert!(validator
            .validate_mission_alignment_scores(&scores)
            .is_err());
    }

    #[test]
    fn test_valid_recommendation() {
        let validator = ResponseValidator::new();
        assert!(validator.validate_recommendation("Priority").is_ok());
        assert!(validator.validate_recommendation("Good").is_ok());
        assert!(validator.validate_recommendation("Consider").is_ok());
        assert!(validator.validate_recommendation("Avoid").is_ok());
    }

    #[test]
    fn test_invalid_recommendation() {
        let validator = ResponseValidator::new();
        assert!(validator.validate_recommendation("Invalid").is_err());
    }

    #[test]
    fn test_two_decimal_places_validation() {
        let validator = ResponseValidator::new();

        // Valid two decimal places
        assert!(validator.validate_two_decimal_places(1.23, "test").is_ok());
        assert!(validator.validate_two_decimal_places(0.00, "test").is_ok());
        assert!(validator.validate_two_decimal_places(10.50, "test").is_ok());

        // Invalid decimal places - 3 decimal places
        assert!(validator
            .validate_two_decimal_places(1.234, "test")
            .is_err());

        // Invalid decimal places - 1 decimal place
        assert!(validator.validate_two_decimal_places(1.2, "test").is_err());

        // Invalid decimal places - no decimal places (whole number)
        assert!(validator.validate_two_decimal_places(5.0, "test").is_err());
    }

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
        "Domain Expertise": "Directly uses existing skills"
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

    #[tokio::test]
    async fn test_fallback_scorer_generation() {
        let idea = "Test idea for fallback scoring";
        let fallback_scorer = FallbackScorer::new();
        let fallback_result = fallback_scorer.generate_fallback_analysis(idea).await;

        // The fallback should produce a valid result
        assert!(fallback_result.final_score >= 0.0);
        assert!(fallback_result.final_score <= 10.0);
        assert!(["Priority", "Good", "Consider", "Avoid"]
            .contains(&fallback_result.recommendation.as_str()));
    }

    #[test]
    fn test_fix_common_validation_issues() {
        let invalid_result = LlmAnalysisResult {
            scores: LlmScores {
                mission_alignment: MissionAlignmentScores {
                    domain_expertise: 2.0, // Invalid - too high
                    ai_alignment: 1.20,
                    execution_support: 0.70,
                    revenue_potential: 0.40,
                    category_total: 4.30, // Doesn't match sum
                },
                anti_challenge_patterns: AntiChallengePatternsScores {
                    avoid_context_switching: 1.00,
                    rapid_prototyping: 0.80,
                    accountability: 0.60,
                    income_anxiety: 0.40,
                    category_total: 2.80,
                },
                strategic_fit: StrategicFitScores {
                    stack_compatibility: 0.80,
                    shipping_habit: 0.60,
                    public_accountability: 0.30,
                    revenue_testing: 0.20,
                    category_total: 1.90,
                },
            },
            weighted_totals: LlmWeightedTotals {
                mission_alignment: 4.30 * 0.4, // Based on invalid total
                anti_challenge_patterns: 2.80,
                strategic_fit: 1.90,
            },
            final_score: 15.0,                     // Invalid - too high
            recommendation: "Invalid".to_string(), // Invalid recommendation
            explanations: HashMap::new(),
        };

        let fixed_result = fix_common_validation_issues(invalid_result);

        // Check that the fixed values are within valid ranges
        assert!(fixed_result.scores.mission_alignment.domain_expertise <= 1.2);
        assert_eq!(fixed_result.recommendation, "Priority"); // Should be fixed based on score
    }

    #[test]
    fn test_weighted_totals_calculation_validation() {
        let result = LlmAnalysisResult {
            scores: LlmScores {
                mission_alignment: MissionAlignmentScores {
                    domain_expertise: 1.00,
                    ai_alignment: 1.20,
                    execution_support: 0.70,
                    revenue_potential: 0.40,
                    category_total: 3.30,
                },
                anti_challenge_patterns: AntiChallengePatternsScores {
                    avoid_context_switching: 1.00,
                    rapid_prototyping: 0.80,
                    accountability: 0.60,
                    income_anxiety: 0.40,
                    category_total: 2.80,
                },
                strategic_fit: StrategicFitScores {
                    stack_compatibility: 0.80,
                    shipping_habit: 0.60,
                    public_accountability: 0.30,
                    revenue_testing: 0.20,
                    category_total: 1.90,
                },
            },
            weighted_totals: LlmWeightedTotals {
                mission_alignment: 99.0,       // Wrong value
                anti_challenge_patterns: 99.0, // Wrong value
                strategic_fit: 99.0,           // Wrong value
            },
            final_score: 2.78, // Expected: 3.30*0.4 + 2.80*0.35 + 1.90*0.25 = 1.32 + 0.98 + 0.475 = 2.775 ≈ 2.78
            recommendation: "Consider".to_string(),
            explanations: HashMap::new(),
        };

        let validator = ResponseValidator::new();
        // This should fail because weighted totals are incorrect
        assert!(validator.validate_analysis_result(&result).is_err());
    }

    #[test]
    fn test_final_score_calculation_validation() {
        let result = LlmAnalysisResult {
            scores: LlmScores {
                mission_alignment: MissionAlignmentScores {
                    domain_expertise: 1.00,
                    ai_alignment: 1.20,
                    execution_support: 0.70,
                    revenue_potential: 0.40,
                    category_total: 3.30,
                },
                anti_challenge_patterns: AntiChallengePatternsScores {
                    avoid_context_switching: 1.00,
                    rapid_prototyping: 0.80,
                    accountability: 0.60,
                    income_anxiety: 0.40,
                    category_total: 2.80,
                },
                strategic_fit: StrategicFitScores {
                    stack_compatibility: 0.80,
                    shipping_habit: 0.60,
                    public_accountability: 0.30,
                    revenue_testing: 0.20,
                    category_total: 1.90,
                },
            },
            weighted_totals: LlmWeightedTotals {
                mission_alignment: 1.32,       // 3.30 * 0.4
                anti_challenge_patterns: 0.98, // 2.80 * 0.35
                strategic_fit: 0.48,           // 1.90 * 0.25 (rounded)
            },
            final_score: 99.0, // Wrong value - should be 2.78
            recommendation: "Consider".to_string(),
            explanations: HashMap::new(),
        };

        let validator = ResponseValidator::new();
        // This should fail because final score doesn't match weighted totals sum
        assert!(validator.validate_analysis_result(&result).is_err());
    }

    #[test]
    fn test_explanation_validation() {
        let mut result = LlmAnalysisResult {
            scores: LlmScores {
                mission_alignment: MissionAlignmentScores {
                    domain_expertise: 1.00,
                    ai_alignment: 1.20,
                    execution_support: 0.70,
                    revenue_potential: 0.40,
                    category_total: 3.30,
                },
                anti_challenge_patterns: AntiChallengePatternsScores {
                    avoid_context_switching: 1.00,
                    rapid_prototyping: 0.80,
                    accountability: 0.60,
                    income_anxiety: 0.40,
                    category_total: 2.80,
                },
                strategic_fit: StrategicFitScores {
                    stack_compatibility: 0.80,
                    shipping_habit: 0.60,
                    public_accountability: 0.30,
                    revenue_testing: 0.20,
                    category_total: 1.90,
                },
            },
            weighted_totals: LlmWeightedTotals {
                mission_alignment: 1.32,
                anti_challenge_patterns: 0.98,
                strategic_fit: 0.48,
            },
            final_score: 2.78,
            recommendation: "Consider".to_string(),
            explanations: HashMap::new(),
        };

        let validator = ResponseValidator::new();

        // Valid explanation
        result.explanations.insert(
            "test".to_string(),
            "This is a valid explanation".to_string(),
        );
        assert!(validator.validate_analysis_result(&result).is_ok());

        // Empty explanation
        let mut result_with_empty = result.clone();
        result_with_empty
            .explanations
            .insert("empty".to_string(), "".to_string());
        assert!(validator
            .validate_analysis_result(&result_with_empty)
            .is_err());

        // Too long explanation
        let mut result_with_long = result.clone();
        let long_explanation = "a".repeat(600); // Exceeds max length (500)
        result_with_long
            .explanations
            .insert("long".to_string(), long_explanation);
        assert!(validator
            .validate_analysis_result(&result_with_long)
            .is_err());
    }

    #[test]
    fn test_two_decimal_places_validation_comprehensive() {
        let validator = ResponseValidator::new();

        // Valid two decimal places
        let valid_result = LlmAnalysisResult {
            scores: LlmScores {
                mission_alignment: MissionAlignmentScores {
                    domain_expertise: 1.23, // Valid 2 decimals
                    ai_alignment: 1.20,
                    execution_support: 0.70,
                    revenue_potential: 0.40,
                    category_total: 3.30,
                },
                anti_challenge_patterns: AntiChallengePatternsScores {
                    avoid_context_switching: 1.00,
                    rapid_prototyping: 0.80,
                    accountability: 0.60,
                    income_anxiety: 0.40,
                    category_total: 2.80,
                },
                strategic_fit: StrategicFitScores {
                    stack_compatibility: 0.80,
                    shipping_habit: 0.60,
                    public_accountability: 0.30,
                    revenue_testing: 0.20,
                    category_total: 1.90,
                },
            },
            weighted_totals: LlmWeightedTotals {
                mission_alignment: 1.32,
                anti_challenge_patterns: 0.98,
                strategic_fit: 0.48,
            },
            final_score: 2.78,
            recommendation: "Consider".to_string(),
            explanations: HashMap::new(),
        };

        assert!(validator.validate_analysis_result(&valid_result).is_ok());

        // Invalid - one decimal place
        let mut invalid_result = valid_result.clone();
        invalid_result.scores.mission_alignment.domain_expertise = 1.2; // Only one decimal
        assert!(validator.validate_analysis_result(&invalid_result).is_err());

        // Invalid - three decimal places
        let mut invalid_result = valid_result.clone();
        invalid_result.scores.mission_alignment.domain_expertise = 1.234; // Three decimals
        assert!(validator.validate_analysis_result(&invalid_result).is_err());
    }

    #[test]
    fn test_response_validator_custom_config() {
        let custom_config = ResponseValidationConfig {
            max_mission_score: 5.0,        // Custom value
            max_anti_challenge_score: 4.0, // Custom value
            max_strategic_score: 3.0,      // Custom value
            allowed_recommendations: vec!["Priority".to_string(), "Avoid".to_string()], // Custom allowed values
            enforce_two_decimal_places: true,
            max_explanation_length: 1000, // Custom max length
            high_confidence_threshold: 0.9,
            medium_confidence_threshold: 0.7,
            low_confidence_threshold: 0.5,
        };

        let validator = ResponseValidator::with_config(custom_config);
        assert_eq!(validator.config.max_mission_score, 5.0);
        assert_eq!(validator.config.max_anti_challenge_score, 4.0);
        assert_eq!(validator.config.max_strategic_score, 3.0);
        assert_eq!(
            validator.config.allowed_recommendations,
            vec!["Priority".to_string(), "Avoid".to_string()]
        );
    }

    #[tokio::test]
    async fn test_process_llm_response_with_valid_json() {
        let valid_json = r#"{
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
                "Domain Expertise": "Valid explanation for domain expertise"
            }
        }"#;

        let result = process_llm_response(valid_json, "Test idea").await;
        assert!(result.is_ok());

        let analysis = result.unwrap();
        assert_eq!(analysis.final_score, 2.78);
        assert_eq!(analysis.recommendation, "Consider");
    }

    #[tokio::test]
    async fn test_process_llm_response_with_json_backticks() {
        let response_with_backticks = r#"Here's the analysis:
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
        "Mission Alignment": 1.32,
        "Anti-Challenge Patterns": 0.98,
        "Strategic Fit": 0.48
    },
    "final_score": 2.78,
    "recommendation": "Consider",
    "explanations": {
        "Domain Expertise": "Valid explanation for domain expertise"
    }
}
```
This was a comprehensive analysis."#;

        let result = process_llm_response(response_with_backticks, "Test idea").await;
        assert!(result.is_ok());

        let analysis = result.unwrap();
        assert_eq!(analysis.final_score, 2.78);
        assert_eq!(analysis.recommendation, "Consider");
    }

    #[tokio::test]
    async fn test_process_llm_response_with_invalid_json() {
        let invalid_json = r#"{
            "scores": {
                "Mission Alignment": {
                    "Domain Expertise": 2.00,
                    "AI Alignment": 1.20,
                    "Execution Support": 0.70,
                    "Revenue Potential": 0.40,
                    "category_total": 4.30
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
                "Mission Alignment": 1.72,
                "Anti-Challenge Patterns": 0.98,
                "Strategic Fit": 0.48
            },
            "final_score": 3.18,
            "recommendation": "Invalid",
            "explanations": {
                "Domain Expertise": "Valid explanation"
            }
        }"#;

        // This should trigger fallback scoring since the JSON has validation errors
        let result = process_llm_response(invalid_json, "Test idea").await;
        assert!(result.is_ok());

        // The fallback scorer should return a valid result
        let analysis = result.unwrap();
        assert!(analysis.final_score >= 0.0 && analysis.final_score <= 10.0);
        assert!(
            ["Priority", "Good", "Consider", "Avoid"].contains(&analysis.recommendation.as_str())
        );
    }

    #[tokio::test]
    async fn test_process_llm_response_with_malformed_json() {
        let malformed_json = r#"{
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
        "#; // Missing closing brace

        // This should trigger fallback scoring due to JSON parsing error
        let result = process_llm_response(malformed_json, "Test idea").await;
        assert!(result.is_ok());

        // The fallback scorer should return a valid result
        let analysis = result.unwrap();
        assert!(analysis.final_score >= 0.0 && analysis.final_score <= 10.0);
        assert!(
            ["Priority", "Good", "Consider", "Avoid"].contains(&analysis.recommendation.as_str())
        );
    }

    #[tokio::test]
    async fn test_process_llm_response_with_empty_response() {
        let empty_response = "";

        // This should trigger fallback scoring
        let result = process_llm_response(empty_response, "Test idea").await;
        assert!(result.is_ok());

        // The fallback scorer should return a valid result
        let analysis = result.unwrap();
        assert!(analysis.final_score >= 0.0 && analysis.final_score <= 10.0);
        assert!(
            ["Priority", "Good", "Consider", "Avoid"].contains(&analysis.recommendation.as_str())
        );
    }

    #[tokio::test]
    async fn test_enhanced_response_processing_with_valid_data() {
        let valid_json = r#"{
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
                "Domain Expertise": "Valid explanation for domain expertise"
            }
        }"#;

        let fallback_scorer = FallbackScorer::new();
        let result =
            process_llm_response_with_quality(valid_json, "Test idea", &fallback_scorer).await;
        assert!(result.is_ok());

        let enhanced_analysis = result.unwrap();
        assert_eq!(enhanced_analysis.confidence_level, ConfidenceLevel::High);
        assert!(enhanced_analysis.quality_metrics.quality_score >= 0.9);
        assert!(enhanced_analysis.quality_metrics.validation_passed);
        assert!(enhanced_analysis.quality_metrics.json_parsing_passed);
        assert!(!enhanced_analysis.quality_metrics.fallback_used);
    }

    #[tokio::test]
    async fn test_enhanced_response_processing_with_invalid_data() {
        let invalid_json = r#"{
            "scores": {
                "Mission Alignment": {
                    "Domain Expertise": 2.00,
                    "AI Alignment": 1.20,
                    "Execution Support": 0.70,
                    "Revenue Potential": 0.40,
                    "category_total": 4.30
                }
            },
            "final_score": 15.0,
            "recommendation": "Invalid"
        }"#;

        let fallback_scorer = FallbackScorer::new();
        let result =
            process_llm_response_with_quality(invalid_json, "Test idea", &fallback_scorer).await;
        assert!(result.is_ok());

        let enhanced_analysis = result.unwrap();
        assert_eq!(enhanced_analysis.confidence_level, ConfidenceLevel::VeryLow);
        assert!(enhanced_analysis.quality_metrics.fallback_used);
        assert!(!enhanced_analysis.processing_notes.is_empty());
    }

    #[test]
    fn test_confidence_level_properties() {
        assert_eq!(ConfidenceLevel::High.emoji(), "🟢");
        assert_eq!(ConfidenceLevel::High.score(), 0.9);
        assert!(ConfidenceLevel::High
            .description()
            .contains("High confidence"));

        assert_eq!(ConfidenceLevel::Medium.emoji(), "🟡");
        assert_eq!(ConfidenceLevel::Medium.score(), 0.7);

        assert_eq!(ConfidenceLevel::Low.emoji(), "🟠");
        assert_eq!(ConfidenceLevel::Low.score(), 0.5);

        assert_eq!(ConfidenceLevel::VeryLow.emoji(), "🔴");
        assert_eq!(ConfidenceLevel::VeryLow.score(), 0.3);
    }

    #[test]
    fn test_response_quality_metrics() {
        let mut metrics = ResponseQualityMetrics::new();

        // Test initial state
        assert!(!metrics.validation_passed);
        assert!(!metrics.json_parsing_passed);
        assert_eq!(metrics.quality_score, 0.0);
        assert_eq!(metrics.confidence_level, ConfidenceLevel::VeryLow);

        // Test with successful validation
        metrics.json_parsing_passed = true;
        metrics.scores_in_range = true;
        metrics.calculations_correct = true;
        metrics.explanations_provided = true;
        metrics.fallback_used = false;

        metrics.calculate_quality_score();
        metrics.determine_confidence_level();

        assert!(metrics.quality_score > 0.8);
        assert_eq!(metrics.confidence_level, ConfidenceLevel::High);
    }

    #[test]
    fn test_enhanced_llm_analysis_result_structure() {
        let base_result = FallbackScorer::generate_default_fallback("Test idea");
        let enhanced_result = EnhancedLlmAnalysisResult {
            base_result,
            confidence_level: ConfidenceLevel::Medium,
            quality_metrics: ResponseQualityMetrics::new(),
            processing_notes: vec!["Test note".to_string()],
        };

        assert_eq!(enhanced_result.confidence_level, ConfidenceLevel::Medium);
        assert_eq!(enhanced_result.processing_notes.len(), 1);
        assert_eq!(enhanced_result.processing_notes[0], "Test note");
    }
}
