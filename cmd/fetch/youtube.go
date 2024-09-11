/*
Copyright © 2024 Nghi Nguyen <hamnghi250699@gmail.com>
*/
package fetch

import (
	"context"
	"fmt"
	"time"

	"github.com/nt54hamnghi/hiku/cmd/flagGroup"
	"github.com/nt54hamnghi/hiku/pkg/loader"
	"github.com/nt54hamnghi/hiku/pkg/loader/youtube"
	"github.com/spf13/cobra"
)

var (
	metadata bool
	start    string
	end      string
)

// youtubeCmd represents the caption command
var youtubeCmd = &cobra.Command{
	Use:          "youtube [url|videoId]",
	Short:        "Get caption and description of a YouTube video",
	Aliases:      []string{"ytb", "y"},
	Args:         validateCaptionArgs,
	SilenceUsage: true,
	PreRunE:      flagGroup.ValidateGroups(&output),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		vid := args[0]

		var err error
		var startTs, endTs *youtube.Timestamp

		if start != "" {
			startTs, err = youtube.ParseTimestamp(start)
			if err != nil {
				return fmt.Errorf("failed to parse start time: %w", err)
			}
		}

		if end != "" {
			endTs, err = youtube.ParseTimestamp(end)
			if err != nil {
				return fmt.Errorf("failed to parse end time: %w", err)
			}
		}

		youtubeLoader := youtube.NewYouTubeLoader(
			youtube.WithVideoId(vid),
			youtube.WithMetadata(metadata),
			youtube.WithStart(startTs),
			youtube.WithEnd(endTs),
		)

		dest, err := output.Writer()
		if err != nil {
			return err
		}
		defer dest.Close()

		return loader.LoadAndWrite(ctx, youtubeLoader, dest, asJson)
	},
}

func init() {
	youtubeCmd.Flags().SortFlags = false

	youtubeCmd.Flags().StringVarP(&start, "start", "s", "", "start time")
	youtubeCmd.Flags().StringVarP(&end, "end", "e", "", "end time")
	youtubeCmd.Flags().BoolVarP(&metadata, "metadata", "m", false, "include metadata in the output")
	youtubeCmd.Flags().BoolVarP(&asJson, "json", "j", false, "output as JSON")

	flagGroup.InitGroups(youtubeCmd, &output)
}

func validateCaptionArgs(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("accepts 1 arg(s), received %d", len(args))
	}

	vid, err := youtube.ResolveVideoId(args[0])
	if err != nil {
		return err
	}

	args[0] = vid

	return nil
}
