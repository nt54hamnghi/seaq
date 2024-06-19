package pattern

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type PatternConfig struct {
	repo    string
	pattern string
}

func WithName(name string) (cfg PatternConfig, err error) {
	// use viper to read the config file
	// get the patterns repo path
	cfg.repo = viper.GetString("patterns.repo")
	if cfg.repo == "" {
		return cfg, fmt.Errorf("patterns.repo is not set in the config file")
	}

	// get the default prompt
	cfg.pattern = name // set to the pattern flag value
	if cfg.pattern == "" {
		cfg.pattern = viper.GetString("patterns.default")
	}

	if cfg.pattern == "" {
		return cfg, fmt.Errorf("no pattern provided")
	}

	return
}

func (cfg PatternConfig) GetAvailablePatterns() ([]string, error) {
	dirs, err := os.ReadDir(cfg.repo)
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

func (cfg PatternConfig) GetPrompt() (string, error) {
	// read the pattern
	prompt, err := os.ReadFile(filepath.Join(cfg.repo, cfg.pattern, "system.md"))

	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("pattern %s does not exist", cfg.pattern)
		}
		return "", err
	}

	return string(prompt), nil
}
