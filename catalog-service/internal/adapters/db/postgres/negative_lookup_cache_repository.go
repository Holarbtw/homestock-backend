package postgres

import (
	"context"
	"time"

	"github.com/Holarbtw/homestock-backend/pkg/apperrors"
	pkgPostgres "github.com/Holarbtw/homestock-backend/pkg/postgres"
	"github.com/jackc/pgx/v5/pgxpool"
)


type NegativeLookupCacheRepository struct {
	pool *pgxpool.Pool
}

func NewNegativeLookupCacheRepository(pool *pgxpool.Pool) *NegativeLookupCacheRepository {
	return &NegativeLookupCacheRepository{pool: pool}
}

func (r *NegativeLookupCacheRepository) getExecutor(ctx context.Context) pkgPostgres.QueryExecutor {
	return pkgPostgres.GetExecutor(ctx, r.pool)
}

func (r *NegativeLookupCacheRepository) IsActive(ctx context.Context, gtin, provider string, at time.Time) (bool, error) {
	query := `
		select exists (
			select 1
			from negative_lookup_cache
			where gtin = $1 and provider = $2 and expires_at > $3
		)
	`

	var active bool
	err := r.getExecutor(ctx).QueryRow(ctx, query, gtin, provider, at).Scan(&active)
	if err != nil {
		return false, apperrors.MapPgError(err, nil)
	}

	return active, nil
}

func (r *NegativeLookupCacheRepository) Upsert(ctx context.Context, gtin, provider string, checkedAt, expiresAt time.Time) error {
	query := `
		insert into negative_lookup_cache (gtin, provider, checked_at, expires_at)
		values ($1, $2, $3, $4)
		on conflict (gtin, provider) do update
		set checked_at = excluded.checked_at,
			expires_at = excluded.expires_at
	`

	_, err := r.getExecutor(ctx).Exec(ctx, query, gtin, provider, checkedAt, expiresAt)
	return apperrors.MapPgError(err, nil)
}
