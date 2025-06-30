package pattern

import (
	"fmt"

	"github.com/nt54hamnghi/seaq/cmd/flag"
	"github.com/nt54hamnghi/seaq/pkg/config"
	"github.com/spf13/cobra"
)

type listOptions struct {
	configFile flag.FilePath
}

func newListCmd() *cobra.Command {
	var opts listOptions

	cmd := &cobra.Command{
		Use:          "list",
		Short:        "List available patterns",
		Aliases:      []string{"ls"},
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		PreRunE:      config.Init,
		RunE: func(cmd *cobra.Command, args []string) error { // nolint: revive
			pats, err := config.ListPatterns()
			if err != nil {
				return err
			}

			for _, p := range pats {
				fmt.Println(p)
			}

			return nil
		},
	}

	// set up flags
	config.AddConfigFlag(cmd, &opts.configFile)

	return cmd
}
