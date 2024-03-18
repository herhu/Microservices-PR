package main

import (
	"flag"
	"log"

	"github.com/herhu/Microservices-PR/pkg/logger"
	"github.com/herhu/Microservices-PR/writer_service/config"
	"github.com/herhu/Microservices-PR/writer_service/internal/server"
)

func main() {
	flag.Parse()

	cfg, err := config.InitConfig()
	if err != nil {
		log.Fatal(err)
	}

	appLogger := logger.NewAppLogger(cfg.Logger)
	appLogger.InitLogger()
	appLogger.WithName("WriterService")

	s := server.NewServer(appLogger, cfg)
	appLogger.Fatal(s.Run())
}
