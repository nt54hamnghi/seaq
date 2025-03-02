/*
Copyright © 2024 Nghi Nguyen <hamnghi250699@gmail.com>
*/
package fetch

import (
	"context"
	"time"

	"github.com/nt54hamnghi/seaq/cmd/flaggroup"
	"github.com/nt54hamnghi/seaq/pkg/loader"
	"github.com/nt54hamnghi/seaq/pkg/loader/udemy"
	"github.com/spf13/cobra"
)

type udemyOptions struct {
	// global fetch options
	fetchGlobalOptions

	url      string
	interval flaggroup.Interval
}

func newUdemyCmd() *cobra.Command {
	var opts udemyOptions

	cmd := &cobra.Command{
		Use:          "udemy [url]",
		Short:        "Get transcript of a Udemy lecture",
		Aliases:      []string{"udm"},
		Args:         cobra.ExactArgs(1),
		PreRunE:      flaggroup.ValidateGroups(&opts.interval, &opts.fetchGlobalOptions),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := opts.parse(cmd, args); err != nil {
				return err
			}
			return udemyRun(cmd.Context(), opts)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false
	flaggroup.InitGroups(cmd, &opts.interval, &opts.fetchGlobalOptions)

	return cmd
}

func (opts *udemyOptions) parse(_ *cobra.Command, args []string) error {
	opts.url = args[0]
	return nil
}

func udemyRun(ctx context.Context, opts udemyOptions) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()

	udemyLoader, err := udemy.NewUdemyLoader(
		udemy.WithURL(opts.url),
		udemy.WithStart(opts.interval.Start),
		udemy.WithEnd(opts.interval.End),
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
		return loader.LoadAndCache(ctx, udemyLoader, dest, opts.asJSON)
	}

	return loader.LoadAndWrite(ctx, udemyLoader, dest, opts.asJSON)
}
