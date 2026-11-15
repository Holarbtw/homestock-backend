package postgres

import (
	"context"
	"errors"

	"github.com/Holarbtw/homestock-backend/inventory-service/internal/core/domain"
	"github.com/Holarbtw/homestock-backend/inventory-service/internal/core/ports"
	"github.com/Holarbtw/homestock-backend/pkg/apperrors"
	pkgPostgres "github.com/Holarbtw/homestock-backend/pkg/postgres"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var _ ports.SpaceRepository = (*SpaceRepository)(nil)

type SpaceRepository struct {
	pool *pgxpool.Pool
}

func NewSpaceRepository(pool *pgxpool.Pool) *SpaceRepository {
	return &SpaceRepository{pool: pool}
}

func (r *SpaceRepository) getExecutor(ctx context.Context) pkgPostgres.QueryExecutor {
	return pkgPostgres.GetExecutor(ctx, r.pool)
}

func (r *SpaceRepository) GetByOwnerUserID(ctx context.Context, ownerUserID uuid.UUID) (*domain.Space, error) {
	query := `
		select id, owner_user_id, created_at
		from spaces
		where owner_user_id = $1
	`

	rows, err := r.getExecutor(ctx).Query(ctx, query, ownerUserID)
	if err != nil {
		return nil, mapSpaceError(err)
	}

	record, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[spaceDB])
	if err != nil {
		return nil, mapSpaceError(err)
	}

	return toDomainSpace(&record), nil
}

func (r *SpaceRepository) GetOrCreate(ctx context.Context, ownerUserID uuid.UUID) (*domain.Space, error) {
	query := `
		insert into spaces (owner_user_id)
		values ($1)
		on conflict (owner_user_id) do nothing
		returning id, owner_user_id, created_at
	`

	rows, err := r.getExecutor(ctx).Query(ctx, query, ownerUserID)
	if err != nil {
		return nil, mapSpaceError(err)
	}

	record, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[spaceDB])
	if errors.Is(err, pgx.ErrNoRows) {
		// A separate statement sees the space committed by a concurrent creator.
		return r.GetByOwnerUserID(ctx, ownerUserID)
	}
	if err != nil {
		return nil, mapSpaceError(err)
	}

	return toDomainSpace(&record), nil
}

func mapSpaceError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrSpaceNotFound
	}
	return apperrors.MapPgError(err, nil)
}
