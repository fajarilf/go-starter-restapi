package handler

import (
	"net/http"
	"strings"

	"github.com/fajarilf/go-starter-api/internal/domain"
	"github.com/fajarilf/go-starter-api/internal/service"
)

type AuthHandler struct {
	svc *service.AuthService
}

func NewAuthHandler(svc *service.AuthService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) error {
	var req domain.RegisterRequest
	if err := decodeJSON(r, &req); err != nil {
		return err
	}

	user, err := h.svc.Register(r.Context(), req.Username, req.Password)
	if err != nil {
		return err
	}

	writeSuccess(w, http.StatusCreated, user)
	return nil
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) error {
	var req domain.LoginRequest
	if err := decodeJSON(r, &req); err != nil {
		return err
	}

	token, err := h.svc.Login(r.Context(), req.Username, req.Password)
	if err != nil {
		return err
	}

	writeSuccess(w, http.StatusOK, domain.LoginResponse{Token: token})
	return nil
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) error {
	header := r.Header.Get("Authorization")
	token, ok := strings.CutPrefix(header, "Bearer ")
	if !ok || token == "" {
		return domain.ErrUnauthorized
	}

	if err := h.svc.Logout(token); err != nil {
		return err
	}

	writeSuccess(w, http.StatusOK, map[string]string{"message": "logged out"})
	return nil
}
