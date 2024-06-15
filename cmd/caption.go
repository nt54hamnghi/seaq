/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// captionCmd represents the caption command
var captionCmd = &cobra.Command{
	Use:     "caption",
	Short:   "Get caption from a YouTube video",
	Aliases: []string{"c"},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("caption called")
	},
}

func init() {
	rootCmd.AddCommand(captionCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// captionCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// captionCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
