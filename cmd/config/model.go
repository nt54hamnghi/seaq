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

// ListModels returns a list of available models
func (hiku *HikuConfig) ListModels() []string {
	models := make([]string, 0, len(llm.Models))
	for _, v := range llm.Models {
		for m := range v {
			models = append(models, m)
		}
	}
	return models
}
