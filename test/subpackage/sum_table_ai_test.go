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
			expected: 8, // ⚠️ !!!ATTENTION!!! a == b returns sum, not one value
		},
		{
			name:     "b greater than a",
			a:        2,
			b:        6,
			expected: 6,
		},
		{
			name:     "both negative, a less than b",
			a:        -3,
			b:        -1,
			expected: -1,
		},
		{
			name:     "both negative, a greater than b",
			a:        -1,
			b:        -3,
			expected: -1,
		},
		{
			name:     "one positive, one negative, positive wins",
			a:        -2,
			b:        3,
			expected: 3,
		},
		{
			name:     "one positive, one negative, negative wins",
			a:        -5,
			b:        2,
			expected: 2,
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
			expected: 0, // ⚠️ !!!ATTENTION!!! 0 == 0 returns 0 + 0 = 0
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
			result := Sum(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("Sum(%d, %d) = %d; expected %d", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}
