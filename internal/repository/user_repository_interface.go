package repository

import (
	"context"

	"github.com/fajarilf/go-starter-api/internal/domain"
)

type UserRepositoryInterface interface {
	FindByUsername(ctx context.Context, username string) (*domain.User, error)
	Create(ctx context.Context, username, passwordHash string) (*domain.User, error)
}
