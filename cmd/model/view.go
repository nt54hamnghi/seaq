/*
Copyright © 2024 Nghi Nguyen
*/
package model

import (
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
				cmd.Println("Model:", name)
				cmd.Println("Provider:", provider)
			}
			return nil
		},
	}

	return cmd
}
