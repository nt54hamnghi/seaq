/*
Copyright © 2024 Nghi Nguyen
*/
package model

import (
	"github.com/nt54hamnghi/hiku/pkg/llm"
	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:          "list",
	Short:        "List all available models",
	Aliases:      []string{"ls"},
	Args:         cobra.NoArgs,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error { // nolint: revive
		for provider, models := range llm.Models {
			cmd.Println(provider)
			cmd.Println("--------------------")
			for name := range models {
				cmd.Println(name)
			}
			cmd.Println()
		}

		return nil
	},
}
