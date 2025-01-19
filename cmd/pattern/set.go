/*
Copyright © 2024 Nghi Nguyen
*/
package pattern

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
		Short:             "Set the default pattern",
		Aliases:           []string{"s", "use"},
		Args:              cobra.ExactArgs(1),
		SilenceUsage:      true,
		ValidArgsFunction: CompletePatternArgs,
		PreRunE:           config.Init,
		RunE: func(cmd *cobra.Command, args []string) error { // nolint: revive
			name := args[0]

			if err := config.UsePattern(name); err != nil {
				return err
			}

			if err := viper.WriteConfig(); err != nil {
				return err
			}

			cmd.Printf("Successfully set default pattern to '%s'\n", name)
			return nil
		},
	}

	// set up flags
	config.AddConfigFlag(cmd, &opts.configFile)

	return cmd
}
