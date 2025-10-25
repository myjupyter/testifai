package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/myjupyter/testifai/src/backend/app/model"
	"github.com/myjupyter/testifai/src/backend/service/router"
)

type Handler struct {
	routerSrv *router.Service
}

func NewHandler(routerSrv *router.Service) *Handler {
	return &Handler{
		routerSrv: routerSrv,
	}
}

// Generate godoc
// @Summary Generate tests
// @Schemes
// @Description Метод генерирует и возвращает тесты
// @Tags Generator
// @Param request body Request true "Реквест"
// @Produce json
// @Accept json
// @Success 200 {object} Response
// @Router /generate [post]
func (h *Handler) Handle(c *gin.Context) {
	req := Request{}
	if err := c.BindJSON(&req); err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	generate, err := h.routerSrv.Generate(c, model.GenerateRequest{
		Context: model.GenerateContext{
			UserCode:        req.Context.UserCode,
			UserCodeContext: req.Context.UserCodeContext,
			PackageName:     req.Context.PackageName,
			ExternalImport:  req.Context.ExternalImport,
			Plarform:        req.Context.Plarform,
		},
		Testifai: model.Testifai{
			TestType: testTypeTo(req.Testifai.TestType),
		},
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"err": err.Error()})
		return
	}

	c.JSON(http.StatusOK, Response{
		Id: req.Id,
		Generated: Generated{
			TestCode: generate.Content,
		},
	})

}

func testTypeTo(strTestType string) model.TestType {
	return model.TestType(strTestType)
}

type Request struct {
	Id       string          `json:"id"`
	Context  GenerateContext `json:"context"`
	Testifai Testifai        `json:"testifai"`
}

type GenerateContext struct {
	UserCode        string   `json:"user_code"`
	Plarform        string   `json:"platform"`
	UserCodeContext string   `json:"user_code_context"`
	PackageName     string   `json:"package_name"`
	ExternalImport  []string `json:"external_import"`
}

type Testifai struct {
	TestType string `json:"test_type"`
}

type Response struct {
	Id        string    `json:"id"`
	Generated Generated `json:"generated"`
}

type Generated struct {
	TestCode string `json:"test_code"`
}
