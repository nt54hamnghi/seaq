package pattern

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/nt54hamnghi/seaq/cmd/flag"
	"github.com/nt54hamnghi/seaq/pkg/config"
	"github.com/nt54hamnghi/seaq/pkg/github"
	"github.com/nt54hamnghi/seaq/pkg/util/log"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type addOptions struct {
	remote      flag.URL
	configFile  flag.FilePath
	patternName string
}

var fs = afero.Afero{
	Fs: afero.NewOsFs(),
}

// nolint: revive
func newAddCmd() *cobra.Command {
	var opts addOptions

	cmd := &cobra.Command{
		Use:               "add [pattern-name]",
		Short:             "Add a pattern from a remote GitHub repository",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completeAddPatternArgs,
		SilenceUsage:      true,
		PreRunE:           config.Init,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := opts.parse(cmd, args); err != nil {
				return err
			}
			return addRun(cmd.Context(), opts)
		},
	}

	// set up flags
	flags := cmd.Flags()
	flags.SortFlags = false

	flags.VarP(&opts.remote, "remote", "r", "remote GitHub repository")
	config.AddConfigFlag(cmd, &opts.configFile)

	return cmd
}

func (opts *addOptions) parse(_ *cobra.Command, args []string) error {
	opts.patternName = args[0]
	return nil
}

func addRun(ctx context.Context, opts addOptions) error {
	remoteURL := viper.GetString("pattern.remote")
	repo, err := github.ParseRepositoryURL(remoteURL)
	if err != nil {
		return fmt.Errorf("parsing repository URL: %w", err)
	}

	// download the pattern content
	log.Info("Downloading pattern", "pattern", opts.patternName)
	content, err := repo.DownloadPattern(ctx, opts.patternName)
	if err != nil {
		return fmt.Errorf("downloading pattern: %w", err)
	}

	patternRepo := viper.GetString("pattern.repo")
	if patternRepo == "" {
		return fmt.Errorf("unexpected: pattern repository is not set")
	}

	// create directory for the requested pattern if not exists
	patternDir := filepath.Join(patternRepo, opts.patternName)
	if err := fs.MkdirAll(patternDir, 0o755); err != nil {
		return fmt.Errorf("creating %s pattern directory: %w", opts.patternName, err)
	}

	// write the pattern content to the file if it doesn't exist
	patternFile := filepath.Join(patternDir, "system.md")
	log.Info("Writing pattern", "pattern", opts.patternName, "file", patternFile)
	if err := fs.SafeWriteReader(patternFile, strings.NewReader(content)); err != nil {
		return fmt.Errorf("writing %s pattern file: %w", opts.patternName, err)
	}

	return nil
}

func completeAddPatternArgs(cmd *cobra.Command, args []string, _ string) ([]string, cobra.ShellCompDirective) {
	// the config file must be loaded before we access "pattern.remote"
	if err := config.EnsureConfig(cmd, args); err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	remote := viper.GetString("pattern.remote")
	repo, err := github.ParseRepositoryURL(remote)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	patterns, err := repo.GetPatternNames(cmd.Context())
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	return patterns, cobra.ShellCompDirectiveNoFileComp
}
