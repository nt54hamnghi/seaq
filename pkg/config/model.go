package config

import (
	"github.com/nt54hamnghi/seaq/pkg/llm"
	"github.com/spf13/viper"
)

func Model() string {
	return viper.GetString("model.name")
}

func UseModel(name string) error {
	if !llm.HasModel(name) {
		return &Unsupported{Type: "model", Key: name}
	}
	viper.Set("model.name", name)
	return nil
}
