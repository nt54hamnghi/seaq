/*
Copyright © 2024 Nghi Nguyen <hamnghi250699@gmail.com>
*/
package fetch

import (
	"context"
	"errors"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/nt54hamnghi/seaq/cmd/flaggroup"
	"github.com/nt54hamnghi/seaq/pkg/loader"
	"github.com/nt54hamnghi/seaq/pkg/loader/html"
	"github.com/spf13/cobra"
	"github.com/tmc/langchaingo/documentloaders"
)

// region: --- flag groups

type recursive struct {
	Recursive bool
	MaxPages  int
}

func (r *recursive) Init(cmd *cobra.Command) { // nolint: revive
	cmd.Flags().BoolVarP(&r.Recursive, "recursive", "r", false, "recursively fetch content")
	cmd.Flags().IntVarP(&r.MaxPages, "max-pages", "m", 5, "maximum number of pages to fetch")
}

func (r *recursive) Validate(cmd *cobra.Command, args []string) error { // nolint: revive
	maxPagesSet := cmd.Flags().Changed("max-pages")
	recursiveSet := cmd.Flags().Changed("recursive")

	if maxPagesSet && (!recursiveSet || !r.Recursive) {
		return errors.New("--max-pages can only be used with --recursive")
	}
	return nil
}

// endregion: --- flag groups

type pageOptions struct {
	url       string
	selector  string
	auto      bool
	recursive recursive
	output    flaggroup.Output
	asJSON    bool
}

func newPageCmd() *cobra.Command {
	var opts pageOptions

	cmd := &cobra.Command{
		Use:          "page [url]",
		Short:        "Get HTML data from a URL and convert it to markdown",
		Aliases:      []string{"pg", "p"},
		Args:         pageArgs,
		SilenceUsage: true,
		PreRunE:      flaggroup.ValidateGroups(&opts.recursive, &opts.output),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := opts.parse(cmd, args); err != nil {
				return err
			}
			return pageRun(cmd.Context(), opts)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false
	flags.BoolVarP(&opts.auto, "auto", "a", false, "automatically detect content")
	flags.StringVarP(&opts.selector, "selector", "s", "", "filter content by selector")
	flags.BoolVarP(&opts.asJSON, "json", "j", false, "output as JSON")
	flaggroup.InitGroups(cmd, &opts.recursive, &opts.output)

	return cmd
}

func pageArgs(cmd *cobra.Command, args []string) error { // nolint: revive
	if err := cobra.ExactArgs(1)(cmd, args); err != nil {
		return err
	}

	if !govalidator.IsURL(args[0]) {
		return errors.New("invalid URL")
	}

	return nil
}

func (opts *pageOptions) parse(_ *cobra.Command, args []string) error {
	opts.url = args[0]
	return nil
}

func pageRun(ctx context.Context, opts pageOptions) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()

	var htmlLoader documentloaders.Loader

	baseLoader := html.NewLoader(
		html.WithURL(opts.url),
		html.WithSelector(opts.selector),
		html.WithAuto(opts.auto),
	)

	if opts.recursive.Recursive {
		htmlLoader = html.NewRecursiveLoader(
			html.WithHTMLLoader(baseLoader),
			html.WithMaxPages(opts.recursive.MaxPages),
		)
	} else {
		htmlLoader = baseLoader
	}

	dest, err := opts.output.Writer()
	if err != nil {
		return err
	}
	defer dest.Close()

	return loader.LoadAndWrite(ctx, htmlLoader, dest, opts.asJSON)
}
