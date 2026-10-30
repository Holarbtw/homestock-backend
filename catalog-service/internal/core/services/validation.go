package services

import (
	"regexp"
	"strings"

	"github.com/Holarbtw/homestock-backend/catalog-service/internal/domain"
)

var gtinPattern = regexp.MustCompile(`^([0-9]{8}|[0-9]{12,14})$`)

func normalizeGTIN(gtin string) (string, error) {
	trimmed := strings.TrimSpace(gtin)
	if !gtinPattern.MatchString(trimmed) {
		return "", domain.ErrInvalidGTIN
	}

	return strings.Repeat("0", 14-len(trimmed)) + trimmed, nil
}
