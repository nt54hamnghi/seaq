/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package fetch

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tmc/langchaingo/documentloaders"
)

var outputFile string

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

	// flags
	// persistent flags
	FetchCmd.PersistentFlags().StringVarP(&outputFile, "output", "o", "", "output file")
}

func fetch(ctx context.Context, l documentloaders.Loader) (string, error) {
	docs, err := l.Load(ctx)
	if err != nil {
		return "", err
	}

	content := ""
	for _, doc := range docs {
		content += fmt.Sprintf("%s\n", doc.PageContent)
	}

	return content, nil
}
