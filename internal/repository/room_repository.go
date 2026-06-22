package repository

import (
	"context"
	"fmt"

	"github.com/fajarilf/go-starter-api/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RoomRepository struct {
	db *pgxpool.Pool
}

var _ RoomRepositoryInterface = (*RoomRepository)(nil)

func NewRoomRepository(db *pgxpool.Pool) *RoomRepository {
	return &RoomRepository{
		db: db,
	}
}

func (r *RoomRepository) Create(ctx context.Context, entity *domain.Room) (*domain.Room, error) {
	rows, err := r.db.Query(ctx,
		`INSERT INTO rooms (name, description)
		 VALUES ($1, $2)
		 RETURNING *`,
		entity.Name, entity.Description,
	)
	if err != nil {
		return nil, fmt.Errorf("Room Repository: %w", err)
	}

	room, err := pgx.CollectOneRow(rows, pgx.RowToAddrOfStructByName[domain.Room])
	if err != nil {
		return nil, fmt.Errorf("Room Repository: %w", err)
	}

	return room, nil
}

func (r *RoomRepository) GetById(ctx context.Context, id int) (*domain.Room, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, name, description, created_at, updated_at
		 FROM rooms WHERE id = $1`,
		id,
	)
	if err != nil {
		return nil, fmt.Errorf("Room Repository: %w", err)
	}

	room, err := pgx.CollectOneRow(rows, pgx.RowToAddrOfStructByName[domain.Room])
	if err != nil {
		return nil, fmt.Errorf("Room Repository: %w", err)
	}

	return room, nil
}

func (r *RoomRepository) Get(ctx context.Context, param *domain.PaginateRequest) ([]*domain.Room, domain.Pagination, error) {
	offset := (param.Page - 1) * param.Limit

	var total int
	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM rooms`).Scan(&total); err != nil {
		return nil, domain.Pagination{}, fmt.Errorf("Room Repository: %w", err)
	}

	rows, err := r.db.Query(ctx,
		`SELECT id, name, description, created_at, updated_at
		 FROM rooms
		 ORDER BY created_at DESC
		 LIMIT $1 OFFSET $2`,
		param.Limit, offset,
	)
	if err != nil {
		return nil, domain.Pagination{}, fmt.Errorf("Room Repository: %w", err)
	}

	rooms, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[domain.Room])
	if err != nil {
		return nil, domain.Pagination{}, fmt.Errorf("Room Repository: %w", err)
	}

	totalPages := 0
	if param.Limit > 0 {
		totalPages = (total + param.Limit - 1) / param.Limit // ceil(total / limit)
	}

	return rooms, domain.Pagination{
		Page:       param.Page,
		Limit:      param.Limit,
		TotalPages: totalPages,
		Total:      total,
		HasPrev:    param.Page > 1,
		HasNext:    param.Page < totalPages,
	}, nil
}
