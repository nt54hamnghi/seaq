package model

import (
	"errors"
	"fmt"
	"text/tabwriter"

	"github.com/nt54hamnghi/seaq/cmd/flag"
	"github.com/nt54hamnghi/seaq/pkg/config"
	"github.com/nt54hamnghi/seaq/pkg/llm"
	"github.com/spf13/cobra"
)

type getOptions struct {
	configFile flag.FilePath
}

func newGetCmd() *cobra.Command {
	var opts getOptions

	cmd := &cobra.Command{
		Use:          "get",
		Short:        "Get the default model",
		Aliases:      []string{"show"},
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		PreRunE:      config.Init,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if provider, name, ok := llm.LookupModel(config.Model()); ok {
				w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
				defer w.Flush()
				fmt.Fprintf(w, "Model:\t%s\n", name)
				fmt.Fprintf(w, "Provider:\t%s\n", provider)
				return nil
			}
			return errors.New("unexpected error: failed to get default model")
		},
	}

	config.AddConfigFlag(cmd, &opts.configFile)

	return cmd
}
