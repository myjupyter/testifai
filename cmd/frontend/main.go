package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/myjupyter/testifai/pkg/path"
	"github.com/myjupyter/testifai/pkg/vcs"
	"github.com/myjupyter/testifai/src/frontend"
	"github.com/myjupyter/testifai/src/frontend/client"
	parserv2 "github.com/myjupyter/testifai/src/frontend/parser/v2"
	"github.com/spf13/cobra"
)

var outputPath string
var testType string
var funcName string
var fileName string

var rootCmd = &cobra.Command{
	Use: "testifai",
	RunE: func(cmd *cobra.Command, args []string) error {
		rootPath, err := path.GetRepoRoot()
		if err != nil {
			return err
		}

		repoContext, err := parserv2.NewRepositoryContext(rootPath)
		if err != nil {
			return err
		}

		if err := repoContext.Parse(); err != nil {
			return err
		}

		collection, err := parserv2.Collect(repoContext, fileName, funcName)
		if err != nil {
			return err
		}

		for _, collect := range collection {
			//fmt.Println()
			//fmt.Println("пакет")
			//fmt.Println(collect.PackageName)
			//fmt.Println(collect.FunctionName)
			//fmt.Println()
			//fmt.Println("тело")
			//fmt.Println(collect.BodyWithReceiver)
			//fmt.Println()
			//fmt.Println("внешние зависимости")
			//fmt.Println(collect.ExternalImports)
			//fmt.Println()
			//fmt.Println("внутренние зависимости")
			//fmt.Println(collect.InternalDependencies)
			//result, err := frontend.SendRequest(frontend.RequestData{
			//	Host:     "localhost:6667",
			//	UserCode: collect.Code,
			//	Provider: "openai",
			//	TestType: collect.Options.TestType,
			//	Platform: "go",
			//})

			response, err := client.SendRequest(client.RequestData{
				Host:            "localhost:6667",
				UserCode:        collect.BodyWithReceiver,
				UserCodeContext: collect.InternalDependencies,
				ExternalImports: collect.ExternalImports,
				PackageName:     collect.PackageName,
				TestType:        testType,
				Platform:        "go",
			})

			if err != nil {
				return err
			}

			err = os.WriteFile(outputPath, []byte(response.GeneratedTest), 0644)
			if err != nil {
				return err
			}

		}

		return nil

		// targets, err := parser.ParseFile(fileName, funcName, testType, outputPath)
		// if err != nil {
		// 	return err
		// }

		// for _, target := range targets {
		// 	result, err := client.SendRequest(client.RequestData{
		// 		Host:     "localhost:6667",
		// 		UserCode: target.Code,
		// 		Provider: "openai",
		// 		TestType: target.Options.TestType,
		// 		Platform: "go",
		// 	})

		// 	fmt.Println()
		// 	fmt.Println()
		// 	fmt.Println(result.GeneratedTest, err)
		// 	fmt.Println()
		// 	fmt.Println()
		// }

		// return nil
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
