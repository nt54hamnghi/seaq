/*
Copyright © 2024 Nghi Nguyen <hamnghi250699@gmail.com>
*/
package scrape

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/nt54hamnghi/hiku/pkg/scraper"
	"github.com/nt54hamnghi/hiku/pkg/util"
	"github.com/spf13/cobra"
)

var tag string
var noFilter bool

// pageCmd represents the scrape command
var pageCmd = &cobra.Command{
	Use:          "page [url]",
	Short:        "Scrape HTML data with a given URL and convert it to markdown",
	Aliases:      []string{"web", "p", "w"},
	Args:         cobra.ExactArgs(1),
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		parsedUrl, err := url.Parse(args[0])
		if err != nil {
			return err
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var scr scraper.Scraper

		if noFilter {
			scr = scraper.WithFullPage()
		} else if tag != "" {
			scr, err = scraper.WithTag(tag)
			if err != nil {
				return err
			}
		} else {
			scr = scraper.New()
		}

		content, err := scraper.Scrape(ctx, parsedUrl.String(), scr)
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
	pageCmd.Flags().StringVarP(&tag, "tag", "t", "", "filter content by tag")
	pageCmd.Flags().BoolVarP(&noFilter, "no-filter", "n", false, "do not filter content")
}
