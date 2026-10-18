package http

import (
	"time"
	"github.com/google/uuid"
)


type RegisterRequest struct {
	Email    string `json:"email" example:"user@example.com"`
	Password string `json:"password" example:"secret123"`
	Nickname string `json:"nickname" example:"Алексей"`
	Timezone string `json:"timezone" example:"Europe/Moscow"`
}

type LoginRequest struct {
	Email    string `json:"email" example:"user@example.com"`
	Password string `json:"password" example:"secret123"`
}

type GuestLoginRequest struct {
	Nickname string `json:"nickname,omitempty" example:"Гость"`
	Timezone string `json:"timezone" example:"UTC"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" example:"d290f1ee6..."`
}

type LinkGuestRequest struct {
	Email    string `json:"email" example:"user@example.com"`
	Password string `json:"password" example:"secret123"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" example:"oldsecret123"`
	NewPassword string `json:"new_password" example:"newsecret123"`
}

type ChangeEmailRequest struct {
	NewEmail string `json:"new_email" example:"new_user@example.com"`
	Password string `json:"password" example:"secret123"`
}

type UpdateProfileRequest struct {
	Nickname string `json:"nickname,omitempty" example:"НовыйНик"`
	Timezone string `json:"timezone,omitempty" example:"Europe/Moscow"`
}


type UserResponse struct {
	ID        uuid.UUID `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Email     *string   `json:"email,omitempty" example:"user@example.com"`
	Nickname  string    `json:"nickname" example:"Алексей"`
	Timezone  string    `json:"timezone" example:"Europe/Moscow"`
	IsGuest   bool      `json:"is_guest" example:"false"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type AuthResponse struct {
	AccessToken     string       `json:"access_token" example:"eyJhbGciOi..."`
	RefreshToken    string       `json:"refresh_token" example:"a1b2c3..."`
	AccessExpiresAt time.Time    `json:"access_expires_at"`
	User            UserResponse `json:"user"`
}

type TokenResponse struct {
	AccessToken     string    `json:"access_token" example:"eyJhbGciOi..."`
	RefreshToken    string    `json:"refresh_token" example:"a1b2c3..."`
	AccessExpiresAt time.Time `json:"access_expires_at"`
}