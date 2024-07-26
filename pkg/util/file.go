package util

import (
	"fmt"
	"os"
)

func WriteFile(filename string, msg string) error {
	content := []byte(msg)

	// write to file with 0644 permission
	// 0644 means owner can read and write, group can read, and others can read
	// writes data to the named file, creating it if necessary, truncate it if it already exists
	if err := os.WriteFile(filename, content, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// NewOverwriteWriter creates a new file
// or opens an existing file for writing.
// It truncates the file if it already exists.
func NewOverwriteWriter(filename string) (*os.File, error) {
	// os.O_WRONLY: Open for writing only.
	// os.O_CREATE: Create the file if it doesn't exist.
	// os.O_TRUNC: If the file exists, truncate it to zero length.
	// 0644: File permissions (read/write for owner, read for others).
	return os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
}

// NewFailExistingWriter creates a new file for writing
// but fails if the file already exists.
func NewFailExistingWriter(filename string) (*os.File, error) {
	// Check if the file already exists
	if _, err := os.Stat(filename); err == nil {
		return nil, fmt.Errorf("file %s already exists", filename)
	}
	// File doesn't exist, so create it
	return os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
}
