/*
Copyright © 2024 Nghi Nguyen
*/
package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/nt54hamnghi/seaq/cmd/chat"
	"github.com/nt54hamnghi/seaq/cmd/compose"
	"github.com/nt54hamnghi/seaq/cmd/fetch"
	"github.com/nt54hamnghi/seaq/cmd/flag"
	"github.com/nt54hamnghi/seaq/cmd/flaggroup"
	"github.com/nt54hamnghi/seaq/cmd/model"
	"github.com/nt54hamnghi/seaq/cmd/pattern"
	"github.com/nt54hamnghi/seaq/pkg/config"
	"github.com/nt54hamnghi/seaq/pkg/llm"
	"github.com/nt54hamnghi/seaq/pkg/util"
	"github.com/nt54hamnghi/seaq/pkg/util/log"

	"github.com/spf13/cobra"
)

const version = "0.2.13"

type rootOptions struct {
	configFile  flag.FilePath
	hint        string
	input       string
	model       string
	noStream    bool
	inputFile   flag.FilePath
	output      flaggroup.Output
	pattern     string
	patternRepo string
	verbose     bool
}

func New() *cobra.Command {
	var opts rootOptions
	cmd := &cobra.Command{
		Use:          "seaq",
		Short:        "A cli tool to make learning more fun",
		Version:      version,
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		PreRunE: compose.SequenceE(
			config.Init,
			flaggroup.ValidateGroups(&opts.output),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			switch err := opts.parse(cmd, args); {
			case errors.Is(err, util.ErrInteractiveInput):
				return cmd.Usage()
			case err != nil:
				return err
			default:
				return run(cmd.Context(), opts)
			}
		},
	}

	setupFlags(cmd, &opts)
	addCommands(cmd)

	return cmd
}

func (opts *rootOptions) parse(_ *cobra.Command, _ []string) error {
	var (
		input string
		err   error
	)

	if opts.inputFile != "" {
		bytes, err := os.ReadFile(opts.inputFile.String())
		if err != nil {
			return err
		}
		input = string(bytes)
	} else {
		input, err = util.ReadPipedStdin()
		if err != nil {
			return err
		}
	}

	seaq := config.Seaq

	opts.input = input
	opts.model = seaq.Model()
	opts.pattern = seaq.Pattern()

	return nil
}

func run(ctx context.Context, opts rootOptions) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()

	if opts.verbose {
		log.Info("completion",
			"config", config.Seaq.ConfigFileUsed(),
			"model", config.Seaq.Model(),
			"pattern", config.Seaq.Pattern(),
		)
		fmt.Fprintln(os.Stderr)
	}

	// construct the prompt from pattern and scraped content
	prompt, err := config.Seaq.GetPrompt()
	if err != nil {
		return err
	}

	// construct the model
	// nolint: contextcheck
	model, err := llm.New(opts.model)
	if err != nil {
		return err
	}

	dest, err := opts.output.Writer()
	if err != nil {
		return err
	}
	defer dest.Close()

	// run the completion
	msgs := llm.PrepareMessages(prompt, opts.input, opts.hint)
	if opts.noStream {
		return llm.CreateCompletion(ctx, model, dest, msgs)
	}
	return llm.CreateStreamCompletion(ctx, model, dest, msgs)
}

func setupFlags(cmd *cobra.Command, opts *rootOptions) {
	// local flags are only available to the current command
	flags := cmd.Flags()
	flags.SortFlags = false
	flags.StringVarP(&opts.model, "model", "m", "", "model to use")
	flags.StringVar(&opts.hint, "hint", "", "optional context to guide the LLM's focus")
	flags.BoolVar(&opts.noStream, "no-stream", false, "disable streaming mode")
	flags.StringVarP(&opts.pattern, "pattern", "p", "", "pattern to use")
	flags.StringVarP(&opts.patternRepo, "repo", "r", "", "path to the pattern repository")
	flags.VarP(&opts.inputFile, "input", "i", "input file")
	config.AddConfigFlag(cmd, &opts.configFile)
	flags.BoolVarP(&opts.verbose, "verbose", "V", false, "verbose output")

	// flag groups
	flaggroup.InitGroups(cmd, &opts.output)

	// register completion function
	err := cmd.RegisterFlagCompletionFunc("pattern", pattern.CompletePatternArgs)
	if err != nil {
		cobra.CheckErr(err)
	}
	err = cmd.RegisterFlagCompletionFunc("model", model.CompleteModelArgs)
	if err != nil {
		cobra.CheckErr(err)
	}
}

func addCommands(cmd *cobra.Command) {
	// add subcommands
	cmd.AddCommand(
		chat.NewChatCmd(),
		model.NewModelCmd(),
		fetch.NewFetchCmd(),
		pattern.NewPatternCmd(),
	)

	// add groups
	// https://github.com/spf13/cobra/blob/main/site/content/user_guide.md#grouping-commands-in-help
	cmd.AddGroup(
		&cobra.Group{Title: "Common Commands:", ID: "common"},
		&cobra.Group{Title: "Management Commands:", ID: "management"},
	)
}
