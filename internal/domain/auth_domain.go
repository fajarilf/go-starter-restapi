package domain

import "github.com/golang-jwt/jwt/v5"

type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

type JWTClaims struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

var ErrInvalidCredentials = &AppError{Kind: KindValidation, Message: "invalid username or password"}
var ErrUnauthorized = &AppError{Kind: KindValidation, Message: "missing or invalid token"}
