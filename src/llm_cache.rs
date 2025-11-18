//! LLM Response Caching System with Semantic Similarity
//!
//! This module provides intelligent caching for LLM responses to reduce API calls
//! and improve performance through semantic similarity matching.

use crate::commands::analyze_llm::{LlmAnalysisResult, LlmProvider};
use crate::errors::Result;
use crate::implementations::InMemoryCache;
use crate::traits::{Cache, CacheStats};
use regex::Regex;
use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use std::sync::Arc;
use std::sync::OnceLock;
use std::time::{Duration, SystemTime, UNIX_EPOCH};
use tokio::sync::RwLock;

/// Semantic similarity threshold for cache hits (0.0 - 1.0)
pub const DEFAULT_SIMILARITY_THRESHOLD: f64 = 0.85;

/// Default TTL for cached LLM responses (24 hours)
pub const DEFAULT_CACHE_TTL: Duration = Duration::from_secs(24 * 60 * 60);

/// Maximum number of cached responses per idea type
pub const MAX_CACHE_SIZE_PER_TYPE: usize = 1000;

/// LLM response cache entry with similarity metadata
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct LlmCacheEntry {
    /// The cached LLM analysis result
    pub analysis_result: LlmAnalysisResult,
    /// The original idea text (normalized)
    pub normalized_idea: String,
    /// Provider that generated this response
    pub provider: LlmProvider,
    /// Timestamp when cached
    pub cached_at: u64,
    /// Number of times this cache entry has been used
    pub hit_count: u64,
    /// Similarity score when last matched (if applicable)
    pub last_similarity_score: Option<f64>,
    /// Confidence level of the original analysis
    pub confidence_level: String,
    /// Quality score of the original analysis
    pub quality_score: f64,
}

impl LlmCacheEntry {
    /// Create a new cache entry
    pub fn new(
        analysis_result: LlmAnalysisResult,
        normalized_idea: String,
        provider: LlmProvider,
        confidence_level: String,
        quality_score: f64,
    ) -> Self {
        let now = SystemTime::now()
            .duration_since(UNIX_EPOCH)
            .unwrap()
            .as_secs();

        Self {
            analysis_result,
            normalized_idea,
            provider,
            cached_at: now,
            hit_count: 0,
            last_similarity_score: None,
            confidence_level,
            quality_score,
        }
    }

    /// Check if this entry is expired
    pub fn is_expired(&self, ttl: Duration) -> bool {
        let now = SystemTime::now()
            .duration_since(UNIX_EPOCH)
            .unwrap()
            .as_secs();

        let age_seconds = now.saturating_sub(self.cached_at);
        age_seconds > ttl.as_secs()
    }

    /// Record a cache hit
    pub fn record_hit(&mut self, similarity_score: Option<f64>) {
        self.hit_count += 1;
        self.last_similarity_score = similarity_score;
    }

    /// Calculate cache entry score for eviction policy (higher is better)
    pub fn cache_score(&self) -> f64 {
        let age_hours = (SystemTime::now()
            .duration_since(UNIX_EPOCH)
            .unwrap()
            .as_secs()
            .saturating_sub(self.cached_at) as f64)
            / 3600.0;

        // Balance recency (lower age) and hit frequency
        let hit_rate = if age_hours > 0.0 {
            self.hit_count as f64 / age_hours
        } else {
            self.hit_count as f64
        };
        let recency_factor = 1.0 / (1.0 + age_hours / 24.0); // Decay over days

        (hit_rate * 0.6) + (recency_factor * 0.4)
    }
}

/// Cache key for LLM responses with semantic grouping
#[derive(Debug, Clone, Hash, Eq, PartialEq, Serialize, Deserialize)]
pub struct LlmCacheKey {
    /// Normalized idea type/category
    pub idea_type: String,
    /// Idea length category (short, medium, long)
    pub length_category: String,
    /// Provider type
    pub provider_type: String,
    /// Prompt version/hash
    pub prompt_version: String,
}

impl LlmCacheKey {
    /// Create cache key from idea and provider
    pub fn from_idea_and_provider(
        idea: &str,
        provider: &LlmProvider,
        prompt_version: &str,
    ) -> Self {
        let idea_type = Self::classify_idea_type(idea);
        let length_category = Self::classify_length(idea);

        Self {
            idea_type,
            length_category,
            provider_type: provider.provider_type(),
            prompt_version: prompt_version.to_string(),
        }
    }

    /// Classify idea into semantic categories
    fn classify_idea_type(idea: &str) -> String {
        let idea_lower = idea.to_lowercase();

        // Technical/Development ideas
        if idea_lower.contains("app")
            || idea_lower.contains("software")
            || idea_lower.contains("code")
            || idea_lower.contains("programming")
            || idea_lower.contains("development")
            || idea_lower.contains("tech")
        {
            return "technical".to_string();
        }

        // Business/Startup ideas
        if idea_lower.contains("business")
            || idea_lower.contains("startup")
            || idea_lower.contains("company")
            || idea_lower.contains("service")
            || idea_lower.contains("product")
            || idea_lower.contains("market")
        {
            return "business".to_string();
        }

        // Content/Creative ideas
        if idea_lower.contains("blog")
            || idea_lower.contains("content")
            || idea_lower.contains("video")
            || idea_lower.contains("write")
            || idea_lower.contains("creative")
            || idea_lower.contains("art")
        {
            return "content".to_string();
        }

        // Learning/Research ideas
        if idea_lower.contains("learn")
            || idea_lower.contains("study")
            || idea_lower.contains("research")
            || idea_lower.contains("course")
            || idea_lower.contains("book")
            || idea_lower.contains("education")
        {
            return "learning".to_string();
        }

        // Personal/Productivity ideas
        if idea_lower.contains("habit")
            || idea_lower.contains("personal")
            || idea_lower.contains("productivity")
            || idea_lower.contains("health")
            || idea_lower.contains("fitness")
            || idea_lower.contains("life")
        {
            return "personal".to_string();
        }

        "general".to_string()
    }

    /// Classify idea length
    fn classify_length(idea: &str) -> String {
        let word_count = idea.split_whitespace().count();

        if word_count <= 10 {
            "short".to_string()
        } else if word_count <= 25 {
            "medium".to_string()
        } else {
            "long".to_string()
        }
    }
}

/// Text normalizer for semantic comparison
pub struct TextNormalizer;

impl TextNormalizer {
    /// Normalize text for semantic comparison
    pub fn normalize(text: &str) -> String {
        static WORD_REGEX: OnceLock<Regex> = OnceLock::new();
        let word_regex = WORD_REGEX.get_or_init(|| Regex::new(r"[a-zA-Z]+").unwrap());

        // Convert to lowercase and extract words
        let words: Vec<String> = word_regex
            .find_iter(text.to_lowercase().as_str())
            .map(|m| m.as_str().to_string())
            .collect();

        words.join(" ")
    }

    /// Extract key terms from text (removes common stop words)
    pub fn extract_key_terms(text: &str) -> Vec<String> {
        static STOP_WORDS: OnceLock<std::collections::HashSet<&str>> = OnceLock::new();
        let stop_words = STOP_WORDS.get_or_init(|| {
            vec![
                "a", "an", "and", "are", "as", "at", "be", "by", "for", "from", "has", "he", "in",
                "is", "it", "its", "of", "on", "that", "the", "to", "was", "were", "will", "with",
                "the", "this", "but", "they", "have", "had", "what", "said", "each", "which",
                "their", "time", "if", "up", "out", "many", "then", "them", "can", "would",
                "there", "all", "so", "also", "her", "much", "more", "very", "she", "may", "these",
                "his", "see", "way", "had", "now", "who", "oil", "sit", "its", "yes", "cold",
                "tell", "try", "take", "why", "help", "put", "say", "much", "too", "how", "our",
                "work", "first", "well", "way", "even", "new", "because", "use", "her", "make",
                "two", "being", "other", "after", "here", "how", "only", "look", "such", "take",
                "time", "think", "come", "made",
            ]
            .into_iter()
            .collect()
        });

        let normalized = Self::normalize(text);
        normalized
            .split_whitespace()
            .filter(|word| !stop_words.contains(*word) && word.len() > 2)
            .map(|word| word.to_string())
            .collect::<std::collections::HashSet<String>>()
            .into_iter()
            .collect()
    }
}

/// Semantic similarity calculator
pub struct SemanticSimilarity;

impl SemanticSimilarity {
    /// Calculate Jaccard similarity between two texts
    pub fn jaccard_similarity(text1: &str, text2: &str) -> f64 {
        let terms1 = TextNormalizer::extract_key_terms(text1);
        let terms2 = TextNormalizer::extract_key_terms(text2);

        if terms1.is_empty() && terms2.is_empty() {
            return 1.0;
        }

        if terms1.is_empty() || terms2.is_empty() {
            return 0.0;
        }

        let set1: std::collections::HashSet<_> = terms1.iter().collect();
        let set2: std::collections::HashSet<_> = terms2.iter().collect();

        let intersection = set1.intersection(&set2).count();
        let union = set1.union(&set2).count();

        if union == 0 {
            0.0
        } else {
            intersection as f64 / union as f64
        }
    }

    /// Calculate cosine similarity between term vectors
    pub fn cosine_similarity(text1: &str, text2: &str) -> f64 {
        let terms1 = TextNormalizer::extract_key_terms(text1);
        let terms2 = TextNormalizer::extract_key_terms(text2);

        if terms1.is_empty() && terms2.is_empty() {
            return 1.0;
        }

        if terms1.is_empty() || terms2.is_empty() {
            return 0.0;
        }

        // Create term frequency maps
        let mut tf1 = HashMap::new();
        let mut tf2 = HashMap::new();

        for term in &terms1 {
            *tf1.entry(term).or_insert(0) += 1;
        }

        for term in &terms2 {
            *tf2.entry(term).or_insert(0) += 1;
        }

        // Calculate dot product and magnitudes
        let mut dot_product = 0.0;
        let mut mag1 = 0.0;
        let mut mag2 = 0.0;

        // Union of all terms
        let all_terms: std::collections::HashSet<_> = tf1.keys().chain(tf2.keys()).collect();

        for term in all_terms {
            let f1 = *tf1.get(term).unwrap_or(&0) as f64;
            let f2 = *tf2.get(term).unwrap_or(&0) as f64;

            dot_product += f1 * f2;
            mag1 += f1 * f1;
            mag2 += f2 * f2;
        }

        if mag1 == 0.0 || mag2 == 0.0 {
            0.0
        } else {
            dot_product / (mag1.sqrt() * mag2.sqrt())
        }
    }

    /// Combined similarity score (weighted average of Jaccard and Cosine)
    pub fn combined_similarity(text1: &str, text2: &str) -> f64 {
        let jaccard = Self::jaccard_similarity(text1, text2);
        let cosine = Self::cosine_similarity(text1, text2);

        // Weight cosine similarity slightly higher as it captures term frequency
        (jaccard * 0.4) + (cosine * 0.6)
    }

    /// Check if two ideas are semantically similar enough
    pub fn is_similar_enough(idea1: &str, idea2: &str, threshold: f64) -> bool {
        let similarity = Self::combined_similarity(idea1, idea2);
        similarity >= threshold
    }
}

/// LLM Response Cache with semantic similarity matching
pub struct LlmResponseCache {
    /// Underlying in-memory cache for exact matches
    exact_cache: InMemoryCache<String, LlmCacheEntry>,
    /// Cache for similar ideas grouped by semantic key
    semantic_cache: Arc<RwLock<HashMap<LlmCacheKey, Vec<LlmCacheEntry>>>>,
    /// Similarity threshold for cache hits
    similarity_threshold: f64,
    /// Default TTL for cache entries
    default_ttl: Duration,
    /// Cache statistics
    stats: Arc<RwLock<LlmCacheStats>>,
}

/// Enhanced cache statistics for LLM responses
#[derive(Debug, Clone, Default)]
pub struct LlmCacheStats {
    /// Basic cache stats
    pub basic_stats: CacheStats,
    /// Semantic hits (similar but not identical ideas)
    pub semantic_hits: u64,
    /// Semantic misses (no similar ideas found)
    pub semantic_misses: u64,
    /// Total similarity score of all semantic hits
    pub total_similarity_score: f64,
    /// Number of cache evictions
    pub evictions: u64,
    /// Number of ideas by type
    pub ideas_by_type: HashMap<String, usize>,
}

impl LlmResponseCache {
    /// Create a new LLM response cache
    pub fn new() -> Self {
        Self::with_config(DEFAULT_SIMILARITY_THRESHOLD, DEFAULT_CACHE_TTL)
    }

    /// Create a cache with custom similarity threshold and TTL
    pub fn with_config(similarity_threshold: f64, default_ttl: Duration) -> Self {
        Self {
            exact_cache: InMemoryCache::new(),
            semantic_cache: Arc::new(RwLock::new(HashMap::new())),
            similarity_threshold: similarity_threshold.clamp(0.0, 1.0),
            default_ttl,
            stats: Arc::new(RwLock::new(LlmCacheStats::default())),
        }
    }

    /// Try to get a cached response for the given idea
    pub async fn get_response(
        &self,
        idea: &str,
        provider: &LlmProvider,
        prompt_version: &str,
    ) -> Result<Option<(LlmAnalysisResult, f64)>> {
        // First try exact match with normalized idea
        let normalized_idea = TextNormalizer::normalize(idea);

        match self.exact_cache.get(&normalized_idea).await {
            Ok(Some(mut entry)) => {
                if !entry.is_expired(self.default_ttl) {
                    entry.record_hit(Some(1.0)); // Exact match has 1.0 similarity

                    // Update the entry in cache
                    let _ = self
                        .exact_cache
                        .set(normalized_idea.clone(), entry, Some(self.default_ttl))
                        .await;

                    self.update_stats(|stats| {
                        stats.basic_stats.hits += 1;
                    })
                    .await;

                    // Get the updated entry
                    if let Ok(Some(updated_entry)) = self.exact_cache.get(&normalized_idea).await {
                        return Ok(Some((updated_entry.analysis_result, 1.0)));
                    }
                } else {
                    // Remove expired entry
                    let _ = self.exact_cache.remove(&normalized_idea).await;
                }
            }
            Ok(None) => {
                // No exact match, try semantic match
                self.update_stats(|stats| {
                    stats.basic_stats.misses += 1;
                })
                .await;
            }
            Err(_) => {
                self.update_stats(|stats| {
                    stats.basic_stats.misses += 1;
                })
                .await;
            }
        }

        // Try semantic similarity match
        let cache_key = LlmCacheKey::from_idea_and_provider(idea, provider, prompt_version);

        let mut semantic_cache = self.semantic_cache.write().await;
        let entries = semantic_cache
            .entry(cache_key.clone())
            .or_insert_with(Vec::new);

        // Remove expired entries
        entries.retain(|entry| !entry.is_expired(self.default_ttl));

        // Find the best matching entry based on similarity
        let mut best_match: Option<(usize, f64, LlmCacheEntry)> = None;

        for (index, entry) in entries.iter().enumerate() {
            let similarity = SemanticSimilarity::combined_similarity(idea, &entry.normalized_idea);

            if similarity >= self.similarity_threshold {
                match &best_match {
                    None => {
                        best_match = Some((index, similarity, entry.clone()));
                    }
                    Some((_, current_similarity, _)) => {
                        if similarity > *current_similarity {
                            best_match = Some((index, similarity, entry.clone()));
                        }
                    }
                }
            }
        }

        if let Some((index, similarity, mut entry)) = best_match {
            // Record the hit and update entry
            entry.record_hit(Some(similarity));
            entries[index] = entry.clone();

            // Update stats
            self.update_stats(|stats| {
                stats.semantic_hits += 1;
                stats.total_similarity_score += similarity;
            })
            .await;

            drop(semantic_cache);

            return Ok(Some((entry.analysis_result, similarity)));
        } else {
            // No semantic match found
            self.update_stats(|stats| {
                stats.semantic_misses += 1;
            })
            .await;
        }

        Ok(None)
    }

    /// Cache a response for the given idea
    pub async fn cache_response(
        &self,
        idea: &str,
        analysis_result: &LlmAnalysisResult,
        provider: &LlmProvider,
        prompt_version: &str,
        confidence_level: &str,
        quality_score: f64,
    ) -> Result<()> {
        let normalized_idea = TextNormalizer::normalize(idea);
        let cache_key = LlmCacheKey::from_idea_and_provider(idea, provider, prompt_version);

        // Create cache entry
        let entry = LlmCacheEntry::new(
            analysis_result.clone(),
            normalized_idea.clone(),
            provider.clone(),
            confidence_level.to_string(),
            quality_score,
        );

        // Cache in exact match cache
        self.exact_cache
            .set(
                normalized_idea.clone(),
                entry.clone(),
                Some(self.default_ttl),
            )
            .await
            .map_err(|_| {
                crate::errors::ApplicationError::Generic(anyhow::anyhow!(
                    "Failed to cache LLM response"
                ))
            })?;

        // Cache in semantic cache
        let mut semantic_cache = self.semantic_cache.write().await;
        let entries = semantic_cache
            .entry(cache_key.clone())
            .or_insert_with(Vec::new);

        // Check if we need to evict entries
        if entries.len() >= MAX_CACHE_SIZE_PER_TYPE {
            // Sort entries by cache score (lowest first) and remove the worst
            entries.sort_by(|a, b| a.cache_score().partial_cmp(&b.cache_score()).unwrap());
            entries.remove(0);

            self.update_stats(|stats| {
                stats.evictions += 1;
            })
            .await;
        }

        entries.push(entry);

        // Update ideas_by_type stats
        self.update_stats(|stats| {
            *stats
                .ideas_by_type
                .entry(cache_key.idea_type.clone())
                .or_insert(0) += 1;
        })
        .await;

        Ok(())
    }

    /// Get enhanced cache statistics
    pub async fn get_stats(&self) -> LlmCacheStats {
        let basic_stats = self.exact_cache.stats().await;
        let mut stats = self.stats.write().await;
        stats.basic_stats = basic_stats;
        stats.clone()
    }

    /// Clear all cached responses
    pub async fn clear(&self) -> Result<()> {
        self.exact_cache.clear().await.map_err(|_| {
            crate::errors::ApplicationError::Generic(anyhow::anyhow!("Failed to clear exact cache"))
        })?;
        self.semantic_cache.write().await.clear();

        self.update_stats(|stats| {
            *stats = LlmCacheStats::default();
        })
        .await;

        Ok(())
    }

    /// Clean up expired entries
    pub async fn cleanup_expired(&self) -> Result<usize> {
        let mut total_removed = 0;

        // Cleanup semantic cache
        let mut semantic_cache = self.semantic_cache.write().await;
        for (_, entries) in semantic_cache.iter_mut() {
            let initial_len = entries.len();
            entries.retain(|entry| !entry.is_expired(self.default_ttl));
            total_removed += initial_len - entries.len();
        }

        // Remove empty cache keys
        semantic_cache.retain(|_, entries| !entries.is_empty());

        Ok(total_removed)
    }

    /// Update cache statistics asynchronously
    async fn update_stats<F>(&self, updater: F)
    where
        F: FnOnce(&mut LlmCacheStats),
    {
        let mut stats = self.stats.write().await;
        updater(&mut stats);
    }

    /// Get cache effectiveness metrics
    pub async fn get_effectiveness_metrics(&self) -> CacheEffectivenessMetrics {
        let stats = self.get_stats().await;

        let total_requests = stats.basic_stats.hits + stats.basic_stats.misses;
        let exact_hit_rate = if total_requests > 0 {
            stats.basic_stats.hits as f64 / total_requests as f64
        } else {
            0.0
        };

        let semantic_requests = stats.semantic_hits + stats.semantic_misses;
        let semantic_hit_rate = if semantic_requests > 0 {
            stats.semantic_hits as f64 / semantic_requests as f64
        } else {
            0.0
        };

        let average_similarity = if stats.semantic_hits > 0 {
            stats.total_similarity_score / stats.semantic_hits as f64
        } else {
            0.0
        };

        let overall_hit_rate = exact_hit_rate + (semantic_hit_rate * (1.0 - exact_hit_rate));

        CacheEffectivenessMetrics {
            exact_hit_rate,
            semantic_hit_rate,
            overall_hit_rate,
            average_similarity,
            total_cache_size: stats.basic_stats.size,
            eviction_rate: if stats.basic_stats.size > 0 {
                stats.evictions as f64 / stats.basic_stats.size as f64
            } else {
                0.0
            },
            ideas_by_type_distribution: stats.ideas_by_type,
        }
    }
}

impl Default for LlmResponseCache {
    fn default() -> Self {
        Self::new()
    }
}

/// Cache effectiveness metrics
#[derive(Debug, Clone)]
pub struct CacheEffectivenessMetrics {
    pub exact_hit_rate: f64,
    pub semantic_hit_rate: f64,
    pub overall_hit_rate: f64,
    pub average_similarity: f64,
    pub total_cache_size: usize,
    pub eviction_rate: f64,
    pub ideas_by_type_distribution: HashMap<String, usize>,
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::commands::analyze_llm::{
        AntiChallengePatternsScores, LlmScores, LlmWeightedTotals, MissionAlignmentScores,
        StrategicFitScores,
    };

    fn create_test_analysis(score: f64) -> LlmAnalysisResult {
        LlmAnalysisResult {
            scores: LlmScores {
                mission_alignment: MissionAlignmentScores {
                    domain_expertise: 1.0,
                    ai_alignment: 1.2,
                    execution_support: 0.7,
                    revenue_potential: 0.4,
                    category_total: 3.3,
                },
                anti_challenge_patterns: AntiChallengePatternsScores {
                    avoid_context_switching: 1.0,
                    rapid_prototyping: 0.8,
                    accountability: 0.6,
                    income_anxiety: 0.4,
                    category_total: 2.8,
                },
                strategic_fit: StrategicFitScores {
                    stack_compatibility: 0.8,
                    shipping_habit: 0.6,
                    public_accountability: 0.3,
                    revenue_testing: 0.2,
                    category_total: 1.9,
                },
            },
            weighted_totals: LlmWeightedTotals {
                mission_alignment: 1.32,
                anti_challenge_patterns: 0.98,
                strategic_fit: 0.48,
            },
            final_score: score,
            recommendation: "Good".to_string(),
            explanations: std::collections::HashMap::new(),
        }
    }

    fn create_test_provider() -> LlmProvider {
        LlmProvider::OpenAi
    }

    #[tokio::test]
    async fn test_text_normalization() {
        let text1 = "Create a mobile app for tracking fitness goals!";
        let text2 = "I want to create a mobile application for tracking fitness goals.";

        let normalized1 = TextNormalizer::normalize(text1);
        let normalized2 = TextNormalizer::normalize(text2);

        assert_eq!(
            normalized1,
            "create a mobile app for tracking fitness goals"
        );
        assert_eq!(
            normalized2,
            "i want to create a mobile application for tracking fitness goals"
        );
    }

    #[test]
    fn test_semantic_similarity() {
        let idea1 = "Build a mobile app for fitness tracking";
        let idea2 = "Create an application for tracking exercise goals";
        let idea3 = "Write a blog post about cooking recipes";

        let sim_1_2 = SemanticSimilarity::combined_similarity(idea1, idea2);
        let sim_1_3 = SemanticSimilarity::combined_similarity(idea1, idea3);

        assert!(sim_1_2 > 0.5); // Should be similar
        assert!(sim_1_3 < 0.3); // Should be dissimilar
    }

    #[test]
    fn test_idea_classification() {
        assert_eq!(
            LlmCacheKey::classify_idea_type("Create a mobile app"),
            "technical"
        );
        assert_eq!(
            LlmCacheKey::classify_idea_type("Start a business"),
            "business"
        );
        assert_eq!(LlmCacheKey::classify_idea_type("Write a blog"), "content");
        assert_eq!(LlmCacheKey::classify_idea_type("Learn Python"), "learning");
        assert_eq!(LlmCacheKey::classify_idea_type("Build a habit"), "personal");
        assert_eq!(LlmCacheKey::classify_idea_type("Random idea"), "general");
    }

    #[test]
    fn test_length_classification() {
        assert_eq!(LlmCacheKey::classify_length("Short idea"), "short");
        assert_eq!(
            LlmCacheKey::classify_length("This is a medium length idea with several words"),
            "medium"
        );
        assert_eq!(LlmCacheKey::classify_idea_type("This is a very long idea with many words that should definitely be classified as a long idea because it has way more than twenty-five words in total and goes on for quite some time"), "long");
    }

    #[tokio::test]
    async fn test_cache_exact_match() {
        let cache = LlmResponseCache::new();
        let provider = create_test_provider();
        let analysis = create_test_analysis(7.5);
        let idea = "Create a mobile app for fitness tracking";

        // Cache the response
        cache
            .cache_response(idea, &analysis, &provider, "v1.0", "High", 0.9)
            .await
            .unwrap();

        // Retrieve exact match
        let cached = cache.get_response(idea, &provider, "v1.0").await.unwrap();
        assert!(cached.is_some());
        let (retrieved_analysis, similarity) = cached.unwrap();
        assert_eq!(retrieved_analysis.final_score, 7.5);
        assert_eq!(similarity, 1.0); // Exact match should have 1.0 similarity
    }

    #[tokio::test]
    async fn test_cache_semantic_match() {
        let cache = LlmResponseCache::with_config(0.7, DEFAULT_CACHE_TTL);
        let provider = create_test_provider();
        let analysis = create_test_analysis(7.5);
        let idea1 = "Create a mobile app for fitness tracking";
        let idea2 = "Build an application for tracking exercise goals"; // Semantically similar

        // Cache the first idea
        cache
            .cache_response(idea1, &analysis, &provider, "v1.0", "High", 0.9)
            .await
            .unwrap();

        // Try to get semantic match for the second idea
        let cached = cache.get_response(idea2, &provider, "v1.0").await.unwrap();
        assert!(cached.is_some());
        let (retrieved_analysis, similarity) = cached.unwrap();
        assert_eq!(retrieved_analysis.final_score, 7.5);
        assert!(similarity >= 0.7); // Should meet our threshold
        assert!(similarity < 1.0); // But shouldn't be exact match
    }

    #[tokio::test]
    async fn test_cache_miss() {
        let cache = LlmResponseCache::new();
        let provider = create_test_provider();
        let analysis = create_test_analysis(7.5);
        let idea1 = "Create a mobile app";
        let idea2 = "Write a blog about cooking"; // Different domain

        // Cache the first idea
        cache
            .cache_response(idea1, &analysis, &provider, "v1.0", "High", 0.9)
            .await
            .unwrap();

        // Try to get match for unrelated idea
        let cached = cache.get_response(idea2, &provider, "v1.0").await.unwrap();
        assert!(cached.is_none());
    }

    #[tokio::test]
    async fn test_cache_stats() {
        let cache = LlmResponseCache::new();
        let provider = create_test_provider();
        let analysis = create_test_analysis(7.5);
        let idea1 = "Create a mobile app";
        let idea2 = "Build an application"; // Similar
        let idea3 = "Write a blog"; // Different

        // Cache one idea
        cache
            .cache_response(idea1, &analysis, &provider, "v1.0", "High", 0.9)
            .await
            .unwrap();

        // Get exact match (hit)
        cache.get_response(idea1, &provider, "v1.0").await.unwrap();

        // Get semantic match (hit)
        cache.get_response(idea2, &provider, "v1.0").await.unwrap();

        // Get miss
        cache.get_response(idea3, &provider, "v1.0").await.unwrap();

        let stats = cache.get_stats().await;
        assert_eq!(stats.basic_stats.hits, 1); // One exact hit
        assert_eq!(stats.semantic_hits, 1); // One semantic hit
        assert_eq!(stats.semantic_misses, 1); // One semantic miss

        let metrics = cache.get_effectiveness_metrics().await;
        assert!(metrics.overall_hit_rate > 0.0);
        assert!(metrics.average_similarity > 0.0);
    }

    #[tokio::test]
    async fn test_cache_cleanup_expired() {
        let cache = LlmResponseCache::with_config(0.7, Duration::from_millis(100));
        let provider = create_test_provider();
        let analysis = create_test_analysis(7.5);
        let idea = "Create a mobile app";

        // Cache the response
        cache
            .cache_response(idea, &analysis, &provider, "v1.0", "High", 0.9)
            .await
            .unwrap();

        // Wait for expiration
        tokio::time::sleep(Duration::from_millis(150)).await;

        // Try to get expired response (should return None)
        let cached = cache.get_response(idea, &provider, "v1.0").await.unwrap();
        assert!(cached.is_none());

        // Cleanup expired entries
        let removed = cache.cleanup_expired().await.unwrap();
        assert!(removed >= 0);
    }

    #[tokio::test]
    async fn test_cache_eviction() {
        // Create a cache with very small max size
        let cache = LlmResponseCache::new();
        let provider = create_test_provider();

        // Fill the cache with many entries for the same type
        for i in 0..MAX_CACHE_SIZE_PER_TYPE + 10 {
            let idea = format!("Idea number {} about creating apps", i);
            let analysis = create_test_analysis(5.0 + (i as f64 * 0.1));
            cache
                .cache_response(&idea, &analysis, &provider, "v1.0", "Medium", 0.7)
                .await
                .unwrap();
        }

        let stats = cache.get_stats().await;
        assert!(stats.evictions > 0); // Should have evicted some entries
    }

    #[test]
    fn test_cache_entry_score() {
        let mut entry = LlmCacheEntry::new(
            create_test_analysis(7.5),
            "test idea".to_string(),
            create_test_provider(),
            "High".to_string(),
            0.9,
        );

        let initial_score = entry.cache_score();

        // Simulate some hits and age
        entry.record_hit(Some(0.9));
        entry.record_hit(Some(0.8));

        let new_score = entry.cache_score();
        assert!(new_score > initial_score); // More hits should increase score
    }
}
