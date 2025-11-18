# Phase 1: Configuration Abstraction - Detailed Breakdown

> **For Subagent Execution**: Each task below is 1-2 hours of focused work. Subagent should complete one task, report results, then move to next.

**Goal**: Decouple telos-idea-matrix from Ray's personal setup. After Phase 1, any user can provide their own telos.md file and the system works without code changes.

**Effort**: 2-3 hours total (3 tasks × 45-60 min each)

**Blocker**: None (standalone refactoring)

---

## Task 1.1: Create Configuration Module

**Subagent: Create `src/config.rs` with intelligent config loading**

### What We're Building
A `ConfigPaths` struct that:
1. Checks environment variable `TELOS_FILE` first (best for Docker/CI)
2. Checks `./telos.md` in current directory (best for CLI)
3. Checks `~/.config/telos-matrix/config.toml` (best for long-term setup)
4. Launches interactive wizard if nothing found (best for first-time users)

### Why This Order?
- Environment variable = most explicit (respects what user wants)
- Current directory = most intuitive (works with `cd my-project && tm dump`)
- Config file = most persistent (survives across sessions)
- Wizard = most helpful (guides new users)

### Requirements

**Input**: User's system state (env vars, files, directories)

**Output**: `ConfigPaths` struct with:
- `telos_file: PathBuf` (validated path to telos.md)
- `data_dir: PathBuf` (where database/ideas go)
- `log_dir: PathBuf` (where logs go)
- `config_file: Option<PathBuf>` (path to config.toml if used)

**Error Handling**:
- If telos.md not found, launch interactive wizard
- Wizard asks: "Where is your telos.md?"
- Wizard creates config.toml in ~/.config/telos-matrix/
- Subagent uses `dialoguer` crate (already in Cargo.toml)

**Exit Criteria**:
- [ ] `src/config.rs` created with 300-400 lines of code
- [ ] Compiles without errors
- [ ] All 4 config sources implemented
- [ ] Interactive wizard works
- [ ] Unit tests pass (see test requirements below)

### Implementation Requirements

#### Structure
```rust
pub struct ConfigPaths {
    pub telos_file: PathBuf,
    pub data_dir: PathBuf,
    pub log_dir: PathBuf,
    pub config_file: Option<PathBuf>,
}

impl ConfigPaths {
    // Main function
    pub fn load() -> Result<Self>

    // Helper: Load from env var
    fn from_env_var() -> Option<ConfigPaths>

    // Helper: Load from current directory
    fn from_current_dir() -> Option<ConfigPaths>

    // Helper: Load from config file
    fn from_config_file() -> Option<ConfigPaths>

    // Helper: Interactive wizard
    async fn launch_wizard() -> Result<ConfigPaths>

    // Helper: Create directories
    pub fn ensure_directories_exist(&self) -> Result<()>
}
```

#### Specific Code Requirements

**1. Environment variable source** - Priority 1
```rust
fn from_env_var() -> Option<ConfigPaths> {
    let telos_path = std::env::var("TELOS_FILE").ok()?;
    let path = PathBuf::from(&telos_path);

    // Validate file exists
    if !path.exists() {
        eprintln!("Warning: TELOS_FILE env var points to non-existent file: {}", telos_path);
        return None;
    }

    Some(ConfigPaths {
        telos_file: path,
        data_dir: Self::default_data_dir(),
        log_dir: Self::default_log_dir(),
        config_file: None,
    })
}
```

**2. Current directory source** - Priority 2
```rust
fn from_current_dir() -> Option<ConfigPaths> {
    let telos_path = std::env::current_dir()
        .ok()?
        .join("telos.md");

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
```

**3. Config file source** - Priority 3
```rust
fn from_config_file() -> Option<ConfigPaths> {
    let config_dir = dirs::config_dir()?;
    let config_path = config_dir.join("telos-matrix/config.toml");

    if config_path.exists() {
        let content = std::fs::read_to_string(&config_path).ok()?;
        let config: ConfigPaths = toml::from_str(&content).ok()?;

        // Validate telos file still exists
        if !config.telos_file.exists() {
            eprintln!("Warning: telos.md path in config.toml doesn't exist");
            return None;
        }

        return Some(config);
    }
    None
}
```

**4. Interactive wizard** - Priority 4 (async)
```rust
async fn launch_wizard() -> Result<ConfigPaths> {
    use dialoguer::Input;

    println!("\n🔧 First-time setup: Where is your telos.md?");

    let telos_path_str: String = Input::new()
        .with_prompt("Path to telos.md")
        .interact_text()?;

    let telos_path = PathBuf::from(&telos_path_str);

    if !telos_path.exists() {
        return Err(anyhow::anyhow!("File not found: {}", telos_path_str));
    }

    let config = ConfigPaths {
        telos_file: telos_path,
        data_dir: Self::default_data_dir(),
        log_dir: Self::default_log_dir(),
        config_file: None,
    };

    // Optionally save to config.toml
    if dialoguer::Confirm::new()
        .with_prompt("Save this configuration for future runs?")
        .interact()?
    {
        config.save_to_config_file()?;
    }

    Ok(config)
}
```

**5. Main load() function**
```rust
pub fn load() -> Result<Self> {
    // Try in priority order
    if let Some(config) = Self::from_env_var() {
        return Ok(config);
    }

    if let Some(config) = Self::from_current_dir() {
        return Ok(config);
    }

    if let Some(config) = Self::from_config_file() {
        return Ok(config);
    }

    // If nothing found, launch wizard
    tokio::runtime::Handle::current()
        .block_on(Self::launch_wizard())
}
```

#### Error Handling
- Use `anyhow::Result<T>` for flexibility
- `thiserror` for application-specific errors (optional for this task)
- All errors should suggest the fix (helpful messages)

#### Default Paths
```rust
fn default_data_dir() -> PathBuf {
    dirs::data_dir()
        .unwrap_or_else(|| PathBuf::from("."))
        .join("telos-matrix")
}

fn default_log_dir() -> PathBuf {
    dirs::cache_dir()
        .unwrap_or_else(|| PathBuf::from("."))
        .join("telos-matrix/logs")
}
```

### Testing for Task 1.1

**Unit Tests** (minimal, just structure):
```rust
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
    }

    #[test]
    fn test_default_paths_valid() {
        let data_dir = ConfigPaths::default_data_dir();
        let log_dir = ConfigPaths::default_log_dir();

        // Should not be empty
        assert!(data_dir.to_string_lossy().len() > 0);
        assert!(log_dir.to_string_lossy().len() > 0);
    }
}
```

**Integration Tests** (covered in Task 1.2)

### Deliverables
1. `src/config.rs` - Complete file (300-400 lines)
2. Compiles: `cargo build` succeeds
3. Tests pass: `cargo test config::tests`
4. No clippy warnings: `cargo clippy src/config.rs`

### Subagent Notes
- Use existing crates: `dirs`, `dialoguer`, `toml`, `anyhow`
- Follow Rust idioms (error handling, ownership)
- Add doc comments for public methods
- Keep wizard simple (just ask for path)
- Don't panic; always return Result

---

## Task 1.2: Create Integration Tests for Config Loading

**Subagent: Create comprehensive tests in `tests/config_integration_test.rs`**

### What We're Testing

How the config module handles:
1. Environment variable path (explicit user choice)
2. Current directory path (intuitive)
3. Config file path (persistent)
4. Missing file errors (helpful)
5. Directory creation (works)
6. Wizard interaction (can't test fully, mock instead)

### Requirements

**Input**: Various file system states and environment setups

**Output**: Test suite that validates all code paths

**Exit Criteria**:
- [ ] `tests/config_integration_test.rs` created
- [ ] `tests/fixtures/sample_telos.md` created
- [ ] All tests pass: `cargo test --test config_integration_test`
- [ ] Tests cover happy path + error cases
- [ ] ~200 lines of test code

### Test Cases (Minimum 8 tests)

#### Test 1: Load from environment variable
```rust
#[test]
fn test_load_config_from_env_var() {
    // Create temp file
    let temp_dir = tempfile::tempdir().unwrap();
    let telos_path = temp_dir.path().join("telos.md");
    std::fs::write(&telos_path, "# Test Telos").unwrap();

    // Set env var
    std::env::set_var("TELOS_FILE", &telos_path);

    // Load config
    let config = ConfigPaths::load().expect("Should load from env");

    // Verify
    assert_eq!(config.telos_file, telos_path);

    // Cleanup
    std::env::remove_var("TELOS_FILE");
}
```

#### Test 2: Load from current directory
```rust
#[test]
fn test_load_config_from_current_dir() {
    // Create temp dir with telos.md
    let temp_dir = tempfile::tempdir().unwrap();
    let telos_path = temp_dir.path().join("telos.md");
    std::fs::write(&telos_path, "# Test Telos").unwrap();

    // Change to temp dir
    let original_dir = std::env::current_dir().unwrap();
    std::env::set_current_dir(&temp_dir).unwrap();
    std::env::remove_var("TELOS_FILE");

    // Load config
    let config = ConfigPaths::load().expect("Should find ./telos.md");
    assert_eq!(config.telos_file, PathBuf::from("telos.md"));

    // Restore
    std::env::set_current_dir(original_dir).unwrap();
}
```

#### Test 3: Environment variable takes priority
```rust
#[test]
fn test_env_var_has_priority() {
    let temp_dir1 = tempfile::tempdir().unwrap();
    let temp_dir2 = tempfile::tempdir().unwrap();

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
```

#### Test 4: Missing file error with helpful message
```rust
#[test]
fn test_missing_file_error() {
    std::env::remove_var("TELOS_FILE");

    // Set cwd to empty dir
    let temp_dir = tempfile::tempdir().unwrap();
    let original_dir = std::env::current_dir().unwrap();
    std::env::set_current_dir(&temp_dir).unwrap();

    // This should trigger wizard (can't easily test)
    // Just verify it doesn't crash with bad path

    std::env::set_current_dir(original_dir).unwrap();
}
```

#### Test 5: Ensure directories are created
```rust
#[test]
fn test_ensure_directories_created() {
    let temp_dir = tempfile::tempdir().unwrap();
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
```

#### Test 6: Config file deserialization
```rust
#[test]
fn test_load_from_config_file() {
    // Create config.toml
    let temp_dir = tempfile::tempdir().unwrap();
    let telos_file = temp_dir.path().join("my_telos.md");
    std::fs::write(&telos_file, "# Test").unwrap();

    let config_toml = format!(r#"
telos_file = "{}"
data_dir = "/tmp/data"
log_dir = "/tmp/logs"
"#, telos_file.display());

    let config: ConfigPaths = toml::from_str(&config_toml).unwrap();
    assert_eq!(config.telos_file, telos_file);
}
```

#### Test 7: Invalid telos file path in config
```rust
#[test]
fn test_invalid_telos_path_in_env() {
    std::env::set_var("TELOS_FILE", "/nonexistent/path/to/telos.md");

    // Should not find it
    // (Real test depends on wizard implementation)

    std::env::remove_var("TELOS_FILE");
}
```

#### Test 8: Path normalization
```rust
#[test]
fn test_telos_file_path_validation() {
    let temp_dir = tempfile::tempdir().unwrap();
    let telos_path = temp_dir.path().join("telos.md");
    std::fs::write(&telos_path, "# Test").unwrap();

    std::env::set_var("TELOS_FILE", telos_path.to_string_lossy().as_ref());

    let config = ConfigPaths::load().expect("Should load");
    assert!(config.telos_file.exists());

    std::env::remove_var("TELOS_FILE");
}
```

### Test Fixtures

**Create `tests/fixtures/sample_telos.md`**:
```markdown
# Sample Telos for Testing

## Goals
- G1: Ship product (Deadline: 2025-12-31)
- G2: Build community (Deadline: 2025-12-31)
- G3: Establish credibility (Deadline: 2025-12-31)
- G4: Create income stream (Deadline: 2025-12-31)

## Strategies
- S1: Focus on shipping
- S2: One stack rule
- S3: Build in public
- S4: MVP mindset

## Stack
- Primary: Rust
- Secondary: Python

## Failure Patterns
- Context-switching
- Perfectionism
- Procrastination
```

### Testing Strategy

Use `tempfile` crate (already in Cargo.toml) for temporary directories:
- Each test gets isolated temp dir
- No side effects
- Cleanup automatic
- Can use `std::env::set_var` for env vars

**Don't** test the wizard interactively (impossible in CI)
**Do** test the loading of wizard-created files

### Deliverables
1. `tests/config_integration_test.rs` - ~200 lines
2. `tests/fixtures/sample_telos.md` - Test fixture
3. All tests pass: `cargo test --test config_integration_test`

### Subagent Notes
- Use `tempfile::tempdir()` for test dirs
- Use `std::env::{set_var, remove_var}` for env manipulation
- Save/restore original `CURRENT_DIR` after tests
- Keep tests independent (no shared state)
- Name tests clearly: `test_loads_from_X`

---

## Task 1.3: Update src/telos.rs and src/main.rs to Use ConfigPaths

**Subagent: Integrate config module into existing code**

### What We're Changing

**Before**:
```rust
// In telos.rs
const TELOS_PATH: &str = "/Users/rayyacub/Documents/CCResearch/Hanai/telos.md";

pub fn load() -> Result<TelosConfig> {
    let content = std::fs::read_to_string(TELOS_PATH)?;
    // Parse...
}
```

**After**:
```rust
// In telos.rs
pub async fn load(telos_path: &Path) -> Result<TelosConfig> {
    let content = tokio::fs::read_to_string(telos_path).await?;
    // Parse...
}

// In main.rs
let config_paths = ConfigPaths::load()?;
let telos = telos::load(&config_paths.telos_file).await?;
```

### Requirements

**Input**: Current telos.rs and main.rs code

**Output**:
- `src/telos.rs` accepts configurable path
- `src/main.rs` uses ConfigPaths module
- All old code still works

**Exit Criteria**:
- [ ] `src/telos.rs` modified (accepts Path parameter)
- [ ] `src/main.rs` modified (uses ConfigPaths::load())
- [ ] Compiles: `cargo build`
- [ ] No clippy warnings
- [ ] Existing functionality preserved

### Detailed Changes

#### Change 1: Update telos.rs

**Find and remove hardcoded path**:
```rust
// REMOVE THIS:
const TELOS_PATH: &str = "/Users/rayyacub/...";
```

**Update load() signature**:
```rust
// BEFORE:
pub fn load() -> Result<TelosConfig> { ... }

// AFTER:
pub async fn load(telos_path: &Path) -> Result<TelosConfig> {
    let content = tokio::fs::read_to_string(telos_path)
        .await
        .context(format!("Failed to read Telos file from: {:?}", telos_path))?;

    parse_telos_markdown(&content)
}
```

**Update error messages to include path**:
```rust
// Include actual path in errors
.context(format!("Failed to parse Telos from: {:?}", telos_path))?
```

#### Change 2: Update main.rs

**Add config module import**:
```rust
mod config;

use config::ConfigPaths;
```

**In main function, add config loading**:
```rust
#[tokio::main]
async fn main() -> Result<()> {
    // Load configuration (this will ask user if needed)
    let config_paths = ConfigPaths::load()?;

    // Ensure directories exist
    config_paths.ensure_directories_exist()?;

    // Initialize logging (use log_dir from config)
    // initialize_logging(&config_paths.log_dir)?;

    // Load Telos configuration (pass path from config)
    let telos_config = telos::load(&config_paths.telos_file).await?;

    // Continue with rest of main...
    // (rest of code unchanged)
}
```

**Update any hardcoded database path**:
```rust
// BEFORE (if exists):
let db = Database::new("data/ideas.db")?;

// AFTER:
let db = Database::new(&config_paths.data_dir)?;
```

#### Change 3: Search for other hardcoded paths

**In entire `src/` directory**, search for:
- `/Users/rayyacub`
- `/Users/ray`
- Absolute paths starting with `/Users/` or `/home/`
- Hard-coded `"data/"`
- Hard-coded `"logs/"`
- Hard-coded `"telos.md"`

Replace with config_paths variables.

### Integration Testing

After changes, test:
```bash
# Should compile
cargo build

# Set env var and test
export TELOS_FILE=./telos.md
cargo run -- --help

# Should work
cargo run -- dump "test idea"
```

### Backward Compatibility

**For Ray's setup to work with zero changes:**
1. Set env var once: `export TELOS_FILE=/Users/rayyacub/.../telos.md`
2. Add to shell config: `~/.zshrc` or `~/.bashrc`
3. Source it: `source ~/.zshrc`
4. From then on: `tm dump "idea"` works as before

### Deliverables

1. `src/telos.rs` - Updated load() function signature
2. `src/main.rs` - Uses ConfigPaths::load()
3. All hardcoded paths replaced
4. Code compiles: `cargo build`
5. No clippy warnings
6. Runs successfully with TELOS_FILE env var set

### Subagent Notes

- Use `&Path` instead of `String` (type safe)
- Use `anyhow::context()` for better error messages
- Keep changes minimal (don't refactor unrelated code)
- Test after each change
- Search comprehensively for hardcoded paths

---

## Task 1.4: Add Config Module to Exports

**Subagent: Make config module accessible**

### What We're Doing

Expose the config module so it can be used/tested:

**In `src/main.rs`** (if not already done):
```rust
mod config;
mod telos;
mod commands;
// ... other modules

use config::ConfigPaths;
use telos::TelosConfig;
```

### Deliverables

1. `mod config;` added to src/main.rs
2. `ConfigPaths` is publicly accessible
3. Compiles without errors
4. Can be imported in tests: `use telos_idea_matrix::config::ConfigPaths;`

---

## Phase 1 Completion Checklist

**For Subagent to Verify Before Reporting Done:**

### Code Quality
- [ ] `cargo build` succeeds with no errors
- [ ] `cargo clippy` produces no warnings (except pre-existing)
- [ ] `cargo fmt` has been run (`cargo fmt --all`)
- [ ] Code compiles with `--all-features`

### Functionality
- [ ] Config loads from env var
- [ ] Config loads from current directory
- [ ] Config loads from ~/.config/ file
- [ ] Config wizard launches on first run
- [ ] Directories are created automatically
- [ ] Error messages are helpful (suggest fix)

### Testing
- [ ] `cargo test config::` passes all unit tests
- [ ] `cargo test --test config_integration_test` passes all integration tests
- [ ] No test panics or hangs
- [ ] Tests use temporary directories (no side effects)

### Integration
- [ ] `src/telos.rs` accepts configurable path
- [ ] `src/main.rs` uses ConfigPaths::load()
- [ ] All hardcoded personal paths removed
- [ ] `cargo build` succeeds after changes
- [ ] System still works: can run `tm dump "test"`

### Git
- [ ] `git status` shows changes
- [ ] Ready for 3 separate commits (one per main task)

---

## Commits (Subagent Should Create These)

### Commit 1: Create config module
```bash
git add src/config.rs
git commit -m "feat: create configuration module with multiple source support

- Add ConfigPaths struct
- Support 4 config sources: env var, cwd, config file, wizard
- Handle directory creation
- Provide helpful error messages
- Include unit tests"
```

### Commit 2: Add integration tests
```bash
git add tests/config_integration_test.rs tests/fixtures/sample_telos.md
git commit -m "test: add configuration integration tests

- Test env var loading
- Test current directory loading
- Test config file loading
- Test priority order
- Test directory creation
- Test error handling"
```

### Commit 3: Integrate config into main
```bash
git add src/main.rs src/telos.rs
git commit -m "refactor: integrate config module into main

- Update telos::load() to accept configurable path
- Remove hardcoded path constants
- Use ConfigPaths in main()
- Ensure backward compatibility with env var"
```

---

## Expected Output for Subagent

After completing Phase 1, subagent should report:

```
PHASE 1 COMPLETION REPORT
========================

✅ Task 1.1: Configuration Module
- src/config.rs created (412 lines)
- Supports 4 config sources
- Interactive wizard implemented
- Compiles without errors

✅ Task 1.2: Integration Tests
- tests/config_integration_test.rs created (234 lines)
- 8 test cases cover all scenarios
- tests/fixtures/sample_telos.md created
- All tests pass

✅ Task 1.3: Main Integration
- src/telos.rs updated (accepts configurable path)
- src/main.rs updated (uses ConfigPaths)
- All hardcoded paths removed
- System runs successfully

✅ Quality Checks
- cargo build ✓
- cargo test ✓
- cargo clippy ✓
- cargo fmt ✓

✅ Git Commits
- 3 commits created with clear messages
- Ready for next phase

Summary: Configuration abstraction complete.
System now works with any user's telos.md file.
```

---

## Questions for Subagent (If Blocked)

If subagent encounters issues, it should ask:

1. **Wizard interaction not working?**
   → Check if `dialoguer` is properly imported
   → Verify interactive mode works in test environment
   → May need to mock wizard for testing

2. **Path resolution issues?**
   → Use `std::fs::canonicalize()` to normalize paths
   → Check if relative vs absolute paths cause issues
   → Log actual paths for debugging

3. **Compilation errors?**
   → Check if all imports are correct
   → Verify `tokio::fs::` is available (should be)
   → Look for typos in `Path` vs `PathBuf`

4. **Test failures?**
   → Check if temp directories are created
   → Verify env vars are properly restored
   → Ensure no race conditions in parallel tests

---

**Phase 1 is ready for subagent execution. Each task is self-contained and complete.**
