package main

import (
	log "github.com/sirupsen/logrus"
	"go.uber.org/zap"

	"github.com/project/library/config"
	"github.com/project/library/internal/app"
)

func main() {
	cfg, err := config.New()

	if err != nil {
		log.Fatalf("can not get application config: %s", err)
	}

	var logger *zap.Logger

	logger, err = zap.NewProduction()

	if err != nil {
		log.Fatalf("can not initialize logger: %s", err)
	}

	app.Run(logger, cfg)
}
