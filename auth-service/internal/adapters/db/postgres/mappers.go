package postgres

import (
	"github.com/Holarbtw/homestock-backend/auth-service/internal/domain"
)

func toDomainUser(u *userDB) *domain.User {
	if u == nil {
		return nil
	}
	return &domain.User{
		ID:           u.ID,
		Email:        u.Email,
		PasswordHash: u.PasswordHash,
		Nickname:     u.Nickname,
		IsGuest:      u.IsGuest,
		CreatedAt:    u.CreatedAt,
		UpdatedAt:    u.UpdatedAt,
		DeletedAt:    u.DeletedAt,
	}
}

func toDomainRefreshToken(t *refreshTokenDB) *domain.RefreshToken {
	if t == nil {
		return nil
	}
	return &domain.RefreshToken{
		ID:        t.ID,
		UserID:    t.UserID,
		TokenHash: t.TokenHash,
		ExpiresAt: t.ExpiresAt,
		CreatedAt: t.CreatedAt,
		RevokedAt: t.RevokedAt,
	}
}
