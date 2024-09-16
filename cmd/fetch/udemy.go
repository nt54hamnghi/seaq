/*
Copyright © 2024 Nghi Nguyen <hamnghi250699@gmail.com>
*/
package fetch

import (
	"context"
	"time"

	"github.com/nt54hamnghi/hiku/cmd/flagGroup"
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
	PreRunE:      flagGroup.ValidateGroups(&output),
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		udemyLoader, err := udemy.NewUdemyLoader(
			udemy.WithUrl(args[0]),
		)
		if err != nil {
			return err
		}

		dest, err := output.Writer()
		if err != nil {
			return err
		}
		defer dest.Close()

		return loader.LoadAndWrite(ctx, udemyLoader, dest, asJson)
	},
}

func init() {
	udemyCmd.Flags().SortFlags = false

	udemyCmd.Flags().BoolVarP(&asJson, "json", "j", false, "output as JSON")
}
