package main

import (
	"context"

	"github.com/k-zavarnitsyn/gophermart/internal/app"
	"github.com/k-zavarnitsyn/gophermart/internal/config"
	"github.com/k-zavarnitsyn/gophermart/internal/container"
	log "github.com/sirupsen/logrus"
)

func main() {
	cfg, err := config.Load(config.DefaultDir, config.WithServerFlags(), config.WithAuth())
	if err != nil {
		log.Fatal(err)
	}

	cnt := container.New(cfg)
	server := app.NewServerApp(cfg, cnt)
	server.Run(context.Background())
}
