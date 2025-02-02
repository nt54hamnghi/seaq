/*
Copyright © 2024 Nghi Nguyen <hamnghi250699@gmail.com>
*/
package fetch

import (
	"context"
	"errors"
	"os"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/nt54hamnghi/seaq/cmd/flaggroup"
	"github.com/nt54hamnghi/seaq/pkg/loader"
	"github.com/nt54hamnghi/seaq/pkg/loader/html"
	"github.com/nt54hamnghi/seaq/pkg/loader/html/jina"
	"github.com/spf13/cobra"
	"github.com/thediveo/enumflag/v2"
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

// region: --- engine options
// https://github.com/thediveo/enumflag?tab=readme-ov-file#cli-flag-with-default

type engine enumflag.Flag

const (
	defaultEngine engine = iota
	jinaEngine
	firecrawlEngine
)

var engineIDs = map[engine][]string{
	defaultEngine:   {"default"},
	jinaEngine:      {"jina"},
	firecrawlEngine: {"firecrawl"},
}

func engineVariants() []string {
	variants := []string{}
	for _, v := range engineIDs {
		variants = append(variants, v...)
	}
	return variants
}

func completeEngineFlag(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	return engineVariants(), cobra.ShellCompDirectiveDefault
}

// endregion: --- engine options

type pageOptions struct {
	url       string
	selector  string
	auto      bool
	recursive recursive
	output    flaggroup.Output
	asJSON    bool
	engine    engine
}

func newPageCmd() *cobra.Command {
	var opts pageOptions

	cmd := &cobra.Command{
		Use:          "page [url]",
		Short:        "Get HTML data from a URL and convert it to markdown",
		Aliases:      []string{"pg"},
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

	flags.VarP(
		enumflag.New(&opts.engine, "engine", engineIDs, enumflag.EnumCaseSensitive),
		"engine", "e",
		"engine to use (default|jina|firecrawl)",
	)
	flags.StringVarP(&opts.selector, "selector", "s", "", "filter content by selector")
	flags.BoolVarP(&opts.auto, "auto", "a", false, "automatically detect content")
	flags.BoolVarP(&opts.asJSON, "json", "j", false, "output as JSON")
	flaggroup.InitGroups(cmd, &opts.recursive, &opts.output)

	// set up completion for engine flag
	err := cmd.RegisterFlagCompletionFunc("engine", completeEngineFlag)
	if err != nil {
		os.Exit(1)
	}

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

func (opts *pageOptions) parse(cmd *cobra.Command, args []string) error {
	opts.url = args[0]

	maxPagesSet := cmd.Flags().Changed("max-pages")
	recursiveSet := cmd.Flags().Changed("recursive")
	autoSet := cmd.Flags().Changed("auto")

	if opts.engine != defaultEngine {
		if autoSet || recursiveSet || maxPagesSet {
			return errors.New("--auto, --recursive, and --max-pages can only be used with --engine=default")
		}
	}

	return nil
}

func pageRun(ctx context.Context, opts pageOptions) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()

	var htmlLoader documentloaders.Loader

	switch opts.engine {
	case defaultEngine:
		baseLoader := html.NewLoader(
			html.WithURL(opts.url),
			html.WithSelector(opts.selector),
			html.WithAuto(opts.auto),
		)

		if opts.recursive.Recursive {
			htmlLoader = html.NewRecursiveLoader(
				html.WithPageLoader(baseLoader),
				html.WithMaxPages(opts.recursive.MaxPages),
			)
		} else {
			htmlLoader = baseLoader
		}
	case jinaEngine:
		htmlLoader = jina.NewLoader(
			jina.WithURL(opts.url),
			jina.WithSelector(opts.selector),
		)
	case firecrawlEngine:
		panic("todo")
	default:
		return errors.New("invalid engine")
	}

	dest, err := opts.output.Writer()
	if err != nil {
		return err
	}
	defer dest.Close()

	return loader.LoadAndWrite(ctx, htmlLoader, dest, opts.asJSON)
}
