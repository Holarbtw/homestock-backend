package http

import (
	"encoding/json"
	"net/http"

	"github.com/Holarbtw/homestock-backend/auth-service/internal/core/ports"
	"github.com/Holarbtw/homestock-backend/pkg/apperrors"
	pkgJWT "github.com/Holarbtw/homestock-backend/pkg/jwt"
)

type AuthHandler struct {
	authService ports.AuthService
}

func NewAuthHandler(authService ports.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

func sendJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

// @Summary Register a new user
// @Description Creates a new user account and returns an auth token pair.
// @Description Email, password and nickname are required. Omitted or blank timezone defaults to UTC. Response timestamps use RFC 3339 UTC (Z).
// @Tags Auth
// @Accept json
// @Produce json
// @Param input body RegisterRequest true "Registration input"
// @Success 201 {object} AuthResponse
// @Failure 400 {object} apperrors.CustomError "INVALID_REQUEST, INVALID_EMAIL, INVALID_PASSWORD, INVALID_NICKNAME, INVALID_TIMEZONE"
// @Failure 409 {object} apperrors.CustomError "EMAIL_ALREADY_EXISTS"
// @Failure 429 {string} string "Gateway rate limit exceeded; the body may be empty or plain text, not CustomError"
// @Header 429 {integer} Retry-After "Seconds to wait before retrying"
// @Failure 500 {object} apperrors.CustomError "INTERNAL_SERVER_ERROR"
// @Router /register [post]
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		apperrors.RespondWithError(w, r, apperrors.ErrInvalidRequest.WithInternal(err))
		return
	}

	tokens, user, err := h.authService.Register(r.Context(), req.Email, req.Password, req.Nickname, req.Timezone)
	if err != nil {
		apperrors.RespondWithError(w, r, err)
		return
	}

	sendJSON(w, http.StatusCreated, AuthResponse{
		AccessToken:     tokens.AccessToken,
		RefreshToken:    tokens.RefreshToken,
		AccessExpiresAt: tokens.AccessExpiresAt.UTC(),
		User:            toUserResponse(user),
	})
}

// @Summary Login user
// @Description Authenticates a user with email and password, returning an auth token pair.
// @Description A 401 INVALID_CREDENTIALS response means the credentials were rejected; refreshing an existing token does not resolve it.
// @Tags Auth
// @Accept json
// @Produce json
// @Param input body LoginRequest true "Login credentials"
// @Success 200 {object} AuthResponse
// @Failure 400 {object} apperrors.CustomError "INVALID_REQUEST, INVALID_EMAIL, INVALID_PASSWORD"
// @Failure 401 {object} apperrors.CustomError "INVALID_CREDENTIALS"
// @Failure 429 {string} string "Gateway rate limit exceeded; the body may be empty or plain text, not CustomError"
// @Header 429 {integer} Retry-After "Seconds to wait before retrying"
// @Failure 500 {object} apperrors.CustomError "INTERNAL_SERVER_ERROR"
// @Router /login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		apperrors.RespondWithError(w, r, apperrors.ErrInvalidRequest.WithInternal(err))
		return
	}

	tokens, user, err := h.authService.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		apperrors.RespondWithError(w, r, err)
		return
	}

	sendJSON(w, http.StatusOK, AuthResponse{
		AccessToken:     tokens.AccessToken,
		RefreshToken:    tokens.RefreshToken,
		AccessExpiresAt: tokens.AccessExpiresAt.UTC(),
		User:            toUserResponse(user),
	})
}

// @Summary Guest login
// @Description Creates an anonymous guest user account and returns an auth token pair.
// @Description The body may be omitted. Omitted or blank nickname defaults to "Гость"; timezone defaults to UTC. Guest user responses omit email instead of returning null.
// @Tags Auth
// @Accept json
// @Produce json
// @Param input body GuestLoginRequest false "Guest login options"
// @Success 201 {object} AuthResponse
// @Failure 400 {object} apperrors.CustomError "INVALID_REQUEST, INVALID_NICKNAME, INVALID_TIMEZONE"
// @Failure 429 {string} string "Gateway rate limit exceeded; the body may be empty or plain text, not CustomError"
// @Header 429 {integer} Retry-After "Seconds to wait before retrying"
// @Failure 500 {object} apperrors.CustomError "INTERNAL_SERVER_ERROR"
// @Router /guest-login [post]
func (h *AuthHandler) LoginGuest(w http.ResponseWriter, r *http.Request) {
	var req GuestLoginRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil && err.Error() != "EOF" {
		apperrors.RespondWithError(w, r, apperrors.ErrInvalidRequest.WithInternal(err))
		return
	}

	tokens, user, err := h.authService.LoginGuest(r.Context(), req.Nickname, req.Timezone)
	if err != nil {
		apperrors.RespondWithError(w, r, err)
		return
	}

	sendJSON(w, http.StatusCreated, AuthResponse{
		AccessToken:     tokens.AccessToken,
		RefreshToken:    tokens.RefreshToken,
		AccessExpiresAt: tokens.AccessExpiresAt.UTC(),
		User:            toUserResponse(user),
	})
}

// @Summary Refresh token pair
// @Description Refreshes the access token and returns a new rotated token pair.
// @Description The supplied refresh token is revoked on success. Store both returned tokens and serialize refresh attempts to avoid reusing an already rotated token. No Authorization header is required.
// @Tags Auth
// @Accept json
// @Produce json
// @Param input body RefreshTokenRequest true "Refresh token input"
// @Success 200 {object} TokenResponse
// @Failure 400 {object} apperrors.CustomError "INVALID_REQUEST"
// @Failure 401 {object} apperrors.CustomError "INVALID_TOKEN (missing, unknown or expired refresh token), SESSION_REVOKED (already revoked or rotated refresh token)"
// @Failure 429 {string} string "Gateway rate limit exceeded; the body may be empty or plain text, not CustomError"
// @Header 429 {integer} Retry-After "Seconds to wait before retrying; the current refresh token was not consumed"
// @Failure 500 {object} apperrors.CustomError "INTERNAL_SERVER_ERROR"
// @Router /refresh [post]
func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req RefreshTokenRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		apperrors.RespondWithError(w, r, apperrors.ErrInvalidRequest.WithInternal(err))
		return
	}

	tokens, err := h.authService.RefreshToken(r.Context(), req.RefreshToken)
	if err != nil {
		apperrors.RespondWithError(w, r, err)
		return
	}

	sendJSON(w, http.StatusOK, TokenResponse{
		AccessToken:     tokens.AccessToken,
		RefreshToken:    tokens.RefreshToken,
		AccessExpiresAt: tokens.AccessExpiresAt.UTC(),
	})
}

// @Summary Logout user
// @Description Revokes the specified refresh token session.
// @Description No Authorization header is required. A non-empty unknown or already revoked token still returns 204. Missing or empty refresh_token returns 401 INVALID_TOKEN. Access tokens remain valid until their expiry.
// @Tags Auth
// @Accept json
// @Produce json
// @Param input body RefreshTokenRequest true "Refresh token to revoke"
// @Success 204 "Successfully logged out"
// @Failure 400 {object} apperrors.CustomError "INVALID_REQUEST"
// @Failure 401 {object} apperrors.CustomError "INVALID_TOKEN"
// @Failure 500 {object} apperrors.CustomError "INTERNAL_SERVER_ERROR"
// @Router /logout [post]
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	var req RefreshTokenRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		apperrors.RespondWithError(w, r, apperrors.ErrInvalidRequest.WithInternal(err))
		return
	}

	if err := h.authService.Logout(r.Context(), req.RefreshToken); err != nil {
		apperrors.RespondWithError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// @Summary Link guest account to email and password
// @Description Converts an active guest user account to a permanent registered account.
// @Description Preserves the user ID and existing access and refresh tokens; no tokens are issued or revoked. Returns the updated user with is_guest=false and email. Update the cached profile and keep using the current tokens until normal expiry, rotation or revocation.
// @Tags Auth
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param input body LinkGuestRequest true "Account credentials"
// @Success 200 {object} UserResponse
// @Failure 400 {object} apperrors.CustomError "INVALID_REQUEST, INVALID_EMAIL, INVALID_PASSWORD, GUEST_ALREADY_LINKED"
// @Failure 401 {object} apperrors.CustomError "UNAUTHORIZED, INVALID_TOKEN"
// @Failure 404 {object} apperrors.CustomError "NOT_FOUND"
// @Failure 409 {object} apperrors.CustomError "EMAIL_ALREADY_EXISTS"
// @Failure 429 {string} string "Gateway rate limit exceeded; the body may be empty or plain text, not CustomError"
// @Header 429 {integer} Retry-After "Seconds to wait before retrying"
// @Failure 500 {object} apperrors.CustomError "INTERNAL_SERVER_ERROR"
// @Router /link-guest [post]
func (h *AuthHandler) LinkGuest(w http.ResponseWriter, r *http.Request) {
	userID, ok := pkgJWT.GetUserIDFromContext(r.Context())
	if !ok {
		apperrors.RespondWithError(w, r, apperrors.ErrUnauthorized)
		return
	}

	var req LinkGuestRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		apperrors.RespondWithError(w, r, apperrors.ErrInvalidRequest.WithInternal(err))
		return
	}

	user, err := h.authService.LinkGuest(r.Context(), userID, req.Email, req.Password)
	if err != nil {
		apperrors.RespondWithError(w, r, err)
		return
	}

	sendJSON(w, http.StatusOK, toUserResponse(user))
}

// @Summary Change password
// @Description Changes the password and revokes every refresh token belonging to the user, including the current session. No new tokens are returned; sign in again to obtain a new pair. Previously issued access tokens remain valid until expiry.
// @Tags Profile
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param input body ChangePasswordRequest true "Old and new password"
// @Success 204 "Password updated successfully"
// @Failure 400 {object} apperrors.CustomError "INVALID_REQUEST, INVALID_PASSWORD (new_password)"
// @Failure 401 {object} apperrors.CustomError "UNAUTHORIZED, INVALID_TOKEN, INVALID_CREDENTIALS (wrong old password or guest account)"
// @Failure 404 {object} apperrors.CustomError "NOT_FOUND"
// @Failure 429 {string} string "Gateway rate limit exceeded; the body may be empty or plain text, not CustomError"
// @Header 429 {integer} Retry-After "Seconds to wait before retrying"
// @Failure 500 {object} apperrors.CustomError "INTERNAL_SERVER_ERROR"
// @Router /password [patch]
func (h *AuthHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	userID, ok := pkgJWT.GetUserIDFromContext(r.Context())
	if !ok {
		apperrors.RespondWithError(w, r, apperrors.ErrUnauthorized)
		return
	}

	var req ChangePasswordRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		apperrors.RespondWithError(w, r, apperrors.ErrInvalidRequest.WithInternal(err))
		return
	}

	err = h.authService.ChangePassword(r.Context(), userID, req.OldPassword, req.NewPassword)
	if err != nil {
		apperrors.RespondWithError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// @Summary Change email
// @Description Changes the email address of the authenticated user.
// @Description Preserves the user ID and existing tokens. The password is checked when the normalized new email differs from the current one. Sending the current email returns the unchanged user without checking the password.
// @Tags Profile
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param input body ChangeEmailRequest true "New email and current password"
// @Success 200 {object} UserResponse
// @Failure 400 {object} apperrors.CustomError "INVALID_REQUEST, INVALID_EMAIL"
// @Failure 401 {object} apperrors.CustomError "UNAUTHORIZED, INVALID_TOKEN, INVALID_CREDENTIALS (wrong password or guest account)"
// @Failure 404 {object} apperrors.CustomError "NOT_FOUND"
// @Failure 409 {object} apperrors.CustomError "EMAIL_ALREADY_EXISTS"
// @Failure 429 {string} string "Gateway rate limit exceeded; the body may be empty or plain text, not CustomError"
// @Header 429 {integer} Retry-After "Seconds to wait before retrying"
// @Failure 500 {object} apperrors.CustomError "INTERNAL_SERVER_ERROR"
// @Router /email [patch]
func (h *AuthHandler) ChangeEmail(w http.ResponseWriter, r *http.Request) {
	userID, ok := pkgJWT.GetUserIDFromContext(r.Context())
	if !ok {
		apperrors.RespondWithError(w, r, apperrors.ErrUnauthorized)
		return
	}

	var req ChangeEmailRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		apperrors.RespondWithError(w, r, apperrors.ErrInvalidRequest.WithInternal(err))
		return
	}

	user, err := h.authService.ChangeEmail(r.Context(), userID, req.NewEmail, req.Password)
	if err != nil {
		apperrors.RespondWithError(w, r, err)
		return
	}

	sendJSON(w, http.StatusOK, toUserResponse(user))
}

// @Summary Update user profile
// @Description Updates nickname and/or timezone for the authenticated user.
// @Description Omitted, null, empty or whitespace-only fields keep their existing values. The time zone preference does not change the UTC timestamps returned by the API.
// @Tags Profile
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param input body UpdateProfileRequest true "Profile update input"
// @Success 200 {object} UserResponse
// @Failure 400 {object} apperrors.CustomError "INVALID_REQUEST, INVALID_NICKNAME, INVALID_TIMEZONE"
// @Failure 401 {object} apperrors.CustomError "UNAUTHORIZED, INVALID_TOKEN"
// @Failure 404 {object} apperrors.CustomError "NOT_FOUND"
// @Failure 500 {object} apperrors.CustomError "INTERNAL_SERVER_ERROR"
// @Router /profile [patch]
func (h *AuthHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := pkgJWT.GetUserIDFromContext(r.Context())
	if !ok {
		apperrors.RespondWithError(w, r, apperrors.ErrUnauthorized)
		return
	}

	var req UpdateProfileRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		apperrors.RespondWithError(w, r, apperrors.ErrInvalidRequest.WithInternal(err))
		return
	}

	user, err := h.authService.UpdateProfile(r.Context(), userID, req.Nickname, req.Timezone)
	if err != nil {
		apperrors.RespondWithError(w, r, err)
		return
	}

	sendJSON(w, http.StatusOK, toUserResponse(user))
}

// @Summary Get current user profile
// @Description Returns the profile information of the authenticated user.
// @Description All fields except email are always present. Email is omitted for a guest. Timestamps are RFC 3339 UTC (Z), independently of the profile time zone.
// @Tags Profile
// @Security ApiKeyAuth
// @Produce json
// @Success 200 {object} UserResponse
// @Failure 401 {object} apperrors.CustomError "UNAUTHORIZED, INVALID_TOKEN"
// @Failure 404 {object} apperrors.CustomError "NOT_FOUND"
// @Failure 500 {object} apperrors.CustomError "INTERNAL_SERVER_ERROR"
// @Router /profile [get]
func (h *AuthHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := pkgJWT.GetUserIDFromContext(r.Context())
	if !ok {
		apperrors.RespondWithError(w, r, apperrors.ErrUnauthorized)
		return
	}

	user, err := h.authService.GetProfile(r.Context(), userID)
	if err != nil {
		apperrors.RespondWithError(w, r, err)
		return
	}

	sendJSON(w, http.StatusOK, toUserResponse(user))
}
