package main

import (
	"flag"
	"log"

	"github.com/herhu/Microservices-PR/api_gateway_service/config"
	"github.com/herhu/Microservices-PR/api_gateway_service/internal/server"
	"github.com/herhu/Microservices-PR/pkg/logger"
)

func main() {
	flag.Parse()

	cfg, err := config.InitConfig()
	if err != nil {
		log.Fatal(err)
	}

	appLogger := logger.NewAppLogger(cfg.Logger)
	appLogger.InitLogger()
	appLogger.WithName("ApiGateway")

	s := server.NewServer(appLogger, cfg)
	appLogger.Fatal(s.Run())
}
