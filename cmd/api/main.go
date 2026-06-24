package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/fajarilf/go-starter-api/internal/app"
	"github.com/fajarilf/go-starter-api/internal/config"
	"github.com/joho/godotenv"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	_ = godotenv.Load()
	cfg, err := config.Load()
	if err != nil {
		slog.Error("config", "error", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	application, err := app.NewApp(ctx, cfg)
	if err != nil {
		slog.Error("init", "error", err)
		os.Exit(1)
	}
	defer application.Close()

	if err := application.Server.Start(ctx); err != nil {
		slog.Error("server", "error", err)
		os.Exit(1)
	}
}
