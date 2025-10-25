// ✅ Tests for countWords
package test

import (
	"testing"
)

func TestCountWords(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"", 0},
		{"hello", 1},
		{"hello world", 2},
		{"   leading spaces", 2},
		{"trailing spaces   ", 2},
		{"  multiple   spaces  ", 2},
		{"one\ttwo\nthree", 3}, // ⚠️ Test with mixed whitespace characters
		{"word", 1},
		{"word word", 2},
		{"word word word", 3},
		{"   ", 0},  // ⚠️ Test with only spaces
		{"\t\n", 0}, // ⚠️ Test with only tab and newline
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			result := countWords(test.input)
			if result != test.expected {
				t.Errorf("countWords(%q) = %d; want %d", test.input, result, test.expected)
			}
		})
	}
}
