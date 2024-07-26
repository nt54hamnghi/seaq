/*
Copyright © 2024 Nghi Nguyen <hamnghi250699@gmail.com>
*/
package fetch

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/nt54hamnghi/hiku/pkg/loader/html"
	"github.com/nt54hamnghi/hiku/pkg/util"
	"github.com/spf13/cobra"
	"github.com/tmc/langchaingo/documentloaders"
)

var (
	selector  string
	auto      bool
	recursive bool
	maxPages  int
)

// pageCmd represents the scrape command
var pageCmd = &cobra.Command{
	Use:          "page [url]",
	Short:        "Fetch HTML from a URL and convert it to markdown",
	Aliases:      []string{"pg", "p"},
	Args:         validatePageArgs,
	SilenceUsage: true,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		maxPagesSet := cmd.Flags().Changed("max-pages")
		recursiveSet := cmd.Flags().Changed("recursive")

		if maxPagesSet && (!recursiveSet || !recursive) {
			return errors.New("--max-pages can only be used with --recursive")
		}

		return validatePersistentFlags(cmd, args)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var loader documentloaders.Loader

		htmlLoader := html.NewHtmlLoader(
			html.WithUrl(args[0]),
			html.WithSelector(selector),
			html.WithAuto(auto),
		)

		if !recursive {
			loader = htmlLoader
		} else {
			loader = html.NewRecursiveHtmlLoader(
				html.WithHtmlLoader(htmlLoader),
				html.WithMaxPages(maxPages),
			)
		}

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
	pageCmd.Flags().StringVarP(&selector, "selector", "s", "", "filter content by selector")
	pageCmd.Flags().BoolVarP(&auto, "auto", "a", false, "automatically detect content")
	pageCmd.Flags().BoolVarP(&recursive, "recursive", "r", false, "recursively fetch content")
	pageCmd.Flags().IntVarP(&maxPages, "max-pages", "m", 5, "maximum number of pages to fetch")
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
