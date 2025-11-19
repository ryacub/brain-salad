# Track 8B: Utilities (Clipboard Integration)

**Phase**: 8 - Polish & Documentation  
**Estimated Time**: 2-3 hours
**Dependencies**: None
**Can Run in Parallel**: Yes (with 7, 8A)

---

## Mission

You are implementing clipboard integration for the Telos Idea Matrix Go application, following Test-Driven Development (TDD).

## Context

- Rust implementation has clipboard support in `src/clipboard_helper.rs`
- Need cross-platform clipboard operations (Linux, macOS, Windows)
- Integration with dump command for quick idea capture

## Reference Implementation

Review `/home/user/brain-salad/src/clipboard_helper.rs`

## Your Task

Implement clipboard integration using strict TDD methodology.

## Directory Structure

Create files in `go/internal/utils/`:
- `clipboard.go` - Clipboard operations
- `clipboard_test.go` - Tests

## TDD Workflow (RED → GREEN → REFACTOR)

### STEP 1 - RED PHASE (Write Failing Tests)

Create `go/internal/utils/clipboard_test.go`:
- `TestClipboard_Copy()`
- `TestClipboard_Paste()`
- `TestClipboard_UnavailableHandler()`

Run: `go test ./internal/utils -v`
Expected: **ALL TESTS FAIL**

### STEP 2 - GREEN PHASE (Implement)

#### A. Install dependency:

```bash
go get github.com/atotto/clipboard
```

#### B. Implement `go/internal/utils/clipboard.go`:

```go
package utils

import (
    "fmt"
    
    "github.com/atotto/clipboard"
)

// CopyToClipboard copies text to system clipboard
func CopyToClipboard(text string) error {
    if err := clipboard.WriteAll(text); err != nil {
        return fmt.Errorf("copy to clipboard: %w", err)
    }
    return nil
}

// PasteFromClipboard retrieves text from system clipboard
func PasteFromClipboard() (string, error) {
    text, err := clipboard.ReadAll()
    if err != nil {
        return "", fmt.Errorf("paste from clipboard: %w", err)
    }
    return text, nil
}

// IsClipboardAvailable checks if clipboard is accessible
func IsClipboardAvailable() bool {
    _, err := clipboard.ReadAll()
    return err == nil
}
```

Run: `go test ./internal/utils -v`
Expected: **ALL TESTS PASS**

### STEP 3 - INTEGRATION

#### Update `go/internal/cli/dump.go`:

```go
// Add flags
var (
    fromClipboard bool
    toClipboard   bool
)

dumpCmd.Flags().BoolVar(&fromClipboard, "from-clipboard", false, "Read idea from clipboard")
dumpCmd.Flags().BoolVar(&toClipboard, "to-clipboard", false, "Copy result to clipboard")

// In RunE:
var ideaContent string

if fromClipboard {
    text, err := utils.PasteFromClipboard()
    if err != nil {
        return fmt.Errorf("read clipboard: %w", err)
    }
    ideaContent = text
} else if len(args) > 0 {
    ideaContent = args[0]
} else {
    return fmt.Errorf("provide idea text or use --from-clipboard")
}

// After analysis...
if toClipboard {
    summary := fmt.Sprintf("Score: %.1f\n%s", result.FinalScore, result.Recommendation)
    if err := utils.CopyToClipboard(summary); err != nil {
        fmt.Printf("Warning: failed to copy to clipboard: %v\n", err)
    } else {
        fmt.Println("✓ Result copied to clipboard")
    }
}
```

## Success Criteria

- ✅ All tests pass with >80% coverage
- ✅ Works on Linux and macOS
- ✅ Graceful handling when clipboard unavailable
- ✅ Integration with dump command works

## Validation

```bash
# Copy idea from clipboard
echo "Build a tool" | xclip -selection clipboard
tm dump --from-clipboard

# Copy result to clipboard
tm dump "Build automation tool" --to-clipboard
xclip -o -selection clipboard
```

## Deliverables

- `go/internal/utils/clipboard.go`
- `go/internal/utils/clipboard_test.go`
- Updated `go/internal/cli/dump.go`
