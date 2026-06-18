package repository

import (
	"context"
	"fmt"

	"github.com/fajarilf/go-starter-api/internal/domain"
	"github.com/fajarilf/go-starter-api/internal/service"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RoomRepository struct {
	db *pgxpool.Pool
}

var _ service.RoomRepository = (*RoomRepository)(nil)

func NewRoomRepository(db *pgxpool.Pool) *RoomRepository {
	return &RoomRepository{
		db: db,
	}
}

func (r *RoomRepository) Create(ctx context.Context, entity *domain.Room) (*domain.Room, error) {
	return nil, fmt.Errorf("Method Unimplemented")
}

func (r *RoomRepository) GetById(ctx context.Context, id int) (*domain.Room, error) {
	return nil, fmt.Errorf("Method Unimplemented")
}
