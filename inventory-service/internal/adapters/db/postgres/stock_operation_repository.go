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

var stockOperationConstraintMap = map[string]*apperrors.CustomError{
	"fk_stock_operations_item": domain.ErrItemNotFound,
}

var _ ports.StockOperationRepository = (*StockOperationRepository)(nil)

type StockOperationRepository struct {
	pool *pgxpool.Pool
}

func NewStockOperationRepository(pool *pgxpool.Pool) *StockOperationRepository {
	return &StockOperationRepository{pool: pool}
}

func (r *StockOperationRepository) CreateReplenishment(ctx context.Context, batch *domain.Batch, actorUserID uuid.UUID) (*domain.StockOperation, error) {
	if batch == nil {
		return nil, apperrors.ErrInternalServer.WithInternal(
			errors.New("replenishment creation requires a batch"),
		)
	}
	tx, ok := pkgPostgres.GetExecutor(ctx, r.pool).(pgx.Tx)
	if !ok {
		return nil, apperrors.ErrInternalServer.WithInternal(
			errors.New("replenishment creation requires WithinTransaction context"),
		)
	}

	const insertOperation = `
		insert into stock_operations (space_id, item_id, actor_user_id, type, quantity)
		values ($1, $2, $3, $4, $5)
		returning id, space_id, item_id, actor_user_id, type, quantity::text as quantity, occurred_at
	`
	rows, err := tx.Query(ctx, insertOperation,
		batch.SpaceID, batch.ItemID, actorUserID, string(domain.Replenish), batch.RemainingQuantity.String(),
	)
	if err != nil {
		return nil, apperrors.MapPgError(err, stockOperationConstraintMap)
	}
	record, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[stockOperationDB])
	if err != nil {
		return nil, apperrors.MapPgError(err, stockOperationConstraintMap)
	}
	created, err := toDomainStockOperation(&record)
	if err != nil {
		return nil, err
	}

	const insertLine = `
		insert into stock_operation_lines (space_id, operation_id, item_id, batch_id, quantity)
		values ($1, $2, $3, $4, $5)
	`
	line := domain.StockOperationLine{
		SpaceID:     created.SpaceID,
		OperationID: created.ID,
		ItemID:      created.ItemID,
		BatchID:     batch.ID,
		Quantity:    created.Quantity,
	}
	if _, err := tx.Exec(ctx, insertLine, line.SpaceID, line.OperationID, line.ItemID, line.BatchID, line.Quantity.String()); err != nil {
		return nil, apperrors.MapPgError(err, nil)
	}
	created.Lines = []domain.StockOperationLine{line}
	return created, nil
}
