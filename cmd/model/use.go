/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package model

import (
	"fmt"

	"github.com/spf13/cobra"
)

// useCmd represents the use command
var useCmd = &cobra.Command{
	Use:          "use",
	Short:        "Set a default model to use",
	Args:         cobra.ExactArgs(1),
	SilenceUsage: true,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("model use called")
	},
}

func init() {}
