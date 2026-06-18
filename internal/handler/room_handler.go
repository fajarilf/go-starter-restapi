package handler

import (
	"net/http"
	"strconv"

	"github.com/fajarilf/go-starter-api/internal/domain"
	"github.com/fajarilf/go-starter-api/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
)

type RoomHandler struct {
	service  *service.RoomService
	validate *validator.Validate
}

func NewRoomHandler(s *service.RoomService, v *validator.Validate) *RoomHandler {
	return &RoomHandler{
		service:  s,
		validate: v,
	}
}

func (h *RoomHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req domain.RoomCreateDto
	if !decodeJSON(w, r, &req) {
		return // validate if request is json or not
	}

	if err := h.validate.Struct(req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	room, err := h.service.Create(r.Context(), &domain.RoomCreateDto{
		Name:        req.Name,
		Description: req.Description,
	})

	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, room)
}

func (h *RoomHandler) GetById(w http.ResponseWriter, r *http.Request) {
	idString := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idString)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
	}

	room, err := h.service.GetById(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusFound, room)
}
