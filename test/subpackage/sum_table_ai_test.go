// ✅ Tests for Sum function
package subpackage

import (
	"testing"
)

func TestSum(t *testing.T) {
	tests := []struct {
		name     string
		a, b     int
		expected int
	}{
		{"a less than b", 1, 2, 2},
		{"a greater than b", 3, 2, 3},
		{"a equals b", 2, 2, 4},
		{"a less than b with a() and b()", a(), b(), 2},    // ⚠️ Using functions a() and b()
		{"a greater than b with a() and b()", b(), a(), 2}, // ⚠️ Using functions a() and b()
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Sum(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("Sum(%d, %d) = %d; want %d", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}
