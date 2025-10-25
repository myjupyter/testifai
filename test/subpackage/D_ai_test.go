// ✅ Tests for D

package subpackage

import (
	"testing"
)

func TestD(t *testing.T) {
	tests := []struct {
		name string
		x    int
		want int
	}{
		{"Test with zero", 0, 6},
		{"Test with positive number", 5, 11},
		{"Test with negative number", -3, 3},
		{"Test with large positive number", 1000, 1006},
		{"Test with large negative number", -1000, -994},
		{"Test with maximum int", int(^uint(0) >> 1), int(^uint(0)>>1) + 6},         // ⚠️ Test case for maximum int value
		{"Test with minimum int", -int(^uint(0)>>1) - 1, -int(^uint(0)>>1) - 1 + 6}, // ⚠️ Test case for minimum int value
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := D(tt.x); got != tt.want {
				t.Errorf("D(%d) = %d, want %d", tt.x, got, tt.want)
			}
		})
	}
}
