package llm

import (
	"context"
	"iter"
	"net/http"
	"net/url"
	"slices"

	"github.com/nt54hamnghi/seaq/pkg/env"
	"github.com/ollama/ollama/api"
)

func listOllamaModels() ([]string, error) {
	// create a new client
	hostURL, err := url.ParseRequestURI(env.OllamaHost())
	if err != nil {
		return nil, err
	}
	client := api.NewClient(hostURL, http.DefaultClient)

	// list all models
	models, err := client.List(context.Background())
	if err != nil {
		return nil, err
	}

	return slices.Collect(toModelNames(models)), nil
}

func toModelNames(r *api.ListResponse) iter.Seq[string] {
	return func(yield func(string) bool) {
		for _, m := range r.Models {
			if !yield(m.Name) {
				return
			}
		}
	}
}
