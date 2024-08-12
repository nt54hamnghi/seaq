/*
Copyright © 2024 Nghi Nguyen
*/
package pattern

import (
	"github.com/nt54hamnghi/hiku/cmd/config"
	"github.com/spf13/cobra"
)

// patternCmd represents the pattern command
var PatternCmd = &cobra.Command{
	Use:          "pattern",
	Short:        "Manage patterns",
	Aliases:      []string{"pat", "p"},
	Args:         cobra.NoArgs,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Usage()
	},
}

func init() {
	PatternCmd.AddCommand(
		useCmd,
		listCmd,
	)
}

func CompletePatternArgs(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	patterns, err := config.Hiku.ListPatterns()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	return patterns, cobra.ShellCompDirectiveNoFileComp
}
