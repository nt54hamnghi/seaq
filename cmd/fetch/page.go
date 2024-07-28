/*
Copyright © 2024 Nghi Nguyen <hamnghi250699@gmail.com>
*/
package fetch

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/nt54hamnghi/hiku/cmd/flagGroup"
	"github.com/nt54hamnghi/hiku/pkg/loader/html"
	"github.com/spf13/cobra"
	"github.com/tmc/langchaingo/documentloaders"
)

// region: --- flag groups

type recursive struct {
	Recursive bool
	MaxPages  int
}

func (r *recursive) Init(cmd *cobra.Command) {
	pageCmd.Flags().BoolVarP(&r.Recursive, "recursive", "r", false, "recursively fetch content")
	pageCmd.Flags().IntVarP(&r.MaxPages, "max-pages", "m", 5, "maximum number of pages to fetch")
}

func (r *recursive) Validate(cmd *cobra.Command, args []string) error {
	maxPagesSet := cmd.Flags().Changed("max-pages")
	recursiveSet := cmd.Flags().Changed("recursive")

	if maxPagesSet && (!recursiveSet || !r.Recursive) {
		return errors.New("--max-pages can only be used with --recursive")
	}
	return nil
}

// endregion: --- flag groups

var (
	selector string
	auto     bool
	rc       recursive
)

// pageCmd represents the scrape command
var pageCmd = &cobra.Command{
	Use:          "page [url]",
	Short:        "Fetch HTML from a URL and convert it to markdown",
	Aliases:      []string{"pg", "p"},
	Args:         validatePageArgs,
	SilenceUsage: true,
	PreRunE:      flagGroup.ValidateGroups(&rc, &output),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var loader documentloaders.Loader

		htmlLoader := html.NewHtmlLoader(
			html.WithUrl(args[0]),
			html.WithSelector(selector),
			html.WithAuto(auto),
		)

		if !rc.Recursive {
			loader = htmlLoader
		} else {
			loader = html.NewRecursiveHtmlLoader(
				html.WithHtmlLoader(htmlLoader),
				html.WithMaxPages(rc.MaxPages),
			)
		}

		dest, err := output.Writer()
		if err != nil {
			return err
		}
		defer dest.Close()

		return fetchAndWrite(ctx, loader, dest)
	},
}

func init() {
	pageCmd.Flags().StringVarP(&selector, "selector", "s", "", "filter content by selector")
	pageCmd.Flags().BoolVarP(&auto, "auto", "a", false, "automatically detect content")

	flagGroup.InitGroups(pageCmd, &rc, &output)
}

func validatePageArgs(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("accepts 1 arg(s), received %d", len(args))
	}

	if !govalidator.IsURL(args[0]) {
		return errors.New("invalid URL")
	}

	return nil
}
