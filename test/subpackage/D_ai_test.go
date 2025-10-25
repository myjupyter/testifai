// ✅ Tests for D function
package subpackage

import (
	"testing"
)

func TestD(t *testing.T) {
	tests := []struct {
		name     string
		input    int
		expected int
	}{
		{
			name:     "zero input",
			input:    0,
			expected: 6, // 0 + 1 + 2 + 3
		},
		{
			name:     "positive input",
			input:    5,
			expected: 11, // 5 + 1 + 2 + 3
		},
		{
			name:     "negative input",
			input:    -10,
			expected: -4, // -10 + 1 + 2 + 3
		},
		{
			name:     "large positive input",
			input:    1000000,
			expected: 1000006, // 1000000 + 1 + 2 + 3
		},
		{
			name:     "large negative input",
			input:    -1000000,
			expected: -999994, // -1000000 + 1 + 2 + 3
		},
		// ⚠️ Test with minimum int value to check for potential overflow behavior
		{
			name:     "minimum int",
			input:    -9223372036854775808,
			expected: -9223372036854775802, // -9223372036854775808 + 1 + 2 + 3
		},
		// ⚠️ Test with maximum int value to check for potential overflow behavior
		{
			name:     "maximum int",
			input:    9223372036854775807,
			expected: 9223372036854775813, // 9223372036854775807 + 1 + 2 + 3
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := D(tt.input)
			if result != tt.expected {
				t.Errorf("D(%d) = %d; expected %d", tt.input, result, tt.expected)
			}
		})
	}
}
