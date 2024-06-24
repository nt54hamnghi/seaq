package config

import "github.com/nt54hamnghi/hiku/pkg/openai"

func (hiku *HikuConfig) Model() string {
	return hiku.GetString("model.name")
}

func (hiku *HikuConfig) GetAvailableModels() ([]string, error) {
	models := openai.Models()
	return models[:], nil
}
