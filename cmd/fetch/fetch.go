/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package fetch

import (
	"context"
	"errors"
	"io"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tmc/langchaingo/documentloaders"
)

var (
	outputFile string
	force      bool
)

// FetchCmd represents the scrape command
var FetchCmd = &cobra.Command{
	Use:          "fetch",
	Short:        "Fetch data and output text",
	Long:         ``,
	Aliases:      []string{"fet", "f"},
	Args:         cobra.NoArgs,
	SilenceUsage: true,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {
	FetchCmd.AddCommand(captionCmd)
	FetchCmd.AddCommand(pageCmd)

	// persistent flags
	FetchCmd.PersistentFlags().StringVarP(&outputFile, "output", "o", "", "output file")
	FetchCmd.PersistentFlags().BoolVarP(&force, "force", "f", false, "overwrite output file if it exists")
}

func validatePersistentFlags(cmd *cobra.Command, _ []string) error {
	outputFileSet := cmd.Flags().Changed("output")
	forceSet := cmd.Flags().Changed("force")

	if forceSet && (!outputFileSet || outputFile == "") {
		return errors.New("--force can only be used with --output")
	}

	return nil
}

// fetch loads documents using the Loader and concatenates their content.
// It separates documents with "\n\n---\n\n".
// Returns the concatenated string or an error if loading fails.
func fetch(ctx context.Context, l documentloaders.Loader) (string, error) {
	docs, err := l.Load(ctx)
	if err != nil {
		return "", err
	}

	sep := "\n\n---\n\n"
	var builder strings.Builder
	for i, doc := range docs {
		// to avoid adding a trailing separator
		if i > 0 {
			builder.WriteString(sep)
		}
		builder.WriteString(doc.PageContent)
	}

	return builder.String(), nil
}

// fetchAndWrite loads documents using the Loader and writes their content
// to the specified io.Writer. Documents are separated by "\n\n---\n\n".
// Returns an error if loading or writing fails.
func fetchAndWrite(ctx context.Context, l documentloaders.Loader, writer io.Writer) error {
	docs, err := l.Load(ctx)

	if err != nil {
		return err
	}

	sep := "\n\n---\n\n"
	for i, doc := range docs {
		if i > 0 {
			if _, err := io.WriteString(writer, sep); err != nil {
				return err
			}
		}
		if _, err := io.WriteString(writer, doc.PageContent); err != nil {
			return err
		}
	}

	return nil
}
