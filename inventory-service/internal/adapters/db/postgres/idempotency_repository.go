package postgres

import (
	"context"
	"crypto/sha256"
	"fmt"

	"github.com/Holarbtw/homestock-backend/inventory-service/internal/core/domain"
	"github.com/Holarbtw/homestock-backend/inventory-service/internal/core/ports"
	"github.com/Holarbtw/homestock-backend/pkg/apperrors"
	pkgPostgres "github.com/Holarbtw/homestock-backend/pkg/postgres"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var idempotencyConstraintMap = map[string]*apperrors.CustomError{
	"idempotency_records_space_id_fkey": domain.ErrSpaceNotFound,
}

var _ ports.IdempotencyRepository = (*IdempotencyRepository)(nil)

type IdempotencyRepository struct {
	pool *pgxpool.Pool
}

func NewIdempotencyRepository(pool *pgxpool.Pool) *IdempotencyRepository {
	return &IdempotencyRepository{pool: pool}
}

func (r *IdempotencyRepository) getExecutor(ctx context.Context) pkgPostgres.QueryExecutor {
	return pkgPostgres.GetExecutor(ctx, r.pool)
}

func (r *IdempotencyRepository) getTransaction(ctx context.Context) (pgx.Tx, error) {
	tx, ok := r.getExecutor(ctx).(pgx.Tx)
	if !ok {
		return nil, apperrors.ErrInternalServer.WithInternal(
			fmt.Errorf("idempotency writes require WithinTransaction context"),
		)
	}
	return tx, nil
}

func (r *IdempotencyRepository) TryReserve(ctx context.Context, spaceID uuid.UUID, requestKey, commandType string, requestHash [sha256.Size]byte) (bool, error) {
	tx, err := r.getTransaction(ctx)
	if err != nil {
		return false, err
	}

	query := `
		insert into idempotency_records (space_id, request_key, command_type, request_hash)
		values ($1, $2, $3, $4)
		on conflict (space_id, request_key) do nothing
	`

	tag, err := tx.Exec(ctx, query, spaceID, requestKey, commandType, requestHash[:])
	if err != nil {
		return false, apperrors.MapPgError(err, idempotencyConstraintMap)
	}
	return tag.RowsAffected() == 1, nil
}

func (r *IdempotencyRepository) GetByKey(ctx context.Context, spaceID uuid.UUID, requestKey string) (*ports.IdempotencyRecord, error) {
	query := `
		select space_id, request_key, command_type, request_hash,
			result_data, created_at
		from idempotency_records
		where space_id = $1 and request_key = $2
	`

	rows, err := r.getExecutor(ctx).Query(ctx, query, spaceID, requestKey)
	if err != nil {
		return nil, apperrors.MapPgError(err, idempotencyConstraintMap)
	}

	record, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[idempotencyRecordDB])
	if err != nil {
		return nil, apperrors.MapPgError(err, idempotencyConstraintMap)
	}

	return toIdempotencyRecord(&record)
}

func (r *IdempotencyRepository) SaveResult(ctx context.Context, spaceID uuid.UUID, requestKey string, result ports.IdempotencyResult) error {
	tx, err := r.getTransaction(ctx)
	if err != nil {
		return err
	}
	data := result.Data
	if data == nil {
		data = []byte{}
	}
	query := `
		update idempotency_records
		set result_data = $3
		where space_id = $1 and request_key = $2
			and result_data is null
	`

	tag, err := tx.Exec(ctx, query, spaceID, requestKey, data)
	if err != nil {
		return apperrors.MapPgError(err, idempotencyConstraintMap)
	}
	if tag.RowsAffected() == 0 {
		return apperrors.ErrInternalServer.WithInternal(
			fmt.Errorf("idempotency reservation is missing or its result is already saved"),
		)
	}
	return nil
}
