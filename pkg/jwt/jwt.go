package jwt

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/Holarbtw/homestock-backend/pkg/apperrors"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type userContextKey struct{}

var userCtxKey = userContextKey{}

type UserClaims struct {
	UserID uuid.UUID `json:"user_id"`
	jwt.RegisteredClaims
}

type TokenVerifier struct {
	secretKey string
}

func NewTokenVerifier(secretKey string) *TokenVerifier {
	return &TokenVerifier{
		secretKey: secretKey,
	}
}

func (v *TokenVerifier) ParseToken(tokenString string) (*UserClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			slog.Warn("suspicious: unexpected signing method", "alg", token.Header["alg"])
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(v.secretKey), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, apperrors.ErrInvalidToken.WithInternal(err)
		}
		return nil, apperrors.ErrInvalidToken.WithInternal(err)
	}

	claims, ok := token.Claims.(*UserClaims)
	if ok && token.Valid {
		return claims, nil
	}

	return nil, apperrors.ErrInvalidToken
}

func (v *TokenVerifier) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		authHeader := req.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			apperrors.RespondWithError(w, apperrors.ErrUnauthorized)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := v.ParseToken(tokenString)
		if err != nil {
			apperrors.RespondWithError(w, apperrors.ErrInvalidToken)
			return
		}

		ctx := context.WithValue(req.Context(), userCtxKey, claims)
		next.ServeHTTP(w, req.WithContext(ctx))
	})
}

func GetUserFromContext(ctx context.Context) (*UserClaims, bool) {
	claims, ok := ctx.Value(userCtxKey).(*UserClaims)
	return claims, ok
}

func GetUserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	claims, ok := GetUserFromContext(ctx)
	if !ok {
		return uuid.Nil, false
	}
	return claims.UserID, true
}
