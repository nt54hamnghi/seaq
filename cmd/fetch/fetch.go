/*
Copyright © 2024 Nghi Nguyen
*/
package fetch

import (
	"github.com/nt54hamnghi/seaq/cmd/flaggroup"
	"github.com/spf13/cobra"
)

type fetchGlobalOptions struct {
	output      flaggroup.Output
	asJSON      bool
	ignoreCache bool
}

func (opts *fetchGlobalOptions) Init(cmd *cobra.Command) {
	cmd.Flags().BoolVarP(&opts.asJSON, "json", "j", false, "output as JSON")
	cmd.Flags().BoolVar(&opts.ignoreCache, "no-cache", false, "ignore cache")
	opts.output.Init(cmd)
}

func (opts *fetchGlobalOptions) Validate(cmd *cobra.Command, args []string) error {
	return opts.output.Validate(cmd, args)
}

func NewFetchCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "fetch",
		Short:        "Fetch data and output text",
		Aliases:      []string{"fet"},
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
