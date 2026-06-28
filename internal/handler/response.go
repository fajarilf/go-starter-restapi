package handler

import (
	"encoding/json"
	"errors"
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

// writeServiceError maps an error returned by the service layer to a semantic
// HTTP status. Known sentinel errors get specific codes; anything else is an
// unexpected failure (500) and is logged rather than leaked to the client.
func writeServiceError(w http.ResponseWriter, err error) {
	var appError *domain.AppError
	if errors.As(err, &appError) {
		switch appError.Kind {
		case domain.KindNotFound:
			writeError(w, http.StatusNotFound, appError.Message)
			return
		case domain.KindValidation:
			writeError(w, http.StatusBadRequest, appError.Message)
			return
		case domain.KindConflict:
			writeError(w, http.StatusConflict, appError.Message)
			return
		case domain.KindUnauthorized:
			writeError(w, http.StatusUnauthorized, appError.Message)
			return
		}
	}

	// Anything else (including KindInternal) is unexpected: log it, don't leak it.
	slog.Error("unexpected service error", "error", err)
	writeError(w, http.StatusInternalServerError, "internal server error")
}

func decodeJSON(r *http.Request, dst any) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(dst); err != nil {
		return domain.NewValidationError("invalid request body")
	}

	return nil
}
