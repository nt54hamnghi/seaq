/*
Copyright © 2024 Nghi Nguyen
*/
package pattern

import (
	"github.com/nt54hamnghi/seaq/pkg/config"
	"github.com/spf13/cobra"
)

func newSetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "set",
		Short:             "Set the default pattern",
		Aliases:           []string{"s", "use"},
		Args:              cobra.ExactArgs(1),
		SilenceUsage:      true,
		ValidArgsFunction: CompletePatternArgs,
		PreRunE:           config.Init,
		RunE: func(cmd *cobra.Command, args []string) error { // nolint: revive
			name := args[0]

			if err := config.Seaq.UsePattern(name); err != nil {
				return err
			}

			if err := config.Seaq.WriteConfig(); err != nil {
				return err
			}

			cmd.Printf("Successfully set default pattern to '%s'\n", name)
			return nil
		},
	}

	return cmd
}
