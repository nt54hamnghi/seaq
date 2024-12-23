package input

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHistoryAppend(t *testing.T) {
	tests := []struct {
		name       string
		inputs     []string
		toAppend   string
		wantInputs []string
		wantCursor int
	}{
		{
			name:       "append to empty",
			inputs:     []string{},
			toAppend:   "cmd1",
			wantInputs: []string{"cmd1"},
			wantCursor: 1,
		},
		{
			name:       "append with whitespace",
			inputs:     []string{"cmd1"},
			toAppend:   "  cmd2  ",
			wantInputs: []string{"cmd1", "cmd2"},
			wantCursor: 2,
		},
		{
			name:       "skip empty input",
			inputs:     []string{"existing"},
			toAppend:   "   ",
			wantInputs: []string{"existing"},
			wantCursor: 1,
		},
		{
			name:       "skip duplicate input",
			inputs:     []string{"cmd1", "cmd2"},
			toAppend:   "cmd2",
			wantInputs: []string{"cmd1", "cmd2"},
			wantCursor: 2,
		},
	}

	a := assert.New(t)

	for _, tt := range tests {
		t.Run(tt.name, func(*testing.T) {
			h := newHistory()

			// Setup initial state
			for _, input := range tt.inputs {
				h.append(input)
			}

			// Perform append
			h.append(tt.toAppend)

			// Assert results
			a.Equal(tt.wantInputs, h.inputs)
			a.Equal(tt.wantCursor, h.cursor)
		})
	}
}

func TestHistoryNext(t *testing.T) {
	tests := []struct {
		name          string
		initialInputs []string
		initialCursor int
		wantOutput    string
		wantCursor    int
	}{
		{
			name:          "empty history",
			initialInputs: []string{},
			initialCursor: 0,
			wantOutput:    "",
			wantCursor:    0,
		},
		{
			name:          "non-empty history",
			initialInputs: []string{"cmd1", "cmd2"},
			initialCursor: 0,
			wantOutput:    "cmd2",
			wantCursor:    1,
		},
		{
			name:          "at last entry",
			initialInputs: []string{"cmd1", "cmd2"},
			initialCursor: 1,
			wantOutput:    "",
			wantCursor:    2,
		},
		{
			name:          "beyond last entry",
			initialInputs: []string{"cmd1", "cmd2"},
			initialCursor: 2,
			wantOutput:    "",
			wantCursor:    2,
		},
	}

	a := assert.New(t)

	for _, tt := range tests {
		t.Run(tt.name, func(*testing.T) {
			h := newHistory()

			// Setup initial state
			for _, input := range tt.initialInputs {
				h.append(input)
			}
			h.cursor = tt.initialCursor

			// Perform next operation
			result := h.next()

			// Assert results
			a.Equal(tt.wantOutput, result)
			a.Equal(tt.wantCursor, h.cursor)
		})
	}
}

func TestHistoryPrevious(t *testing.T) {
	tests := []struct {
		name          string
		initialInputs []string
		initialCursor int
		wantOutput    string
		wantCursor    int
	}{
		{
			name:          "empty history",
			initialInputs: []string{},
			initialCursor: 0,
			wantOutput:    "",
			wantCursor:    0,
		},
		{
			name:          "non-empty history",
			initialInputs: []string{"cmd1", "cmd2"},
			initialCursor: 1,
			wantOutput:    "cmd1",
			wantCursor:    0,
		},
		{
			name:          "at ready for new input",
			initialInputs: []string{"cmd1"},
			initialCursor: 1, // cursor at len(inputs)
			wantOutput:    "cmd1",
			wantCursor:    0,
		},
		{
			name:          "at first entry",
			initialInputs: []string{"cmd1", "cmd2"},
			initialCursor: 0,
			wantOutput:    "cmd1", // stays at first entry
			wantCursor:    0,
		},
	}

	a := assert.New(t)

	for _, tt := range tests {
		t.Run(tt.name, func(*testing.T) {
			h := newHistory()

			// Setup initial state
			for _, input := range tt.initialInputs {
				h.append(input)
			}
			h.cursor = tt.initialCursor

			// Perform previous operation
			result := h.previous()

			// Assert results
			a.Equal(tt.wantOutput, result)
			a.Equal(tt.wantCursor, h.cursor)
		})
	}
}
