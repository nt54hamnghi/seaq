/*
Copyright © 2024 Nghi Nguyen
*/

package config

import (
	_ "embed"
	"fmt"
	"path/filepath"
	"slices"

	"github.com/charmbracelet/huh"
	"github.com/nt54hamnghi/seaq/pkg/config"
	"github.com/nt54hamnghi/seaq/pkg/llm"
	"github.com/nt54hamnghi/seaq/pkg/util/log"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var fs = afero.Afero{
	Fs: afero.NewOsFs(),
}

func newSetupCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "setup",
		Short:        "Setup a new config file interactively",
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE: func(_ *cobra.Command, _ []string) error {
			input, err := collect()
			if err != nil {
				return err
			}

			return input.setup()
		},
	}

	return cmd
}

type setupInput struct {
	modelName   string
	patternName string
	patternRepo string
}

// setup writes the configuration to a file and initializes the pattern repository.
// If using the default pattern repository, it will also create the directory structure with the built-in patterns.
func (i setupInput) setup() error {
	// Validate inputs
	if i.modelName == "" || i.patternName == "" || i.patternRepo == "" {
		return fmt.Errorf(
			"invalid configuration: model name, pattern name, and pattern repository are required",
		)
	}

	configDir, configFile := getConfig()

	// Create config directory if not exists
	if err := fs.MkdirAll(configDir, 0o755); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	log.Info("setting up configuration",
		"model", i.modelName,
		"pattern", i.patternName,
		"pattern_repo", i.patternRepo,
	)

	viper.SetConfigType("yaml")

	// Set values
	viper.Set("model.name", i.modelName)
	viper.Set("pattern.name", i.patternName)
	viper.Set("pattern.repo", i.patternRepo)

	// SafeWriteConfigAs will error if the file already exists
	log.Info("writing configuration", "config_file", configFile)
	if err := viper.SafeWriteConfigAs(configFile); err != nil {
		return fmt.Errorf("writing config file: %w", err)
	}

	if err := i.createDefaultPatternRepo(); err != nil {
		return fmt.Errorf("setting up default patterns: %w", err)
	}

	return nil
}

//go:embed patterns/prime_mind/system.md
var primeMindPattern []byte

//go:embed patterns/improve_prompt/system.md
var improvePromptPattern []byte

// createDefaultPatternRepo creates the default pattern directories.
// It will create directories for built-in patterns only if:
// 1. The configured pattern repository is the default one
// 2. The repository directory doesn't already exist
// The default patterns created are "prime_mind" and "improve_prompt".
func (i setupInput) createDefaultPatternRepo() error {
	if i.patternRepo != getDefaultPatternsRepo() {
		return nil
	}

	if exists, err := fs.DirExists(i.patternRepo); err != nil {
		return err
	} else if exists {
		return nil // directory already exists
	}

	log.Info("setting up default patterns", "pattern_repo", i.patternRepo)

	defaultPatterns := map[string][]byte{
		"prime_mind":     primeMindPattern,
		"improve_prompt": improvePromptPattern,
	}
	for name, data := range defaultPatterns {
		patDir := filepath.Join(i.patternRepo, name)
		patFile := filepath.Join(patDir, "system.md")

		log.Info("creating pattern directory", "pattern", name)
		if err := fs.MkdirAll(patDir, 0o755); err != nil {
			return fmt.Errorf("creating pattern directory %s: %w", name, err)
		}
		if err := fs.WriteFile(patFile, data, 0o600); err != nil {
			return fmt.Errorf("writing pattern file %s: %w", name, err)
		}
	}

	return nil
}

func getDefaultPatternsRepo() string {
	configDir, _ := getConfig()
	return filepath.Join(configDir, "patterns")
}

// collect interactively gathers the setup input from the user.
// It validates that the config file doesn't already exist and ensures
// the pattern repository is valid if specified.
func collect() (setupInput, error) {
	var input setupInput

	// Check if config already exists
	_, configFile := getConfig()
	if exists, err := fs.Exists(configFile); err != nil {
		return input, err
	} else if exists {
		return input, fmt.Errorf("config file %s already exists", configFile)
	}

	defaultPatternRepo := getDefaultPatternsRepo()

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Default model").
				Height(20).
				OptionsFunc(func() []huh.Option[string] {
					models := slices.Collect(llm.BuiltinModels())
					return huh.NewOptions(models...)
				}, nil).
				Value(&input.modelName),
		),
		huh.NewGroup(
			huh.NewInput().
				Title("Pattern repository").
				Description(fmt.Sprintf("The folder containing prompts. Leave empty to use %s.", defaultPatternRepo)).
				Placeholder(defaultPatternRepo).
				Value(&input.patternRepo).
				Validate(func(s string) error {
					// empty is ok, default will be used
					if s == "" {
						return nil
					}

					if exists, err := fs.DirExists(s); err != nil {
						return err
					} else if !exists {
						return fmt.Errorf("%s does not exist", s)
					}

					// A valid pattern repository must contain at least one subfolder
					// where each subfolder represents a pattern
					entries, err := fs.ReadDir(s)
					if err != nil {
						return err
					}
					for _, entry := range entries {
						if entry.IsDir() {
							return nil // found at least one pattern directory
						}
					}
					return fmt.Errorf("%s contains no pattern folders", s)
				}),
		),
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Default pattern").
				OptionsFunc(func() []huh.Option[string] {
					if input.patternRepo == "" {
						return huh.NewOptions("prime_mind", "improve_prompt")
					}
					patterns, err := config.ListPatternsInRepo(input.patternRepo)
					if err != nil {
						return nil
					}
					return huh.NewOptions(patterns...)
				}, &input.patternRepo).
				Value(&input.patternName),
		),
	)

	if err := form.Run(); err != nil {
		return setupInput{}, err
	}

	// if pattern repo is not set, use default
	if input.patternRepo == "" {
		input.patternRepo = defaultPatternRepo
	}

	return input, nil
}
