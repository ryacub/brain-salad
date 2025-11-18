//! # Telos Idea Matrix Library
//!
//! This library provides core functionality for the Telos Idea Matrix application.
//! It includes configuration loading, idea scoring, database operations, and more.
//!
//! The library exposes the main components for external use while keeping
//! the CLI interface in the binary target.

pub mod config;
pub mod errors;
pub mod scoring;
pub mod telos;

// Re-export important items for external use
pub use config::ConfigPaths;
pub use errors::Result;
pub use scoring::{
    AntiChallengeScores, MissionScores, Recommendation, Score, ScoringEngine, StrategicScores,
    TelosConfig,
};
pub use telos::TelosParser;
