/*
Copyright © 2024 Nghi Nguyen <hamnghi250699@gmail.com>
*/
package cmd

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/nt54hamnghi/hoc/util"
	"github.com/spf13/cobra"
)

// scrapeCmd represents the scrape command
var scrapeCmd = &cobra.Command{
	Use:     "scrape",
	Short:   "Scrape data with a given URL",
	Aliases: []string{"s"},
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("expecting 1 argument, got %d", len(args))
		}

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

func init() {
	rootCmd.AddCommand(scrapeCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// scrapeCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// scrapeCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
