/*
Copyright © 2024 Nghi Nguyen
*/
package model

import (
	"errors"
	"fmt"
	"text/tabwriter"

	"github.com/nt54hamnghi/seaq/cmd/config"
	"github.com/nt54hamnghi/seaq/pkg/llm"
	"github.com/spf13/cobra"
)

func newViewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "view",
		Short:        "View the current default model",
		Aliases:      []string{"v"},
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if provider, name, ok := llm.LookupModel(config.Seaq.Model()); ok {
				w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
				defer w.Flush()
				fmt.Fprintf(w, "Model:\t%s\n", name)
				fmt.Fprintf(w, "Provider:\t%s\n", provider)
				return nil
			}
			return errors.New("unexpected error: failed to get default model")
		},
	}

	return cmd
}
