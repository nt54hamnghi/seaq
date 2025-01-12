package set

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAdd(t *testing.T) {
	testCases := []struct {
		name  string
		toAdd []int
		want  Set[int]
	}{
		{
			name:  "add once",
			toAdd: []int{1},
			// manually created set as NewSet uses Add internally
			want: Set[int]{1: {}},
		},
		{
			name:  "add multiple",
			toAdd: []int{1, 2, 3},
			// manually created set as NewSet uses Add internally
			want: Set[int]{1: {}, 2: {}, 3: {}},
		},
		{
			name:  "add duplicate",
			toAdd: []int{1, 1, 1},
			// manually created set as NewSet uses Add internally
			want: Set[int]{1: {}},
		},
	}

	a := assert.New(t)

	for _, tt := range testCases {
		t.Run(tt.name, func(*testing.T) {
			set := make(Set[int])
			for _, e := range tt.toAdd {
				set.Add(e)
			}
			a.Equal(tt.want, set)
		})
	}
}

func TestNewSet(t *testing.T) {
	testCases := []struct {
		name     string
		elements []int
		want     Set[int]
	}{
		{
			name:     "empty set",
			elements: nil,
			want:     Set[int]{},
		},
		{
			name:     "single",
			elements: []int{1},
			want:     Set[int]{1: {}},
		},
		{
			name:     "multiple",
			elements: []int{1, 2, 3},
			want:     Set[int]{1: {}, 2: {}, 3: {}},
		},
		{
			name:     "duplicate",
			elements: []int{1, 1, 2, 2, 3, 3},
			want:     Set[int]{1: {}, 2: {}, 3: {}},
		},
	}

	a := assert.New(t)

	for _, tt := range testCases {
		t.Run(tt.name, func(*testing.T) {
			got := New(tt.elements...)
			a.Equal(tt.want, got)
		})
	}
}

func TestContains(t *testing.T) {
	testCases := []struct {
		name    string
		set     Set[int]
		element int
		want    bool
	}{
		{
			name:    "empty set",
			set:     New[int](),
			element: 1,
			want:    false,
		},
		{
			name:    "contains",
			set:     New(1, 2, 3),
			element: 2,
			want:    true,
		},
		{
			name:    "not contains",
			set:     New(1, 2, 3),
			element: 4,
			want:    false,
		},
	}

	a := assert.New(t)

	for _, tt := range testCases {
		t.Run(tt.name, func(*testing.T) {
			got := tt.set.Contains(tt.element)
			a.Equal(tt.want, got)
		})
	}
}

func TestIter(t *testing.T) {
	testCases := []struct {
		name string
		set  Set[int]
		want []int
	}{
		{
			name: "empty set",
			set:  New[int](),
			want: []int(nil),
		},
		{
			name: "single",
			set:  New(1),
			want: []int{1},
		},
		{
			name: "multiple",
			set:  New(1, 2, 3),
			want: []int{1, 2, 3},
		},
	}

	a := assert.New(t)

	for _, tt := range testCases {
		t.Run(tt.name, func(*testing.T) {
			var got []int
			for v := range tt.set.Iter() {
				got = append(got, v)
			}
			slices.Sort(got)

			a.Equal(tt.want, got)
		})
	}
}
