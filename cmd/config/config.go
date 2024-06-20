package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// region: --- errors

var ErrEmptyPattern = errors.New("pattern name is empty")
var ErrEmptyRepo = errors.New("repo is empty")

func patternNotFound(name string) error {
	return fmt.Errorf("pattern '%s' does not exist", name)
}

// endregion: --- errors

var Hiku *HikuConfig

type HikuConfig struct {
	*viper.Viper
}

func New() *HikuConfig {
	return &HikuConfig{
		viper.New(),
	}
}

func (hiku *HikuConfig) Repo() string {
	return hiku.GetString("pattern.repo")
}

func (hiku *HikuConfig) Pattern() string {
	return hiku.GetString("pattern.name")
}

func (hiku *HikuConfig) HasPattern(name string) bool {
	pats, err := hiku.GetAvailablePatterns()
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

func (hiku *HikuConfig) UsePattern(name string) error {
	if !hiku.HasPattern(name) {
		return patternNotFound(name)
	}
	hiku.Set("pattern.name", name)
	return nil
}

func (hiku *HikuConfig) GetPrompt() (string, error) {
	pat := hiku.Pattern()
	if pat == "" {
		return "", ErrEmptyPattern
	}

	repo := hiku.Repo()
	if repo == "" {
		return "", ErrEmptyRepo
	}

	path := filepath.Join(repo, pat, "system.md")

	prompt, err := os.ReadFile(path) // read the pattern
	if err != nil {
		if os.IsNotExist(err) {
			return "", patternNotFound(pat)
		}
		return "", err
	}

	return string(prompt), nil
}

func (hiku *HikuConfig) GetAvailablePatterns() ([]string, error) {
	repo := hiku.Repo()
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
