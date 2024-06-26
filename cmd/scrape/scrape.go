/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package scrape

import (
	"github.com/spf13/cobra"
)

var outputFile string

// ScrapeCmd represents the scrape command
var ScrapeCmd = &cobra.Command{
	Use:          "scrape",
	Short:        "Scrape web data and output text",
	Long:         ``,
	Aliases:      []string{"scr", "s"},
	Args:         cobra.NoArgs,
	SilenceUsage: true,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {
	ScrapeCmd.AddCommand(captionCmd)
	ScrapeCmd.AddCommand(pageCmd)

	// flags
	// persistent flags
	ScrapeCmd.PersistentFlags().StringVarP(&outputFile, "output", "o", "", "output file")
}
