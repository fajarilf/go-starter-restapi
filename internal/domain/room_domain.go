package domain

import (
	"database/sql"
	"time"
)

type Room struct {
	Id          int          `db:"id"`
	Name        string       `db:"name"`
	Description string       `db:"description"`
	CreatedAt   time.Time    `db:"created_at"`
	UpdatedAt   sql.NullTime `db:"updated_at"`
	DeletedAt   sql.NullTime `db:"deleted_at"`
}

type RoomDto struct {
	Id          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type RoomPaginateDto struct {
	Data       []*RoomDto
	Pagination Pagination
}

type CursorPaginateRequest struct {
	Cursor int
	Limit  int
}

type RoomCursorPaginateDto struct {
	Data       []*RoomDto
	Pagination CursorPagination
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
