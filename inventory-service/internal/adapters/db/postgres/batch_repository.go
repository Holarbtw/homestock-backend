package postgres

import (
	"context"
	"fmt"

	"github.com/Holarbtw/homestock-backend/inventory-service/internal/core/domain"
	"github.com/Holarbtw/homestock-backend/inventory-service/internal/core/ports"
	"github.com/Holarbtw/homestock-backend/pkg/apperrors"
	pkgPostgres "github.com/Holarbtw/homestock-backend/pkg/postgres"
	"github.com/Holarbtw/homestock-backend/pkg/quantity"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const batchColumns = `id, space_id, item_id, remaining_quantity::text as remaining_quantity,
	purchased_on, added_at, expires_on`

var batchConstraintMap = map[string]*apperrors.CustomError{
	"fk_batches_item": domain.ErrItemNotFound,
}

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

func (r *BatchRepository) Create(ctx context.Context, batch domain.Batch) (*domain.Batch, error) {
	query := `
		insert into batches (space_id, item_id, remaining_quantity, purchased_on, expires_on)
		values ($1, $2, $3, $4, $5)
		returning ` + batchColumns

	record, err := toDBBatch(batch)
	if err != nil {
		return nil, err
	}
	rows, err := r.getExecutor(ctx).Query(ctx, query,
		record.SpaceID, record.ItemID, record.RemainingQuantity, record.PurchasedOn, record.ExpiresOn,
	)
	if err != nil {
		return nil, apperrors.MapPgError(err, batchConstraintMap)
	}

	created, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[batchDB])
	if err != nil {
		return nil, apperrors.MapPgError(err, batchConstraintMap)
	}

	return toDomainBatch(&created)
}

func (r *BatchRepository) ListByItem(ctx context.Context, spaceID, itemID uuid.UUID) ([]domain.Batch, error) {
	query := `
		select ` + batchColumns + `
		from batches
		where space_id = $1 and item_id = $2
		order by added_at desc, id desc
	`
	rows, err := r.getExecutor(ctx).Query(ctx, query, spaceID, itemID)
	if err != nil {
		return nil, apperrors.MapPgError(err, nil)
	}
	records, err := pgx.CollectRows(rows, pgx.RowToStructByName[batchDB])
	if err != nil {
		return nil, apperrors.MapPgError(err, nil)
	}

	batches := make([]domain.Batch, 0, len(records))
	for i := range records {
		batch, err := toDomainBatch(&records[i])
		if err != nil {
			return nil, err
		}
		batches = append(batches, *batch)
	}
	return batches, nil
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
