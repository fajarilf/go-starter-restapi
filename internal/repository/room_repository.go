package repository

import (
	"context"

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
		return nil, err
	}

	room, err := pgx.CollectOneRow(rows, pgx.RowToAddrOfStructByName[domain.Room])
	if err != nil {
		return nil, err
	}

	return room, nil
}

func (r *RoomRepository) GetById(ctx context.Context, id int) (*domain.Room, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, name, description, created_at, updated_at, deleted_at
		 FROM rooms WHERE id = $1 AND deleted_at IS NULL`,
		id,
	)
	if err != nil {
		return nil, err
	}

	room, err := pgx.CollectOneRow(rows, pgx.RowToAddrOfStructByName[domain.Room])
	if err != nil {
		return nil, err
	}

	return room, nil
}

func (r *RoomRepository) Update(ctx context.Context, entity *domain.Room) (*domain.Room, error) {
	rows, err := r.db.Query(ctx,
		`UPDATE rooms
		 SET name = $1, description = $2, updated_at = now()
		 WHERE id = $3 AND deleted_at IS NULL
		 RETURNING *`,
		entity.Name, entity.Description, entity.Id,
	)
	if err != nil {
		return nil, err
	}

	room, err := pgx.CollectOneRow(rows, pgx.RowToAddrOfStructByName[domain.Room])
	if err != nil {
		return nil, err
	}

	return room, nil
}

func (r *RoomRepository) Delete(ctx context.Context, id int) (int64, error) {
	tag, err := r.db.Exec(ctx,
		`UPDATE rooms SET deleted_at = now() WHERE id = $1 AND deleted_at IS NULL`,
		id,
	)
	if err != nil {
		return tag.RowsAffected(), err
	}

	return tag.RowsAffected(), nil
}

func (r *RoomRepository) Get(ctx context.Context, param *domain.PaginateRequest) ([]*domain.Room, domain.Pagination, error) {
	offset := (param.Page - 1) * param.Limit

	var total int
	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM rooms WHERE deleted_at IS NULL`).Scan(&total); err != nil {
		return nil, domain.Pagination{}, err
	}

	rows, err := r.db.Query(ctx,
		`SELECT id, name, description, created_at, updated_at, deleted_at
		 FROM rooms
		 WHERE deleted_at IS NULL
		 ORDER BY created_at DESC
		 LIMIT $1 OFFSET $2`,
		param.Limit, offset,
	)
	if err != nil {
		return nil, domain.Pagination{}, err
	}

	rooms, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[domain.Room])
	if err != nil {
		return nil, domain.Pagination{}, err
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

func (r *RoomRepository) GetByCursor(ctx context.Context, param *domain.CursorPaginateRequest) ([]*domain.Room, domain.CursorPagination, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, name, description, created_at, updated_at, deleted_at
		 FROM rooms
		 WHERE deleted_at IS NULL AND id < $1
		 ORDER BY id DESC
		 LIMIT $2`,
		param.Cursor, param.Limit+1,
	)
	if err != nil {
		return nil, domain.CursorPagination{}, err
	}

	rooms, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[domain.Room])
	if err != nil {
		return nil, domain.CursorPagination{}, err
	}

	hasNext := len(rooms) > param.Limit
	if hasNext {
		rooms = rooms[:param.Limit]
	}

	var nextCursor *int
	if hasNext && len(rooms) > 0 {
		last := rooms[len(rooms)-1]
		nextCursor = &last.Id
	}

	return rooms, domain.CursorPagination{
		NextCursor: nextCursor,
		HasNext:    hasNext,
		Limit:      param.Limit,
	}, nil
}

func (r *RoomRepository) Recover(ctx context.Context, id int) (*domain.Room, error) {
	rows, err := r.db.Query(ctx,
		`UPDATE rooms
		 SET deleted_at = NULL, updated_at = now()
		 WHERE id = $1
		 RETURNING *`,
		id,
	)
	if err != nil {
		return nil, err
	}

	room, err := pgx.CollectOneRow(rows, pgx.RowToAddrOfStructByName[domain.Room])
	if err != nil {
		return nil, err
	}

	return room, nil
}
