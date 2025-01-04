/*
Copyright © 2024 Nghi Nguyen
*/
package model

import (
	"maps"
	"slices"

	"github.com/nt54hamnghi/seaq/pkg/llm"
	"github.com/spf13/cobra"
)

func newListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "list",
		Short:        "List all available models",
		Aliases:      []string{"ls"},
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			providers := slices.Collect(maps.Keys(llm.Models))
			slices.Sort(providers)

			for _, p := range providers {
				cmd.Println(p)
				cmd.Println("--------------------")

				models := slices.Collect(maps.Keys(llm.Models[p]))
				slices.Sort(models)

				for _, name := range models {
					cmd.Println(name)
				}
				cmd.Println()
			}

			return nil
		},
	}

	return cmd
}
