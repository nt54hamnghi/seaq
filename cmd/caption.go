/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/nt54hamnghi/hoc/youtube"
	"github.com/spf13/cobra"
)

// captionCmd represents the caption command
var captionCmd = &cobra.Command{
	Use:     "caption",
	Short:   "Get caption from a YouTube video",
	Aliases: []string{"c", "cap"},
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("expecting 1 argument, got %d", len(args))
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		caption, err := youtube.WithVideoUrl(ctx, args[0])
		if err != nil && err.Error() == "invalid YouTube video URL" {
			caption, err = youtube.WithVideoId(ctx, args[0])
		}

		if err != nil {
			return err
		}

		fmt.Println(caption)

		return nil
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
