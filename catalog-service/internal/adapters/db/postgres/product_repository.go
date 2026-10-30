package postgres

import (
	"context"

	"github.com/Holarbtw/homestock-backend/catalog-service/internal/domain"
	"github.com/Holarbtw/homestock-backend/pkg/apperrors"
	pkgPostgres "github.com/Holarbtw/homestock-backend/pkg/postgres"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ProductRepository struct {
	pool *pgxpool.Pool
}

func NewProductRepository(pool *pgxpool.Pool) *ProductRepository {
	return &ProductRepository{pool: pool}
}

func (r *ProductRepository) getExecutor(ctx context.Context) pkgPostgres.QueryExecutor {
	return pkgPostgres.GetExecutor(ctx, r.pool)
}

func (r *ProductRepository) GetByGTIN(ctx context.Context, gtin string) (*domain.Product, error) {
	query := `select id, gtin, name, brand, category, package_quantity::text as package_quantity, 
				package_unit, image_url, created_at, updated_at
			  from products
			  where gtin = $1`

	rows, err := r.getExecutor(ctx).Query(ctx, query, gtin)
	if err != nil {
		return nil, apperrors.MapPgError(err, nil)
	}

	record, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[productDB])
	if err != nil {
		return nil, apperrors.MapPgError(err, nil)
	}

	return toDomainProduct(&record)
}

func (r *ProductRepository) CreateProduct(ctx context.Context, product domain.Product) (*domain.Product, error) {
	query := `
		insert into products (gtin, name, brand, category, package_quantity, package_unit, image_url)
		values ($1, $2, $3, $4, $5, $6, $7)
		returning id, gtin, name, brand, category, package_quantity::text as package_quantity,
			package_unit, image_url, created_at, updated_at
	`

	var packageQuantity, packageUnit, imageURL *string
	if product.PackageSize != nil {
		quantity := product.PackageSize.Quantity.String()
		unit := string(product.PackageSize.Unit)
		packageQuantity = &quantity
		packageUnit = &unit
	}
	if product.ImageURL != nil {
		value := product.ImageURL.String()
		imageURL = &value
	}

	rows, err := r.getExecutor(ctx).Query(ctx, query,
		product.GTIN, product.Name, product.Brand, product.Category,
		packageQuantity, packageUnit, imageURL,
	)
	if err != nil {
		return nil, apperrors.MapPgError(err, nil)
	}

	record, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[productDB])
	if err != nil {
		return nil, apperrors.MapPgError(err, nil)
	}

	return toDomainProduct(&record)
}

func (r *ProductRepository) UpsertSource(ctx context.Context, source domain.ProductSource) error {
	query := `
		insert into product_sources (
			product_id, provider, external_code, raw_payload, fetched_at, source_updated_at
		)
		values ($1, $2, $3, $4, $5, $6)
		on conflict (provider, external_code) do update
		set product_id = excluded.product_id,
			raw_payload = excluded.raw_payload,
			fetched_at = excluded.fetched_at,
			source_updated_at = excluded.source_updated_at
	`

	_, err := r.getExecutor(ctx).Exec(ctx, query,
		source.ProductID, source.Provider, source.ExternalCode,
		source.RawPayload, source.FetchedAt, source.SourceUpdatedAt,
	)
	return apperrors.MapPgError(err, nil)
}
