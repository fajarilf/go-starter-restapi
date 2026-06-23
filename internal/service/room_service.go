package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/fajarilf/go-starter-api/internal/domain"
	"github.com/fajarilf/go-starter-api/internal/repository"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5"
)

type RoomService struct {
	repo     repository.RoomRepositoryInterface
	validate *validator.Validate
}

func NewRoomService(r repository.RoomRepositoryInterface, v *validator.Validate) *RoomService {
	return &RoomService{
		repo:     r,
		validate: v,
	}
}

func (s *RoomService) Create(ctx context.Context, req *domain.RoomCreateDto) (*domain.RoomDto, error) {
	if err := s.validate.Struct(req); err != nil {
		return nil, domain.NewValidationError(err.Error())
	}

	result, err := s.repo.Create(ctx, &domain.Room{
		Name:        req.Name,
		Description: req.Description,
	})

	if err != nil {
		return nil, domain.NewInternalError(err.Error())
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
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.NewNotFoundError(err.Error())
		}

		return nil, domain.NewInternalError(err.Error())
	}

	return &domain.RoomDto{
		Id:          result.Id,
		Name:        result.Name,
		Description: result.Description,
	}, nil
}

func (s *RoomService) Update(ctx context.Context, id int, req *domain.RoomUpdateDto) (*domain.RoomDto, error) {
	if err := s.validate.Struct(req); err != nil {
		return nil, domain.NewValidationError(err.Error())
	}

	result, err := s.repo.Update(ctx, &domain.Room{
		Id:          id,
		Name:        req.Name,
		Description: req.Description,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.NewNotFoundError(err.Error())
		}

		return nil, domain.NewInternalError(err.Error())
	}

	return domain.ToRoomDto(result), nil
}

func (s *RoomService) Delete(ctx context.Context, id int) error {
	rowsAffected, err := s.repo.Delete(ctx, id)

	if err != nil {
		return domain.NewInternalError(err.Error())
	}

	if rowsAffected == 0 {
		message := fmt.Sprintf("room id: %d not found", id)
		return domain.NewNotFoundError(message)
	}

	return nil
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
