/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package pattern

import (
	"fmt"

	"github.com/nt54hamnghi/hiku/cmd/config"
	"github.com/spf13/cobra"
)

// useCmd represents the use command
var useCmd = &cobra.Command{
	Use:               "use",
	Short:             "Set a default pattern to use",
	Args:              cobra.ExactArgs(1),
	SilenceUsage:      true,
	ValidArgsFunction: CompletePatternArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		if err := config.Hiku.UsePattern(name); err != nil {
			return err
		}
		config.Hiku.WriteConfig()
		fmt.Printf("Successfully set the default pattern to '%s'\n", name)
		return nil
	},
}
