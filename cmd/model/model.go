/*
Copyright © 2024 Nghi Nguyen
*/
package model

import (
	"slices"

	"github.com/nt54hamnghi/seaq/pkg/llm"
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
		newGetCmd(),
		newListCmd(),
		newSetCmd(),
	)

	return cmd
}

// nolint: revive
func CompleteModelArgs(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return listModels(), cobra.ShellCompDirectiveNoFileComp
}

func listModels() []string {
	models := slices.Collect(llm.Models())
	slices.Sort(models)
	return models
}
