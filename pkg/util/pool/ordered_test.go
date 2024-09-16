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

func TestOrderedGoFunc(t *testing.T) {
	var errTest = errors.New("test")

	tests := []struct {
		name     string
		input    []int
		taskFunc func(int) (int, error)
		want     []Result[int]
	}{
		{
			name:  "empty",
			input: []int{},
			want:  []Result[int]{},
		},
		{
			name:  "one",
			input: []int{1},
			taskFunc: func(i int) (int, error) {
				return i, nil
			},
			want: []Result[int]{
				{id: 0, Output: 1, Err: nil},
			},
		},
		{
			name:  "multiple",
			input: []int{0, 1, 2},
			taskFunc: func(i int) (int, error) {
				if i == 0 {
					return i, errTest
				}
				return i, nil
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
			res := OrderedGoFunc(tt.input, tt.taskFunc)
			asserts.Equal(tt.want, res)
		})
	}
}

func TestOrderedRun(t *testing.T) {
	taskFunc := func(i int) (int, error) {
		return i, nil
	}

	tests := []struct {
		name  string
		input []int
		want  []int
	}{
		{
			name:  "empty",
			input: []int{},
			want:  []int{},
		},
		{
			name:  "one",
			input: []int{1},
			want:  []int{1},
		},
		{
			name:  "multiple",
			input: []int{0, 1, 2},
			want:  []int{0, 1, 2},
		},
	}

	asserts := assert.New(t)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := OrderedRun(tt.input, taskFunc)
			asserts.Nil(err)
			asserts.Equal(tt.want, res)
		})
	}
}

func TestOrderedRun_Error(t *testing.T) {
	errTest := errors.New("test")
	taskFunc := func(i int) (int, error) {
		return 0, errTest
	}

	tests := []struct {
		name    string
		input   []int
		wantErr error
	}{
		{
			name:    "empty",
			input:   []int{},
			wantErr: nil,
		},
		{
			name:    "multiple",
			input:   []int{0, 1, 2},
			wantErr: errTest,
		},
	}

	asserts := assert.New(t)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := OrderedRun(tt.input, taskFunc)
			asserts.Equal(tt.wantErr, err)
		})
	}
}
