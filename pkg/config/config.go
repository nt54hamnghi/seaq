package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/spf13/viper"
)

var Seaq = New()

// SeaqConfig is a slim wrapper around an instance of viper.Viper
type SeaqConfig struct {
	*viper.Viper
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

type Unsupported struct {
	Type string
	Key  string
}

func (e *Unsupported) Error() string {
	return fmt.Sprintf("unsupported %s: '%s'", e.Type, e.Key)
}
