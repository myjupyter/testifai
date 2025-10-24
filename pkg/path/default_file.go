package path

import (
	"os"
	"path/filepath"
	"strings"
)

const gofileEnv = "GOFILE"

func GetInFilepath(fname string) string {
	if fname != "" {
		return fname
	}
	// default case
	sourceCodeFile := os.Getenv(gofileEnv)
	if sourceCodeFile == "" {
		return ""
	}
	dname, err := os.Getwd()
	if err != nil {
		return ""
	}
	return filepath.Join(dname, sourceCodeFile)
}

func GetOutFilepath(fname string) string {
	dname, err := os.Getwd()
	if err != nil {
		return ""
	}

	if fname != "" {
		dir := filepath.Dir(fname)
		if dir == "." {
			return filepath.Join(dname, fname)
		}
		return fname
	}

	sourceCodeFile := os.Getenv(gofileEnv)

	const testSuffix = "_ai_test.go"
	const ext = ".go"

	sourceCodeFile = strings.TrimSuffix(sourceCodeFile, testSuffix) + testSuffix

	return filepath.Join(dname, sourceCodeFile)
}
