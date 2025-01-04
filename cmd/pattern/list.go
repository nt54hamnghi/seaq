/*
Copyright © 2024 Nghi Nguyen
*/
package pattern

import (
	"github.com/nt54hamnghi/seaq/cmd/config"
	"github.com/spf13/cobra"
)

func newListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "list",
		Short:        "List available patterns",
		Aliases:      []string{"ls"},
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error { // nolint: revive
			pats, err := config.Seaq.ListPatterns()
			if err != nil {
				return err
			}

			for _, p := range pats {
				cmd.Println(p)
			}

			return nil
		},
	}

	return cmd
}
