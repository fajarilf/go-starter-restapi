package service

import (
	"context"
	"fmt"

	"github.com/fajarilf/go-starter-api/internal/domain"
)

type RoomRepository interface {
	Create(ctx context.Context, entity *domain.Room) (*domain.Room, error)
	GetById(ctx context.Context, id int) (*domain.Room, error)
}

type RoomService struct {
	repo RoomRepository
}

func NewRoomService(r RoomRepository) *RoomService {
	return &RoomService{
		repo: r,
	}
}

func (s *RoomService) Create(ctx context.Context, req *domain.RoomCreateDto) (*domain.RoomDto, error) {
	result, err := s.repo.Create(ctx, &domain.Room{
		Name:        req.Name,
		Description: req.Description,
	})

	if err != nil {
		return nil, fmt.Errorf("error: %w", err)
	}

	return &domain.RoomDto{
		Id:          result.Id,
		Name:        result.Name,
		Description: result.Description,
	}, nil
}

func (s *RoomService) GetById(ctx context.Context, id int) (*domain.RoomDto, error) {
	result, err := s.repo.GetById(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("error: %w", err)
	}

	return &domain.RoomDto{
		Id:          result.Id,
		Name:        result.Name,
		Description: result.Description,
	}, nil
}
