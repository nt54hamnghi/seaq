/*
Copyright © 2024 Nghi Nguyen <hamnghi250699@gmail.com>
*/
package fetch

import (
	"context"
	"fmt"
	"time"

	"github.com/nt54hamnghi/hiku/cmd/flaggroup"
	"github.com/nt54hamnghi/hiku/pkg/loader"
	"github.com/nt54hamnghi/hiku/pkg/loader/youtube"
	"github.com/nt54hamnghi/hiku/pkg/util/timestamp"
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
	Args:         youTubeArgs,
	SilenceUsage: true,
	PreRunE:      flaggroup.ValidateGroups(&output),
	RunE: func(cmd *cobra.Command, args []string) error { // nolint: revive
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		vid := args[0]

		var err error
		// using pointers as an indicator for optional value
		var startTs, endTs *timestamp.Timestamp

		if start != "" {
			ts, err := timestamp.ParseTimestamp(start)
			if err != nil {
				return fmt.Errorf("failed to parse start time: %w", err)
			}
			startTs = &ts
		}

		if end != "" {
			ts, err := timestamp.ParseTimestamp(end)
			if err != nil {
				return fmt.Errorf("failed to parse end time: %w", err)
			}
			endTs = &ts
		}

		youtubeLoader := youtube.NewYouTubeLoader(
			youtube.WithVideoID(vid),
			youtube.WithMetadata(metadata),
			youtube.WithStart(startTs),
			youtube.WithEnd(endTs),
		)

		dest, err := output.Writer()
		if err != nil {
			return err
		}
		defer dest.Close()

		return loader.LoadAndWrite(ctx, youtubeLoader, dest, asJSON)
	},
}

func init() {
	youtubeCmd.Flags().SortFlags = false

	youtubeCmd.Flags().StringVarP(&start, "start", "s", "", "start time")
	youtubeCmd.Flags().StringVarP(&end, "end", "e", "", "end time")
	youtubeCmd.Flags().BoolVarP(&metadata, "metadata", "m", false, "include metadata in the output")
	youtubeCmd.Flags().BoolVarP(&asJSON, "json", "j", false, "output as JSON")

	flaggroup.InitGroups(youtubeCmd, &output)
}

func youTubeArgs(_ *cobra.Command, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("accepts 1 arg(s), received %d", len(args))
	}

	vid, err := youtube.ResolveVideoID(args[0])
	if err != nil {
		return err
	}

	args[0] = vid

	return nil
}
