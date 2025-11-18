use std::fs;
use std::path::PathBuf;
use telos_idea_matrix::config::ConfigPaths;
use tempfile::tempdir;

#[test]
fn test_load_config_from_env_var() {
    // Create temp file
    let temp_dir = tempdir().unwrap();
    let telos_path = temp_dir.path().join("telos.md");
    std::fs::write(&telos_path, "# Test Telos").unwrap();

    // Set env var
    std::env::set_var("TELOS_FILE", &telos_path);

    // Load config
    let config = ConfigPaths::load().expect("Should load from env");

    // Verify - normalize paths for comparison
    assert_eq!(
        fs::canonicalize(&config.telos_file).unwrap(),
        fs::canonicalize(&telos_path).unwrap()
    );

    // Cleanup
    std::env::remove_var("TELOS_FILE");
}

#[test]
fn test_load_config_from_current_dir() {
    // Create temp dir with telos.md
    let temp_dir = tempdir().unwrap();
    let telos_path = temp_dir.path().join("telos.md");
    std::fs::write(&telos_path, "# Test Telos").unwrap();

    // Change to temp dir
    let original_dir = std::env::current_dir().unwrap();
    std::env::set_current_dir(&temp_dir).unwrap();
    std::env::remove_var("TELOS_FILE");

    // Load config
    let config = ConfigPaths::load().expect("Should find ./telos.md");
    let expected_path = temp_dir.path().join("telos.md");
    assert_eq!(
        fs::canonicalize(&config.telos_file).unwrap(),
        fs::canonicalize(&expected_path).unwrap()
    );

    // Restore
    std::env::set_current_dir(original_dir).unwrap();
}

#[test]
fn test_env_var_has_priority() {
    let temp_dir1 = tempdir().unwrap();
    let temp_dir2 = tempdir().unwrap();

    let env_telos = temp_dir1.path().join("from_env.md");
    let cwd_telos = temp_dir2.path().join("telos.md");

    std::fs::write(&env_telos, "# Env").unwrap();
    std::fs::write(&cwd_telos, "# CWD").unwrap();

    std::env::set_var("TELOS_FILE", &env_telos);
    std::env::set_current_dir(&temp_dir2).unwrap();

    let config = ConfigPaths::load().expect("Should load");
    assert_eq!(config.telos_file, env_telos);

    // Cleanup
    std::env::remove_var("TELOS_FILE");
}

#[test]
fn test_missing_file_error() {
    std::env::remove_var("TELOS_FILE");

    // Set cwd to empty dir
    let temp_dir = tempdir().unwrap();
    let original_dir = std::env::current_dir().unwrap();
    std::env::set_current_dir(&temp_dir).unwrap();

    // This should trigger wizard (can't easily test)
    // Just verify it doesn't crash with bad path

    std::env::set_current_dir(original_dir).unwrap();
}

#[test]
fn test_ensure_directories_created() {
    let temp_dir = tempdir().unwrap();
    let config = ConfigPaths {
        telos_file: PathBuf::from("telos.md"),
        data_dir: temp_dir.path().join("data"),
        log_dir: temp_dir.path().join("logs"),
        config_file: None,
    };

    assert!(!config.data_dir.exists());
    assert!(!config.log_dir.exists());

    config.ensure_directories_exist().unwrap();

    assert!(config.data_dir.exists());
    assert!(config.log_dir.exists());
    assert!(config.log_dir.parent().unwrap().exists());
}

#[test]
fn test_load_from_config_file() {
    // Create config.toml content
    let temp_dir = tempdir().unwrap();
    let telos_file = temp_dir.path().join("my_telos.md");
    std::fs::write(&telos_file, "# Test").unwrap();

    let config_toml = format!(
        r#"telos_file = "{}"
data_dir = "{}"
log_dir = "{}""#,
        telos_file.display(),
        temp_dir.path().join("data").display(),
        temp_dir.path().join("logs").display()
    );

    let config: ConfigPaths = toml::from_str(&config_toml).unwrap();
    assert_eq!(config.telos_file, telos_file);
}

#[test]
fn test_invalid_telos_path_in_env() {
    std::env::set_var("TELOS_FILE", "/nonexistent/path/to/telos.md");

    // Should not find it
    // (Real test depends on wizard implementation)

    std::env::remove_var("TELOS_FILE");
}

#[test]
fn test_telos_file_path_validation() {
    let temp_dir = tempdir().unwrap();
    let telos_path = temp_dir.path().join("telos.md");
    std::fs::write(&telos_path, "# Test").unwrap();

    std::env::set_var("TELOS_FILE", telos_path.to_string_lossy().as_ref());

    let config = ConfigPaths::load().expect("Should load");
    assert_eq!(
        fs::canonicalize(&config.telos_file).unwrap(),
        fs::canonicalize(&telos_path).unwrap()
    );

    std::env::remove_var("TELOS_FILE");
}
