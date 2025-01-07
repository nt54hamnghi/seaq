/*
Copyright © 2024 Nghi Nguyen
*/
package chat

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/nt54hamnghi/seaq/cmd/config"
	"github.com/nt54hamnghi/seaq/cmd/flag"
	"github.com/nt54hamnghi/seaq/cmd/model"
	"github.com/nt54hamnghi/seaq/pkg/repl"
	"github.com/nt54hamnghi/seaq/pkg/util"
	"github.com/spf13/cobra"
	"github.com/tmc/langchaingo/documentloaders"
	"github.com/tmc/langchaingo/textsplitter"
)

type chatOptions struct {
	input     string
	model     string
	noStream  bool
	inputFile flag.FilePath
	verbose   bool
}

func NewChatCmd() *cobra.Command {
	var opts chatOptions

	cmd := &cobra.Command{
		Use:     "chat",
		Short:   "Open a chat session",
		GroupID: "common",
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

	// set up flags
	flags := cmd.Flags()
	flags.SortFlags = false
	flags.StringVarP(&opts.model, "model", "m", "", "model to use")
	flags.BoolVar(&opts.noStream, "no-stream", false, "disable streaming mode")
	flags.VarP(&opts.inputFile, "input", "i", "input file")

	// set up completion for model flag
	err := cmd.RegisterFlagCompletionFunc("model", model.CompleteModelArgs)
	if err != nil {
		os.Exit(1)
	}

	return cmd
}

func (opts *chatOptions) parse(cmd *cobra.Command, _ []string) error {
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

	verbose, err := cmd.Root().PersistentFlags().GetBool("verbose")
	if err != nil {
		return err
	}

	opts.input = input
	opts.verbose = verbose
	if opts.model == "" {
		opts.model = config.Seaq.Model()
	}

	return nil
}

func run(ctx context.Context, opts chatOptions) error {
	if opts.verbose {
		fmt.Println("Using model:", opts.model)
	}

	// load the document
	loader := documentloaders.NewText(strings.NewReader(opts.input))
	docs, err := loader.LoadAndSplit(ctx,
		textsplitter.NewRecursiveCharacter(
			textsplitter.WithChunkSize(750),
			textsplitter.WithChunkOverlap(100),
		),
	)
	if err != nil {
		return err
	}

	// initialize chatREPL
	// nolint: contextcheck
	chatREPL, err := repl.New(docs,
		repl.WithContext(ctx),
		repl.WithModelName(opts.model),
		repl.WithNoStream(opts.noStream),
	)
	if err != nil {
		return err
	}

	return chatREPL.Run()
}
