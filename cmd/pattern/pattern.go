package pattern

import (
	"github.com/nt54hamnghi/seaq/pkg/config"
	"github.com/spf13/cobra"
)

func NewPatternCmd() *cobra.Command {
	// patternCmd represents the pattern command
	cmd := &cobra.Command{
		Use:          "pattern",
		Short:        "Manage patterns",
		Aliases:      []string{"pat"},
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		GroupID:      "management",
		RunE: func(cmd *cobra.Command, args []string) error { // nolint: revive
			return cmd.Usage()
		},
	}

	cmd.AddCommand(
		newGetCmd(),
		newListCmd(),
		newSetCmd(),
		newAddCmd(),
	)

	return cmd
}

func CompletePatternArgs(cmd *cobra.Command, args []string, _ string) ([]string, cobra.ShellCompDirective) {
	// ListPatterns reads `pattern.repo` from the config file to provide completions
	// so it relies on the config file being fully loaded
	if err := config.EnsureConfig(cmd, args); err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	patterns, err := config.ListPatterns()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	return patterns, cobra.ShellCompDirectiveNoFileComp
}
