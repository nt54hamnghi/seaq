package pattern

import (
	"fmt"
	"text/tabwriter"

	"github.com/nt54hamnghi/seaq/cmd/flag"
	"github.com/nt54hamnghi/seaq/pkg/config"
	"github.com/spf13/cobra"
)

type getOptions struct {
	configFile flag.FilePath
}

func newGetCmd() *cobra.Command {
	var opts getOptions

	cmd := &cobra.Command{
		Use:          "get",
		Short:        "Get the default pattern",
		Aliases:      []string{"g", "show"},
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		PreRunE:      config.Init,
		RunE: func(cmd *cobra.Command, _ []string) error {
			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
			defer w.Flush()

			fmt.Fprintf(w, "Pattern:\t%s\n", config.Pattern())
			fmt.Fprintf(w, "Repo:\t%s\n", config.Repo())

			return nil
		},
	}

	// set up flags
	config.AddConfigFlag(cmd, &opts.configFile)

	return cmd
}
