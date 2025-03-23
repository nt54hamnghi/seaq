/*
Copyright © 2024 Nghi Nguyen
*/

package config

import (
	_ "embed"
	"errors"
	"fmt"
	"path/filepath"
	"slices"

	"github.com/charmbracelet/huh"
	"github.com/nt54hamnghi/seaq/cmd/flag"
	"github.com/nt54hamnghi/seaq/pkg/config"
	"github.com/nt54hamnghi/seaq/pkg/llm"
	"github.com/nt54hamnghi/seaq/pkg/util/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	DefaultRemotePatternsRepo = "https://github.com/danielmiessler/fabric"
)

type setupOptions struct {
	dir flag.DirPath
}

func newSetupCmd() *cobra.Command {
	var opts setupOptions

	cmd := &cobra.Command{
		Use:          "setup",
		Short:        "Setup a new config file interactively",
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE: func(_ *cobra.Command, _ []string) error {
			input, err := collect(opts)
			if err != nil {
				return err
			}

			return input.setup(opts)
		},
	}

	// set up flags
	flags := cmd.Flags()
	flags.SortFlags = false
	flags.VarP(&opts.dir, "dir", "d", "directory to create the config file in")

	return cmd
}

type setupInput struct {
	modelName     string
	patternName   string
	patternRepo   string
	patternRemote string
}

// setup writes the configuration to a file and initializes the pattern repository.
// If using the default pattern repository, it will also create the directory structure with the built-in patterns.
func (i setupInput) setup(opts setupOptions) error {
	// Validate inputs
	if i.modelName == "" || i.patternName == "" || i.patternRepo == "" || i.patternRemote == "" {
		return errors.New(
			"model name, pattern name, pattern repository, and pattern remote URL are required",
		)
	}

	// --dir is required, so empty means it's not set
	isDirSet := opts.dir != ""

	configDir, configFile := getConfig()

	// only create default config directory if --dir is not set
	if !isDirSet {
		// create config directory if not exists
		if err := fs.MkdirAll(configDir, 0o755); err != nil {
			return fmt.Errorf("creating config directory: %w", err)
		}
	}

	log.Info("setting up configuration",
		"model", i.modelName,
		"pattern", i.patternName,
		"pattern_repo", i.patternRepo,
		"pattern_remote", i.patternRemote,
	)

	viper.SetConfigType("yaml")

	// Set values
	viper.Set("model.name", i.modelName)
	viper.Set("pattern.name", i.patternName)
	viper.Set("pattern.repo", i.patternRepo)
	viper.Set("pattern.remote", i.patternRemote)

	// if --dir is set, use it as the base for the config file path
	if isDirSet {
		configFile = filepath.Join(opts.dir.String(), "seaq.yaml")
	}

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

// createDefaultPatternRepo creates the default pattern directory.
// It will create directories for built-in patterns only if:
//  1. i.patternRepo is the default config directory
//  2. i.patternRepo doesn't already exist
//
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
func collect(opts setupOptions) (setupInput, error) {
	var input setupInput

	_, configFile := getConfig()
	// --dir is required, so not empty means it's set
	if opts.dir != "" {
		configFile = filepath.Join(opts.dir.String(), "seaq.yaml")
	}
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
				Description(
					fmt.Sprintf(
						"The folder containing patterns.\n"+
							"Leave empty to use %s.",
						defaultPatternRepo,
					),
				).
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
				// This OptionsFunc will be re-evaluated when the binding of the OptionsFunc changes.
				OptionsFunc(func() []huh.Option[string] {
					useDefault := input.patternRepo == ""

					if useDefault {
						input.patternRepo = defaultPatternRepo
					}

					if exists, err := fs.DirExists(input.patternRepo); err != nil {
						return nil
					} else if !exists && useDefault {
						return huh.NewOptions("prime_mind", "improve_prompt")
					}

					patterns, err := config.ListPatternsInRepo(input.patternRepo)
					if err != nil {
						return nil
					}
					return huh.NewOptions(patterns...)
				}, nil).
				Value(&input.patternName),
		),
		huh.NewGroup(
			huh.NewInput().
				Title("Remote pattern repository").
				Description(
					fmt.Sprintf(
						"The remote GitHub repository containing patterns to download.\n"+
							"Leave empty to use %s.",
						DefaultRemotePatternsRepo,
					),
				).
				Placeholder(DefaultRemotePatternsRepo).
				Value(&input.patternRemote),
		),
	)

	if err := form.Run(); err != nil {
		return setupInput{}, err
	}

	if input.patternRemote == "" {
		input.patternRemote = DefaultRemotePatternsRepo
	}

	return input, nil
}
