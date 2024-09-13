/*
Copyright © 2024 Nghi Nguyen
*/
package fetch

import (
	"github.com/nt54hamnghi/hiku/cmd/flagGroup"
	"github.com/spf13/cobra"
)

var (
	output flagGroup.Output
	asJson bool
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
	FetchCmd.AddCommand(pageCmd)
	FetchCmd.AddCommand(xCmd)
	FetchCmd.AddCommand(youtubeCmd)
}
