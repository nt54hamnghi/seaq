package fetch

import (
	"context"
	"time"

	"github.com/nt54hamnghi/seaq/cmd/flaggroup"
	"github.com/nt54hamnghi/seaq/pkg/loader"
	"github.com/nt54hamnghi/seaq/pkg/loader/reddit"
	"github.com/spf13/cobra"
)

type redditOptions struct {
	// global fetch options
	fetchGlobalOptions

	url string
}

func newRedditCmd() *cobra.Command {
	var opts redditOptions

	cmd := &cobra.Command{
		Use:          "reddit [url]",
		Short:        "Get content from a Reddit post or comment",
		Aliases:      []string{"rdt"},
		Args:         cobra.ExactArgs(1),
		PreRunE:      flaggroup.ValidateGroups(&opts.fetchGlobalOptions),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := opts.parse(cmd, args); err != nil {
				return err
			}
			return redditRun(cmd.Context(), opts)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false
	flaggroup.InitGroups(cmd, &opts.fetchGlobalOptions)

	return cmd
}

func (opts *redditOptions) parse(_ *cobra.Command, args []string) error {
	opts.url = args[0]
	return nil
}

func redditRun(ctx context.Context, opts redditOptions) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()

	redditLoader, err := reddit.NewRedditLoader(
		reddit.WithURL(opts.url),
	)
	if err != nil {
		return err
	}

	dest, err := opts.output.Writer()
	if err != nil {
		return err
	}
	defer dest.Close()

	if !opts.ignoreCache {
		return loader.LoadAndCache(ctx, redditLoader, dest, opts.asJSON)
	}

	return loader.LoadAndWrite(ctx, redditLoader, dest, opts.asJSON)
}
