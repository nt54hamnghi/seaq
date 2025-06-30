package connection

import (
	"fmt"

	"github.com/nt54hamnghi/seaq/cmd/flag"
	"github.com/nt54hamnghi/seaq/pkg/config"
	"github.com/spf13/cobra"
)

type removeOptions struct {
	configFile flag.FilePath
}

func newRemoveCmd() *cobra.Command {
	var opts removeOptions

	cmd := &cobra.Command{
		Use:          "remove",
		Short:        "Remove a connection",
		Aliases:      []string{"rm"},
		Args:         cobra.MinimumNArgs(1),
		SilenceUsage: true,
		PreRunE:      config.Init,
		RunE: func(cmd *cobra.Command, args []string) error { //nolint:revive
			if err := config.RemoveConnection(args...); err != nil {
				return fmt.Errorf("remove connection: %w", err)
			}
			return nil
		},
	}

	// set up flags
	config.AddConfigFlag(cmd, &opts.configFile)

	return cmd
}
