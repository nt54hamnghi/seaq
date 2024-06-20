/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package pattern

import (
	"github.com/nt54hamnghi/hiku/cmd/config"
	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:          "list",
	Short:        "List all available patterns",
	Aliases:      []string{"ls"},
	Args:         cobra.NoArgs,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		pats, err := config.Hiku.GetAvailablePatterns()
		if err != nil {
			return err
		}

		for _, p := range pats {
			cmd.Println(p)
		}

		return nil
	},
}

func init() {}
