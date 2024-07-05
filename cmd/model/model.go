/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package model

import (
	"fmt"

	"github.com/nt54hamnghi/hiku/cmd/config"
	"github.com/spf13/cobra"
)

// modelCmd represents the model command
var ModelCmd = &cobra.Command{
	Use:          "model",
	Short:        "Manage models",
	Aliases:      []string{"mdl"},
	Args:         cobra.NoArgs,
	SilenceUsage: true,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("model called")
	},
}

func init() {
	ModelCmd.AddCommand(
		useCmd,
		listCmd,
	)
}

func CompleteModelArgs(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return config.Hiku.ListModels(), cobra.ShellCompDirectiveNoFileComp
}
