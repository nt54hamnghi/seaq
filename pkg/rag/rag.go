package rag

import (
	chroma_go "github.com/amikos-tech/chroma-go/types"
	"github.com/nt54hamnghi/hiku/pkg/env"
	"github.com/tmc/langchaingo/vectorstores/chroma"
)

func NewChromaStore() (*chroma.Store, error) {
	chromaURL, err := env.ChromaURL()
	if err != nil {
		return nil, err
	}

	// TODO: consider other embedding models
	apiKey, err := env.OpenAIAPIKey()
	if err != nil {
		return nil, err
	}

	// create a new Chroma store
	store, err := chroma.New(
		chroma.WithChromaURL(chromaURL),
		chroma.WithOpenAIAPIKey(apiKey),
		chroma.WithDistanceFunction(chroma_go.COSINE),
		chroma.WithNameSpace("hiku-dev"), // TODO: use UUID
	)
	if err != nil {
		return nil, err
	}

	return &store, nil
}
