package main

import (
	"log"
	"testifai/src/backend/config"
	"testifai/src/backend/server"
)

// @title          Testifai backend
// @version         1.0
// @termsOfService  http://swagger.io/terms/
// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  iauglov@gmail.com
// @BasePath  /
func main() {
	cfg := &config.Config{
		Host:       "127.0.0.1",
		ListenAddr: ":6667",
	}
	handler := server.NewHandler()
	srv := server.NewServer(cfg, handler)
	if err := srv.Run(); err != nil {
		log.Fatal(err)
	}
}
