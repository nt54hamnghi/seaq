/*
Copyright © 2024 Nghi Nguyen
*/
package chat

import (
	"errors"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/nt54hamnghi/hiku/pkg/repl"
	"github.com/nt54hamnghi/hiku/pkg/util"
	"github.com/spf13/cobra"
)

// ChatCmd represents the chat command
var ChatCmd = &cobra.Command{
	Use:   "chat",
	Short: "Open a chat session",
	RunE: func(cmd *cobra.Command, args []string) error {
		_, err := util.ReadPipedStdin()
		if err != nil {
			if errors.Is(err, util.ErrInteractiveInput) {
				cmd.Help()
				return nil
			}
			return err
		}

		p := tea.NewProgram(repl.New())
		if _, err := p.Run(); err != nil {
			return err
		}
		return nil
	},
}

func init() {}
