/*
Copyright © 2024 Nghi Nguyen <hamnghi250699@gmail.com>
*/
package fetch

import (
	"context"
	"fmt"
	"time"

	"github.com/nt54hamnghi/hiku/cmd/flagGroup"
	"github.com/nt54hamnghi/hiku/pkg/loader"
	"github.com/nt54hamnghi/hiku/pkg/loader/x"
	"github.com/spf13/cobra"
)

var (
	onlyTweet bool
)

// xCmd represents the x command
var xCmd = &cobra.Command{
	Use:          "x [url|videoId]",
	Short:        "Get thread or tweet from x.com",
	Args:         xArgs,
	PreRunE:      flagGroup.ValidateGroups(&output),
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		tid := args[0]

		xLoader, err := x.NewXLoader(
			x.WithTweetId(tid),
			x.WithoutReply(onlyTweet),
		)
		if err != nil {
			return err
		}

		dest, err := output.Writer()
		if err != nil {
			return err
		}
		defer dest.Close()

		return loader.LoadAndWrite(ctx, xLoader, dest, asJson)
	},
}

func init() {
	xCmd.Flags().SortFlags = false

	xCmd.Flags().BoolVar(&onlyTweet, "tweet", false, "get a single tweet")
	xCmd.Flags().BoolVarP(&asJson, "json", "j", false, "output as JSON")
}

func xArgs(_ *cobra.Command, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("accepts 1 arg(s), received %d", len(args))
	}

	tid, err := x.ResolveTweetId(args[0])
	if err != nil {
		return err
	}

	args[0] = tid

	return nil
}
