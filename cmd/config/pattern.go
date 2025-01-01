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

func New() *SeaqConfig {
	return &SeaqConfig{
		viper.New(),
	}
}

func (sc *SeaqConfig) Repo() string {
	return sc.GetString("pattern.repo")
}

func (sc *SeaqConfig) Pattern() string {
	return sc.GetString("pattern.name")
}

func (sc *SeaqConfig) HasPattern(name string) bool {
	pats, err := sc.ListPatterns()
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

func (sc *SeaqConfig) UsePattern(name string) error {
	if !sc.HasPattern(name) {
		return &Unsupported{Type: "pattern", Key: name}
	}
	sc.Set("pattern.name", name)
	return nil
}

func (sc *SeaqConfig) GetPrompt() (string, error) {
	pat := sc.Pattern()
	if pat == "" {
		return "", ErrEmptyPattern
	}

	repo := sc.Repo()
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
func (sc *SeaqConfig) ListPatterns() ([]string, error) {
	repo := sc.Repo()
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
