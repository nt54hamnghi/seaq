package config

import (
	"github.com/nt54hamnghi/hiku/pkg/llm"
)

func (hiku *HikuConfig) Model() string {
	return hiku.GetString("model.name")
}

func (hiku *HikuConfig) HasModel(name string) bool {
	_, _, ok := llm.LookupModel(name)
	return ok
}

func (hiku *HikuConfig) UseModel(name string) error {
	if !hiku.HasModel(name) {
		return &Unsupported{Type: "model", Key: name}
	}
	hiku.Set("model.name", name)
	return nil
}
