/*
Copyright © 2024 Nghi Nguyen
*/
package pattern

import (
	"github.com/nt54hamnghi/seaq/cmd/config"
	"github.com/spf13/cobra"
)

// patternCmd represents the pattern command
var PatternCmd = &cobra.Command{
	Use:          "pattern",
	Short:        "Manage patterns",
	Aliases:      []string{"pat", "p"},
	Args:         cobra.NoArgs,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error { // nolint: revive
		return cmd.Usage()
	},
}

func init() {
	PatternCmd.AddCommand(
		useCmd,
		listCmd,
	)
}

func CompletePatternArgs(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) { // nolint: revive
	patterns, err := config.Seaq.ListPatterns()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	return patterns, cobra.ShellCompDirectiveNoFileComp
}
