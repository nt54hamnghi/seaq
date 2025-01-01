package config

import (
	"github.com/nt54hamnghi/seaq/pkg/llm"
)

func (sc *SeaqConfig) Model() string {
	return sc.GetString("model.name")
}

func (sc *SeaqConfig) HasModel(name string) bool {
	_, _, ok := llm.LookupModel(name)
	return ok
}

func (sc *SeaqConfig) UseModel(name string) error {
	if !sc.HasModel(name) {
		return &Unsupported{Type: "model", Key: name}
	}
	sc.Set("model.name", name)
	return nil
}

// ListModels returns a list of available models
func (sc *SeaqConfig) ListModels() []string {
	models := make([]string, 0, len(llm.Models))
	for _, v := range llm.Models {
		for m := range v {
			models = append(models, m)
		}
	}
	return models
}
