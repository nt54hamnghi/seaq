package repl

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/uuid"
)

type role string

const (
	roleUser      role = "user"
	roleAssistant role = "assistant"
)

type conversation struct {
	CreatedAt int64     `json:"created_at"` // timestamp in epoch seconds
	ID        uuid.UUID `json:"id"`
	Model     string    `json:"model"`
	Messages  []message `json:"messages"`
}

type message struct {
	ID        uuid.UUID  `json:"id"`
	ParentID  *uuid.UUID `json:"parentId"` // optional - pointer to handle null
	ChildID   *uuid.UUID `json:"childId"`  // optional - pointer to handle null
	Content   string     `json:"content"`
	Role      role       `json:"role"`
	Timestamp int64      `json:"timestamp"` // timestamp in epoch seconds
}

func newConversation(model string) *conversation {
	return &conversation{
		CreatedAt: time.Now().Unix(),
		ID:        uuid.New(),
		Model:     model,
		Messages:  []message{},
	}
}

func (c *conversation) addMessage(content string, role role) error {
	if content == "" {
		return errors.New("content is empty")
	}

	if role != roleUser && role != roleAssistant {
		return errors.New("invalid role")
	}

	msg := message{
		ID:        uuid.New(),
		Content:   content,
		Role:      role,
		Timestamp: time.Now().Unix(),
	}

	if len(c.Messages) > 0 {
		// create a copy of the ID
		msgID := msg.ID
		lastMsgID := c.Messages[len(c.Messages)-1].ID

		msg.ParentID = &lastMsgID
		c.Messages[len(c.Messages)-1].ChildID = &msgID
	}

	c.Messages = append(c.Messages, msg)
	return nil
}

func (c *conversation) toText() string {
	var b strings.Builder

	for _, msg := range c.Messages {
		b.WriteString("--- " + strings.ToUpper(string(msg.Role)) + " ---\n")
		b.WriteString(msg.Content + "\n\n")
	}

	return b.String()
}

func (c *conversation) toJSON() (string, error) {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

type saveConversationMsg struct {
	path string
}

func (c *conversation) save(path string, data []byte) error {
	if len(c.Messages) == 0 {
		return errors.New("conversation is empty")
	}

	// owner: read (4) + write (2) = 6
	// group: no permissions (0)
	// others: no permissions (0)
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return err
	}

	return nil
}

func (c *conversation) saveText() tea.Cmd {
	return func() tea.Msg {
		filename := fmt.Sprintf("chat-export-%d.txt", time.Now().UnixMilli())
		absPath, err := filepath.Abs(filename)
		if err != nil {
			return err
		}
		if err := c.save(absPath, []byte(c.toText())); err != nil {
			return err
		}
		return saveConversationMsg{path: absPath}
	}
}

func (c *conversation) saveJSON() tea.Cmd {
	return func() tea.Msg {
		filename := fmt.Sprintf("chat-export-%d.json", time.Now().UnixMilli())
		// $PWD/filename
		absPath, err := filepath.Abs(filename)
		if err != nil {
			return err
		}
		json, err := c.toJSON()
		if err != nil {
			return err
		}
		if err := c.save(absPath, []byte(json)); err != nil {
			return err
		}
		return saveConversationMsg{path: absPath}
	}
}
