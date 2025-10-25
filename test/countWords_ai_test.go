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
		{"", 0},                            // ⚠️ !!!ATTENTION!!! Empty string case
		{"hello", 1},                       // Single word
		{"hello world", 2},                 // Two words
		{"  hello   world  ", 2},           // ⚠️ Extra spaces around and between words
		{"\t\nhello\rworld\f", 2},          // ⚠️ !!!ATTENTION!!! Whitespace characters: tab, newline, carriage return, form feed
		{"a b c d e", 5},                   // Multiple words
		{"   ", 0},                         // ⚠️ Only whitespace
		{"word", 1},                        // No extra spaces
		{" word ", 1},                      // Leading and trailing space
		{"  word1  word2  word3  ", 3},     // ⚠️ Multiple leading/trailing spaces
		{"\n\t\r\f", 0},                    // ⚠️ !!!ATTENTION!!! Only escape sequence whitespaces
		{"single", 1},                      // Single word without any space
		{"one two three four five six", 6}, // Several words
	}

	for _, tt := range tests {
		result := countWords(tt.input)
		if result != tt.expected {
			t.Errorf("countWords(%q) = %d; expected %d", tt.input, result, tt.expected)
		}
	}
}
