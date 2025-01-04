/*
Copyright © 2024 Nghi Nguyen
*/
package model

import (
	"github.com/nt54hamnghi/seaq/cmd/config"
	"github.com/spf13/cobra"
)

func NewModelCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "model",
		Short:        "Manage models",
		Aliases:      []string{"mdl", "m"},
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		GroupID:      "management",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Usage()
		},
	}

	cmd.AddCommand(
		newUseCmd(),
		newListCmd(),
		newViewCmd(),
	)

	return cmd
}

// nolint: revive
func CompleteModelArgs(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return config.Seaq.ListModels(), cobra.ShellCompDirectiveNoFileComp
}
