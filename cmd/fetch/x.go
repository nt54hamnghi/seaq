/*
Copyright © 2024 Nghi Nguyen <hamnghi250699@gmail.com>
*/
package fetch

import (
	"context"
	"time"

	"github.com/nt54hamnghi/seaq/cmd/flaggroup"
	"github.com/nt54hamnghi/seaq/pkg/loader"
	"github.com/nt54hamnghi/seaq/pkg/loader/x"
	"github.com/spf13/cobra"
)

type xOptions struct {
	tweetID string
	single  bool
	output  flaggroup.Output
	asJSON  bool
}

func newXCmd() *cobra.Command {
	var opts xOptions

	// cmd represents the x command
	cmd := &cobra.Command{
		Use:          "x [url|videoId]",
		Short:        "Get thread or tweet from x.com",
		Args:         xArgs,
		PreRunE:      flaggroup.ValidateGroups(&opts.output),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := opts.parse(cmd, args); err != nil {
				return err
			}
			return xRun(cmd.Context(), opts)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false
	flags.BoolVar(&opts.single, "tweet", false, "get a single tweet")
	flags.BoolVarP(&opts.asJSON, "json", "j", false, "output as JSON")

	return cmd
}

func xArgs(cmd *cobra.Command, args []string) error {
	if err := cobra.ExactArgs(1)(cmd, args); err != nil {
		return err
	}

	tid, err := x.ResolveTweetID(args[0])
	if err != nil {
		return err
	}

	args[0] = tid

	return nil
}

func (opts *xOptions) parse(_ *cobra.Command, args []string) error {
	opts.tweetID = args[0]
	return nil
}

func xRun(ctx context.Context, opts xOptions) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()

	xLoader, err := x.NewXLoader(
		x.WithTweetID(opts.tweetID),
		x.WithoutReply(opts.single),
	)
	if err != nil {
		return err
	}

	dest, err := opts.output.Writer()
	if err != nil {
		return err
	}
	defer dest.Close()

	return loader.LoadAndWrite(ctx, xLoader, dest, opts.asJSON)
}
