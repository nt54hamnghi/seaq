/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package pattern

import (
	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:     "list",
	Short:   "List all available patterns",
	Aliases: []string{"ls"},
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := WithName("")
		if err != nil {
			return err
		}

		pats, err := cfg.GetAvailablePatterns()
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
