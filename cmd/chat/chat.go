/*
Copyright © 2024 Nghi Nguyen
*/
package chat

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/nt54hamnghi/hiku/pkg/repl"
	"github.com/nt54hamnghi/hiku/pkg/util"
	"github.com/spf13/cobra"
	"github.com/tmc/langchaingo/documentloaders"
	"github.com/tmc/langchaingo/textsplitter"
)

// ChatCmd represents the chat command
var ChatCmd = &cobra.Command{
	Use:   "chat",
	Short: "Open a chat session",
	RunE: func(cmd *cobra.Command, args []string) error {
		input, err := util.ReadPipedStdin()
		if err != nil {
			if errors.Is(err, util.ErrInteractiveInput) {
				cmd.Help()
				return nil
			}
			return err
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		reader := strings.NewReader(input)
		loader := documentloaders.NewText(reader)
		docs, err := loader.LoadAndSplit(
			ctx,
			textsplitter.NewRecursiveCharacter(
				textsplitter.WithChunkSize(1000),
				textsplitter.WithChunkOverlap(100),
			),
		)
		if err != nil {
			return err
		}

		return repl.Run(ctx, docs)
	},
}

func init() {}
