package pool

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOrderedGo(t *testing.T) {
	var errTest = errors.New("test")

	tests := []struct {
		name  string
		tasks []Task[int]
		want  []Result[int]
	}{
		{
			name:  "empty",
			tasks: []Task[int]{},
			want:  []Result[int]{},
		},
		{
			name: "one",
			tasks: []Task[int]{
				func() (int, error) { return 1, nil },
			},
			want: []Result[int]{
				{id: 0, Output: 1, Err: nil},
			},
		},
		{
			name: "multiple",
			tasks: []Task[int]{
				func() (int, error) { return 0, errTest },
				func() (int, error) { return 1, nil },
				func() (int, error) { return 2, nil },
			},
			want: []Result[int]{
				{id: 0, Output: 0, Err: errTest},
				{id: 1, Output: 1, Err: nil},
				{id: 2, Output: 2, Err: nil},
			},
		},
	}

	asserts := assert.New(t)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := OrderedGo(tt.tasks)

			asserts.Equal(tt.want, res)
		})
	}
}
