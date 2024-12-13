/*
Copyright © 2024 Nghi Nguyen <hamnghi250699@gmail.com>
*/
package fetch

import (
	"context"
	"time"

	"github.com/nt54hamnghi/hiku/cmd/flaggroup"
	"github.com/nt54hamnghi/hiku/pkg/loader"
	"github.com/nt54hamnghi/hiku/pkg/loader/udemy"
	"github.com/spf13/cobra"
)

// udemyCmd represents the udemy command
var udemyCmd = &cobra.Command{
	Use:          "udemy [url]",
	Short:        "Get transcript of a Udemy lecture video",
	Aliases:      []string{"udm", "u"},
	Args:         cobra.ExactArgs(1),
	PreRunE:      flaggroup.ValidateGroups(&output),
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error { // nolint: revive
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		udemyLoader, err := udemy.NewUdemyLoader(
			udemy.WithURL(args[0]),
			udemy.WithStart(interval.Start),
			udemy.WithEnd(interval.End),
		)
		if err != nil {
			return err
		}

		dest, err := output.Writer()
		if err != nil {
			return err
		}
		defer dest.Close()

		return loader.LoadAndWrite(ctx, udemyLoader, dest, asJSON)
	},
}

func init() {
	flags := udemyCmd.Flags()

	flags.SortFlags = false
	flags.BoolVarP(&asJSON, "json", "j", false, "output as JSON")

	flaggroup.InitGroups(udemyCmd, &output, &interval)
}
