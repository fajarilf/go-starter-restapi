package service

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/fajarilf/go-starter-api/internal/config"
	"github.com/fajarilf/go-starter-api/internal/domain"
	"github.com/fajarilf/go-starter-api/internal/repository"
	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// ponytail: in-memory blacklist, lost on restart. Replace with Redis/DB table
// if persistent revocation is needed.
type AuthService struct {
	userRepo  repository.UserRepositoryInterface
	cfg       config.Config
	validate  *validator.Validate
	blacklist sync.Map // map[string]bool — jti → revoked
}

func NewAuthService(userRepo repository.UserRepositoryInterface, v *validator.Validate, cfg config.Config) *AuthService {
	return &AuthService{userRepo: userRepo, cfg: cfg, validate: v}
}

// IsRevoked returns true if the jti has been logged out.
func (s *AuthService) IsRevoked(jti string) bool {
	_, ok := s.blacklist.Load(jti)
	return ok
}

func (s *AuthService) Register(ctx context.Context, username, password string) (*domain.RegisterResponse, error) {
	req := domain.RegisterRequest{Username: username, Password: password}
	if err := s.validate.Struct(req); err != nil {
		return nil, domain.NewValidationError(err.Error())
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, domain.NewInternalError("failed to hash password")
	}

	user, err := s.userRepo.Create(ctx, username, string(hash))
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, domain.ErrUsernameTaken
		}
		return nil, domain.NewInternalError(err.Error())
	}

	return &domain.RegisterResponse{ID: user.ID, Username: user.Username}, nil
}

func (s *AuthService) Login(ctx context.Context, username, password string) (string, error) {
	req := domain.LoginRequest{Username: username, Password: password}
	if err := s.validate.Struct(req); err != nil {
		return "", domain.NewValidationError(err.Error())
	}

	user, err := s.userRepo.FindByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", domain.ErrInvalidCredentials
		}
		return "", domain.NewInternalError(err.Error())
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", domain.ErrInvalidCredentials
	}

	claims := domain.JWTClaims{
		UserID:   user.ID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.NewString(),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(s.cfg.JWTExpiryHours) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(s.cfg.JWTSecret))
	if err != nil {
		return "", domain.NewInternalError("failed to sign token")
	}

	return signed, nil
}

func (s *AuthService) ValidateToken(tokenStr string) (*domain.JWTClaims, error) {
	claims := &domain.JWTClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims,
		func(t *jwt.Token) (any, error) { return []byte(s.cfg.JWTSecret), nil },
	)
	if err != nil || !token.Valid {
		return nil, domain.ErrUnauthorized
	}
	if s.IsRevoked(claims.ID) {
		return nil, domain.ErrUnauthorized
	}
	return claims, nil
}

// Logout revokes the token by its JWT ID. Tokens are blacklisted until
// restart. The token must be valid (signature + expiry) to be revoked.
func (s *AuthService) Logout(tokenStr string) error {
	claims, err := s.ValidateToken(tokenStr)
	if err != nil {
		return err
	}
	s.blacklist.Store(claims.ID, true)
	return nil
}
