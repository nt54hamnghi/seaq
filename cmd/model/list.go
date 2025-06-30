package model

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
		Short:        "List available models",
		Aliases:      []string{"ls"},
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		PreRunE:      config.Init,
		RunE: func(_ *cobra.Command, _ []string) error {
			for _, m := range listModels() {
				fmt.Println(m)
			}
			return nil
		},
	}

	// set up flags
	config.AddConfigFlag(cmd, &opts.configFile)

	return cmd
}
