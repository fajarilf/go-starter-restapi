package domain

import "github.com/golang-jwt/jwt/v5"

type RegisterRequest struct {
	Username string `json:"username" validate:"required,min=3,max=50"`
	Password string `json:"password" validate:"required,min=6"`
}

type RegisterResponse struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
}

type User struct {
	ID           int    `gorm:"column:id;primaryKey;autoIncrement"`
	Username     string `gorm:"column:username"`
	PasswordHash string `gorm:"column:password_hash"`
}

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

var ErrUsernameTaken = &AppError{Kind: KindConflict, Message: "username already taken"}
var ErrInvalidCredentials = &AppError{Kind: KindValidation, Message: "invalid username or password"}
var ErrUnauthorized = &AppError{Kind: KindUnauthorized, Message: "missing or invalid token"}
