package apperrors

import (
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const (
	PgErrUniqueViolation     = "23505" // UNIQUE
	PgErrForeignKeyViolation = "23503" // FOREIGN KEY
	PgErrCheckViolation      = "23514" // CHECK
	PgErrNotNullViolation    = "23502" // NOT NULL
)

func MapPgError(err error, constraintMap map[string]*CustomError) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case PgErrUniqueViolation:
			if constraintMap != nil {
				if customErr, ok := constraintMap[pgErr.ConstraintName]; ok {
					return customErr
				}
			}
			return ErrConflict

		case PgErrForeignKeyViolation:
			if constraintMap != nil {
				if customErr, ok := constraintMap[pgErr.ConstraintName]; ok {
					return customErr
				}
			}
			return ErrNotFound

		case PgErrNotNullViolation, PgErrCheckViolation:
			return ErrValidationFailed.WithInternal(err)

		default:
			return ErrInternalServer.WithInternal(err)
		}
	}

	return ErrInternalServer.WithInternal(err)
}
