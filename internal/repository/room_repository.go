package repository

import (
	"context"
	"time"

	"github.com/fajarilf/go-starter-api/internal/domain"
	"gorm.io/gorm"
)

type RoomRepository struct {
	db *gorm.DB
}

var _ RoomRepositoryInterface = (*RoomRepository)(nil)

func NewRoomRepository(db *gorm.DB) *RoomRepository {
	return &RoomRepository{
		db: db,
	}
}

func (r *RoomRepository) Create(ctx context.Context, entity *domain.Room) (*domain.Room, error) {
	if err := r.db.WithContext(ctx).Create(entity).Error; err != nil {
		return nil, err
	}
	return entity, nil
}

func (r *RoomRepository) GetById(ctx context.Context, id int) (*domain.Room, error) {
	var room domain.Room
	if err := r.db.WithContext(ctx).First(&room, id).Error; err != nil {
		return nil, err
	}
	return &room, nil
}

func (r *RoomRepository) Update(ctx context.Context, entity *domain.Room) (*domain.Room, error) {
	now := time.Now()
	entity.UpdatedAt = &now
	tx := r.db.WithContext(ctx).Model(&domain.Room{}).
		Where("id = ? AND deleted_at IS NULL", entity.Id).
		Updates(map[string]interface{}{
			"name":        entity.Name,
			"description": entity.Description,
			"updated_at":  entity.UpdatedAt,
		})
	if tx.Error != nil {
		return nil, tx.Error
	}
	if tx.RowsAffected == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return entity, nil
}

func (r *RoomRepository) Delete(ctx context.Context, id int) (int64, error) {
	tx := r.db.WithContext(ctx).Delete(&domain.Room{}, id)
	return tx.RowsAffected, tx.Error
}

func (r *RoomRepository) Get(ctx context.Context, param *domain.PaginateRequest) ([]*domain.Room, domain.Pagination, error) {
	var total int64
	if err := r.db.WithContext(ctx).Model(&domain.Room{}).Count(&total).Error; err != nil {
		return nil, domain.Pagination{}, err
	}

	offset := (param.Page - 1) * param.Limit
	var rooms []*domain.Room
	if err := r.db.WithContext(ctx).
		Order("created_at DESC").
		Offset(offset).Limit(param.Limit).
		Find(&rooms).Error; err != nil {
		return nil, domain.Pagination{}, err
	}

	t := int(total)
	totalPages := 0
	if param.Limit > 0 {
		totalPages = (t + param.Limit - 1) / param.Limit
	}

	return rooms, domain.Pagination{
		Page:       param.Page,
		Limit:      param.Limit,
		TotalPages: totalPages,
		Total:      t,
		HasPrev:    param.Page > 1,
		HasNext:    param.Page < totalPages,
	}, nil
}

func (r *RoomRepository) GetByCursor(ctx context.Context, param *domain.CursorPaginateRequest) ([]*domain.Room, domain.CursorPagination, error) {
	var rooms []*domain.Room
	if err := r.db.WithContext(ctx).
		Where("id < ?", param.Cursor).
		Order("id DESC").
		Limit(param.Limit + 1).
		Find(&rooms).Error; err != nil {
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
	var room domain.Room
	if err := r.db.WithContext(ctx).Unscoped().First(&room, id).Error; err != nil {
		return nil, err
	}
	room.DeletedAt = gorm.DeletedAt{}
	if err := r.db.WithContext(ctx).Unscoped().Save(&room).Error; err != nil {
		return nil, err
	}
	return &room, nil
}
