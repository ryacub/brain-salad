// Package utils provides utility functions for clipboard operations and other common tasks.
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
