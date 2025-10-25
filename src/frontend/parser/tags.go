package parser

import (
	"go/ast"
	"strings"
)

const testifaiTag = "//testifai:"

type TestStyle string

const (
	testStyleXUnit TestStyle = "xunit"
	testStyleTable TestStyle = "table-driven"
)

func ParseTestifyTags(cmtGp *ast.CommentGroup) (*PromptBuilder, bool) {
	if cmtGp == nil {
		return nil, false
	}
	for i := range cmtGp.List {
		if cmtGp.List[i] == nil {
			continue
		}
		if strings.HasPrefix(cmtGp.List[i].Text, testifaiTag) {
			return parseLine(cmtGp.List[i].Text)
		}
	}
	return nil, false
}

func parseLine(line string) (*PromptBuilder, bool) {
	const del = " "
	splits := strings.Split(line, del)
	s, ok := strings.CutPrefix(splits[0], testifaiTag)
	if !ok {
		return nil, false
	}
	b := NewPromptBuilder()
	switch s {
	case string(testStyleXUnit), string(testStyleTable):
		b.SetTestStyle(testStyleXUnit)
		return b, true
	}
	return nil, false
}
