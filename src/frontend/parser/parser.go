package parser

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var testFileNamePattern = regexp.MustCompile(`_test\.go$`)

const goExt = ".go"

func WalkGoFiles(root string, callback func(path string) error) error {
	err := filepath.WalkDir(root, func(path string, info os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || testFileNamePattern.MatchString(info.Name()) {
			return nil
		}
		if filepath.Ext(path) == goExt {
			if err := callback(path); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func ParseFile(
	path,
	funcName,
	testType,
	output string,
) ([]Target, error) {
	fset := token.NewFileSet()

	isSpecificFunction := funcName != ""

	f, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}
	src, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	currentOptions := TestifyOptions{
		TestFunc:   funcName,
		TestType:   testType,
		OutputFile: output,
	}

	var targets []Target
	ast.Inspect(f, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.FuncDecl:

			options, isOverriden := OverrideTestify(x.Doc)
			if isOverriden && options != currentOptions {
				return true
			}
			if isSpecificFunction && options.TestFunc != funcName {
				return true
			}
			if isSpecificFunction && options.TestFunc != funcName {
				options = currentOptions
			}

			var target Target

			tf := fset.File(x.Pos())
			start := tf.Offset(x.Pos())
			end := tf.Offset(x.End())

			target.Code = string(src[start:end])
			target.Options = currentOptions
			targets = append(targets, target)

		}
		return true
	})

	return targets, nil
}

func OverrideTestify(doc *ast.CommentGroup) (options TestifyOptions, overriden bool) {
	config := TestifyOptions{}
	if doc == nil {
		return config, false
	}

	const generatePrefix = "//go:generate"

	var commandAndArgs string
	for _, commentLine := range doc.List {
		line := commentLine.Text
		if !strings.HasPrefix(line, generatePrefix) {
			continue
		}
		commandAndArgs = strings.TrimPrefix(line, generatePrefix)
		break
	}
	if commandAndArgs == "" {
		return config, false
	}

	parts := strings.Fields(commandAndArgs)

	if len(parts) == 0 {
		return config, false
	}

	for i := 1; i < len(parts); i++ {
		part := parts[i]

		flagParts := strings.SplitN(part, "=", 2)

		flagName := flagParts[0]
		var flagValue string
		if len(flagParts) > 1 {
			flagValue = flagParts[1]
		}

		switch flagName {
		case "--type", "-t":
			config.TestType = flagValue
		case "--output", "-o":
			config.OutputFile = flagValue
		case "--func", "-f":
			config.TestFunc = flagValue
		}
	}

	return config, config != TestifyOptions{}
}
