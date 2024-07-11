/*
Copyright © 2024 Nghi Nguyen <hamnghi250699@gmail.com>
*/
package fetch

import (
	"context"
	"fmt"
	"time"

	"github.com/nt54hamnghi/hiku/pkg/loader/youtube"
	"github.com/nt54hamnghi/hiku/pkg/util"
	"github.com/spf13/cobra"
)

var (
	metadata bool
	start    string
	end      string
)

// captionCmd represents the caption command
var captionCmd = &cobra.Command{
	Use:          "caption [url|videoId]",
	Short:        "Get caption from a YouTube video",
	Aliases:      []string{"cap", "c"},
	Args:         cobra.ExactArgs(1),
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		src := args[0]
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var err error
		var startTs, endTs *youtube.Timestamp

		if start != "" {
			startTs, err = youtube.ParseTimestamp(start)
			if err != nil {
				return fmt.Errorf("failed to parse start time: %w", err)
			}
		}

		if end != "" {
			endTs, err = youtube.ParseTimestamp(end)
			if err != nil {
				return fmt.Errorf("failed to parse end time: %w", err)
			}
		}

		loader := youtube.NewYouTubeCaption(
			youtube.WithSource(src),
			youtube.WithMetadata(metadata),
			youtube.WithStart(startTs),
			youtube.WithEnd(endTs),
		)

		docs, err := loader.Load(ctx)
		if err != nil {
			return err
		}

		content := ""
		for _, doc := range docs {
			content += fmt.Sprintf("%s\n", doc.PageContent)
		}

		fmt.Print(content)

		if outputFile != "" {
			if err := util.WriteFile(outputFile, content); err != nil {
				return err
			}
		}

		return nil
	},
}

func init() {
	captionCmd.Flags().BoolVarP(&metadata, "metadata", "m", false, "include metadata in the output")
	captionCmd.Flags().StringVar(&start, "start", "", "start time")
	captionCmd.Flags().StringVar(&end, "end", "", "end time")
}
