package frontend

type Provider string

type TestType string

const (
	XUnitTestType TestType = "xunit"
	TableTestType TestType = "table"
	SuiteTestType TestType = "suite"
)

func (t TestType) String() string {
	return string(t)
}

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

type TestGenResponse struct {
	TestedCode string
}
