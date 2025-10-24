package server

import "github.com/gin-gonic/gin"

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
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
	c.JSON(200, gin.H{"Hello": "World"})
}

type Request struct {
	ApiKey   string          `json:"api_key"`
	Provider string          `json:"provider" binding:"required,oneOf=openAi gemini"`
	Id       string          `json:"id"`
	Context  GenerateContext `json:"context"`
}

type GenerateContext struct {
	UserCode string `json:"user_code"`
}

type Testifai struct {
	TestType string `json:"test_type" binding:"required,oneOf=xunit table suite"`
}

type Response struct {
	Id        string    `json:"id"`
	Generated Generated `json:"generated"`
}

type Generated struct {
	TestCode string `json:"test_code"`
}
