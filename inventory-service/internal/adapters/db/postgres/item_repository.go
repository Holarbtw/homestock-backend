package postgres

import (
	"context"
	"errors"

	"github.com/Holarbtw/homestock-backend/inventory-service/internal/domain"
	"github.com/Holarbtw/homestock-backend/inventory-service/internal/ports"
	"github.com/Holarbtw/homestock-backend/pkg/apperrors"
	pkgPostgres "github.com/Holarbtw/homestock-backend/pkg/postgres"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const itemColumns = `id, space_id, catalog_product_id, gtin, name, brand, category, unit,
	minimum_quantity::text as minimum_quantity, package_quantity::text as package_quantity,
	external_image_url, image_object_key, created_at, updated_at, deleted_at`

var itemConstraintMap = map[string]*apperrors.CustomError{
	"inventory_items_space_id_fkey": domain.ErrSpaceNotFound,
}

var _ ports.ItemRepository = (*ItemRepository)(nil)

type ItemRepository struct {
	pool *pgxpool.Pool
}

func NewItemRepository(pool *pgxpool.Pool) *ItemRepository {
	return &ItemRepository{pool: pool}
}

func (r *ItemRepository) getExecutor(ctx context.Context) pkgPostgres.QueryExecutor {
	return pkgPostgres.GetExecutor(ctx, r.pool)
}

func (r *ItemRepository) Create(ctx context.Context, item domain.InventoryItem) (*domain.InventoryItem, error) {
	query := `
		insert into inventory_items (
			space_id, catalog_product_id, gtin, name, brand, category, unit,
			minimum_quantity, package_quantity, external_image_url, image_object_key
		)
		values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		returning ` + itemColumns

	record := toDBItem(item)
	rows, err := r.getExecutor(ctx).Query(ctx, query,
		record.SpaceID, record.CatalogProductID, record.GTIN, record.Name, record.Brand,
		record.Category, record.Unit, record.MinimumQuantity, record.PackageQuantity,
		record.ExternalImageURL, record.ImageObjectKey,
	)
	if err != nil {
		return nil, mapItemError(err)
	}

	created, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[itemDB])
	if err != nil {
		return nil, mapItemError(err)
	}

	return toDomainItem(&created)
}

func (r *ItemRepository) GetByID(ctx context.Context, spaceID, itemID uuid.UUID) (*domain.InventoryItem, error) {
	query := `
		select ` + itemColumns + `
		from inventory_items
		where space_id = $1 and id = $2 and deleted_at is null
	`

	rows, err := r.getExecutor(ctx).Query(ctx, query, spaceID, itemID)
	if err != nil {
		return nil, mapItemError(err)
	}

	record, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[itemDB])
	if err != nil {
		return nil, mapItemError(err)
	}

	return toDomainItem(&record)
}

func (r *ItemRepository) GetByIDForUpdate(ctx context.Context, spaceID, itemID uuid.UUID) (*domain.InventoryItem, error) {
	query := `
		select ` + itemColumns + `
		from inventory_items
		where space_id = $1 and id = $2 and deleted_at is null
		for update
	`

	rows, err := r.getExecutor(ctx).Query(ctx, query, spaceID, itemID)
	if err != nil {
		return nil, mapItemError(err)
	}

	record, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[itemDB])
	if err != nil {
		return nil, mapItemError(err)
	}

	return toDomainItem(&record)
}

func (r *ItemRepository) ListBySpace(ctx context.Context, spaceID uuid.UUID, limit, offset int) ([]domain.InventoryItem, error) {
	query := `
		select ` + itemColumns + `
		from inventory_items
		where space_id = $1 and deleted_at is null
		order by created_at desc, id desc
		limit $2 offset $3
	`

	rows, err := r.getExecutor(ctx).Query(ctx, query, spaceID, limit, offset)
	if err != nil {
		return nil, mapItemError(err)
	}

	records, err := pgx.CollectRows(rows, pgx.RowToStructByName[itemDB])
	if err != nil {
		return nil, mapItemError(err)
	}

	items := make([]domain.InventoryItem, 0, len(records))
	for i := range records {
		item, err := toDomainItem(&records[i])
		if err != nil {
			return nil, err
		}
		items = append(items, *item)
	}

	return items, nil
}

func (r *ItemRepository) Update(ctx context.Context, spaceID, itemID uuid.UUID, params ports.UpdateItemParams) (*domain.InventoryItem, error) {
	query := `
		update inventory_items
		set name = $3,
			brand = $4,
			category = $5,
			minimum_quantity = $6,
			updated_at = now()
		where space_id = $1 and id = $2 and deleted_at is null
		returning ` + itemColumns

	var minimumQuantity *string
	if params.MinimumQuantity != nil {
		value := params.MinimumQuantity.String()
		minimumQuantity = &value
	}
	rows, err := r.getExecutor(ctx).Query(ctx, query,
		spaceID, itemID, params.Name, params.Brand, string(params.Category), minimumQuantity,
	)
	if err != nil {
		return nil, mapItemError(err)
	}

	updated, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[itemDB])
	if err != nil {
		return nil, mapItemError(err)
	}

	return toDomainItem(&updated)
}

func (r *ItemRepository) Delete(ctx context.Context, spaceID, itemID uuid.UUID) error {
	query := `
		update inventory_items
		set deleted_at = now(), updated_at = now()
		where space_id = $1 and id = $2 and deleted_at is null
	`

	tag, err := r.getExecutor(ctx).Exec(ctx, query, spaceID, itemID)
	if err != nil {
		return mapItemError(err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrItemNotFound
	}

	return nil
}

func mapItemError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrItemNotFound
	}
	return apperrors.MapPgError(err, itemConstraintMap)
}
