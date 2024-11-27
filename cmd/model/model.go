/*
Copyright © 2024 Nghi Nguyen
*/
package model

import (
	"github.com/nt54hamnghi/hiku/cmd/config"
	"github.com/spf13/cobra"
)

// modelCmd represents the model command
var ModelCmd = &cobra.Command{
	Use:          "model",
	Short:        "Manage models",
	Aliases:      []string{"mdl", "m"},
	Args:         cobra.NoArgs,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error { // nolint: revive
		return cmd.Usage()
	},
}

func init() {
	ModelCmd.AddCommand(
		useCmd,
		listCmd,
	)
}

func CompleteModelArgs(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) { // nolint: revive
	return config.Hiku.ListModels(), cobra.ShellCompDirectiveNoFileComp
}
