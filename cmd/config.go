package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

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

func (hiku *HikuConfig) UsePattern(name string) error {
	if name == "" {
		return fmt.Errorf("pattern name is empty")
	}
	hiku.Set("pattern.name", name)
	return nil
}

func (hiku *HikuConfig) GetPrompt() (string, error) {
	pat := hiku.Pattern()
	if pat == "" {
		return "", fmt.Errorf("pattern name is empty")
	}

	repo := hiku.Repo()
	if repo == "" {
		return "", fmt.Errorf("repo is empty")
	}

	path := filepath.Join(repo, pat, "system.md")

	prompt, err := os.ReadFile(path) // read the pattern
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("pattern %s does not exist", pat)
		}
		return "", err
	}

	return string(prompt), nil
}

func (hiku *HikuConfig) GetAvailablePatterns() ([]string, error) {
	dirs, err := os.ReadDir(hiku.Repo())
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
