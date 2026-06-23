package repository

import (
	"context"

	"github.com/fajarilf/go-starter-api/internal/domain"
)

type RoomRepositoryInterface interface {
	Create(ctx context.Context, entity *domain.Room) (*domain.Room, error)
	GetById(ctx context.Context, id int) (*domain.Room, error)
	Get(ctx context.Context, param *domain.PaginateRequest) ([]*domain.Room, domain.Pagination, error)
	Update(ctx context.Context, entity *domain.Room) (*domain.Room, error)
	Delete(ctx context.Context, id int) (int64, error)
	Recover(ctx context.Context, id int) (*domain.Room, error)
}
