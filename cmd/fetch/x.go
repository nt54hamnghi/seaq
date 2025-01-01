/*
Copyright © 2024 Nghi Nguyen <hamnghi250699@gmail.com>
*/
package fetch

import (
	"context"
	"fmt"
	"time"

	"github.com/nt54hamnghi/seaq/cmd/flaggroup"
	"github.com/nt54hamnghi/seaq/pkg/loader"
	"github.com/nt54hamnghi/seaq/pkg/loader/x"
	"github.com/spf13/cobra"
)

var onlyTweet bool

// xCmd represents the x command
var xCmd = &cobra.Command{
	Use:          "x [url|videoId]",
	Short:        "Get thread or tweet from x.com",
	Args:         xArgs,
	PreRunE:      flaggroup.ValidateGroups(&output),
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error { // nolint: revive
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		tid := args[0]

		xLoader, err := x.NewXLoader(
			x.WithTweetID(tid),
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

		return loader.LoadAndWrite(ctx, xLoader, dest, asJSON)
	},
}

func init() {
	flags := xCmd.Flags()

	flags.SortFlags = false

	flags.BoolVar(&onlyTweet, "tweet", false, "get a single tweet")
	flags.BoolVarP(&asJSON, "json", "j", false, "output as JSON")
}

func xArgs(_ *cobra.Command, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("accepts 1 arg(s), received %d", len(args))
	}

	tid, err := x.ResolveTweetID(args[0])
	if err != nil {
		return err
	}

	args[0] = tid

	return nil
}
