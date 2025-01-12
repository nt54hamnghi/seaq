package config

import (
	"github.com/nt54hamnghi/seaq/pkg/llm"
)

func (sc *SeaqConfig) Model() string {
	return sc.GetString("model.name")
}

func (sc *SeaqConfig) UseModel(name string) error {
	if !llm.HasModel(name) {
		return &Unsupported{Type: "model", Key: name}
	}
	sc.Set("model.name", name)
	return nil
}
