package model

type GenerateRequest struct {
	Context  GenerateContext
	Testifai Testifai
}

type GenerateContext struct {
	UserCode        string
	UserCodeContext string
	PackageName     string
	ExternalImport  []string
	Plarform        string
}

type Testifai struct {
	TestType TestType
}
