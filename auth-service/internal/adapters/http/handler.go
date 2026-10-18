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
// @Tags Auth
// @Accept json
// @Produce json
// @Param input body RegisterRequest true "Registration input"
// @Success 201 {object} AuthResponse
// @Failure 400 {object} apperrors.CustomError "Validation error"
// @Failure 409 {object} apperrors.CustomError "User with this email already exists"
// @Failure 500 {object} apperrors.CustomError "Internal server error"
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
		AccessExpiresAt: tokens.AccessExpiresAt,
		User:            toUserResponse(user),
	})
}

// @Summary Login user
// @Description Authenticates a user with email and password, returning an auth token pair.
// @Tags Auth
// @Accept json
// @Produce json
// @Param input body LoginRequest true "Login credentials"
// @Success 200 {object} AuthResponse
// @Failure 400 {object} apperrors.CustomError "Validation error"
// @Failure 401 {object} apperrors.CustomError "Invalid credentials"
// @Failure 500 {object} apperrors.CustomError "Internal server error"
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
		AccessExpiresAt: tokens.AccessExpiresAt,
		User:            toUserResponse(user),
	})
}

// @Summary Guest login
// @Description Creates an anonymous guest user account and returns an auth token pair.
// @Tags Auth
// @Accept json
// @Produce json
// @Param input body GuestLoginRequest false "Guest login options"
// @Success 201 {object} AuthResponse
// @Failure 400 {object} apperrors.CustomError "Validation error"
// @Failure 500 {object} apperrors.CustomError "Internal server error"
// @Router /guest-login [post]
func (h *AuthHandler) LoginGuest(w http.ResponseWriter, r *http.Request) {
	var req GuestLoginRequest
	err := json.NewDecoder(r.Body).Decode(&req);
	if  err != nil && err.Error() != "EOF" {
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
		AccessExpiresAt: tokens.AccessExpiresAt,
		User:            toUserResponse(user),
	})
}

// @Summary Refresh token pair
// @Description Refreshes the access token and returns a new rotated token pair.
// @Tags Auth
// @Accept json
// @Produce json
// @Param input body RefreshTokenRequest true "Refresh token input"
// @Success 200 {object} TokenResponse
// @Failure 400 {object} apperrors.CustomError "Invalid request"
// @Failure 401 {object} apperrors.CustomError "Invalid or revoked token"
// @Failure 500 {object} apperrors.CustomError "Internal server error"
// @Router /refresh [post]
func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req RefreshTokenRequest
	err := json.NewDecoder(r.Body).Decode(&req);
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
		AccessExpiresAt: tokens.AccessExpiresAt,
	})
}

// @Summary Logout user
// @Description Revokes the specified refresh token session.
// @Tags Auth
// @Accept json
// @Produce json
// @Param input body RefreshTokenRequest true "Refresh token to revoke"
// @Success 204 "Successfully logged out"
// @Failure 400 {object} apperrors.CustomError "Invalid request"
// @Failure 500 {object} apperrors.CustomError "Internal server error"
// @Router /logout [post]
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	var req RefreshTokenRequest
	err := json.NewDecoder(r.Body).Decode(&req);
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
// @Tags Auth
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param input body LinkGuestRequest true "Account credentials"
// @Success 200 {object} UserResponse
// @Failure 400 {object} apperrors.CustomError "Validation error or account already linked"
// @Failure 401 {object} apperrors.CustomError "Unauthorized"
// @Failure 409 {object} apperrors.CustomError "Email already in use"
// @Failure 500 {object} apperrors.CustomError "Internal server error"
// @Router /link-guest [post]
func (h *AuthHandler) LinkGuest(w http.ResponseWriter, r *http.Request) {
	userID, ok := pkgJWT.GetUserIDFromContext(r.Context())
	if !ok {
		apperrors.RespondWithError(w, r, apperrors.ErrUnauthorized)
		return
	}

	var req LinkGuestRequest
	err := json.NewDecoder(r.Body).Decode(&req);
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
// @Description Changes the password of the authenticated user and revokes other active sessions.
// @Tags Profile
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param input body ChangePasswordRequest true "Old and new password"
// @Success 204 "Password updated successfully"
// @Failure 400 {object} apperrors.CustomError "Validation error"
// @Failure 401 {object} apperrors.CustomError "Unauthorized or wrong old password"
// @Failure 500 {object} apperrors.CustomError "Internal server error"
// @Router /password [patch]
func (h *AuthHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	userID, ok := pkgJWT.GetUserIDFromContext(r.Context())
	if !ok {
		apperrors.RespondWithError(w, r, apperrors.ErrUnauthorized)
		return
	}

	var req ChangePasswordRequest
	err := json.NewDecoder(r.Body).Decode(&req);
	if err != nil {
		apperrors.RespondWithError(w, r, apperrors.ErrInvalidRequest.WithInternal(err))
		return
	}

	err = h.authService.ChangePassword(r.Context(), userID, req.OldPassword, req.NewPassword); 
	if err != nil {
		apperrors.RespondWithError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// @Summary Change email
// @Description Changes the email address of the authenticated user.
// @Tags Profile
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param input body ChangeEmailRequest true "New email and current password"
// @Success 200 {object} UserResponse
// @Failure 400 {object} apperrors.CustomError "Validation error"
// @Failure 401 {object} apperrors.CustomError "Unauthorized or wrong password"
// @Failure 409 {object} apperrors.CustomError "Email already in use"
// @Failure 500 {object} apperrors.CustomError "Internal server error"
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
// @Tags Profile
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param input body UpdateProfileRequest true "Profile update input"
// @Success 200 {object} UserResponse
// @Failure 400 {object} apperrors.CustomError "Validation error"
// @Failure 401 {object} apperrors.CustomError "Unauthorized"
// @Failure 500 {object} apperrors.CustomError "Internal server error"
// @Router /profile [patch]
func (h *AuthHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := pkgJWT.GetUserIDFromContext(r.Context())
	if !ok {
		apperrors.RespondWithError(w, r, apperrors.ErrUnauthorized)
		return
	}

	var req UpdateProfileRequest
	err := json.NewDecoder(r.Body).Decode(&req);
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
// @Tags Profile
// @Security ApiKeyAuth
// @Produce json
// @Success 200 {object} UserResponse
// @Failure 401 {object} apperrors.CustomError "Unauthorized"
// @Failure 404 {object} apperrors.CustomError "User not found"
// @Failure 500 {object} apperrors.CustomError "Internal server error"
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
