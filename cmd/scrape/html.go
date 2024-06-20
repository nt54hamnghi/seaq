/*
Copyright © 2024 Nghi Nguyen <hamnghi250699@gmail.com>
*/
package scrape

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/nt54hamnghi/hiku/pkg/util"
	"github.com/spf13/cobra"
)

// htmlCmd represents the scrape command
var htmlCmd = &cobra.Command{
	Use:          "html [url]",
	Short:        "Scrape HTML data with a given URL and convert it to markdown",
	Aliases:      []string{"p", "page", "web"},
	Args:         cobra.ExactArgs(1),
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		parsedUrl, err := url.Parse(args[0])
		if err != nil {
			return err
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		html, err := util.GetRaw(ctx, parsedUrl.String(), nil)
		if err != nil {
			return err
		}

		safeHTML := util.SanitizeHTML(html)
		markdown, err := util.HTMLToMarkdown(safeHTML)
		if err != nil {
			return err
		}

		fmt.Println(string(markdown))

		return nil
	},
}

func init() {}
