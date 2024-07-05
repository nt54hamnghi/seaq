/*
Copyright © 2024 Nghi Nguyen <hamnghi250699@gmail.com>
*/
package scrape

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/nt54hamnghi/hiku/pkg/util"
	"github.com/nt54hamnghi/hiku/pkg/youtube"
	"github.com/spf13/cobra"
)

var (
	metadata bool
	start    string
	end      string
)

type task = func(context.Context, string) (string, error)
type result struct {
	id   int
	data string
	err  error
}

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

		ch := make(chan result, 2)
		wg := sync.WaitGroup{}

		tasks := make([]task, 0, 2)
		if metadata {
			tasks = append(tasks, youtube.FetchMetadata)
		}
		tasks = append(tasks, func(ctx context.Context, s string) (string, error) {
			var startTs, endTs *youtube.Timestamp

			if start != "" {
				s, err := youtube.ParseTimestamp(start)
				if err != nil {
					return "", fmt.Errorf("failed to parse start time: %w", err)
				}
				startTs = &s
			}

			if end != "" {
				e, err := youtube.ParseTimestamp(end)
				if err != nil {
					return "", fmt.Errorf("failed to parse end time: %w", err)
				}
				endTs = &e
			}

			return youtube.FetchCaption(ctx, s,
				youtube.WithStart(startTs),
				youtube.WithEnd(endTs),
			)
		})

		for i, f := range tasks {
			wg.Add(1)
			go func(id int, f task) {
				defer wg.Done()
				data, err := f(ctx, src)
				ch <- result{id: i, data: data, err: err}
			}(i, f)
		}

		go func() {
			wg.Wait()
			close(ch)
		}()

		res := make([]string, 2)
		for r := range ch {
			if r.err != nil {
				return r.err
			}
			res[r.id] = r.data
		}

		for _, r := range res {
			fmt.Print(r)
		}

		if outputFile != "" {
			content := fmt.Sprintf("%s\n%s", res[0], res[1])
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
