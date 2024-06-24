package config

import (
	"github.com/spf13/viper"
)

var Hiku *HikuConfig

type HikuConfig struct {
	*viper.Viper
}
