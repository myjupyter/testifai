package model

type GenerateRequest struct {
	ApiKey   string
	Provider string
	Id       string
	Context  GenerateContext
	Testifai Testifai
}

type GenerateContext struct {
	UserCode string
	Plarform string
}

type Testifai struct {
	TestType TestType
}
