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
		{name: "large", taskCount: 100, expected: runtime.NumCPU()},
	}

	assert := assert.New(t)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := GetThreadCount(tc.taskCount)
			assert.Equal(actual, tc.expected)
		})
	}

}
