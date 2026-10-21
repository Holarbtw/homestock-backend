package services

import (
	"net/mail"
	"strings"
	"time"
	_ "time/tzdata"
	"unicode/utf8"

	"github.com/Holarbtw/homestock-backend/auth-service/internal/domain"
)

const (
	minPasswordLen       = 8
	maxPasswordLen       = 72
	minNicknameLen       = 2
	maxNicknameLen       = 32
	maxEmailLen          = 255
	defaultGuestNickname = "Гость"
	defaultTimezone      = "UTC"
)

func validateEmail(email string) (string, error) {
	trimmed := strings.TrimSpace(email)
	if trimmed == "" || len(trimmed) > maxEmailLen {
		return "", domain.ErrInvalidEmail
	}

	addr, err := mail.ParseAddress(trimmed)
	if err != nil || addr.Address != trimmed {
		return "", domain.ErrInvalidEmail
	}

	return strings.ToLower(addr.Address), nil
}

func validatePassword(password string) error {
	if utf8.RuneCountInString(password) < minPasswordLen || len(password) > maxPasswordLen {
		return domain.ErrInvalidPassword
	}
	return nil
}

func validateNickname(nickname string) (string, error) {
	trimmed := strings.TrimSpace(nickname)
	runeCount := utf8.RuneCountInString(trimmed)
	if runeCount < minNicknameLen || runeCount > maxNicknameLen {
		return "", domain.ErrInvalidNickname
	}
	return trimmed, nil
}

func validateTimezone(tz string) (string, error) {
	trimmed := strings.TrimSpace(tz)
	if trimmed == "" {
		return defaultTimezone, nil
	}

	if _, err := time.LoadLocation(trimmed); err != nil {
		return "", domain.ErrInvalidTimezone
	}

	return trimmed, nil
}
