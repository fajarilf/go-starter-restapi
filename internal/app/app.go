package app

import (
	"context"

	"github.com/fajarilf/go-starter-api/internal/config"
	"github.com/fajarilf/go-starter-api/internal/handler"
	"github.com/fajarilf/go-starter-api/internal/repository"
	"github.com/fajarilf/go-starter-api/internal/server"
	"github.com/fajarilf/go-starter-api/internal/service"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"
)

type App struct {
	Server *server.Server
	pool   *pgxpool.Pool
}

func NewApp(ctx context.Context, cfg config.Config) (*App, error) {
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}

	v := validator.New()

	roomRepo := repository.NewRoomRepository(pool)
	roomService := service.NewRoomService(roomRepo, v)
	roomHandler := handler.NewRoomHandler(roomService)

	authService := service.NewAuthService(pool, cfg)
	authHandler := handler.NewAuthHandler(authService)

	srv := server.New(cfg, roomHandler, authHandler, authService)
	return &App{Server: srv, pool: pool}, nil
}

func (a *App) Close() { a.pool.Close() }
