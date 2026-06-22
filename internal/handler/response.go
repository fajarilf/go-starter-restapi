package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"

	"github.com/fajarilf/go-starter-api/internal/domain"
)

// queryInt reads a positive integer query param, falling back to def when the
// value is missing, non-numeric, or not positive.
func queryInt(values url.Values, key string, def int) int {
	if n, err := strconv.Atoi(values.Get(key)); err == nil && n > 0 {
		return n
	}
	return def
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if v == nil {
		return
	}

	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("Failed to encode response", "error", err)
	}
}

func writePaginate[T any](w http.ResponseWriter, status int, data []T, pagination domain.Pagination) {
	writeJSON(w, status, domain.PaginateResponse[T]{
		Status:     status,
		Data:       data,
		Pagination: pagination,
	})
}

func writeSuccess[T any](w http.ResponseWriter, status int, data T) {
	writeJSON(w, status, domain.SuccessResponse[T]{
		Status: status,
		Data:   data,
	})
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, domain.ErrorResponse{Error: message})
}

func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(dst); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return false
	}

	return true
}
