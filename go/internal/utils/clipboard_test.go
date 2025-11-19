package utils

import (
	"strings"
	"testing"
)

func TestClipboard_Copy(t *testing.T) {
	testText := "Test clipboard content"

	err := CopyToClipboard(testText)
	if err != nil {
		// In headless environments, clipboard might not be available
		// This is acceptable for CI/CD
		t.Logf("Clipboard not available: %v", err)
		return
	}

	// Verify we can read back what we wrote
	retrieved, err := PasteFromClipboard()
	if err != nil {
		t.Fatalf("Failed to paste from clipboard: %v", err)
	}

	if retrieved != testText {
		t.Errorf("Expected %q, got %q", testText, retrieved)
	}
}

func TestClipboard_Paste(t *testing.T) {
	// First write known content
	testText := "Test paste operation"

	err := CopyToClipboard(testText)
	if err != nil {
		t.Logf("Clipboard not available: %v", err)
		return
	}

	// Now test paste
	retrieved, err := PasteFromClipboard()
	if err != nil {
		t.Fatalf("Failed to paste from clipboard: %v", err)
	}

	if retrieved != testText {
		t.Errorf("Paste failed: expected %q, got %q", testText, retrieved)
	}
}

func TestClipboard_MultilineContent(t *testing.T) {
	testText := "Line 1\nLine 2\nLine 3"

	err := CopyToClipboard(testText)
	if err != nil {
		t.Logf("Clipboard not available: %v", err)
		return
	}

	retrieved, err := PasteFromClipboard()
	if err != nil {
		t.Fatalf("Failed to paste multiline content: %v", err)
	}

	if retrieved != testText {
		t.Errorf("Multiline paste failed: expected %q, got %q", testText, retrieved)
	}
}

func TestClipboard_EmptyString(t *testing.T) {
	err := CopyToClipboard("")
	if err != nil {
		t.Logf("Clipboard not available: %v", err)
		return
	}

	retrieved, err := PasteFromClipboard()
	if err != nil {
		t.Fatalf("Failed to paste empty string: %v", err)
	}

	if retrieved != "" {
		t.Errorf("Expected empty string, got %q", retrieved)
	}
}

func TestClipboard_LargeContent(t *testing.T) {
	// Test with larger content (1KB)
	testText := strings.Repeat("Test content with some variety. ", 30)

	err := CopyToClipboard(testText)
	if err != nil {
		t.Logf("Clipboard not available: %v", err)
		return
	}

	retrieved, err := PasteFromClipboard()
	if err != nil {
		t.Fatalf("Failed to paste large content: %v", err)
	}

	if retrieved != testText {
		t.Errorf("Large content paste failed: lengths %d vs %d", len(testText), len(retrieved))
	}
}

func TestClipboard_UnavailableHandler(t *testing.T) {
	// This test verifies that IsClipboardAvailable provides a way
	// to check clipboard availability gracefully

	available := IsClipboardAvailable()

	// If clipboard is available, test operations
	if available {
		testText := "Availability test"
		err := CopyToClipboard(testText)
		if err != nil {
			t.Errorf("Clipboard marked as available but copy failed: %v", err)
		}

		retrieved, err := PasteFromClipboard()
		if err != nil {
			t.Errorf("Clipboard marked as available but paste failed: %v", err)
		}

		if retrieved != testText {
			t.Errorf("Expected %q, got %q", testText, retrieved)
		}
	} else {
		t.Log("Clipboard not available on this system")
	}
}

func TestClipboard_SpecialCharacters(t *testing.T) {
	testText := "Special chars: 日本語 émojis 🎉 symbols €£¥"

	err := CopyToClipboard(testText)
	if err != nil {
		t.Logf("Clipboard not available: %v", err)
		return
	}

	retrieved, err := PasteFromClipboard()
	if err != nil {
		t.Fatalf("Failed to paste special characters: %v", err)
	}

	if retrieved != testText {
		t.Errorf("Special characters paste failed: expected %q, got %q", testText, retrieved)
	}
}
