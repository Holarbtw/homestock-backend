package quantity

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/Holarbtw/homestock-backend/pkg/apperrors"
)

const (
	scale     = 1000
	fractions = 3
)

type Quantity struct {
	thousandths int64
}

func NewQuantity(thousandths int64) (*Quantity, error) {
	if thousandths < 0 {
		return nil, apperrors.ErrQuantityOutOfRange
	}

	return &Quantity{
		thousandths: thousandths,
	}, nil
}

func ParseQuantity(text string) (*Quantity, error) {
	trimmed := strings.TrimSpace(text)

	parts := strings.Split(trimmed, ".")
	if len(parts) > 2 {
		return nil, apperrors.ErrInvalidQuantityFormat
	}
	if len(parts) == 1 {
		parts = append(parts, "")
	}
	if parts[0] == "" && parts[1] == "" {
		return nil, apperrors.ErrInvalidQuantityFormat
	}

	for _, part := range parts {
		for _, digit := range part {
			if digit < '0' || digit > '9' {
				return nil, apperrors.ErrInvalidQuantityFormat
			}
		}
	}
	if parts[0] == "" {
		parts[0] = "0"
	}

	if len(parts[1]) > fractions {
		return nil, apperrors.ErrQuantityPrecisionExceeded
	}

	var fraction strings.Builder

	amount := fractions - len(parts[1])

	fraction.WriteString(parts[1])

	for range amount {
		fraction.WriteString("0")
	}

	whole := parts[0] + fraction.String()

	parsed, err := strconv.ParseInt(whole, 10, 64)
	if err != nil {
		if errors.Is(err, strconv.ErrRange) {
			return nil, apperrors.ErrQuantityOutOfRange.WithInternal(err)
		}
		return nil, apperrors.ErrInvalidQuantityFormat.WithInternal(err)
	}

	return NewQuantity(parsed)
}

func (q *Quantity) String() string{
	return fmt.Sprintf("%d.%03d", q.thousandths / scale, q.thousandths % scale)
}

func (q *Quantity) Thousandths() int64{
	return q.thousandths
}