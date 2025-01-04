/*
Copyright © 2024 Nghi Nguyen <hamnghi250699@gmail.com>
*/
package fetch

import (
	"context"
	"time"

	"github.com/nt54hamnghi/seaq/cmd/flaggroup"
	"github.com/nt54hamnghi/seaq/pkg/loader"
	"github.com/nt54hamnghi/seaq/pkg/loader/youtube"
	"github.com/spf13/cobra"
)

type youtubeOptions struct {
	videoID  string
	metadata bool
	output   flaggroup.Output
	interval flaggroup.Interval
	asJSON   bool
}

func newYoutubeCmd() *cobra.Command {
	var opts youtubeOptions

	cmd := &cobra.Command{
		Use:          "youtube [url|videoId]",
		Short:        "Get captions and metadata from YouTube videos",
		Aliases:      []string{"ytb", "y"},
		Args:         youtubeArgs,
		SilenceUsage: true,
		PreRunE:      flaggroup.ValidateGroups(&opts.output, &opts.interval),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := opts.parse(cmd, args); err != nil {
				return err
			}
			return youtubeRun(cmd.Context(), opts)
		},
	}

	// set up flags
	flags := cmd.Flags()
	flags.SortFlags = false
	flags.BoolVarP(&opts.metadata, "metadata", "m", false, "to include metadata")
	flags.BoolVarP(&opts.asJSON, "json", "j", false, "output as JSON")
	flaggroup.InitGroups(cmd, &opts.output, &opts.interval)

	return cmd
}

func youtubeArgs(cmd *cobra.Command, args []string) error {
	if err := cobra.ExactArgs(1)(cmd, args); err != nil {
		return err
	}

	vid, err := youtube.ResolveVideoID(args[0])
	if err != nil {
		return err
	}

	args[0] = vid

	return nil
}

func (opts *youtubeOptions) parse(_ *cobra.Command, args []string) error {
	opts.videoID = args[0]
	return nil
}

func youtubeRun(ctx context.Context, opts youtubeOptions) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()

	youtubeLoader := youtube.NewYouTubeLoader(
		youtube.WithVideoID(opts.videoID),
		youtube.WithMetadata(opts.metadata),
		youtube.WithStart(opts.interval.Start),
		youtube.WithEnd(opts.interval.End),
	)

	dest, err := opts.output.Writer()
	if err != nil {
		return err
	}
	defer dest.Close()

	return loader.LoadAndWrite(ctx, youtubeLoader, dest, opts.asJSON)
}
