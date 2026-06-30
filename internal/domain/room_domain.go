package domain

import (
	"time"

	"gorm.io/gorm"
)

type Room struct {
	Id          int            `gorm:"column:id;primaryKey;autoIncrement"`
	Name        string         `gorm:"column:name"`
	Description string         `gorm:"column:description"`
	CreatedAt   time.Time      `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt   *time.Time     `gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt   gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

type RoomDto struct {
	Id          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type RoomPaginateDto[T any] struct {
	Data       []*RoomDto
	Pagination T
}

type CursorPaginateRequest struct {
	Cursor int
	Limit  int
}

type RoomCreateDto struct {
	Name        string `json:"name" validate:"required,min=3,max=50"`
	Description string `json:"description" validate:"required,gt=0"`
}

type RoomUpdateDto struct {
	Name        string `json:"name" validate:"required,min=3,max=50"`
	Description string `json:"description" validate:"required,gt=0"`
}

func ToRoomDto(room *Room) *RoomDto {
	return &RoomDto{
		Id:          room.Id,
		Name:        room.Name,
		Description: room.Description,
	}
}
