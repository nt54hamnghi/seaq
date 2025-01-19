/*
Copyright © 2024 Nghi Nguyen
*/
package model

import (
	"github.com/nt54hamnghi/seaq/cmd/flag"
	"github.com/nt54hamnghi/seaq/pkg/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type setOptions struct {
	configFile flag.FilePath
}

func newSetCmd() *cobra.Command {
	var opts setOptions

	cmd := &cobra.Command{
		Use:               "set",
		Short:             "Set the default model",
		Aliases:           []string{"s", "use"},
		Args:              cobra.ExactArgs(1),
		SilenceUsage:      true,
		ValidArgsFunction: CompleteModelArgs,
		PreRunE:           config.Init,
		RunE: func(cmd *cobra.Command, args []string) error { // nolint: revive
			name := args[0]

			if err := config.UseModel(name); err != nil {
				return err
			}

			if err := viper.WriteConfig(); err != nil {
				return err
			}

			cmd.Printf("Successfully set default model to '%s'\n", name)
			return nil
		},
	}

	config.AddConfigFlag(cmd, &opts.configFile)

	return cmd
}
