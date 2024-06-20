/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package pattern

import (
	"fmt"

	"github.com/spf13/cobra"
)

// useCmd represents the use command
var useCmd = &cobra.Command{
	Use:          "use",
	Short:        "Set a default pattern to use",
	Args:         cobra.ExactArgs(1),
	SilenceUsage: true,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("use called")
	},
}

func init() {}
