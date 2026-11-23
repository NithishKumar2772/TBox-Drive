package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/NithishKumar2772/TBox-Drive/internal/app"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := app.Run(ctx); err != nil {
		log.Printf("tbox-drive stopped: %v", err)
		os.Exit(1)
	}
}
