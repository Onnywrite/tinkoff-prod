package main

import (
	"context"
	"os"
	"os/signal"

	"github.com/Onnywrite/tinkoff-prod/internal/app"
	"github.com/Onnywrite/tinkoff-prod/internal/config"
)

func main() {
	cfg := config.MustLoad("/etc/service/config/ignore-config.yaml")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	application := app.New(cfg)
	application.MustStart(ctx)

	// gracefull shutdown
	shut := make(chan os.Signal, 1)
	signal.Notify(shut, os.Interrupt)
	<-shut

	application.MustStop()
}
