package server

import (
	"testifai/api"
	"testifai/src/backend/config"

	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type Server struct {
	cfg     *config.Config
	engine  *gin.Engine
	handler *Handler
}

func NewServer(cfg *config.Config, handler *Handler) *Server {
	engine := gin.Default()
	return &Server{
		cfg:     cfg,
		engine:  engine,
		handler: handler,
	}
}

func (s *Server) Run() error {
	err := s.registerHandlers()
	s.handleSwagger()
	if err != nil {
		return err
	}
	return s.engine.Run(s.cfg.ListenAddr)
}

func (s *Server) registerHandlers() error {
	s.engine.POST("/generate", s.handler.Handle)
	return nil
}

func (s *Server) handleSwagger() {
	api.SwaggerInfo.BasePath = "/"
	api.SwaggerInfo.Host = s.cfg.Host + s.cfg.ListenAddr
	s.engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
}
