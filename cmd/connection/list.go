package connection

import (
	"fmt"
	"text/tabwriter"

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
		Short:        "List available connections",
		Aliases:      []string{"ls"},
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		PreRunE:      config.Init,
		RunE: func(cmd *cobra.Command, args []string) error { //nolint:revive
			conns, err := config.ListConnections()
			if err != nil {
				return fmt.Errorf("list connections: %w", err)
			}

			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 4, ' ', 0)
			defer w.Flush()

			const format = "%s\t%s\t%s\n"
			fmt.Fprintf(w, format, "PROVIDER", "BASE URL", "ENV KEY")
			for _, conn := range conns {
				fmt.Fprintf(w, format, conn.Provider, conn.BaseURL, conn.EnvKey)
			}

			return nil
		},
	}

	// set up flags
	config.AddConfigFlag(cmd, &opts.configFile)

	return cmd
}
