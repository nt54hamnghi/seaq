package config

import (
	"fmt"

	configPkg "github.com/nt54hamnghi/seaq/pkg/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func newShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "show",
		Short:        "Show the contents of the config file being used",
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		PreRunE:      configPkg.Init,
		RunE: func(_ *cobra.Command, _ []string) error {
			path := viper.ConfigFileUsed()
			content, err := fs.ReadFile(path)
			if err != nil {
				return err
			}
			fmt.Println(string(content))

			return nil
		},
	}

	return cmd
}
