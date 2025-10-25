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
		// ⚠️ Test case for potential integer overflow on 32-bit systems if input is near MaxInt32
		{
			name:     "near max int32",
			input:    2147483640,
			expected: 2147483646, // 2147483640 + 1 + 2 + 3
		},
		// ⚠️ Test case for potential integer underflow on 32-bit systems if input is near MinInt32
		{
			name:     "near min int32",
			input:    -2147483640,
			expected: -2147483634, // -2147483640 + 1 + 2 + 3
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