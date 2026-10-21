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
	LinkGuestAccount(ctx context.Context, userID uuid.UUID, email, passwordHash string) (*domain.User, error)

	CreateRefreshToken(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) error
	GetRefreshTokenForUpdate(ctx context.Context, tokenHash string) (*domain.RefreshToken, error)
	RevokeRefreshToken(ctx context.Context, tokenHash string) error
	RevokeAllUserTokens(ctx context.Context, userID uuid.UUID) error

	UpdatePassword(ctx context.Context, userID uuid.UUID, passwordHash string) error
	UpdateEmail(ctx context.Context, userID uuid.UUID, email string) (*domain.User, error)
	UpdateProfile(ctx context.Context, userID uuid.UUID, nickname, timezone string) (*domain.User, error)
}

type AuthService interface {
	Register(ctx context.Context, email, password, nickname, timezone string) (*domain.TokenPair, *domain.User, error)
	Login(ctx context.Context, email, password string) (*domain.TokenPair, *domain.User, error)
	LoginGuest(ctx context.Context, nickname, timezone string) (*domain.TokenPair, *domain.User, error)
	RefreshToken(ctx context.Context, refreshToken string) (*domain.TokenPair, error)
	Logout(ctx context.Context, refreshToken string) error
	LinkGuest(ctx context.Context, userID uuid.UUID, email, password string) (*domain.User, error)

	ChangePassword(ctx context.Context, userID uuid.UUID, oldPassword, newPassword string) error
	ChangeEmail(ctx context.Context, userID uuid.UUID, newEmail, password string) (*domain.User, error)
	UpdateProfile(ctx context.Context, userID uuid.UUID, nickname, timezone string) (*domain.User, error)
	GetProfile(ctx context.Context, userID uuid.UUID) (*domain.User, error)
}

type PasswordHasher interface {
    Hash(password string) (string, error)
    Compare(hash, password string) error
}

type TokenManager interface {
    GenerateAccessToken(userID uuid.UUID) (token string, expiresAt time.Time, err error)
    GenerateRefreshToken() (rawToken string, tokenHash string, expiresAt time.Time, err error)
	HashRefreshToken(rawToken string) string
}
