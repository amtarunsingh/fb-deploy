package main

import (
	"github.bumble.dev/shcherbanich/user-votes-storage/config"
	"github.bumble.dev/shcherbanich/user-votes-storage/internal/app/di"
)

func main() {
	conf := config.Load()
	app, _ := di.InitializeApiWebServer(conf)
	app.Serve()
}
