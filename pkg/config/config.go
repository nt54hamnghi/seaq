package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var Seaq = New()

// SeaqConfig is a slim wrapper around an instance of viper.Viper
type SeaqConfig struct {
	*viper.Viper
}

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
	if err := Seaq.EnsureConfig(cmd, args); err != nil {
		return err
	}

	flags := cmd.Flags()
	for flag, key := range flagBindings {
		if f := flags.Lookup(flag); f != nil {
			// https://github.com/spf13/viper#working-with-flags
			// the config value is not set at binding time but at access time
			if err := Seaq.BindPFlag(key, f); err != nil {
				return fmt.Errorf("binding flag %s to key %s: %w", flag, key, err)
			}
		}
	}

	return nil
}

// EnsureConfig ensures that the config file is loaded.
//
// It will reads the config flags from the provided command to load the config file.
// If this flag is not available (the command doesn't have it or it's not set), it will search for the config file.
func (sc *SeaqConfig) EnsureConfig(cmd *cobra.Command, args []string) error { //nolint:revive
	var configFile string
	if f := cmd.Flags().Lookup("config"); f != nil {
		configFile = f.Value.String()
	}
	// UseConfigFile will search for the config file if the path is empty
	if err := sc.UseConfigFile(configFile); err != nil {
		return err
	}

	sc.AutomaticEnv()
	if err := sc.ReadInConfig(); err != nil {
		return fmt.Errorf("reading config: %w", err)
	}

	return nil
}

// UseConfigFile sets the config file to use.
//
// If the path is empty, it will search for the config file.
func (sc *SeaqConfig) UseConfigFile(path string) error {
	if path == "" {
		return sc.SearchConfigFile()
	}

	// use Viper to avoid recursive calls
	sc.SetConfigFile(path)
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
func (sc *SeaqConfig) SearchConfigFile() error {
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
	sc.SetConfigName("seaq")
	sc.SetConfigType("yaml")

	// paths to search for the config file
	// paths are searched in the order they are added
	sc.AddConfigPath(curDir)
	sc.AddConfigPath(appDir)
	if runtime.GOOS == "linux" {
		sc.AddConfigPath("/etc/seaq")
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
