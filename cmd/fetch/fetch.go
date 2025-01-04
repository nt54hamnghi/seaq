/*
Copyright © 2024 Nghi Nguyen
*/
package fetch

import (
	"github.com/spf13/cobra"
)

func NewFetchCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "fetch",
		Short:        "Fetch data and output text",
		Aliases:      []string{"fet", "f"},
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		GroupID:      "common",
		RunE: func(cmd *cobra.Command, args []string) error { // nolint: revive
			return cmd.Usage()
		},
	}

	cmd.AddCommand(
		newPageCmd(),
		newUdemyCmd(),
		newXCmd(),
		newYoutubeCmd(),
	)

	return cmd
}
