package config

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// region: --- errors

var (
	ErrEmptyPattern = errors.New("pattern name is empty")
	ErrEmptyRepo    = errors.New("repo is empty")
)

// endregion: --- errors

func Repo() string {
	return viper.GetString("pattern.repo")
}

func Pattern() string {
	return viper.GetString("pattern.name")
}

func HasPattern(name string) bool {
	pats, err := ListPatterns()
	if err != nil {
		return false
	}

	for _, p := range pats {
		if p == name {
			return true
		}
	}

	return false
}

func UsePattern(name string) error {
	if !HasPattern(name) {
		return &Unsupported{Type: "pattern", Key: name}
	}
	viper.Set("pattern.name", name)
	return nil
}

func GetPrompt() (string, error) {
	pat := Pattern()
	if pat == "" {
		return "", ErrEmptyPattern
	}

	repo := Repo()
	if repo == "" {
		return "", ErrEmptyRepo
	}

	path := filepath.Join(repo, pat, "system.md")

	prompt, err := os.ReadFile(path) // read the pattern
	if err != nil {
		if os.IsNotExist(err) {
			return "", &Unsupported{Type: "pattern", Key: pat}
		}
		return "", err
	}

	return string(prompt), nil
}

// ListPatterns returns a list of available patterns
// in the repository specified in the config.
func ListPatterns() ([]string, error) {
	return ListPatternsInRepo(Repo())
}

// ListPatternsInRepo returns a list of available patterns in a given repository
func ListPatternsInRepo(repo string) ([]string, error) {
	if repo == "" {
		return nil, ErrEmptyRepo
	}

	dirs, err := os.ReadDir(repo)
	if err != nil {
		return nil, err
	}

	pats := make([]string, 0, len(dirs))
	for _, d := range dirs {
		if d.IsDir() {
			pats = append(pats, d.Name())
		}
	}
	return pats, nil
}
