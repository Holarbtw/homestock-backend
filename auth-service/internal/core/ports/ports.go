package ports

import (
	"context"
	"time"

	"github.com/Holarbtw/homestock-backend/auth-service/internal/domain"
	"github.com/google/uuid"
)

type TxManager interface {
	WithinTransaction(ctx context.Context, f func(ctx context.Context) error) error
}

type UserRepository interface {
	Create(ctx context.Context, email, passwordHash, nickname, timezone string) (*domain.User, error)
	CreateGuest(ctx context.Context, nickname, timezone string) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	GetByIDForUpdate(ctx context.Context, id uuid.UUID) (*domain.User, error)
	LinkGuestAccount(ctx context.Context, userID uuid.UUID, email, passwordHash, nickname string) error

	CreateRefreshToken(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) (*domain.RefreshToken, error)
	GetRefreshToken(ctx context.Context, tokenHash string) (*domain.RefreshToken, error)
	RevokeRefreshToken(ctx context.Context, tokenHash string) error
	RevokeAllUserTokens(ctx context.Context, userID uuid.UUID) error
}

type AuthService interface {
	Register(ctx context.Context, email, password, nickname, timezone string) (*domain.TokenPair, *domain.User, error)
	Login(ctx context.Context, email, password string) (*domain.TokenPair, *domain.User, error)
	LoginGuest(ctx context.Context, nickname, timezone string) (*domain.TokenPair, *domain.User, error)
	RefreshToken(ctx context.Context, refreshToken string) (*domain.TokenPair, error)
	Logout(ctx context.Context, refreshToken string) error
}

type PasswordHasher interface {
    Hash(password string) (string, error)
    Compare(hash, password string) error
}

type TokenManager interface {
    GenerateAccessToken(user *domain.User) (token string, expiresAt time.Time, err error)
    GenerateRefreshToken() (rawToken string, tokenHash string, expiresAt time.Time, err error)
	HashRefreshToken(rawToken string) string
}