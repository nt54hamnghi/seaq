/*
Copyright © 2024 Nghi Nguyen
*/

package config

import (
	"fmt"

	configPkg "github.com/nt54hamnghi/seaq/pkg/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func newPathCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "path",
		Short:        "Show path of the config file being used",
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		PreRunE:      configPkg.Init,
		RunE: func(_ *cobra.Command, _ []string) error {
			fmt.Println(viper.ConfigFileUsed())
			return nil
		},
	}

	return cmd
}
