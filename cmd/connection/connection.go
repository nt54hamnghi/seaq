package connection

import (
	"github.com/spf13/cobra"
)

func NewConnectionCmd() *cobra.Command {
	// connectionCmd represents the connection command
	cmd := &cobra.Command{
		Use:          "connection",
		Short:        "Manage connections",
		Aliases:      []string{"conn"},
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		GroupID:      "management",
		RunE: func(cmd *cobra.Command, args []string) error { // nolint: revive
			return cmd.Usage()
		},
	}

	cmd.AddCommand(
		newCreateCmd(),
		newListCmd(),
		newRemoveCmd(),
	)

	return cmd
}
