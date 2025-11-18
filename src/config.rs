use anyhow::{Context, Result};
use serde::{Deserialize, Serialize};
use std::path::PathBuf;

/// Configuration paths for the Telos Idea Matrix system.
///
/// This struct manages the location of key configuration files and directories:
/// - `telos_file`: Path to the user's telos.md file (YAML format)
/// - `data_dir`: Directory for database and persistent data
/// - `log_dir`: Directory for log files
/// - `config_file`: Optional path to the config.toml file if one was used
///
/// # Loading Priority
///
/// Configuration is loaded in the following priority order:
/// 1. Environment variable `TELOS_FILE` (highest priority)
/// 2. `./telos.md` in current working directory
/// 3. `~/.config/telos-matrix/config.toml`
/// 4. Interactive wizard (if nothing else found)
///
/// # Example
///
/// ```no_run
/// use telos_idea_matrix::config::ConfigPaths;
///
/// let config = ConfigPaths::load()?;
/// config.ensure_directories_exist()?;
/// println!("Telos file: {:?}", config.telos_file);
/// ```
#[derive(Debug, Clone, Serialize, Deserialize, PartialEq, Eq)]
pub struct ConfigPaths {
    /// Path to the telos.md file (YAML format)
    pub telos_file: PathBuf,

    /// Directory for storing database and persistent data
    pub data_dir: PathBuf,

    /// Directory for storing log files
    pub log_dir: PathBuf,

    /// Path to the config.toml file, if one was used
    #[serde(skip_serializing_if = "Option::is_none")]
    pub config_file: Option<PathBuf>,
}

impl ConfigPaths {
    /// Load configuration from the first available source.
    ///
    /// Tries sources in priority order:
    /// 1. TELOS_FILE environment variable
    /// 2. ./telos.md in current directory
    /// 3. ~/.config/telos-matrix/config.toml
    /// 4. Interactive wizard (prompts user)
    ///
    /// # Errors
    ///
    /// Returns an error if:
    /// - The wizard is required but user input is invalid
    /// - File I/O operations fail
    /// - Configuration file is malformed
    ///
    /// # Example
    ///
    /// ```no_run
    /// let config = ConfigPaths::load()?;
    /// println!("Using telos file: {:?}", config.telos_file);
    /// ```
    pub fn load() -> Result<Self> {
        // Try in priority order
        if let Some(config) = Self::from_env_var() {
            tracing::info!("Loaded config from TELOS_FILE environment variable");
            return Ok(config);
        }

        if let Some(config) = Self::from_current_dir() {
            tracing::info!("Loaded config from current directory (./telos.md)");
            return Ok(config);
        }

        if let Some(config) = Self::from_config_file() {
            tracing::info!("Loaded config from ~/.config/telos-matrix/config.toml");
            return Ok(config);
        }

        // If nothing found, launch wizard
        tracing::info!("No configuration found, launching interactive wizard");

        // Create a minimal tokio runtime if we're not already in one
        let config = if tokio::runtime::Handle::try_current().is_ok() {
            // We're already in a tokio runtime
            tokio::task::block_in_place(|| {
                tokio::runtime::Handle::current().block_on(Self::launch_wizard())
            })
        } else {
            // Create a new runtime just for the wizard
            tokio::runtime::Runtime::new()
                .context("Failed to create tokio runtime for wizard")?
                .block_on(Self::launch_wizard())
        }?;

        Ok(config)
    }

    /// Load configuration from the TELOS_FILE environment variable.
    ///
    /// This is the highest priority source and is ideal for:
    /// - Docker/container deployments
    /// - CI/CD pipelines
    /// - Explicit user overrides
    ///
    /// # Returns
    ///
    /// `Some(ConfigPaths)` if TELOS_FILE is set and points to an existing file,
    /// `None` otherwise.
    fn from_env_var() -> Option<ConfigPaths> {
        let telos_path = std::env::var("TELOS_FILE").ok()?;
        let path = PathBuf::from(&telos_path);

        // Validate file exists
        if !path.exists() {
            eprintln!(
                "Warning: TELOS_FILE env var points to non-existent file: {}",
                telos_path
            );
            return None;
        }

        Some(ConfigPaths {
            telos_file: path,
            data_dir: Self::default_data_dir(),
            log_dir: Self::default_log_dir(),
            config_file: None,
        })
    }

    /// Load configuration by checking for ./telos.md in the current directory.
    ///
    /// This source is ideal for:
    /// - Project-specific telos files
    /// - CLI usage: `cd my-project && tm dump "idea"`
    /// - Quick testing and development
    ///
    /// # Returns
    ///
    /// `Some(ConfigPaths)` if ./telos.md exists in the current directory,
    /// `None` otherwise.
    fn from_current_dir() -> Option<ConfigPaths> {
        let telos_path = std::env::current_dir().ok()?.join("telos.md");

        if telos_path.exists() {
            Some(ConfigPaths {
                telos_file: telos_path,
                data_dir: Self::default_data_dir(),
                log_dir: Self::default_log_dir(),
                config_file: None,
            })
        } else {
            None
        }
    }

    /// Load configuration from ~/.config/telos-matrix/config.toml.
    ///
    /// This source is ideal for:
    /// - Persistent user configuration
    /// - System-wide settings
    /// - Multi-project setups
    ///
    /// The config.toml file should be in TOML format:
    /// ```toml
    /// telos_file = "/path/to/telos.md"
    /// data_dir = "/path/to/data"
    /// log_dir = "/path/to/logs"
    /// ```
    ///
    /// # Returns
    ///
    /// `Some(ConfigPaths)` if config.toml exists and is valid,
    /// `None` otherwise.
    fn from_config_file() -> Option<ConfigPaths> {
        let config_dir = dirs::config_dir()?;
        let config_path = config_dir.join("telos-matrix/config.toml");

        if !config_path.exists() {
            return None;
        }

        let content = std::fs::read_to_string(&config_path).ok()?;
        let mut config: ConfigPaths = toml::from_str(&content).ok()?;

        // Validate telos file still exists
        if !config.telos_file.exists() {
            eprintln!(
                "Warning: telos.md path in config.toml doesn't exist: {:?}",
                config.telos_file
            );
            return None;
        }

        // Store the config file path
        config.config_file = Some(config_path);

        Some(config)
    }

    /// Launch an interactive wizard to set up configuration.
    ///
    /// This is triggered when no other configuration source is found.
    /// The wizard:
    /// 1. Prompts user for the path to their telos.md file
    /// 2. Validates the file exists
    /// 3. Asks if they want to save the configuration
    /// 4. Optionally creates ~/.config/telos-matrix/config.toml
    ///
    /// # Errors
    ///
    /// Returns an error if:
    /// - User provides invalid input
    /// - The specified file doesn't exist
    /// - Config file creation fails
    async fn launch_wizard() -> Result<ConfigPaths> {
        use dialoguer::{Confirm, Input};

        println!("\n🔧 First-time setup: Let's configure Telos Idea Matrix");
        println!("─────────────────────────────────────────────────────────\n");

        let telos_path_str: String = Input::new()
            .with_prompt("Path to your telos.md file")
            .interact_text()
            .context("Failed to read user input")?;

        let telos_path = PathBuf::from(&telos_path_str);

        // Validate file exists
        if !telos_path.exists() {
            return Err(anyhow::anyhow!(
                "File not found: {}\n\nPlease check the path and try again.",
                telos_path_str
            ));
        }

        let config = ConfigPaths {
            telos_file: telos_path,
            data_dir: Self::default_data_dir(),
            log_dir: Self::default_log_dir(),
            config_file: None,
        };

        // Ask if user wants to save configuration
        let should_save = Confirm::new()
            .with_prompt("Save this configuration for future runs?")
            .default(true)
            .interact()
            .context("Failed to read confirmation")?;

        if should_save {
            config
                .save_to_config_file()
                .context("Failed to save configuration")?;
            println!("✓ Configuration saved to ~/.config/telos-matrix/config.toml\n");
        }

        Ok(config)
    }

    /// Save this configuration to ~/.config/telos-matrix/config.toml.
    ///
    /// Creates the config directory if it doesn't exist.
    ///
    /// # Errors
    ///
    /// Returns an error if:
    /// - Unable to determine config directory
    /// - Directory creation fails
    /// - File write fails
    /// - TOML serialization fails
    pub fn save_to_config_file(&self) -> Result<()> {
        let config_dir = dirs::config_dir()
            .ok_or_else(|| anyhow::anyhow!("Unable to determine config directory"))?;

        let telos_config_dir = config_dir.join("telos-matrix");
        std::fs::create_dir_all(&telos_config_dir).context("Failed to create config directory")?;

        let config_path = telos_config_dir.join("config.toml");
        let content = toml::to_string_pretty(self).context("Failed to serialize config")?;

        std::fs::write(&config_path, content)
            .context(format!("Failed to write config to {:?}", config_path))?;

        Ok(())
    }

    /// Ensure that data and log directories exist.
    ///
    /// Creates directories if they don't exist, including parent directories.
    ///
    /// # Errors
    ///
    /// Returns an error if directory creation fails.
    ///
    /// # Example
    ///
    /// ```no_run
    /// let config = ConfigPaths::load()?;
    /// config.ensure_directories_exist()?;
    /// // Now data_dir and log_dir are guaranteed to exist
    /// ```
    pub fn ensure_directories_exist(&self) -> Result<()> {
        std::fs::create_dir_all(&self.data_dir).context(format!(
            "Failed to create data directory: {:?}",
            self.data_dir
        ))?;

        std::fs::create_dir_all(&self.log_dir).context(format!(
            "Failed to create log directory: {:?}",
            self.log_dir
        ))?;

        Ok(())
    }

    /// Get the default data directory.
    ///
    /// Uses the platform-specific data directory:
    /// - Linux: `~/.local/share/telos-matrix`
    /// - macOS: `~/Library/Application Support/telos-matrix`
    /// - Windows: `C:\Users\<user>\AppData\Roaming\telos-matrix`
    ///
    /// Falls back to `./telos-matrix` if platform directory cannot be determined.
    pub fn default_data_dir() -> PathBuf {
        dirs::data_dir()
            .unwrap_or_else(|| PathBuf::from("."))
            .join("telos-matrix")
    }

    /// Get the default log directory.
    ///
    /// Uses the platform-specific cache directory:
    /// - Linux: `~/.cache/telos-matrix/logs`
    /// - macOS: `~/Library/Caches/telos-matrix/logs`
    /// - Windows: `C:\Users\<user>\AppData\Local\telos-matrix\logs`
    ///
    /// Falls back to `./telos-matrix/logs` if platform directory cannot be determined.
    pub fn default_log_dir() -> PathBuf {
        dirs::cache_dir()
            .unwrap_or_else(|| PathBuf::from("."))
            .join("telos-matrix/logs")
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_struct_creation() {
        let config = ConfigPaths {
            telos_file: PathBuf::from("test.md"),
            data_dir: PathBuf::from("."),
            log_dir: PathBuf::from("."),
            config_file: None,
        };
        assert_eq!(config.telos_file, PathBuf::from("test.md"));
        assert_eq!(config.data_dir, PathBuf::from("."));
        assert_eq!(config.log_dir, PathBuf::from("."));
        assert_eq!(config.config_file, None);
    }

    #[test]
    fn test_struct_creation_with_config_file() {
        let config = ConfigPaths {
            telos_file: PathBuf::from("test.md"),
            data_dir: PathBuf::from("./data"),
            log_dir: PathBuf::from("./logs"),
            config_file: Some(PathBuf::from("config.toml")),
        };
        assert_eq!(config.telos_file, PathBuf::from("test.md"));
        assert_eq!(config.config_file, Some(PathBuf::from("config.toml")));
    }

    #[test]
    fn test_default_paths_valid() {
        let data_dir = ConfigPaths::default_data_dir();
        let log_dir = ConfigPaths::default_log_dir();

        // Should not be empty
        assert!(data_dir.to_string_lossy().len() > 0);
        assert!(log_dir.to_string_lossy().len() > 0);

        // Should contain "telos-matrix"
        assert!(data_dir.to_string_lossy().contains("telos-matrix"));
        assert!(log_dir.to_string_lossy().contains("telos-matrix"));
    }

    #[test]
    fn test_default_paths_structure() {
        let data_dir = ConfigPaths::default_data_dir();
        let log_dir = ConfigPaths::default_log_dir();

        // Data dir should end with "telos-matrix"
        assert_eq!(data_dir.file_name().unwrap(), "telos-matrix");

        // Log dir should end with "logs"
        assert_eq!(log_dir.file_name().unwrap(), "logs");
    }

    #[test]
    fn test_serialization() {
        let config = ConfigPaths {
            telos_file: PathBuf::from("/path/to/telos.md"),
            data_dir: PathBuf::from("/path/to/data"),
            log_dir: PathBuf::from("/path/to/logs"),
            config_file: None,
        };

        let toml_str = toml::to_string(&config).unwrap();
        assert!(toml_str.contains("telos_file"));
        assert!(toml_str.contains("data_dir"));
        assert!(toml_str.contains("log_dir"));
        assert!(!toml_str.contains("config_file")); // Should be skipped when None
    }

    #[test]
    fn test_deserialization() {
        let toml_str = r#"
            telos_file = "/path/to/telos.md"
            data_dir = "/path/to/data"
            log_dir = "/path/to/logs"
        "#;

        let config: ConfigPaths = toml::from_str(toml_str).unwrap();
        assert_eq!(config.telos_file, PathBuf::from("/path/to/telos.md"));
        assert_eq!(config.data_dir, PathBuf::from("/path/to/data"));
        assert_eq!(config.log_dir, PathBuf::from("/path/to/logs"));
        assert_eq!(config.config_file, None);
    }

    #[test]
    fn test_ensure_directories_exist() {
        // Use tempfile for isolated test
        let temp_dir = tempfile::tempdir().unwrap();
        let config = ConfigPaths {
            telos_file: PathBuf::from("telos.md"),
            data_dir: temp_dir.path().join("data"),
            log_dir: temp_dir.path().join("logs"),
            config_file: None,
        };

        // Directories should not exist yet
        assert!(!config.data_dir.exists());
        assert!(!config.log_dir.exists());

        // Create them
        config.ensure_directories_exist().unwrap();

        // Now they should exist
        assert!(config.data_dir.exists());
        assert!(config.log_dir.exists());
    }

    #[test]
    fn test_from_env_var_with_nonexistent_file() {
        // Set env var to nonexistent file
        std::env::set_var("TELOS_FILE", "/nonexistent/path/telos.md");

        // Should return None with warning
        let result = ConfigPaths::from_env_var();
        assert!(result.is_none());

        // Cleanup
        std::env::remove_var("TELOS_FILE");
    }

    #[test]
    fn test_from_current_dir_when_file_missing() {
        // Save original directory
        let original_dir = std::env::current_dir().unwrap();

        // Create temp dir without telos.md
        let temp_dir = tempfile::tempdir().unwrap();
        std::env::set_current_dir(&temp_dir).unwrap();

        // Should return None
        let result = ConfigPaths::from_current_dir();
        assert!(result.is_none());

        // Restore
        std::env::set_current_dir(original_dir).unwrap();
    }

    #[test]
    fn test_config_file_equality() {
        let config1 = ConfigPaths {
            telos_file: PathBuf::from("test.md"),
            data_dir: PathBuf::from("./data"),
            log_dir: PathBuf::from("./logs"),
            config_file: None,
        };

        let config2 = ConfigPaths {
            telos_file: PathBuf::from("test.md"),
            data_dir: PathBuf::from("./data"),
            log_dir: PathBuf::from("./logs"),
            config_file: None,
        };

        assert_eq!(config1, config2);
    }
}
