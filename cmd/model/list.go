/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package model

import (
	"github.com/nt54hamnghi/hiku/cmd/config"
	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:          "list",
	Short:        "List all available models",
	Aliases:      []string{"ls"},
	Args:         cobra.NoArgs,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		models, err := config.Hiku.GetAvailableModels()
		if err != nil {
			return err
		}

		for _, p := range models {
			cmd.Println(p)
		}

		return nil
	},
}

func init() {}
