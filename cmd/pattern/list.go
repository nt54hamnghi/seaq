/*
Copyright © 2024 Nghi Nguyen
*/
package pattern

import (
	"github.com/nt54hamnghi/hiku/cmd/config"
	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:          "list",
	Short:        "List all available patterns",
	Aliases:      []string{"ls"},
	Args:         cobra.NoArgs,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error { // nolint: revive
		pats, err := config.Hiku.ListPatterns()
		if err != nil {
			return err
		}

		for _, p := range pats {
			cmd.Println(p)
		}

		return nil
	},
}
