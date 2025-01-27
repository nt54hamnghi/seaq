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

// flagBindings maps CLI flags to their corresponding config keys
var flagBindings = map[string]string{
	"pattern": "pattern.name",
	"repo":    "pattern.repo",
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
// It will reads the config flags from the provided command to load the config file.
// If this flag is not available (the command doesn't have it or it's not set), it will search for the config file.
func EnsureConfig(cmd *cobra.Command, args []string) error { //nolint:revive
	var configFile string
	if f := cmd.Flags().Lookup("config"); f != nil {
		configFile = f.Value.String()
	}
	// UseConfigFile will search for the config file if the path is empty
	if err := UseConfigFile(configFile); err != nil {
		return err
	}

	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("reading config: %w.\n%s", err, "Run `seaq config setup` to generate a new config file")
	}

	return nil
}

// UseConfigFile sets the config file to use.
//
// If the path is empty, it will search for the config file.
func UseConfigFile(path string) error {
	if path == "" {
		return SearchConfigFile()
	}

	// use Viper to avoid recursive calls
	viper.SetConfigFile(path)
	return nil
}

// SearchConfigFile searches for the config file to use.
//
// The search order is:
//  1. Current working directory ($PWD/seaq.yaml on Unix)
//  2. User config directory ($XDG_CONFIG_HOME/seaq/seaq.yaml on Unix)
//  3. /etc/seaq (Linux only)
//
// If no config file is found, it will return an error.
func SearchConfigFile() error {
	// get the current working directory
	curDir, err := os.Getwd()
	if err != nil {
		return err
	}

	// find home directory and get the app directory
	configDir, err := os.UserConfigDir()
	if err != nil {
		return err
	}
	appDir := filepath.Join(configDir, "seaq")

	// set config file name and type
	viper.SetConfigName("seaq")
	viper.SetConfigType("yaml")

	// paths to search for the config file
	// paths are searched in the order they are added
	viper.AddConfigPath(curDir)
	viper.AddConfigPath(appDir)
	if runtime.GOOS == "linux" {
		viper.AddConfigPath("/etc/seaq")
	}

	return nil
}

type Unsupported struct {
	Type string
	Key  string
}

func (e *Unsupported) Error() string {
	return fmt.Sprintf("unsupported %s: '%s'", e.Type, e.Key)
}
