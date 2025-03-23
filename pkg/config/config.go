package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const App = "seaq"

// Complete configuration in Go struct
//
// ```go
//	type SeaqConfig struct {
//		Connections []struct {
//			BaseURL  string `yaml:"base_url"`
//			Provider string `yaml:"provider"`
//		} `yaml:"connections"`
//		Model struct {
//			Name string `yaml:"name"`
//		} `yaml:"model"`
//		Pattern struct {
//			Name   string `yaml:"name"`
//			Repo   string `yaml:"repo"`
//			Remote string `yaml:"remote"`
//		} `yaml:"pattern"`
//	}
// ```
//
// Example in yaml format
//
// ```yaml
// connections:
//   - base_url: https://api.groq.com/openai/v1
//     provider: groq
//   - base_url: https://openrouter.ai/api/v1
//     provider: openrouter
// model:
//   name: anthropic/claude-3-5-sonnet-latest
// pattern:
//   name: take_note
//   repo: /home/user/.config/seaq/patterns
//   remote: https://github.com/danielmiessler/fabric
// ```

// flagBindings maps CLI flags to their corresponding config keys
var flagBindings = map[string]string{
	"pattern": "pattern.name",
	"repo":    "pattern.repo",
	"remote":  "pattern.remote",
	"model":   "model.name",
}

// Init loads the config file and binds flags to their corresponding config keys.
//
// It should be used as a PreRunE for commands that require configuration.
func Init(cmd *cobra.Command, args []string) error { //nolint:revive
	if err := EnsureConfig(cmd, args); err != nil {
		return err
	}

	flags := cmd.Flags()
	for flag, key := range flagBindings {
		if f := flags.Lookup(flag); f != nil {
			// https://github.com/spf13/viper#working-with-flags
			// the config value is not set at binding time but at access time
			if err := viper.BindPFlag(key, f); err != nil {
				return fmt.Errorf("binding flag %s to key %s: %w", flag, key, err)
			}
		}
	}

	return nil
}

// AddConfigFlag adds the standard config file flag to a command
func AddConfigFlag(cmd *cobra.Command, configFile pflag.Value) {
	cmd.Flags().VarP(configFile, "config", "c", "config file (default is $HOME/.config/seaq.yaml)")
}

// EnsureConfig ensures that the config file is loaded.
//
// It will reads the config file path from the "config" flag of the provided command.
// If this flag is not available (the command doesn't have it or it's not set), it will search for the config file.
func EnsureConfig(cmd *cobra.Command, args []string) error { //nolint:revive
	var configFile string
	if f := cmd.Flags().Lookup("config"); f != nil {
		configFile = f.Value.String()
	}

	// SetConfigFile will set the config file to use
	// or configure search paths if the path is empty
	if err := SetConfigFile(configFile); err != nil {
		return err
	}

	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("reading config: %w.\n%s", err, "Run `seaq config setup` to generate a new config file")
	}

	return nil
}

// SetConfigFile sets up Viper's configuration source:
// - If path is provided, configures Viper to use that specific file
// - If path is empty, sets up default search paths in multiple locations
func SetConfigFile(path string) error {
	if path == "" {
		return SetupSearchPaths()
	}

	// explicitly set the config file, so viper doesn't search for it
	viper.SetConfigFile(path)
	return nil
}

// SetupSearchPaths configures Viper's search paths in the following order:
//  1. Current working directory ($PWD/seaq.yaml on Unix)
//  2. User config directory ($XDG_CONFIG_HOME/seaq/seaq.yaml on Unix)
//  3. /etc/seaq (Linux only)
//
// Note: This function only sets up the search paths. The actual search and loading
// of the config file happens when viper.ReadInConfig() is called.
func SetupSearchPaths() error {
	// get the current working directory
	curDir, err := os.Getwd()
	if err != nil {
		return err
	}

	// find home directory and get the app directory
	appDir, _, err := AppConfig()
	if err != nil {
		return err
	}

	// set config file name and type
	viper.SetConfigName(App)
	viper.SetConfigType("yaml")

	// paths to search for the config file
	// paths are searched in the order they are added
	viper.AddConfigPath(curDir)
	viper.AddConfigPath(appDir)
	if runtime.GOOS == "linux" {
		viper.AddConfigPath("/etc/" + App)
	}

	return nil
}

// AppConfig is a convenience function that returns the config directory and file for the app.
// Conceptually, it returns the equivalent of:
//   - configDir: $HOME/.config/seaq
//   - configFile: $HOME/.config/seaq/seaq.yaml
//
// It doesn't check if the directory or file exists.
func AppConfig() (configDir string, configFile string, err error) {
	configDir, err = os.UserConfigDir()
	if err != nil {
		return
	}
	configDir = filepath.Join(configDir, App)
	configFile = filepath.Join(configDir, App+".yaml")
	return
}

type Unsupported struct {
	Type string
	Key  string
}

func (e *Unsupported) Error() string {
	return fmt.Sprintf("unsupported %s: '%s'", e.Type, e.Key)
}
