/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package model

import (
	"fmt"

	"github.com/nt54hamnghi/hiku/cmd/config"
	"github.com/spf13/cobra"
)

// useCmd represents the use command
var useCmd = &cobra.Command{
	Use:          "use",
	Short:        "Set a default model to use",
	Args:         cobra.ExactArgs(1),
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		if err := config.Hiku.UseModel(name); err != nil {
			return err
		}
		config.Hiku.WriteConfig()
		fmt.Printf("Successfully set the default model to '%s'\n", name)
		return nil
	},
}

func init() {}
