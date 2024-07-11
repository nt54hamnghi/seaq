/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package fetch

import (
	"github.com/spf13/cobra"
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
