package flag

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/afero"
)

// TESTME: test this file

var fs = afero.Afero{
	Fs: afero.NewOsFs(),
}

type FilePath string

// String implements the pflag.Value interface
// It returns the string representation of the FilePath
func (f *FilePath) String() string {
	return string(*f)
}

// Set implements the pflag.Value interface
// It parses the input string and sets the FilePath
func (f *FilePath) Set(s string) error {
	// get absolute path
	absPath, err := filepath.Abs(s)
	if err != nil {
		return err
	}

	if exists, err := fs.Exists(absPath); err != nil {
		return fmt.Errorf("error checking file %q: %w", absPath, err)
	} else if !exists {
		return fmt.Errorf("%q does not exist", absPath)
	}

	if isDir, err := fs.IsDir(absPath); err != nil {
		return fmt.Errorf("error checking file %q: %w", absPath, err)
	} else if isDir {
		return fmt.Errorf("%q is a directory", absPath)
	}

	*f = FilePath(absPath)
	return nil
}

// Type implements the pflag.Value interface
// It returns the type of the FilePath flag in help message
func (f *FilePath) Type() string {
	return "string"
}

type DirPath string

// String implements the pflag.Value interface
// It returns the string representation of the DirPath
func (d *DirPath) String() string {
	return string(*d)
}

// Set implements the pflag.Value interface
// It parses the input string and sets the FilePath
func (d *DirPath) Set(s string) error {
	// get absolute path
	absPath, err := filepath.Abs(s)
	if err != nil {
		return err
	}

	if ok, err := fs.DirExists(absPath); err != nil {
		return fmt.Errorf("error checking %q: %w", absPath, err)
	} else if !ok {
		return fmt.Errorf("%q does not exist or is not a directory", absPath)
	}

	*d = DirPath(absPath)
	return nil
}

// Type implements the pflag.Value interface
// It returns the type of the FilePath flag in help message
func (d *DirPath) Type() string {
	return "string"
}
