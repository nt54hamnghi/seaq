package loader

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/tmc/langchaingo/documentloaders"
)

// LoadAndJoin loads documents using the Loader and joins their content into a single string.
func LoadAndJoin(ctx context.Context, l documentloaders.Loader) (string, error) {
	docs, err := l.Load(ctx)
	if err != nil {
		return "", err
	}

	var buidler strings.Builder
	for i, doc := range docs {
		if i > 0 {
			buidler.WriteString("\n")
		}
		buidler.WriteString(doc.PageContent)
	}
	return buidler.String(), nil
}

// LoadAndMarshal loads documents using the Loader and marshals them into a JSON string.
func LoadAndMarshal(ctx context.Context, l documentloaders.Loader) (string, error) {
	docs, err := l.Load(ctx)
	if err != nil {
		return "", err
	}

	data, err := json.MarshalIndent(docs, "", "\t")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// LoadAndWrite loads documents using the Loader and writes them to the writer.
// If asJson is true, the documents are marshaled into a JSON string before writing.
// Otherwise, the documents are joined into a single string using "\n" as the separator before writing.
func LoadAndWrite(ctx context.Context, l documentloaders.Loader, writer io.Writer, asJSON bool) error {
	var (
		docs string
		err  error
	)

	if asJSON {
		docs, err = LoadAndMarshal(ctx, l)
	} else {
		docs, err = LoadAndJoin(ctx, l)
	}

	if err != nil {
		return err
	}

	_, err = fmt.Fprintln(writer, docs)
	return err
}
