package main

import (
	"context"

	"github.com/ole-larsen/binance-subscriber/internal/server"
	"github.com/ole-larsen/binance-subscriber/internal/server/config"
)

// main runs the build information and prints it to the provided writer.
func main() {
	ctx, cancel := context.WithCancel(context.Background())

	settings := config.GetConfig()

	srv, err := server.SetupFunc(settings)
	if err != nil {
		panic(err)
	}

	srv.Run(ctx, cancel)
}
