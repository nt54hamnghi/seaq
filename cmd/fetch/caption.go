/*
Copyright © 2024 Nghi Nguyen <hamnghi250699@gmail.com>
*/
package fetch

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/nt54hamnghi/hiku/pkg/loader/youtube"
	"github.com/nt54hamnghi/hiku/pkg/util"
	"github.com/spf13/cobra"
)

var (
	metadata bool
	start    string
	end      string
)

// captionCmd represents the caption command
var captionCmd = &cobra.Command{
	Use:          "caption [url|videoId]",
	Short:        "Get caption from a YouTube video",
	Aliases:      []string{"cap", "c"},
	Args:         validateCaptionArgs,
	SilenceUsage: true,
	PreRunE:      validatePersistentFlags,
	RunE: func(cmd *cobra.Command, args []string) error {
		vid := args[0]
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

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

		loader := youtube.NewYouTubeLoader(
			youtube.WithVideoId(vid),
			youtube.WithMetadata(metadata),
			youtube.WithStart(startTs),
			youtube.WithEnd(endTs),
		)

		var writer io.Writer

		if outputFile == "" {
			writer = os.Stdout
		} else {
			var err error

			if force {
				writer, err = util.NewOverwriteWriter(outputFile)
			} else {
				writer, err = util.NewFailExistingWriter(outputFile)
			}

			if err != nil {
				return err
			}
		}

		return fetchAndWrite(ctx, loader, writer)
	},
}

func init() {
	captionCmd.Flags().BoolVarP(&metadata, "metadata", "m", false, "include metadata in the output")
	captionCmd.Flags().StringVar(&start, "start", "", "start time")
	captionCmd.Flags().StringVar(&end, "end", "", "end time")
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
