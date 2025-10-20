package main

import (
	"context"
	"github.bumble.dev/shcherbanich/user-votes-storage/config"
	"github.bumble.dev/shcherbanich/user-votes-storage/internal/app/di"
	"os/signal"
	"syscall"
)

func main() {
	conf := config.Load()
	worker, _ := di.InitializeMessageProcessor(conf)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := worker.Start(ctx); err != nil {
		panic(err.Error())
	}
}
