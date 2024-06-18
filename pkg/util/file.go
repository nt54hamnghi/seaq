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
