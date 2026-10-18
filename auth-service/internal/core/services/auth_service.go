package services

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/Holarbtw/homestock-backend/auth-service/internal/core/ports"
	"github.com/Holarbtw/homestock-backend/auth-service/internal/domain"
	"github.com/Holarbtw/homestock-backend/pkg/apperrors"
)

type AuthService struct {
	userRepo     ports.UserRepository
	txManager    ports.TxManager
	hasher       ports.PasswordHasher
	tokenManager ports.TokenManager
}

func NewAuthService(
	userRepo ports.UserRepository,
	txManager ports.TxManager,
	hasher ports.PasswordHasher,
	tokenManager ports.TokenManager,
) *AuthService {
	return &AuthService{
		userRepo:     userRepo,
		txManager:    txManager,
		hasher:       hasher,
		tokenManager: tokenManager,
	}
}

func (s *AuthService) issueTokenPair(ctx context.Context, userID uuid.UUID) (*domain.TokenPair, error){
	access, accessExpiresAt, err := s.tokenManager.GenerateAccessToken(userID)
	if err != nil {
		return nil, err
	}

	refresh, tokenHash, refreshExpiresAt, err := s.tokenManager.GenerateRefreshToken()
	if err != nil {
		return nil, err
	}

	err = s.userRepo.CreateRefreshToken(ctx, userID, tokenHash, refreshExpiresAt)
	if err != nil {
		return nil, err
	}

	return &domain.TokenPair{
		AccessToken:      access,
        RefreshToken:     refresh, 
        AccessExpiresAt:  accessExpiresAt,
        RefreshExpiresAt: refreshExpiresAt,
	}, nil
}

func (s *AuthService) Register(ctx context.Context, email, password, nickname, timezone string) (*domain.TokenPair, *domain.User, error){
	validEmail, err := validateEmail(email)
	if err != nil {
		return nil, nil, err
	}

	err = validatePassword(password)
	if err != nil {
		return nil, nil, err
	}

	validNickname, err := validateNickname(nickname)
	if err != nil {
		return nil, nil, err
	}

	validTimezone, err := validateTimezone(timezone)
	if err != nil {
		return nil, nil, err
	}
	
	passwordHash, err := s.hasher.Hash(password)
	if err != nil {
		return nil, nil, err
	}

	var user *domain.User
	var tokens *domain.TokenPair

	err = s.txManager.WithinTransaction(ctx, func(txCtx context.Context) error{
		var txErr error
		user, txErr = s.userRepo.Create(txCtx, validEmail, passwordHash, validNickname, validTimezone)
		if txErr != nil {
			return txErr
		}

		tokens, txErr = s.issueTokenPair(txCtx, user.ID)
		if txErr != nil {
			return txErr
		}

		return nil
	})

	if err != nil {
		return nil, nil, err
	}

	return tokens, user, nil
}

func (s *AuthService) Login(ctx context.Context, email, password string) (*domain.TokenPair, *domain.User, error){
	validEmail, err := validateEmail(email)
	if err != nil {
		return nil, nil, err
	}

	err = validatePassword(password)
	if err != nil {
		return nil, nil, err
	}

	user, err := s.userRepo.GetByEmail(ctx, validEmail)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound){
			return nil, nil, domain.ErrInvalidCredentials
		}

		return nil, nil, err
	}

	if user.PasswordHash == nil{
		return nil, nil, domain.ErrInvalidCredentials
	}	
	
	err = s.hasher.Compare(*user.PasswordHash, password)

	if err != nil{
		return nil, nil, domain.ErrInvalidCredentials
	}

	tokens, err := s.issueTokenPair(ctx, user.ID)
	if err != nil {
		return nil, nil, err
	}

	return tokens, user, nil
}

func (s *AuthService) LoginGuest(ctx context.Context, nickname, timezone string) (*domain.TokenPair, *domain.User, error){
	validNickname, err := validateNickname(nickname)
	if err != nil {
		return nil, nil, err
	}

	validTimezone, err := validateTimezone(timezone)
	if err != nil {
		return nil, nil, err
	}

	var user *domain.User
	var tokens *domain.TokenPair

	err = s.txManager.WithinTransaction(ctx, func(txCtx context.Context) error{
		var txErr error
		user, txErr = s.userRepo.CreateGuest(txCtx, validNickname, validTimezone)
		if txErr != nil {
			return txErr
		}

		tokens, txErr = s.issueTokenPair(txCtx, user.ID)
		if txErr != nil {
			return txErr
		}

		return nil
	})

	if err != nil {
		return nil, nil, err
	}

	return tokens, user, nil
}

func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (*domain.TokenPair, error){
	if refreshToken == ""{
		return nil, apperrors.ErrInvalidToken
	}

	tokenHash := s.tokenManager.HashRefreshToken(refreshToken)

	token, err := s.userRepo.GetRefreshToken(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound){
			return nil, apperrors.ErrInvalidToken
		}
		return nil, err
	}

	if token.RevokedAt != nil{
		if revokeErr := s.userRepo.RevokeAllUserTokens(ctx, token.UserID); revokeErr != nil {
			slog.Error("failed to revoke all user tokens on reuse detection",
				"user_id", token.UserID,
				"err", revokeErr,
			)
		}
		return nil, domain.ErrSessionRevoked
	}

	if time.Now().After(token.ExpiresAt){
		return nil, apperrors.ErrInvalidToken
	}

	var tokens *domain.TokenPair

	err = s.txManager.WithinTransaction(ctx, func(txCtx context.Context) error{
		var txErr error

		txErr = s.userRepo.RevokeRefreshToken(txCtx, tokenHash)
		if txErr != nil {
			return txErr
		}
	
		tokens, txErr = s.issueTokenPair(txCtx, token.UserID)
		if txErr != nil {
			return txErr
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return tokens, nil
}

func (s *AuthService) Logout(ctx context.Context, refreshToken string) error{
	if refreshToken == ""{
		return apperrors.ErrInvalidToken
	}

	tokenHash := s.tokenManager.HashRefreshToken(refreshToken)

	return s.userRepo.RevokeRefreshToken(ctx, tokenHash)
}

func (s *AuthService) LinkGuest(ctx context.Context, userID uuid.UUID, email, password string) (*domain.User, error){
	validEmail, err := validateEmail(email)
	if err != nil {
		return nil, err
	}

	err = validatePassword(password)
	if err != nil {
		return nil, err
	}

	var user *domain.User

	err = s.txManager.WithinTransaction(ctx, func(txCtx context.Context) error{
		var txErr error
		
		user, txErr = s.userRepo.GetByIDForUpdate(txCtx, userID)
		if txErr != nil {
			return txErr
		}

		if !user.IsGuest{
			return domain.ErrGuestAlreadyLinked
		}

		passwordHash, txErr := s.hasher.Hash(password)
		if txErr != nil {
			return txErr
		}

		txErr = s.userRepo.LinkGuestAccount(txCtx, userID, validEmail, passwordHash)
		if txErr != nil {
			return txErr
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	user.Email = &validEmail
	user.IsGuest = false

	return user, nil
}

func (s *AuthService) ChangePassword(ctx context.Context, userID uuid.UUID, oldPassword, newPassword string) error {
	err := validatePassword(newPassword)
	if err != nil {
		return err
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	if user.PasswordHash == nil {
		return domain.ErrInvalidCredentials
	}

	err = s.hasher.Compare(*user.PasswordHash, oldPassword) 
	if err != nil {
		return domain.ErrInvalidCredentials
	}

	newPasswordHash, err := s.hasher.Hash(newPassword)
	if err != nil {
		return err
	}

	return s.txManager.WithinTransaction(ctx, func(txCtx context.Context) error {
		if txErr := s.userRepo.UpdatePassword(txCtx, userID, newPasswordHash); txErr != nil {
			return txErr
		}

		if txErr := s.userRepo.RevokeAllUserTokens(txCtx, userID); txErr != nil {
			slog.Error("failed to revoke all tokens on password change", "user_id", userID, "err", txErr)
			return txErr
		}

		return nil
	})
}

func (s *AuthService) ChangeEmail(ctx context.Context, userID uuid.UUID, newEmail, password string) (*domain.User, error) {
	validEmail, err := validateEmail(newEmail)
	if err != nil {
		return nil, err
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if user.PasswordHash == nil {
		return nil, domain.ErrInvalidCredentials
	}

	if user.Email != nil && *user.Email == validEmail {
		return user, nil
	}

	if err := s.hasher.Compare(*user.PasswordHash, password); err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	if err := s.userRepo.UpdateEmail(ctx, userID, validEmail); err != nil {
		return nil, err
	}

	user.Email = &validEmail
	return user, nil
}

func (s *AuthService) UpdateProfile(ctx context.Context, userID uuid.UUID, nickname, timezone string) (*domain.User, error) {
	var updatedUser *domain.User

	err := s.txManager.WithinTransaction(ctx, func(txCtx context.Context) error {
		user, txErr := s.userRepo.GetByIDForUpdate(txCtx, userID)
		if txErr != nil {
			return txErr
		}

		targetNickname := user.Nickname
		if strings.TrimSpace(nickname) != "" {
			validNickname, err := validateNickname(nickname)
			if err != nil {
				return err
			}
			targetNickname = validNickname
		}

		targetTimezone := user.Timezone
		if strings.TrimSpace(timezone) != "" {
			validTimezone, err := validateTimezone(timezone)
			if err != nil {
				return err
			}
			targetTimezone = validTimezone
		}

		var updateErr error
		updatedUser, updateErr = s.userRepo.UpdateProfile(txCtx, userID, targetNickname, targetTimezone)
		if updateErr != nil {
			return updateErr
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return updatedUser, nil
}
