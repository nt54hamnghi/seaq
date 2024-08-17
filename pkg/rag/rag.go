package rag

import (
	"context"

	chroma_go "github.com/amikos-tech/chroma-go/types"
	"github.com/nt54hamnghi/hiku/pkg/env"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/vectorstores/chroma"
)

type DocumentStore struct {
	*chroma.Store
	Docs []schema.Document
}

func NewStoreWithDocuments(ctx context.Context, docs []schema.Document) (*DocumentStore, error) {
	store, err := NewStore()
	if err != nil {
		return nil, err
	}

	store.Docs = docs
	_, err = store.AddDocuments(ctx, docs)
	if err != nil {
		return nil, err
	}

	return store, nil
}

func NewStore() (*DocumentStore, error) {
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

	return &DocumentStore{Store: &store}, nil
}
