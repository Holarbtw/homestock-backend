package http

import (
	"time"

	"github.com/google/uuid"
)

type RegisterRequest struct {
	// Email address, trimmed and converted to lowercase; at most 255 UTF-8 bytes.
	Email string `json:"email" format:"email" validate:"required" example:"user@example.com"`
	// At least 8 Unicode code points and at most 72 UTF-8 bytes. Not trimmed.
	Password string `json:"password" format:"password" validate:"required" minLength:"8" example:"secret123"`
	// Between 2 and 32 Unicode code points after trimming surrounding whitespace.
	Nickname string `json:"nickname" validate:"required" example:"Алексей"`
	// IANA time zone. Omitted, null, empty or whitespace-only values use UTC.
	Timezone string `json:"timezone" default:"UTC" example:"Europe/Moscow"`
}

type LoginRequest struct {
	// Email address, trimmed and converted to lowercase.
	Email string `json:"email" format:"email" validate:"required" example:"user@example.com"`
	// At least 8 Unicode code points and at most 72 UTF-8 bytes. Not trimmed.
	Password string `json:"password" format:"password" validate:"required" minLength:"8" example:"secret123"`
}

type GuestLoginRequest struct {
	// Omitted, null, empty or whitespace-only values use "Гость"; otherwise 2 to 32 Unicode code points after trimming.
	Nickname string `json:"nickname,omitempty" default:"Гость" example:"Гость"`
	// IANA time zone. Omitted, null, empty or whitespace-only values use UTC.
	Timezone string `json:"timezone" default:"UTC" example:"UTC"`
}

type RefreshTokenRequest struct {
	// Opaque refresh token from the last successful login or refresh. Missing or empty values produce INVALID_TOKEN (401).
	RefreshToken string `json:"refresh_token" validate:"required" minLength:"1" example:"d290f1ee6..."`
}

type LinkGuestRequest struct {
	// Email address, trimmed and converted to lowercase; at most 255 UTF-8 bytes.
	Email string `json:"email" format:"email" validate:"required" example:"user@example.com"`
	// At least 8 Unicode code points and at most 72 UTF-8 bytes. Not trimmed.
	Password string `json:"password" format:"password" validate:"required" minLength:"8" example:"secret123"`
}

type ChangePasswordRequest struct {
	// Current password, sent without trimming.
	OldPassword string `json:"old_password" format:"password" validate:"required" example:"oldsecret123"`
	// At least 8 Unicode code points and at most 72 UTF-8 bytes. Not trimmed.
	NewPassword string `json:"new_password" format:"password" validate:"required" minLength:"8" example:"newsecret123"`
}

type ChangeEmailRequest struct {
	// New email address, trimmed and converted to lowercase; at most 255 UTF-8 bytes.
	NewEmail string `json:"new_email" format:"email" validate:"required" example:"new_user@example.com"`
	// Current password. Checked when the normalized email differs from the current email.
	Password string `json:"password" format:"password" validate:"required" example:"secret123"`
}

type UpdateProfileRequest struct {
	// Omitted, null, empty or whitespace-only values leave the nickname unchanged; otherwise 2 to 32 Unicode code points after trimming.
	Nickname string `json:"nickname,omitempty" example:"НовыйНик"`
	// IANA time zone. Omitted, null, empty or whitespace-only values leave the time zone unchanged.
	Timezone string `json:"timezone,omitempty" example:"Europe/Moscow"`
}

type UserResponse struct {
	ID uuid.UUID `json:"id" format:"uuid" validate:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
	// Omitted for guest accounts; a string for registered accounts. Never serialized as null.
	Email    *string `json:"email,omitempty" format:"email" example:"user@example.com"`
	Nickname string  `json:"nickname" validate:"required" example:"Алексей"`
	// User's IANA time zone preference. Does not change the UTC timestamps in API responses.
	Timezone string `json:"timezone" validate:"required" example:"Europe/Moscow"`
	IsGuest  bool   `json:"is_guest" validate:"required" example:"false"`
	// RFC 3339 timestamp in UTC (Z). Fractional seconds may be present.
	CreatedAt time.Time `json:"created_at" format:"date-time" validate:"required" example:"2026-10-26T10:00:00Z"`
	// RFC 3339 timestamp in UTC (Z). Fractional seconds may be present.
	UpdatedAt time.Time `json:"updated_at" format:"date-time" validate:"required" example:"2026-10-26T10:00:00Z"`
}

type AuthResponse struct {
	AccessToken  string `json:"access_token" validate:"required" example:"eyJhbGciOi..."`
	RefreshToken string `json:"refresh_token" validate:"required" example:"a1b2c3..."`
	// Access token expiry as an RFC 3339 timestamp in UTC (Z). Fractional seconds may be present.
	AccessExpiresAt time.Time    `json:"access_expires_at" format:"date-time" validate:"required" example:"2026-10-26T10:15:00Z"`
	User            UserResponse `json:"user" validate:"required"`
}

type TokenResponse struct {
	AccessToken  string `json:"access_token" validate:"required" example:"eyJhbGciOi..."`
	RefreshToken string `json:"refresh_token" validate:"required" example:"a1b2c3..."`
	// Access token expiry as an RFC 3339 timestamp in UTC (Z). Fractional seconds may be present.
	AccessExpiresAt time.Time `json:"access_expires_at" format:"date-time" validate:"required" example:"2026-10-26T10:15:00Z"`
}