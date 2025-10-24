package server

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/myjupyter/testifai/api"
	"github.com/myjupyter/testifai/src/backend/config"
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
	engine.Use(gin.Recovery()).
		Use(gin.Logger()).
		Use(cors.Default())
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
	port := s.cfg.ListenAddr
	if len(port) > 0 && port[0] == ':' {
		port = port[1:]
	}
	api.SwaggerInfo.Host = s.cfg.Host + ":" + port
	s.engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
}
