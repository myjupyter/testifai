package parser

type Target struct {
	Code    string
	Options TestifyOptions
}

type TestifyOptions struct {
	TestFunc   string
	TestType   string
	OutputFile string
}
