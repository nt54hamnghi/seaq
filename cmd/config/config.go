/*
Copyright © 2024 Nghi Nguyen
*/

package config

import (
	"fmt"
	"os"
	"sync"

	"github.com/nt54hamnghi/seaq/pkg/config"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

var fs = afero.Afero{
	Fs: afero.NewOsFs(),
}

func NewConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "config",
		Short:        "Manage config file",
		Aliases:      []string{"cfg"},
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		GroupID:      "management",
		RunE: func(cmd *cobra.Command, args []string) error { // nolint: revive
			return cmd.Usage()
		},
	}

	cmd.AddCommand(
		newSetupCmd(),
		newPathCmd(),
		newShowCmd(),
	)

	return cmd
}

var (
	once       sync.Once
	configDir  string
	configFile string
)

// getConfig returns the application's config directory and file paths.
// Essentially (on Unix-like systems):
//   - $HOME/.config/seaq
//   - $HOME/.config/seaq/seaq.yaml
//
// It doesn't check if the directory or file exists.
//
// Since these values don't change throughout the lifetime of the config command,
// the first call to this function will "initialize" them. Subsequent calls
// will return the initialized values.
//
// Failure to initialize the config will cause the program to exit.
func getConfig() (string, string) {
	once.Do(func() {
		var err error
		configDir, configFile, err = config.AppConfig()
		if err != nil {
			fmt.Println("Error: failed to get config:", err)
			os.Exit(1)
		}
	})
	return configDir, configFile
}
