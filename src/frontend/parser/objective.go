package parser

const (
	XUnitTestType = "xunit"
	TableTestType = "table"
	SuiteTestType = "suite"
)

type TestType string

type Objective struct {
	TestType TestType
	Code     string
}
