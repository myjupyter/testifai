package parser

import (
	"os"
	"path/filepath"
	"regexp"
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
