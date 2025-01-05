package flag

import (
	"fmt"
	"os"
	"path/filepath"
)

// TESTME: test this file

type FilePath string

// String implements the pflag.Value interface
// It returns the string representation of the FilePath
func (fp *FilePath) String() string {
	return string(*fp)
}

// Set implements the pflag.Value interface
// It parses the input string and sets the FilePath
func (fp *FilePath) Set(s string) error {
	// get absolute path
	absPath, err := filepath.Abs(s)
	if err != nil {
		return err
	}

	switch info, err := os.Stat(absPath); {
	case os.IsNotExist(err):
		return fmt.Errorf("%q does not exist", absPath)
	case err != nil:
		return fmt.Errorf("error checking file %q: %w", absPath, err)
	case info.IsDir():
		return fmt.Errorf("%q is a directory", absPath)
	default:
		*fp = FilePath(absPath)
		return nil
	}
}

// Type implements the pflag.Value interface
// It returns the type of the FilePath flag in help message
func (fp *FilePath) Type() string {
	return "string"
}
