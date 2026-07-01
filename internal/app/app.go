package app

import (
	"context"
	"log/slog"

	"github.com/fajarilf/go-starter-api/internal/config"
	"github.com/fajarilf/go-starter-api/internal/handler"
	"github.com/fajarilf/go-starter-api/internal/logger"
	"github.com/fajarilf/go-starter-api/internal/repository"
	"github.com/fajarilf/go-starter-api/internal/server"
	"github.com/fajarilf/go-starter-api/internal/service"
	"github.com/go-playground/validator/v10"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type App struct {
	Server *server.Server
	db     *gorm.DB
}

func NewApp(ctx context.Context, cfg config.Config) (*App, error) {
	log, err := logger.Setup(cfg.Environment, cfg.LogLevel, cfg.LogDir)
	if err != nil {
		return nil, err
	}
	slog.SetDefault(log)

	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	if err := sqlDB.PingContext(ctx); err != nil {
		sqlDB.Close()
		return nil, err
	}

	v := validator.New()

	roomRepo := repository.NewRoomRepository(db)
	roomService := service.NewRoomService(roomRepo, v)
	roomHandler := handler.NewRoomHandler(roomService)

	userRepo := repository.NewUserRepository(db)
	authService := service.NewAuthService(userRepo, v, cfg)
	authHandler := handler.NewAuthHandler(authService)

	srv := server.New(cfg, roomHandler, authHandler, authService)
	return &App{Server: srv, db: db}, nil
}

func (a *App) Close() {
	if sqlDB, err := a.db.DB(); err == nil {
		sqlDB.Close()
	}
}
