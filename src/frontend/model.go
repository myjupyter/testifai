package frontend

type Provider string

type TestType string

const (
	XUnitTestType TestType = "xunit"
	TableTestType TestType = "table"
	SuiteTestType TestType = "suite"
)

const (
	OpenAi Provider = "openai"
)

type TestGenRequest struct {
	Provider Provider
	ApiKey   string
}

type Context struct {
	UserCode string
}

type TestifyInstructions struct {
	Type TestType
}
