package handler

import (
	"net/http"
	"strconv"

	"github.com/fajarilf/go-starter-api/internal/domain"
	"github.com/fajarilf/go-starter-api/internal/service"
	"github.com/go-chi/chi/v5"
)

type RoomHandler struct {
	service *service.RoomService
}

func NewRoomHandler(s *service.RoomService) *RoomHandler {
	return &RoomHandler{
		service: s,
	}
}

func (h *RoomHandler) Create(w http.ResponseWriter, r *http.Request) error {
	var req domain.RoomCreateDto
	if err := decodeJSON(r, &req); err != nil {
		return err
	}

	room, err := h.service.Create(r.Context(), &req)
	if err != nil {
		return err
	}

	writeSuccess(w, http.StatusCreated, room)
	return nil
}

func (h *RoomHandler) GetById(w http.ResponseWriter, r *http.Request) error {
	id, err := roomID(r)
	if err != nil {
		return err
	}

	room, err := h.service.GetById(r.Context(), id)
	if err != nil {
		return err
	}

	writeSuccess(w, http.StatusOK, room)
	return nil
}

func (h *RoomHandler) Update(w http.ResponseWriter, r *http.Request) error {
	id, err := roomID(r)
	if err != nil {
		return err
	}

	var req domain.RoomUpdateDto
	if err := decodeJSON(r, &req); err != nil {
		return err
	}

	room, err := h.service.Update(r.Context(), id, &req)
	if err != nil {
		return err
	}

	writeSuccess(w, http.StatusOK, room)
	return nil
}

func (h *RoomHandler) Delete(w http.ResponseWriter, r *http.Request) error {
	id, err := roomID(r)
	if err != nil {
		return err
	}

	if err := h.service.Delete(r.Context(), id); err != nil {
		return err
	}

	writeSuccess(w, http.StatusOK, map[string]string{"message": "room deleted"})
	return nil
}

func (h *RoomHandler) Get(w http.ResponseWriter, r *http.Request) error {
	q := r.URL.Query()
	param := domain.PaginateRequest{
		Page:  queryInt(q, "page", 1),
		Limit: queryInt(q, "limit", 10),
	}

	rooms, err := h.service.Get(r.Context(), &param)
	if err != nil {
		return err
	}

	writePaginate(w, http.StatusOK, rooms.Data, rooms.Pagination)
	return nil
}

func (h *RoomHandler) GetByCursor(w http.ResponseWriter, r *http.Request) error {
	q := r.URL.Query()
	param := domain.CursorPaginateRequest{
		Cursor: queryInt(q, "cursor", 0),
		Limit:  queryInt(q, "limit", 10),
	}

	// When cursor is 0 (first page), use MaxInt so all rows qualify for id < $1
	if param.Cursor == 0 {
		param.Cursor = 1<<31 - 1
	}

	rooms, err := h.service.GetByCursor(r.Context(), &param)
	if err != nil {
		return err
	}

	writeCursorPaginate(w, http.StatusOK, rooms.Data, rooms.Pagination)
	return nil
}

func (h *RoomHandler) Recover(w http.ResponseWriter, r *http.Request) error {
	id, err := roomID(r)
	if err != nil {
		return err
	}

	room, err := h.service.Recover(r.Context(), id)
	if err != nil {
		return err
	}

	writeSuccess(w, http.StatusOK, room)
	return nil
}

// roomID parses the {id} path param, returning a validation error on a
// non-numeric value so Wrap maps it to a 400.
func roomID(r *http.Request) (int, error) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		return 0, domain.NewValidationError("invalid id")
	}
	return id, nil
}
