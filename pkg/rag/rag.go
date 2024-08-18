package rag

import (
	chroma_go "github.com/amikos-tech/chroma-go/types"
	"github.com/nt54hamnghi/hiku/pkg/env"
	"github.com/tmc/langchaingo/vectorstores/chroma"
)

func NewChromaStore() (*chroma.Store, error) {
	chromaUrl, err := env.ChromaURL()
	if err != nil {
		return nil, err
	}

	// TODO: consider other embedding models
	openAiApiKey, err := env.OpenAIAPIKey()
	if err != nil {
		return nil, err
	}

	// create a new Chroma store
	store, err := chroma.New(
		chroma.WithChromaURL(chromaUrl),
		chroma.WithOpenAIAPIKey(openAiApiKey),
		chroma.WithDistanceFunction(chroma_go.COSINE),
		chroma.WithNameSpace("hiku-dev"), // TODO: use UUID
	)

	if err != nil {
		return nil, err
	}

	return &store, nil
}
