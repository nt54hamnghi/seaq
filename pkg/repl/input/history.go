package input

// TESTME: add tests for history.go
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

func (h *history) append(input string) *history {
	if len(h.inputs) > 0 {
		latest := h.inputs[len(h.inputs)-1]
		if latest == input {
			h.cursor = len(h.inputs)
			return h
		}
	}

	h.inputs = append(h.inputs, input)
	h.cursor = len(h.inputs)

	return h
}

func (h *history) next() string {
	// at latest history entry or ready for new input
	if h.cursor >= len(h.inputs)-1 {
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
