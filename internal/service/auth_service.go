package service

import (
	"context"
	"sync"
	"time"

	"github.com/fajarilf/go-starter-api/internal/config"
	"github.com/fajarilf/go-starter-api/internal/domain"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

// ponytail: in-memory blacklist, lost on restart. Replace with Redis/DB table
// if persistent revocation is needed.
type AuthService struct {
	db        *pgxpool.Pool
	cfg       config.Config
	blacklist sync.Map // map[string]bool — jti → revoked
}

func NewAuthService(db *pgxpool.Pool, cfg config.Config) *AuthService {
	return &AuthService{db: db, cfg: cfg}
}

// IsRevoked returns true if the jti has been logged out.
func (s *AuthService) IsRevoked(jti string) bool {
	_, ok := s.blacklist.Load(jti)
	return ok
}

func (s *AuthService) Login(ctx context.Context, username, password string) (string, error) {
	var id int
	var hash string
	err := s.db.QueryRow(ctx,
		`SELECT id, password_hash FROM users WHERE username = $1`, username,
	).Scan(&id, &hash)
	if err != nil {
		return "", domain.ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		return "", domain.ErrInvalidCredentials
	}

	claims := domain.JWTClaims{
		UserID:   id,
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
