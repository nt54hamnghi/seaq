package util

import (
	"fmt"
	"os"
)

// NewTruncateFileWriter creates a new file
// or opens an existing file for writing.
// It truncates the file if it already exists.
func NewTruncateFileWriter(filename string) (*os.File, error) {
	// os.O_WRONLY: Open for writing only.
	// os.O_CREATE: Create the file if it doesn't exist.
	// os.O_TRUNC: If the file exists, truncate it to zero length.
	// 0644: File permissions (read/write for owner, read for others).
	return os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
}

// NewCreateOnlyFileWriter creates a new file for writing
// but fails if the file already exists.
func NewCreateOnlyFileWriter(filename string) (*os.File, error) {
	// Check if the file already exists
	if _, err := os.Stat(filename); err == nil {
		return nil, fmt.Errorf("file %s already exists", filename)
	}
	// File doesn't exist, so create it
	return os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
}

// StdinIsPiped checks if the standard input is piped.
// It returns true if input is from a pipe or redirection, false if it's from an interactive terminal.
// An error is returned if there's an issue accessing stdin.
func StdinIsPiped() (bool, error) {
	stat, err := os.Stdin.Stat()
	if err != nil {
		return false, fmt.Errorf("failed to stat stdin: %w", err)
	}

	// Check if stdin is a character device (interactive terminal)
	//
	// stat.Mode()&os.ModeCharDevice == 0 when:
	// - Input is from a pipe or redirection (non-interactive)
	//
	// stat.Mode()&os.ModeCharDevice != 0 when:
	// - Input is from a terminal (interactive)
	return stat.Mode()&os.ModeCharDevice == 0, nil
}
