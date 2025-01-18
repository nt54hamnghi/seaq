/*
Copyright © 2024 Nghi Nguyen
*/
package model

import (
	"github.com/nt54hamnghi/seaq/pkg/config"
	"github.com/spf13/cobra"
)

func newUseCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "use",
		Short:             "Set default model",
		Args:              cobra.ExactArgs(1),
		SilenceUsage:      true,
		ValidArgsFunction: CompleteModelArgs,
		PreRunE:           config.Init,
		RunE: func(cmd *cobra.Command, args []string) error { // nolint: revive
			name := args[0]

			if err := config.Seaq.UseModel(name); err != nil {
				return err
			}

			if err := config.Seaq.WriteConfig(); err != nil {
				return err
			}

			cmd.Printf("Successfully set default model to '%s'\n", name)
			return nil
		},
	}

	return cmd
}
