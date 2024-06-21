/*
Copyright © 2024 Nghi Nguyen <hamnghi250699@gmail.com>
*/
package scrape

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/nt54hamnghi/hiku/pkg/youtube"
	"github.com/spf13/cobra"
)

// captionCmd represents the caption command
var captionCmd = &cobra.Command{
	Use:          "caption [url|videoId]",
	Short:        "Get caption from a YouTube video",
	Aliases:      []string{"c", "cap"},
	Args:         cobra.ExactArgs(1),
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		caption, err := youtube.WithVideoUrl(ctx, args[0])
		if err != nil && errors.Is(err, youtube.ErrInValidYouTubeURL) {
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
	captionCmd.Flags().BoolP("metadata", "m", false, "include metadata in the output")
}
