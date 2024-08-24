/*
Copyright © 2024 Nghi Nguyen
*/
package chat

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/nt54hamnghi/hiku/cmd/config"
	"github.com/nt54hamnghi/hiku/cmd/model"
	"github.com/nt54hamnghi/hiku/pkg/llm"
	"github.com/nt54hamnghi/hiku/pkg/repl"
	"github.com/nt54hamnghi/hiku/pkg/util"
	"github.com/spf13/cobra"
	"github.com/tmc/langchaingo/documentloaders"
	"github.com/tmc/langchaingo/textsplitter"
)

var (
	modelName string
)

// ChatCmd represents the chat command
var ChatCmd = &cobra.Command{
	Use:   "chat",
	Short: "Open a chat session",
	RunE: func(cmd *cobra.Command, args []string) error {
		hiku := config.Hiku

		input, err := util.ReadPipedStdin()
		if err != nil {
			if errors.Is(err, util.ErrInteractiveInput) {
				cmd.Help()
				return nil
			}
			return err
		}

		ctx := context.Background()

		// load the document
		reader := strings.NewReader(input)
		loader := documentloaders.NewText(reader)
		docs, err := loader.LoadAndSplit(
			ctx,
			textsplitter.NewRecursiveCharacter(
				textsplitter.WithChunkSize(750),
				textsplitter.WithChunkOverlap(100),
			),
		)
		if err != nil {
			return err
		}

		verbose, err := cmd.Root().PersistentFlags().GetBool("verbose")
		if err != nil {
			return err
		}
		if verbose {
			fmt.Println("Using model:", modelName)
		}

		// construct model
		model, err := llm.New(hiku.Model())
		if err != nil {
			return err
		}

		// initialize repl
		repl, err := repl.NewRepl(docs,
			repl.WithContext(ctx),
			repl.WithModel(model),
			repl.WithDefaultStore(),
		)
		if err != nil {
			return err
		}

		return repl.Run()
	},
}

func init() {
	ChatCmd.Flags().StringVarP(&modelName, "model", "m", "", "model to use")
	ChatCmd.RegisterFlagCompletionFunc("model", model.CompleteModelArgs)

	err := config.Hiku.BindPFlag("model.name", ChatCmd.Flags().Lookup("model"))
	cobra.CheckErr(err)
}
