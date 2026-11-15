package postgres

import (
	"context"
	"fmt"

	"github.com/Holarbtw/homestock-backend/inventory-service/internal/core/ports"
	"github.com/Holarbtw/homestock-backend/pkg/apperrors"
	pkgPostgres "github.com/Holarbtw/homestock-backend/pkg/postgres"
	"github.com/Holarbtw/homestock-backend/pkg/quantity"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

var _ ports.BatchRepository = (*BatchRepository)(nil)

type BatchRepository struct {
	pool *pgxpool.Pool
}

func NewBatchRepository(pool *pgxpool.Pool) *BatchRepository {
	return &BatchRepository{pool: pool}
}

func (r *BatchRepository) getExecutor(ctx context.Context) pkgPostgres.QueryExecutor {
	return pkgPostgres.GetExecutor(ctx, r.pool)
}

func (r *BatchRepository) GetTotalQuantity(ctx context.Context, spaceID, itemID uuid.UUID) (quantity.Quantity, error) {
	query := `
		select coalesce(sum(remaining_quantity), 0)::text
		from batches
		where space_id = $1 and item_id = $2
	`

	var totalText string
	if err := r.getExecutor(ctx).QueryRow(ctx, query, spaceID, itemID).Scan(&totalText); err != nil {
		return quantity.Quantity{}, apperrors.MapPgError(err, nil)
	}

	total, err := quantity.ParseQuantity(totalText)
	if err != nil {
		return quantity.Quantity{}, apperrors.ErrInternalServer.WithInternal(
			fmt.Errorf("parse total batch quantity for item %s: %w", itemID, err),
		)
	}

	return *total, nil
}
