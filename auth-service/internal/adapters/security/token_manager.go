package security

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/Holarbtw/homestock-backend/auth-service/internal/domain"
	pkgJWT "github.com/Holarbtw/homestock-backend/pkg/jwt"
	"github.com/golang-jwt/jwt/v5"
)

type JWTTokenManager struct {
	secret         []byte
	accessTokenTTL time.Duration
	refreshTokenTTL time.Duration
}

func NewJWTTokenManager(secret string, accessTokenTTL, refreshTokenTTL time.Duration) *JWTTokenManager{
	return &JWTTokenManager{
		secret: []byte(secret),
		accessTokenTTL: accessTokenTTL,
		refreshTokenTTL: refreshTokenTTL,
	}
}

func (t *JWTTokenManager) GenerateAccessToken(user *domain.User) (token string, expiresAt time.Time, err error){
	now := time.Now()
	expiresAt = now.Add(t.accessTokenTTL)

	claims := pkgJWT.UserClaims{
		UserID: user.ID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt: jwt.NewNumericDate(now),
		},
	}

	rawToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err = rawToken.SignedString(t.secret)

	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to sign access token: %w", err)
	}
	
	return token, expiresAt, nil
}

func (t *JWTTokenManager) GenerateRefreshToken() (rawToken string, tokenHash string, expiresAt time.Time, err error) {
	b := make([]byte, 32)

	if _, err := rand.Read(b); err != nil {
		return "", "", time.Time{}, fmt.Errorf("failed to generate random bytes for refresh token: %w", err)
	}

	rawToken = hex.EncodeToString(b)

	hash := sha256.Sum256([]byte(rawToken))
	tokenHash = hex.EncodeToString(hash[:])

	expiresAt = time.Now().Add(t.refreshTokenTTL)
	return rawToken, tokenHash, expiresAt, nil
}

func (t *JWTTokenManager) HashRefreshToken(rawToken string) string {
    hash := sha256.Sum256([]byte(rawToken))
    return hex.EncodeToString(hash[:])
}