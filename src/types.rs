//! Strongly-typed wrappers for common values to improve type safety

use serde::{Deserialize, Serialize};
use std::borrow::Borrow;
use std::cmp::Ordering;
use std::convert::AsRef;
use std::fmt;
use std::hash::Hash;
use std::str::FromStr;

/// Newtype wrapper for idea IDs to prevent mixing with other strings
#[derive(Debug, Clone, PartialEq, Eq, Hash, Serialize, Deserialize)]
pub struct IdeaId(String);

impl IdeaId {
    /// Create a new IdeaId with the given value
    pub fn new(id: impl Into<String>) -> Self {
        Self(id.into())
    }

    /// Get the underlying string value
    pub fn as_str(&self) -> &str {
        &self.0
    }

    /// Consume and return the inner String
    pub fn into_inner(self) -> String {
        self.0
    }

    /// Generate a new random IdeaId using UUID v4
    pub fn generate() -> Self {
        Self(uuid::Uuid::new_v4().to_string())
    }

    /// Validate that the string looks like a valid idea ID
    pub fn is_valid(&self) -> bool {
        // Basic validation - check if it's a valid UUID or has reasonable length
        uuid::Uuid::parse_str(&self.0).is_ok() || self.0.len() >= 8
    }
}

impl fmt::Display for IdeaId {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "{}", self.0)
    }
}

impl AsRef<str> for IdeaId {
    fn as_ref(&self) -> &str {
        &self.0
    }
}

impl Borrow<str> for IdeaId {
    fn borrow(&self) -> &str {
        &self.0
    }
}

impl From<String> for IdeaId {
    fn from(s: String) -> Self {
        Self(s)
    }
}

impl From<&str> for IdeaId {
    fn from(s: &str) -> Self {
        Self(s.to_string())
    }
}

impl FromStr for IdeaId {
    type Err = crate::errors::ValidationError;

    fn from_str(s: &str) -> Result<Self, Self::Err> {
        let id = Self::new(s);
        if id.is_valid() {
            Ok(id)
        } else {
            Err(crate::errors::ValidationError::invalid_format("idea_id", s))
        }
    }
}

/// Newtype wrapper for score values with validation
#[derive(Debug, Clone, Copy, PartialEq, Serialize, Deserialize)]
pub struct Score(f64);

impl Score {
    /// Minimum allowed score value
    pub const MIN: f64 = 0.0;

    /// Maximum allowed score value
    pub const MAX: f64 = 10.0;

    /// Create a new Score with validation
    pub fn new(value: f64) -> Result<Self, ScoreError> {
        if !(Self::MIN..=Self::MAX).contains(&value) {
            Err(ScoreError::InvalidRange {
                value,
                min: Self::MIN,
                max: Self::MAX,
            })
        } else {
            Ok(Self(value))
        }
    }

    /// Create a new Score without validation (for internal use)
    pub fn new_unchecked(value: f64) -> Self {
        Self(value)
    }

    /// Get the underlying f64 value
    pub fn value(self) -> f64 {
        self.0
    }

    /// Get the underlying f64 value as reference
    pub fn as_f64(&self) -> f64 {
        self.0
    }

    /// Check if the score is in the priority range (>= 8.0)
    pub fn is_priority(&self) -> bool {
        self.0 >= 8.0
    }

    /// Check if the score is in the good range (>= 6.0)
    pub fn is_good(&self) -> bool {
        self.0 >= 6.0
    }

    /// Check if the score is in the avoid range (< 4.0)
    pub fn is_avoid(&self) -> bool {
        self.0 < 4.0
    }

    /// Get the recommendation based on score value
    pub fn recommendation(self) -> Recommendation {
        if self.is_priority() {
            Recommendation::Priority
        } else if self.is_good() {
            Recommendation::Good
        } else if self.is_avoid() {
            Recommendation::Avoid
        } else {
            Recommendation::Consider
        }
    }

    /// Clamp the score to valid range
    pub fn clamped(value: f64) -> Self {
        Self(value.clamp(Self::MIN, Self::MAX))
    }
}

impl fmt::Display for Score {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "{:.1}", self.0)
    }
}

impl From<Score> for f64 {
    fn from(score: Score) -> Self {
        score.0
    }
}

impl From<f64> for Score {
    fn from(value: f64) -> Self {
        Self::clamped(value)
    }
}

impl PartialOrd for Score {
    fn partial_cmp(&self, other: &Self) -> Option<Ordering> {
        self.0.partial_cmp(&other.0)
    }
}

/// Score-related errors
#[derive(Debug, Clone, thiserror::Error)]
pub enum ScoreError {
    #[error("Score {value} is out of valid range [{min}, {max}]")]
    InvalidRange { value: f64, min: f64, max: f64 },
}

/// Recommendation levels based on score
#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash, Serialize, Deserialize)]
pub enum Recommendation {
    Priority,
    Good,
    Consider,
    Avoid,
}

impl Recommendation {
    pub fn emoji(&self) -> &'static str {
        match self {
            Recommendation::Priority => "🔥",
            Recommendation::Good => "✅",
            Recommendation::Consider => "⚠️",
            Recommendation::Avoid => "🚫",
        }
    }

    pub fn text(&self) -> &'static str {
        match self {
            Recommendation::Priority => "PRIORITIZE NOW",
            Recommendation::Good => "GOOD ALIGNMENT",
            Recommendation::Consider => "CONSIDER LATER",
            Recommendation::Avoid => "AVOID FOR NOW",
        }
    }
}

impl fmt::Display for Recommendation {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "{} {}", self.emoji(), self.text())
    }
}

/// Newtype wrapper for pattern types with stronger typing
#[derive(Debug, Clone, PartialEq, Eq, Hash, Serialize, Deserialize)]
pub struct PatternType(String);

impl PatternType {
    /// Create a new PatternType
    pub fn new(pattern_type: impl Into<String>) -> Self {
        Self(pattern_type.into())
    }

    /// Get the underlying string value
    pub fn as_str(&self) -> &str {
        &self.0
    }

    /// Known pattern types
    pub const CONTEXT_SWITCHING: &'static str = "ContextSwitching";
    pub const PERFECTIONISM: &'static str = "Perfectionism";
    pub const PROCRASTINATION: &'static str = "Procrastination";
    pub const ACCOUNTABILITY_AVOIDANCE: &'static str = "AccountabilityAvoidance";
    pub const SCOPE_CREEP: &'static str = "ScopeCreep";

    /// Create a context switching pattern type
    pub fn context_switching() -> Self {
        Self(Self::CONTEXT_SWITCHING.to_string())
    }

    /// Create a perfectionism pattern type
    pub fn perfectionism() -> Self {
        Self(Self::PERFECTIONISM.to_string())
    }

    /// Create a procrastination pattern type
    pub fn procrastination() -> Self {
        Self(Self::PROCRASTINATION.to_string())
    }

    /// Create an accountability avoidance pattern type
    pub fn accountability_avoidance() -> Self {
        Self(Self::ACCOUNTABILITY_AVOIDANCE.to_string())
    }

    /// Create a scope creep pattern type
    pub fn scope_creep() -> Self {
        Self(Self::SCOPE_CREEP.to_string())
    }

    /// Check if this is a known pattern type
    pub fn is_known(&self) -> bool {
        matches!(
            self.as_str(),
            Self::CONTEXT_SWITCHING
                | Self::PERFECTIONISM
                | Self::PROCRASTINATION
                | Self::ACCOUNTABILITY_AVOIDANCE
                | Self::SCOPE_CREEP
        )
    }

    /// Get emoji for this pattern type
    pub fn emoji(&self) -> &'static str {
        match self.as_str() {
            Self::CONTEXT_SWITCHING => "🔄",
            Self::PERFECTIONISM => "⚡",
            Self::PROCRASTINATION => "🕰️",
            Self::ACCOUNTABILITY_AVOIDANCE => "👤",
            Self::SCOPE_CREEP => "📏",
            _ => "❓",
        }
    }
}

impl fmt::Display for PatternType {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "{} {}", self.emoji(), self.0)
    }
}

impl AsRef<str> for PatternType {
    fn as_ref(&self) -> &str {
        &self.0
    }
}

impl From<String> for PatternType {
    fn from(s: String) -> Self {
        Self(s)
    }
}

impl From<&str> for PatternType {
    fn from(s: &str) -> Self {
        Self(s.to_string())
    }
}

/// Newtype wrapper for database query limits to prevent invalid values
#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash, Serialize, Deserialize)]
pub struct QueryLimit(usize);

impl QueryLimit {
    /// Default query limit
    pub const DEFAULT: usize = 50;

    /// Maximum allowed query limit
    pub const MAX: usize = 1000;

    /// Create a new QueryLimit with validation
    pub fn new(limit: usize) -> Result<Self, QueryLimitError> {
        if limit == 0 {
            Err(QueryLimitError::Zero)
        } else if limit > Self::MAX {
            Err(QueryLimitError::TooLarge {
                limit,
                max: Self::MAX,
            })
        } else {
            Ok(Self(limit))
        }
    }

    /// Create a new QueryLimit without validation
    pub fn new_unchecked(limit: usize) -> Self {
        Self(limit)
    }

    /// Get the underlying usize value
    pub fn value(self) -> usize {
        self.0
    }

    /// Get the underlying usize value as reference
    pub fn as_usize(&self) -> usize {
        self.0
    }

    /// Get the default query limit
    pub fn default() -> Self {
        Self(Self::DEFAULT)
    }
}

impl fmt::Display for QueryLimit {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "{}", self.0)
    }
}

impl From<QueryLimit> for usize {
    fn from(limit: QueryLimit) -> Self {
        limit.0
    }
}

impl From<usize> for QueryLimit {
    fn from(value: usize) -> Self {
        Self(value.min(Self::MAX).max(1))
    }
}

/// Query limit related errors
#[derive(Debug, Clone, thiserror::Error)]
pub enum QueryLimitError {
    #[error("Query limit cannot be zero")]
    Zero,

    #[error("Query limit {limit} exceeds maximum allowed value of {max}")]
    TooLarge { limit: usize, max: usize },
}

/// Newtype wrapper for file paths to ensure they're valid
#[derive(Debug, Clone, PartialEq, Eq, Hash, Serialize, Deserialize)]
pub struct FilePath(std::path::PathBuf);

impl FilePath {
    /// Create a new FilePath with validation
    pub fn new<P: AsRef<std::path::Path>>(path: P) -> Result<Self, FilePathError> {
        let path_buf = path.as_ref().to_path_buf();

        // Basic validation
        if path_buf.as_os_str().is_empty() {
            return Err(FilePathError::Empty);
        }

        // Check for invalid characters on Windows
        #[cfg(windows)]
        {
            let path_str = path_buf.to_string_lossy();
            if path_str
                .chars()
                .any(|c| matches!(c, '<' | '>' | ':' | '"' | '|' | '?' | '*'))
            {
                return Err(FilePathError::InvalidCharacters);
            }
        }

        Ok(Self(path_buf))
    }

    /// Get the underlying PathBuf
    pub fn as_path(&self) -> &std::path::Path {
        &self.0
    }

    /// Get the underlying PathBuf as reference
    pub fn as_path_buf(&self) -> &std::path::PathBuf {
        &self.0
    }

    /// Consume and return the inner PathBuf
    pub fn into_inner(self) -> std::path::PathBuf {
        self.0
    }

    /// Check if the file exists
    pub async fn exists(&self) -> bool {
        tokio::fs::metadata(&self.0).await.is_ok()
    }

    /// Get the parent directory
    pub fn parent(&self) -> Option<FilePath> {
        self.0.parent().map(|p| FilePath(p.to_path_buf()))
    }

    /// Get the file name
    pub fn file_name(&self) -> Option<&std::ffi::OsStr> {
        self.0.file_name()
    }

    /// Get the file extension
    pub fn extension(&self) -> Option<&std::ffi::OsStr> {
        self.0.extension()
    }
}

impl fmt::Display for FilePath {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "{}", self.0.display())
    }
}

impl AsRef<std::path::Path> for FilePath {
    fn as_ref(&self) -> &std::path::Path {
        &self.0
    }
}

impl TryFrom<std::path::PathBuf> for FilePath {
    type Error = FilePathError;

    fn try_from(path: std::path::PathBuf) -> Result<Self, Self::Error> {
        Self::new(path)
    }
}

impl TryFrom<&std::path::Path> for FilePath {
    type Error = FilePathError;

    fn try_from(path: &std::path::Path) -> Result<Self, Self::Error> {
        Self::new(path)
    }
}

/// File path related errors
#[derive(Debug, Clone, thiserror::Error)]
pub enum FilePathError {
    #[error("File path cannot be empty")]
    Empty,

    #[error("File path contains invalid characters")]
    InvalidCharacters,

    #[error("File path is not UTF-8 compatible")]
    InvalidUtf8,
}

/// Structure representing an analysis result
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct AnalysisResult {
    pub idea: String,
    pub score: f64,
    pub recommendation: String,
    pub timestamp: chrono::DateTime<chrono::Utc>,
    pub analysis_method: String, // "rule-based", "llm", etc.
    pub raw_analysis: String,    // Raw analysis output in JSON format
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_idea_id_creation() {
        let id = IdeaId::new("test-id");
        assert_eq!(id.as_str(), "test-id");
        assert!(id.is_valid());
    }

    #[test]
    fn test_idea_id_generation() {
        let id1 = IdeaId::generate();
        let id2 = IdeaId::generate();
        assert_ne!(id1, id2);
        assert!(id1.is_valid());
        assert!(id2.is_valid());
    }

    #[test]
    fn test_score_validation() {
        assert!(Score::new(5.0).is_ok());
        assert!(Score::new(0.0).is_ok());
        assert!(Score::new(10.0).is_ok());
        assert!(Score::new(-1.0).is_err());
        assert!(Score::new(11.0).is_err());
    }

    #[test]
    fn test_score_methods() {
        let priority = Score::new(8.5).unwrap();
        assert!(priority.is_priority());
        assert_eq!(priority.recommendation(), Recommendation::Priority);

        let avoid = Score::new(2.0).unwrap();
        assert!(avoid.is_avoid());
        assert_eq!(avoid.recommendation(), Recommendation::Avoid);
    }

    #[test]
    fn test_pattern_type() {
        let pattern = PatternType::context_switching();
        assert_eq!(pattern.as_str(), PatternType::CONTEXT_SWITCHING);
        assert!(pattern.is_known());
    }

    #[test]
    fn test_query_limit() {
        assert!(QueryLimit::new(50).is_ok());
        assert!(QueryLimit::new(0).is_err());
        assert!(QueryLimit::new(2000).is_err());
    }

    #[test]
    fn test_file_path() {
        let path = FilePath::new("/tmp/test.txt").unwrap();
        assert_eq!(path.file_name(), Some(std::ffi::OsStr::new("test.txt")));
        assert_eq!(path.extension(), Some(std::ffi::OsStr::new("txt")));
    }
}
