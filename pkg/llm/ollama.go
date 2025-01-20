package llm

import (
	"context"
	"net/http"
	"net/url"

	"github.com/nt54hamnghi/seaq/pkg/env"
	"github.com/ollama/ollama/api"
)

var ollamaLister = SimpleModelLister{
	ProviderName: "ollama",
	Lister:       listOllamaModels,
}

func listOllamaModels(ctx context.Context) ([]string, error) {
	// create a new client
	hostURL, err := url.ParseRequestURI(env.OllamaHost())
	if err != nil {
		return nil, err
	}
	client := api.NewClient(hostURL, http.DefaultClient)

	res, err := client.List(ctx)
	if err != nil {
		return nil, err
	}

	models := make([]string, len(res.Models))
	for i, model := range res.Models {
		models[i] = model.Name
	}

	return models, nil
}
