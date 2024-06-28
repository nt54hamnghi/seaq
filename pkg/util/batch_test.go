package util

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetThreadCount(t *testing.T) {
	var testCases = []struct {
		name      string
		taskCount int
		expected  int
	}{
		{name: "zero", taskCount: 0, expected: 1},
		{name: "negative", taskCount: -1, expected: 1},
		{name: "one", taskCount: 1, expected: 1},
		{name: "large", taskCount: 1024, expected: runtime.NumCPU()},
	}

	asserts := assert.New(t)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := GetThreadCount(tc.taskCount)
			asserts.Equal(actual, tc.expected)
		})
	}

}

func TestBatchProcess(t *testing.T) {

	var testCases = []struct {
		name     string
		nThreads int
		in       []int
		op       func([]int) int
		expected []int
	}{
		{
			name:     "emptySlice",
			nThreads: 1,
			in:       []int{},
			op:       func(x []int) int { return 0 },
			expected: []int{},
		},
		{
			name:     "oneElement",
			nThreads: 1,
			in:       []int{1},
			op:       func(x []int) int { return 0 },
			expected: []int{0},
		},
		{
			name:     "oneThread",
			nThreads: 1,
			in:       []int{1, 2, 3, 4, 5},
			op:       func(x []int) int { return 0 },
			expected: []int{0},
		},
		{
			name:     "multipleThreads",
			nThreads: 2,
			in:       []int{1, 2, 3, 4, 5},
			op:       func(x []int) int { return 0 },
			expected: []int{0, 0},
		},
		{
			name:     "nonTrivialOp",
			nThreads: 2,
			in:       []int{1, 2, 3, 4, 5},
			op:       func(x []int) int { return x[0] },
			expected: []int{1, 4},
		},
	}

	asserts := assert.New(t)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res := BatchProcess(tc.nThreads, tc.in, tc.op)
			if len(tc.in) == 0 {
				asserts.Empty(res)
			} else {
				asserts.Len(res, tc.nThreads)
			}

			asserts.Equal(res, tc.expected)
		})
	}

}
