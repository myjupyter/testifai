package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/myjupyter/testifai/pkg/vcs"
	"github.com/spf13/cobra"
)

var outputPath string
var testType string

var rootCmd = &cobra.Command{
	Use: "testifai",
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

func init() {
	rootCmd.Flags().StringVarP(&testType, "type", "t", "xunit", "Type of test (xunit|table|suite)")
	rootCmd.Flags().StringVarP(&outputPath, "output", "o", "", "Output path")
}

func main() {
	rootCmd.AddCommand(initCmd)
	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}
