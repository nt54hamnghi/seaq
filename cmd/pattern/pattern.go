/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
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
	Aliases:      []string{"pat"},
	Args:         cobra.NoArgs,
	SilenceUsage: true,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
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
