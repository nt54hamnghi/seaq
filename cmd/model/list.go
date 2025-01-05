/*
Copyright © 2024 Nghi Nguyen
*/
package model

import (
	"fmt"
	"maps"
	"slices"
	"text/tabwriter"

	"github.com/nt54hamnghi/seaq/pkg/llm"
	"github.com/spf13/cobra"
)

func newListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "list",
		Short:        "List available models",
		Aliases:      []string{"ls"},
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
			defer w.Flush()

			// print headers
			fmt.Fprintf(w, "NAME\tPROVIDER\n")

			providers := slices.Collect(maps.Keys(llm.Models))
			slices.Sort(providers)

			for _, p := range providers {
				models := slices.Collect(maps.Keys(llm.Models[p]))
				slices.Sort(models)

				for _, name := range models {
					fmt.Fprintf(w, "%s\t%s\n", name, p)
				}
			}

			return nil
		},
	}

	return cmd
}
