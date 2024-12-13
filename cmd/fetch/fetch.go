/*
Copyright © 2024 Nghi Nguyen
*/
package fetch

import (
	"github.com/nt54hamnghi/hiku/cmd/flaggroup"
	"github.com/spf13/cobra"
)

var (
	output   flaggroup.Output
	interval flaggroup.Interval
	asJSON   bool
)

// FetchCmd represents the scrape command
var FetchCmd = &cobra.Command{
	Use:          "fetch",
	Short:        "Fetch data and output text",
	Long:         ``,
	Aliases:      []string{"fet", "f"},
	Args:         cobra.NoArgs,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error { // nolint: revive
		return cmd.Help()
	},
}

func init() {
	FetchCmd.AddCommand(pageCmd)
	FetchCmd.AddCommand(udemyCmd)
	FetchCmd.AddCommand(xCmd)
	FetchCmd.AddCommand(youtubeCmd)
}
