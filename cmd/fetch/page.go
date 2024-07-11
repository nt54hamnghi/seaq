/*
Copyright © 2024 Nghi Nguyen <hamnghi250699@gmail.com>
*/
package fetch

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/nt54hamnghi/hiku/pkg/loader/html"
	"github.com/nt54hamnghi/hiku/pkg/util"
	"github.com/spf13/cobra"
)

var (
	selector string
	auto     bool
)

// pageCmd represents the scrape command
var pageCmd = &cobra.Command{
	Use:          "page [url]",
	Short:        "Fetch HTML from a URL and convert it to markdown",
	Aliases:      []string{"pg", "p"},
	Args:         cobra.ExactArgs(1),
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		parsedUrl, err := url.Parse(args[0])
		if err != nil {
			return err
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		loader := html.NewHtmlLoader(
			html.WithUrl(parsedUrl.String()),
			html.WithSelector(selector),
			html.WithAuto(auto),
		)

		content, err := fetch(ctx, loader)
		if err != nil {
			return err
		}

		fmt.Println(content)

		if outputFile != "" {
			if err := util.WriteFile(outputFile, content); err != nil {
				return err
			}
		}

		return nil
	},
}

func init() {
	pageCmd.Flags().StringVarP(&selector, "selector", "s", "", "filter content by selector")
	pageCmd.Flags().BoolVarP(&auto, "auto", "a", false, "automatically detect content")
}
