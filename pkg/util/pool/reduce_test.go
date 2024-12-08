package pool

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetThreadCount(t *testing.T) {
	testCases := []struct {
		name      string
		taskCount int
		want      int
	}{
		{name: "zero", taskCount: 0, want: 1},
		{name: "negative", taskCount: -1, want: 1},
		{name: "one", taskCount: 1, want: 1},
		{name: "large", taskCount: 1024, want: runtime.NumCPU()},
	}

	a := assert.New(t)

	for _, tc := range testCases {
		t.Run(tc.name, func(*testing.T) {
			actual := GetThreadCount(tc.taskCount)
			a.Equal(tc.want, actual)
		})
	}
}

func TestBatchReduce(t *testing.T) {
	testCases := []struct {
		name     string
		nThreads int
		in       []int
		op       func([]int) int
		want     []int
	}{
		{
			name:     "emptySlice",
			nThreads: 1,
			in:       []int{},
			op:       func([]int) int { return 0 },
			want:     []int{},
		},
		{
			name:     "oneElement",
			nThreads: 1,
			in:       []int{1},
			op:       func([]int) int { return 0 },
			want:     []int{0},
		},
		{
			name:     "oneThread",
			nThreads: 1,
			in:       []int{1, 2, 3, 4, 5},
			op:       func([]int) int { return 0 },
			want:     []int{0},
		},
		{
			name:     "multipleThreads",
			nThreads: 2,
			in:       []int{1, 2, 3, 4, 5},
			op:       func([]int) int { return 0 },
			want:     []int{0, 0},
		},
		{
			name:     "nonTrivialOp",
			nThreads: 2,
			in:       []int{1, 2, 3, 4, 5},
			op:       func(x []int) int { return x[0] },
			want:     []int{1, 4},
		},
	}

	a := assert.New(t)

	for _, tc := range testCases {
		t.Run(tc.name, func(*testing.T) {
			res := BatchReduce(tc.nThreads, tc.in, tc.op)
			if len(tc.in) == 0 {
				a.Empty(res)
			} else {
				a.Len(res, tc.nThreads)
			}

			a.Equal(tc.want, res)
		})
	}
}
