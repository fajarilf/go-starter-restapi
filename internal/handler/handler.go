package handler

import "net/http"

// APIFunc is a handler that returns an error instead of writing it. The Wrap
// adapter centralizes error -> HTTP mapping via writeServiceError, so handlers
// stay free of repetitive error-writing boilerplate.
type APIFunc func(w http.ResponseWriter, r *http.Request) error

func Wrap(fn APIFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := fn(w, r); err != nil {
			writeServiceError(w, err)
		}
	}
}
