package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

var Hiku = New()

// HikuConfig is a slim wrapper around an instance of viper.Viper
type HikuConfig struct {
	*viper.Viper
}

func (hiku *HikuConfig) SearchConfigFile() error {
	// Find home directory.
	config, err := os.UserConfigDir()
	if err != nil {
		return err
	}
	hikuConfig := filepath.Join(config, "hiku")

	// Set config file name and type
	hiku.SetConfigName("hiku")
	hiku.SetConfigType("yaml")

	// Path to look for the config file in
	// The order of paths listed is the order in which they will be searched
	hiku.AddConfigPath("/etc/hiku")
	hiku.AddConfigPath(hikuConfig)
	hiku.AddConfigPath(".")

	return nil
}

type Unsupported struct {
	Type string
	Key  string
}

func (e *Unsupported) Error() string {
	return fmt.Sprintf("unsupported %s: '%s'", e.Type, e.Key)
}
