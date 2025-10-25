package test

import (
	"fmt"
	"strings"
)

//go:generate testifai --func=countWords --type=xunit --output=countWords_ai_test.go

func countWords(s string) int {
	if s == "" {
		return 0
	}
	return len(strings.Fields(s))
}

func bar() {}

func main() {
	var s = "sklfdj as;dlkfj asd;kfjasdf;klj"
	fmt.Println(countWords(s))
	bar()
}
