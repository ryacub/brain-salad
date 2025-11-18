//! Concrete implementations of the trait abstractions

use crate::errors::ValidationError;
use crate::traits::*;
use crate::types::PatternType;
use async_trait::async_trait;

/// Implementation of the Scorer trait for the existing scoring engine
pub struct TelosScorer {
    config: TelosScoringConfig,
}

#[derive(Debug, Clone)]
pub struct TelosScoringConfig {
    pub current_stack: Vec<String>,
    pub domain_keywords: Vec<String>,
    pub weights: (f64, f64, f64), // (mission, anti_challenge, strategic)
}

impl Default for TelosScoringConfig {
    fn default() -> Self {
        Self {
            current_stack: vec![
                "python".to_string(),
                "langchain".to_string(),
                "openai".to_string(),
            ],
            domain_keywords: vec![
                "hotel".to_string(),
                "hospitality".to_string(),
                "mobile".to_string(),
            ],
            weights: (0.4, 0.35, 0.25),
        }
    }
}

#[derive(Debug, Clone)]
pub struct TelosScore {
    pub raw_score: f64,
    pub mission_score: f64,
    pub anti_challenge_score: f64,
    pub strategic_score: f64,
    pub patterns: Vec<String>,
}

#[async_trait]
impl Scorer<String> for TelosScorer {
    type Score = TelosScore;
    type Config = TelosScoringConfig;

    fn new(config: Self::Config) -> Self {
        Self { config }
    }

    async fn score(
        &self,
        idea: &String,
    ) -> std::result::Result<Self::Score, Box<dyn std::error::Error + Send + Sync>> {
        self.validate_input(idea)?;

        let idea_lower = idea.to_lowercase();

        // Simple scoring logic for demonstration
        let mission_score = self.score_mission(&idea_lower);
        let anti_challenge_score = self.score_anti_challenge(&idea_lower);
        let strategic_score = self.score_strategic(&idea_lower);

        let raw_score = mission_score + anti_challenge_score + strategic_score;
        let patterns = self.detect_patterns(&idea_lower);

        Ok(TelosScore {
            raw_score,
            mission_score,
            anti_challenge_score,
            strategic_score,
            patterns,
        })
    }

    fn validate_input(
        &self,
        input: &String,
    ) -> std::result::Result<(), Box<dyn std::error::Error + Send + Sync>> {
        if input.trim().is_empty() {
            return Err(Box::new(ValidationError::empty_field("idea")));
        }
        if input.len() > 10000 {
            return Err(Box::new(ValidationError::too_long(
                "idea",
                input.len(),
                10000,
            )));
        }
        Ok(())
    }
}

impl TelosScorer {
    fn score_mission(&self, idea: &str) -> f64 {
        let mut score = 0.0;

        // Check for AI-related terms
        if idea.contains("ai") || idea.contains("artificial intelligence") {
            score += 1.5;
        }

        // Check for shipping terms
        if idea.contains("ship") || idea.contains("launch") {
            score += 0.8;
        }

        // Check for domain relevance
        if self
            .config
            .domain_keywords
            .iter()
            .any(|kw| idea.contains(kw))
        {
            score += 1.2;
        }

        score
    }

    fn score_anti_challenge(&self, idea: &str) -> f64 {
        let mut score: f64 = 0.0;

        // Check for stack compliance
        if self
            .config
            .current_stack
            .iter()
            .any(|tech| idea.contains(tech))
        {
            score += 1.2;
        }

        // Check for perfectionism indicators (negative)
        if idea.contains("perfect") || idea.contains("comprehensive") {
            score -= 0.5;
        }

        // Check for accountability
        if idea.contains("public") || idea.contains("share") {
            score += 0.8;
        }

        score.max(0.0)
    }

    fn score_strategic(&self, idea: &str) -> f64 {
        let mut score = 0.0;

        // Check for simplicity
        if idea.contains("simple") || idea.contains("mvp") {
            score += 1.0;
        }

        // Check for timeline
        if idea.contains("week") || idea.contains("month") {
            score += 0.8;
        }

        score
    }

    fn detect_patterns(&self, idea: &str) -> Vec<String> {
        let mut patterns = Vec::new();

        if idea.contains("rust") || idea.contains("javascript") {
            patterns.push("Context Switching".to_string());
        }

        if idea.contains("learn") && idea.contains("before") {
            patterns.push("Procrastination".to_string());
        }

        if idea.contains("perfect") {
            patterns.push("Perfectionism".to_string());
        }

        patterns
    }
}

/// Implementation of the PatternDetector trait
pub struct TelosPatternDetector {
    patterns: Vec<TelosPattern>,
}

#[derive(Debug, Clone)]
pub struct TelosPattern {
    pub pattern_type: PatternType,
    pub keywords: Vec<String>,
    pub regex_patterns: Vec<regex::Regex>,
    pub severity: f64,
}

impl TelosPatternDetector {
    pub fn new() -> std::result::Result<Self, regex::Error> {
        let patterns = vec![
            TelosPattern {
                pattern_type: PatternType::context_switching(),
                keywords: vec![
                    "rust".to_string(),
                    "javascript".to_string(),
                    "react".to_string(),
                ],
                regex_patterns: vec![regex::Regex::new(
                    r"\b(new|latest)\s+(framework|library|technology)\b",
                )?],
                severity: 0.8,
            },
            TelosPattern {
                pattern_type: PatternType::procrastination(),
                keywords: vec![
                    "learn before".to_string(),
                    "someday".to_string(),
                    "eventually".to_string(),
                ],
                regex_patterns: vec![
                    regex::Regex::new(r"\blearn\b.*\bbefore\b.*\b(start|build)\b")?,
                    regex::Regex::new(r"\b(need to|have to|should)\s+learn\b")?,
                ],
                severity: 0.9,
            },
            TelosPattern {
                pattern_type: PatternType::perfectionism(),
                keywords: vec![
                    "perfect".to_string(),
                    "comprehensive".to_string(),
                    "complete".to_string(),
                ],
                regex_patterns: vec![regex::Regex::new(
                    r"\b(build.*from scratch|custom.*implementation|reinvent.*wheel)\b",
                )?],
                severity: 0.7,
            },
        ];

        Ok(Self { patterns })
    }
}

#[derive(Debug, Clone)]
pub struct TelosPatternMatch {
    pub pattern_type: PatternType,
    pub severity: f64,
    pub matches: Vec<String>,
    pub message: String,
    pub suggestion: Option<String>,
}

#[async_trait]
impl PatternDetector<String> for TelosPatternDetector {
    type Pattern = TelosPatternMatch;
    type Config = ();

    fn new(_config: Self::Config) -> Self {
        Self::new().expect("Failed to create pattern detector")
    }

    async fn detect_patterns(
        &self,
        input: &String,
    ) -> std::result::Result<Vec<Self::Pattern>, Box<dyn std::error::Error + Send + Sync>> {
        self.validate_input(input)?;

        let idea_lower = input.to_lowercase();
        let mut matches = Vec::new();

        for pattern in &self.patterns {
            let mut pattern_matches = Vec::new();

            // Check keyword matches
            for keyword in &pattern.keywords {
                if idea_lower.contains(keyword) {
                    pattern_matches.push(keyword.clone());
                }
            }

            // Check regex matches
            for regex in &pattern.regex_patterns {
                for mat in regex.find_iter(&idea_lower) {
                    pattern_matches.push(mat.as_str().to_string());
                }
            }

            if !pattern_matches.is_empty() {
                matches.push(TelosPatternMatch {
                    pattern_type: pattern.pattern_type.clone(),
                    severity: pattern.severity,
                    matches: pattern_matches,
                    message: format!("Detected {} pattern", pattern.pattern_type.as_str()),
                    suggestion: self.get_suggestion(&pattern.pattern_type),
                });
            }
        }

        Ok(matches)
    }

    async fn detect_pattern_type(
        &self,
        input: &String,
        pattern_type: &PatternType,
    ) -> std::result::Result<Option<Self::Pattern>, Box<dyn std::error::Error + Send + Sync>> {
        let all_patterns = self.detect_patterns(input).await?;
        Ok(all_patterns
            .into_iter()
            .find(|p| p.pattern_type.as_str() == pattern_type.as_str()))
    }

    fn supported_patterns(&self) -> Vec<PatternType> {
        self.patterns
            .iter()
            .map(|p| p.pattern_type.clone())
            .collect()
    }

    fn validate_input(
        &self,
        input: &String,
    ) -> std::result::Result<(), Box<dyn std::error::Error + Send + Sync>> {
        if input.trim().is_empty() {
            return Err(Box::new(ValidationError::empty_field("input")));
        }
        Ok(())
    }
}

impl TelosPatternDetector {
    fn get_suggestion(&self, pattern_type: &PatternType) -> Option<String> {
        match pattern_type.as_str() {
            "ContextSwitching" => {
                Some("Focus on your current stack: Python + LangChain + OpenAI".to_string())
            }
            "Procrastination" => {
                Some("Build first, learn as needed. Avoid the learning trap.".to_string())
            }
            "Perfectionism" => Some("Start with MVP, iterate based on feedback.".to_string()),
            _ => None,
        }
    }
}

/// Simple in-memory cache implementation
pub struct InMemoryCache<K, V> {
    data: std::sync::RwLock<std::collections::HashMap<K, CacheEntry<V>>>,
    hits: std::sync::atomic::AtomicU64,
    misses: std::sync::atomic::AtomicU64,
}

#[derive(Clone)]
struct CacheEntry<V> {
    value: V,
    created_at: std::time::Instant,
    ttl: Option<std::time::Duration>,
}

impl<K, V> InMemoryCache<K, V>
where
    K: Clone + Eq + std::hash::Hash + Send + Sync + 'static,
    V: Clone + Send + Sync + 'static,
{
    pub fn new() -> Self {
        Self {
            data: std::sync::RwLock::new(std::collections::HashMap::new()),
            hits: std::sync::atomic::AtomicU64::new(0),
            misses: std::sync::atomic::AtomicU64::new(0),
        }
    }
}

impl<K, V> Default for InMemoryCache<K, V>
where
    K: Clone + Eq + std::hash::Hash + Send + Sync + 'static,
    V: Clone + Send + Sync + 'static,
{
    fn default() -> Self {
        Self::new()
    }
}

#[async_trait]
impl<K, V> Cache<K, V> for InMemoryCache<K, V>
where
    K: Clone + Eq + std::hash::Hash + Send + Sync + 'static,
    V: Clone + Send + Sync + 'static,
{
    type Error = CacheError;

    async fn get(&self, key: &K) -> std::result::Result<Option<V>, Self::Error> {
        let data = self.data.read().map_err(|_| CacheError::LockError)?;

        if let Some(entry) = data.get(key) {
            // Check TTL
            if let Some(ttl) = entry.ttl {
                if entry.created_at.elapsed() > ttl {
                    return Ok(None);
                }
            }
            self.hits.fetch_add(1, std::sync::atomic::Ordering::Relaxed);
            Ok(Some(entry.value.clone()))
        } else {
            self.misses
                .fetch_add(1, std::sync::atomic::Ordering::Relaxed);
            Ok(None)
        }
    }

    async fn set(
        &self,
        key: K,
        value: V,
        ttl: Option<std::time::Duration>,
    ) -> std::result::Result<(), Self::Error> {
        let mut data = self.data.write().map_err(|_| CacheError::LockError)?;

        data.insert(
            key,
            CacheEntry {
                value,
                created_at: std::time::Instant::now(),
                ttl,
            },
        );

        Ok(())
    }

    async fn remove(&self, key: &K) -> std::result::Result<bool, Self::Error> {
        let mut data = self.data.write().map_err(|_| CacheError::LockError)?;
        Ok(data.remove(key).is_some())
    }

    async fn clear(&self) -> std::result::Result<(), Self::Error> {
        let mut data = self.data.write().map_err(|_| CacheError::LockError)?;
        data.clear();
        Ok(())
    }

    async fn stats(&self) -> CacheStats {
        let hits = self.hits.load(std::sync::atomic::Ordering::Relaxed);
        let misses = self.misses.load(std::sync::atomic::Ordering::Relaxed);
        let data = self.data.read().map_err(|_| CacheError::LockError).unwrap();
        let size = data.len();

        CacheStats::new(hits, misses, size)
    }
}

#[derive(Debug, thiserror::Error)]
pub enum CacheError {
    #[error("Failed to acquire cache lock")]
    LockError,
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::time::Duration;

    #[tokio::test]
    async fn test_telos_scorer() {
        let scorer = TelosScorer::new(TelosScoringConfig::default());

        let result = scorer
            .score(&"Build an AI hotel booking system using Python".to_string())
            .await
            .unwrap();
        assert!(result.raw_score > 0.0);
        assert!(result.mission_score > 0.0);
    }

    #[tokio::test]
    async fn test_pattern_detector() {
        let detector = TelosPatternDetector::new().unwrap();

        let patterns = detector
            .detect_patterns(&"I want to learn Rust before building my AI project".to_string())
            .await
            .unwrap();
        assert!(!patterns.is_empty());

        let procrastination_pattern = detector
            .detect_pattern_type(
                &"I want to learn Rust before building my AI project".to_string(),
                &PatternType::procrastination(),
            )
            .await
            .unwrap();

        assert!(procrastination_pattern.is_some());
    }

    #[tokio::test]
    async fn test_in_memory_cache() {
        let cache = InMemoryCache::<String, String>::new();

        // Test set and get
        cache
            .set("key1".to_string(), "value1".to_string(), None)
            .await
            .unwrap();
        let value = cache.get(&"key1".to_string()).await.unwrap();
        assert_eq!(value, Some("value1".to_string()));

        // Test TTL
        cache
            .set(
                "key2".to_string(),
                "value2".to_string(),
                Some(Duration::from_millis(10)),
            )
            .await
            .unwrap();
        tokio::time::sleep(Duration::from_millis(20)).await;
        let value = cache.get(&"key2".to_string()).await.unwrap();
        assert_eq!(value, None);

        // Test stats
        let stats = cache.stats().await;
        assert_eq!(stats.hits, 1);
        assert_eq!(stats.misses, 1);
        assert_eq!(stats.size, 1);
    }

    #[test]
    fn test_telos_pattern_detector_creation() {
        let detector = TelosPatternDetector::new().unwrap();
        let patterns = detector.supported_patterns();
        assert_eq!(patterns.len(), 3);

        let pattern_types: Vec<_> = patterns.iter().map(|p| p.as_str()).collect();
        assert!(pattern_types.contains(&"ContextSwitching"));
        assert!(pattern_types.contains(&"Procrastination"));
        assert!(pattern_types.contains(&"Perfectionism"));
    }
}
