/*
Copyright © 2024 Nghi Nguyen
*/
package pattern

import (
	"fmt"

	"github.com/nt54hamnghi/seaq/cmd/config"
	"github.com/spf13/cobra"
)

// useCmd represents the use command
var useCmd = &cobra.Command{
	Use:               "use",
	Short:             "Set a default pattern to use",
	Args:              cobra.ExactArgs(1),
	SilenceUsage:      true,
	ValidArgsFunction: CompletePatternArgs,
	RunE: func(cmd *cobra.Command, args []string) error { // nolint: revive
		name := args[0]

		if err := config.Seaq.UsePattern(name); err != nil {
			return err
		}

		if err := config.Seaq.WriteConfig(); err != nil {
			return err
		}

		fmt.Printf("Successfully set the default pattern to '%s'\n", name)
		return nil
	},
}
