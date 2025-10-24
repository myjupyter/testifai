package client

import (
	"errors"
	"fmt"
	"net/http"

	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	"github.com/google/uuid"
	"github.com/myjupyter/testifai/src/frontend/client/client"
	"github.com/myjupyter/testifai/src/frontend/client/client/generator"
	"github.com/myjupyter/testifai/src/frontend/client/models"
)

type RequestData struct {
	Host     string
	UserCode string
	Token    string
	Provider string
	TestType string
	APIKey   string
	Platform string
}

type Result struct {
	GeneratedTest string
}

func SendRequest(reqData RequestData) (Result, error) {
	transport := httptransport.New(reqData.Host, "", []string{"http"})
	apiClient := client.New(transport, strfmt.Default)
	generate, err := apiClient.Generator.PostGenerate(&generator.PostGenerateParams{
		Request: &models.ServerRequest{
			APIKey: reqData.APIKey,
			Context: &models.ServerGenerateContext{
				Platform: reqData.Platform,
				UserCode: reqData.UserCode,
			},
			ID:       uuid.New().String(),
			Provider: reqData.Provider,
			Testifai: &models.ServerTestifai{
				TestType: reqData.TestType,
			},
		},
	})
	if err != nil {
		return Result{}, fmt.Errorf("error sending request: %w", err)
	}

	if generate == nil {
		return Result{}, errors.New("nil response")
	}

	if !generate.IsCode(http.StatusOK) {
		return Result{}, fmt.Errorf("error code %d is not 200", generate.Code())
	}

	if generate.Payload == nil {
		return Result{}, errors.New("nil payload")
	}

	if generate.Payload.Generated == nil {
		return Result{}, errors.New("nil generated payload")
	}

	return Result{
		GeneratedTest: generate.Payload.Generated.TestCode,
	}, nil

}
