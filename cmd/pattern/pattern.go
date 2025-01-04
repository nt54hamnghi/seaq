/*
Copyright © 2024 Nghi Nguyen
*/
package pattern

import (
	"github.com/nt54hamnghi/seaq/cmd/config"
	"github.com/spf13/cobra"
)

func NewPatternCmd() *cobra.Command {
	// patternCmd represents the pattern command
	cmd := &cobra.Command{
		Use:          "pattern",
		Short:        "Manage patterns",
		Aliases:      []string{"pat", "p"},
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		GroupID:      "management",
		RunE: func(cmd *cobra.Command, args []string) error { // nolint: revive
			return cmd.Usage()
		},
	}

	cmd.AddCommand(
		newUseCmd(),
		newListCmd(),
	)

	return cmd
}

// nolint: revive
func CompletePatternArgs(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	patterns, err := config.Seaq.ListPatterns()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	return patterns, cobra.ShellCompDirectiveNoFileComp
}
