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
	"github.com/spf13/cobra"
)

var metadata bool

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

		youtubeLoader := youtube.NewYouTubeLoader(
			youtube.WithVideoID(args[0]),
			youtube.WithMetadata(metadata),
			youtube.WithStart(interval.Start),
			youtube.WithEnd(interval.End),
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
	flags := youtubeCmd.Flags()

	flags.SortFlags = false
	flags.BoolVarP(&metadata, "metadata", "m", false, "to include metadata")
	flags.BoolVarP(&asJSON, "json", "j", false, "output as JSON")

	flaggroup.InitGroups(youtubeCmd, &output, &interval)
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
