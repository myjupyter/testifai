package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/myjupyter/testifai/pkg/vcs"
	"github.com/myjupyter/testifai/src/frontend/parser"
	"github.com/spf13/cobra"
)

var outputPath string
var testType string
var funcName string
var fileName string

var rootCmd = &cobra.Command{
	Use: "testifai",
	RunE: func(cmd *cobra.Command, args []string) error {
		_, err := parser.ParseFile(fileName, funcName, testType, outputPath)
		if err != nil {
			return err
		}

		// TODO

		return nil
	},
}

var initCmd = &cobra.Command{
	Use: "init",
	RunE: func(cmd *cobra.Command, args []string) error {
		repoURL, err := vcs.GetGitRemoteURL()
		if err != nil {
			fmt.Println(err)
			return err
		}

		repoURL, _ = strings.CutPrefix(repoURL, "git@")
		repoURL = strings.Replace(repoURL, ":", "/", 1)
		repoURL, _ = strings.CutSuffix(repoURL, ".git")

		return os.WriteFile("testifai.yaml", []byte(fmt.Sprintf(`project: '%s'

provider:
	openai:
		apiKey: '<YOUR_OPENAI_API_KEY>'
`, repoURL)), 0644)
	},
}

func main() {
	rootCmd.Flags().StringVarP(&testType, "type", "t", "xunit", "Type of test (xunit|table|suite)")
	rootCmd.Flags().StringVarP(&outputPath, "output", "o", "", "Output path")
	rootCmd.Flags().StringVarP(&funcName, "func", "f", "", "Specific unction/method name to test")

	fname := os.Getenv("GOFILE")
	dirPath, _ := os.Getwd()

	rootCmd.Flags().StringVarP(&fileName, "file", "p", filepath.Join(dirPath, fname), "Specific file to test")

	rootCmd.AddCommand(initCmd)
	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}
