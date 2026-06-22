package service

import (
	"context"
	"fmt"

	"github.com/fajarilf/go-starter-api/internal/domain"
	"github.com/fajarilf/go-starter-api/internal/repository"
)

type RoomService struct {
	repo repository.RoomRepositoryInterface
}

func NewRoomService(r repository.RoomRepositoryInterface) *RoomService {
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

func (s *RoomService) Get(ctx context.Context, param *domain.PaginateRequest) (*domain.RoomPaginateDto, error) {
	result, pagination, err := s.repo.Get(ctx, param)
	if err != nil {
		return nil, fmt.Errorf("error: %w", err)
	}

	rooms := make([]*domain.RoomDto, 0, len(result))
	for _, val := range result {
		rooms = append(rooms, domain.ToRoomDto(val))
	}

	return &domain.RoomPaginateDto{
		Data:       rooms,
		Pagination: pagination,
	}, nil
}
