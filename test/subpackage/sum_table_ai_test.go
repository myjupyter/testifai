// ✅ Tests for Sum function covering all branches and edge cases
package subpackage

import (
	"testing"
)

func TestSum(t *testing.T) {
	tests := []struct {
		name     string
		a        int
		b        int
		expected int
	}{
		{
			name:     "a greater than b",
			a:        5,
			b:        3,
			expected: 5,
		},
		{
			name:     "a equals b",
			a:        4,
			b:        4,
			expected: 8, // ⚠️ !!!ATTENTION!!! a + b when equal, not just a or b
		},
		{
			name:     "b greater than a",
			a:        2,
			b:        7,
			expected: 7,
		},
		{
			name:     "both negative, a greater",
			a:        -1,
			b:        -3,
			expected: -1,
		},
		{
			name:     "both negative, b greater",
			a:        -5,
			b:        -2,
			expected: -2,
		},
		{
			name:     "one positive, one negative, positive wins",
			a:        -1,
			b:        1,
			expected: 1,
		},
		{
			name:     "one positive, one negative, negative wins",
			a:        -5,
			b:        -3,
			expected: -3,
		},
		{
			name:     "zero and positive",
			a:        0,
			b:        5,
			expected: 5,
		},
		{
			name:     "zero and negative",
			a:        0,
			b:        -3,
			expected: -3,
		},
		{
			name:     "both zero",
			a:        0,
			b:        0,
			expected: 0, // ⚠️ !!!ATTENTION!!! 0 + 0 = 0, but still a == b case
		},
		{
			name:     "large positive numbers",
			a:        1000000,
			b:        999999,
			expected: 1000000,
		},
		{
			name:     "large negative numbers",
			a:        -1000000,
			b:        -999999,
			expected: -999999,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Sum(tt.a, tt.b); got != tt.expected {
				t.Errorf("Sum(%d, %d) = %d; want %d", tt.a, tt.b, got, tt.expected)
			}
		})
	}
}