package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

var Seaq = New()

// SeaqConfig is a slim wrapper around an instance of viper.Viper
type SeaqConfig struct {
	*viper.Viper
}

func (sc *SeaqConfig) SearchConfigFile() error {
	// Find home directory.
	config, err := os.UserConfigDir()
	if err != nil {
		return err
	}
	cfgPath := filepath.Join(config, "seaq")

	// Set config file name and type
	sc.SetConfigName("seaq")
	sc.SetConfigType("yaml")

	// Path to look for the config file in
	// The order of paths listed is the order in which they will be searched
	sc.AddConfigPath("/etc/seaq")
	sc.AddConfigPath(cfgPath)
	sc.AddConfigPath(".")

	return nil
}

type Unsupported struct {
	Type string
	Key  string
}

func (e *Unsupported) Error() string {
	return fmt.Sprintf("unsupported %s: '%s'", e.Type, e.Key)
}
