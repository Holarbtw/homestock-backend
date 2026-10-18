package http

import (
	"github.com/Holarbtw/homestock-backend/auth-service/internal/domain"
)

func toUserResponse(u *domain.User) UserResponse {
	return UserResponse{
		ID:        u.ID,
		Email:     u.Email,
		Nickname:  u.Nickname,
		Timezone:  u.Timezone,
		IsGuest:   u.IsGuest,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}