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

var modelName string

// ChatCmd represents the chat command
var ChatCmd = &cobra.Command{
	Use:   "chat",
	Short: "Open a chat session",
	RunE: func(cmd *cobra.Command, args []string) error { // nolint: revive
		hiku := config.Hiku

		input, err := util.ReadPipedStdin()
		if err != nil {
			if errors.Is(err, util.ErrInteractiveInput) {
				_ = cmd.Help()
				return nil
			}
			return err
		}

		ctx := context.Background()

		verbose, err := cmd.Root().PersistentFlags().GetBool("verbose")
		if err != nil {
			return err
		}
		if verbose {
			fmt.Println("Using model:", modelName)
		}

		// load the document
		loader := documentloaders.NewText(strings.NewReader(input))
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

		// construct model
		model, err := llm.New(hiku.Model())
		if err != nil {
			return err
		}

		// initialize chatREPL
		chatREPL, err := repl.New(docs,
			repl.WithContext(ctx),
			repl.WithModel(model),
		)
		if err != nil {
			return err
		}

		return chatREPL.Run()
	},
}

func init() {
	flags := ChatCmd.Flags()

	flags.SortFlags = false
	flags.StringVarP(&modelName, "model", "m", "", "model to use")
	_ = ChatCmd.RegisterFlagCompletionFunc("model", model.CompleteModelArgs)

	err := config.Hiku.BindPFlag("model.name", flags.Lookup("model"))
	cobra.CheckErr(err)
}
