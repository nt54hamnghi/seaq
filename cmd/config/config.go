package config

import (
	"fmt"

	"github.com/spf13/viper"
)

var Hiku *HikuConfig

type HikuConfig struct {
	*viper.Viper
}

type Unsupported struct {
	Type string
	Key  string
}

func (e *Unsupported) Error() string {
	return fmt.Sprintf("unsupported %s: '%s'", e.Type, e.Key)
}
