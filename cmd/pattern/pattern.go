/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package pattern

import (
	"github.com/spf13/cobra"
)

// patternCmd represents the pattern command
var PatternCmd = &cobra.Command{
	Use:     "pattern",
	Short:   "Manage patterns",
	Aliases: []string{"pat"},
	Args:    cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {
	PatternCmd.AddCommand(useCmd)
	PatternCmd.AddCommand(listCmd)
}
