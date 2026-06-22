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

	room, err := h.service.Create(r.Context(), &req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeSuccess(w, http.StatusCreated, room)
}

func (h *RoomHandler) GetById(w http.ResponseWriter, r *http.Request) {
	idString := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idString)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	room, err := h.service.GetById(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeSuccess(w, http.StatusOK, room)
}

func (h *RoomHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	var req domain.RoomUpdateDto
	if !decodeJSON(w, r, &req) {
		return
	}

	if err := h.validate.Struct(req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	room, err := h.service.Update(r.Context(), id, &req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeSuccess(w, http.StatusOK, room)
}

func (h *RoomHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.service.Delete(r.Context(), id); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeSuccess(w, http.StatusOK, map[string]string{"message": "room deleted"})
}

func (h *RoomHandler) Get(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	param := domain.PaginateRequest{
		Page:  queryInt(q, "page", 1),
		Limit: queryInt(q, "limit", 10),
	}

	rooms, err := h.service.Get(r.Context(), &param)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writePaginate(w, http.StatusOK, rooms.Data, rooms.Pagination)
}
