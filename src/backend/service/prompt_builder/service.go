package prompt_builder

import (
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"text/template"

	"github.com/myjupyter/testifai/src/backend/app/model"
)

//go:embed prompt_template.tmpl
var promptTemplate string

type Service struct {
	tmpl *template.Template
}

func New() (*Service, error) {
	tmpl, err := template.New("test_template").Parse(promptTemplate)
	if err != nil {
		return nil, err
	}
	return &Service{
		tmpl: tmpl,
	}, nil
}

type templateModel struct {
	TestType string
	Language string
	UserCode string
}

func (s *Service) BuildPrompt(request model.GenerateRequest) (string, error) {
	buffer := new(bytes.Buffer)

	testType, err := s.toAiTestType(request.Testifai.TestType)
	if err != nil {
		return "", fmt.Errorf("error build prompt template: %v", err)
	}

	err = s.tmpl.Execute(buffer, templateModel{
		TestType: testType,
		Language: request.Context.Plarform,
		UserCode: request.Context.UserCode,
	})
	if err != nil {
		return "", fmt.Errorf("error build prompt template: %v", err)
	}
	return buffer.String(), nil
}

func (s *Service) toAiTestType(testType model.TestType) (string, error) {
	switch testType {
	case model.TestTypeTable:
		return "Table-Driven", nil
	case model.TestTypeSuite:
		return "Test Suite", nil
	case model.TestTypeXUnit:
		return "XUnit", nil
	default:
		return "", errors.New("unknown test type")
	}
}
