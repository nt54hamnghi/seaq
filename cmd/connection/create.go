package connection

import (
	"fmt"
	"os"

	"github.com/nt54hamnghi/seaq/cmd/flag"
	"github.com/nt54hamnghi/seaq/pkg/config"
	"github.com/nt54hamnghi/seaq/pkg/llm"
	"github.com/spf13/cobra"
)

type createOptions struct {
	baseURL    flag.URL
	configFile flag.FilePath
	envKey     string
}

func newCreateCmd() *cobra.Command {
	var opts createOptions

	cmd := &cobra.Command{
		Use:          "create",
		Short:        "Create a new connection",
		Aliases:      []string{"add"},
		Args:         createArgs,
		SilenceUsage: true,
		PreRunE:      config.Init,
		RunE: func(cmd *cobra.Command, args []string) error {
			provider := args[0]

			conn := llm.NewConnection(provider, opts.baseURL.String(), opts.envKey)
			if err := config.AddConnection(conn); err != nil {
				return fmt.Errorf("add connection: %w", err)
			}

			fmt.Fprintln(cmd.OutOrStdout(), provider)
			return nil
		},
	}

	// set up flags
	flags := cmd.Flags()
	flags.SortFlags = false
	flags.Var(&opts.baseURL, "url", "Base URL")
	if err := cmd.MarkFlagRequired("url"); err != nil {
		os.Exit(1)
	}
	flags.StringVar(&opts.envKey, "env", "", "Environment variable name for API key")
	config.AddConfigFlag(cmd, &opts.configFile)

	return cmd
}

func createArgs(cmd *cobra.Command, args []string) error {
	if err := cobra.ExactArgs(1)(cmd, args); err != nil {
		return err
	}

	if !llm.IsIdent(args[0]) {
		return fmt.Errorf("provider name must be a valid identifier")
	}

	return nil
}
