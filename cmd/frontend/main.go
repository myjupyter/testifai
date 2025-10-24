package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/myjupyter/testifai/pkg/path"
	"github.com/myjupyter/testifai/pkg/vcs"
	"github.com/myjupyter/testifai/src/frontend"
	"github.com/myjupyter/testifai/src/frontend/client"
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
		targets, err := parser.ParseFile(fileName, funcName, testType, outputPath)
		if err != nil {
			return err
		}

		for _, target := range targets {
			result, err := client.SendRequest(client.RequestData{
				Host:     "localhost:6667",
				UserCode: target.Code,
				Provider: "openai",
				TestType: target.Options.TestType,
				Platform: "go",
			})
			if err != nil {
				return err
			}

			// fmt.Println()
			// fmt.Println()
			// fmt.Println(target)
			// fmt.Println()
			// fmt.Println()
			os.WriteFile(outputPath, []byte(result.GeneratedTest), 0644)
		}

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
	rootCmd.Flags().StringVarP(&testType, "type", "t", frontend.XUnitTestType.String(), "Type of test (xunit|table|suite)")
	rootCmd.Flags().StringVarP(&funcName, "func", "f", "", "Specific unction/method name to test")

	rootCmd.Flags().StringVarP(&outputPath, "output", "o", "", "Output path")
	outputPath = path.GetOutFilepath(outputPath)

	rootCmd.Flags().StringVarP(&fileName, "file", "p", "", "Specific file to test")
	fileName = path.GetInFilepath(fileName)

	rootCmd.AddCommand(initCmd)
	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}
