/*
Copyright © 2024 Nghi Nguyen
*/
package model

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "list",
		Short:        "List available models",
		Aliases:      []string{"ls"},
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE: func(_ *cobra.Command, _ []string) error {
			for _, m := range listModels() {
				fmt.Println(m)
			}
			return nil
		},
	}

	return cmd
}
