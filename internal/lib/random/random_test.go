package random

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewRandomString(t *testing.T) {
	tests := []struct {
		name string
		size int
	}{
		{
			name: "size = 1",
			size: 1,
		},
		{
			name: "size = 5",
			size: 5,
		},
		{
			name: "size = 10",
			size: 10,
		},
		{
			name: "size = 20",
			size: 20,
		},
		{
			name: "size = 100",
			size: 100,
		},
		{
			name: "size = 0",
			size: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			str1, err := NewRandomString(tt.size)
			assert.NoError(t, err)
			str2, err := NewRandomString(tt.size)
			assert.NoError(t, err)

			assert.Len(t, str1, tt.size)
			assert.Len(t, str2, tt.size)

			// Check that two generated strings are different
			// This is not an absolute guarantee that the function works correctly,
			// but this is a good heuristic for a simple random generator.
			if tt.size != 0 {
				assert.NotEqual(t, str1, str2)
			}
		})
	}
}
