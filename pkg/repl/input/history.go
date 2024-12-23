package input

import "strings"

// TESTME: add testCases for history.go
type history struct {
	inputs []string
	cursor int
}

func newHistory() history {
	return history{
		inputs: make([]string, 0),
		cursor: 0,
	}
}

func (h *history) append(input string) {
	input = strings.TrimSpace(input)

	// Skip empty inputs
	if input == "" {
		return
	}

	// Skip duplicate inputs
	if len(h.inputs) > 0 {
		if last := h.inputs[len(h.inputs)-1]; last == input {
			h.cursor = len(h.inputs)
			return
		}
	}

	h.inputs = append(h.inputs, input)
	h.cursor = len(h.inputs)
}

func (h *history) next() string {
	// at latest history entry or ready for new input
	if h.cursor >= len(h.inputs)-1 {
		h.cursor = len(h.inputs)
		return ""
	}

	h.cursor++
	return h.inputs[h.cursor]
}

func (h *history) previous() string {
	// at earliest history entry
	if len(h.inputs) == 0 {
		return ""
	}

	if h.cursor > 0 {
		h.cursor--
	}

	return h.inputs[h.cursor]
}
