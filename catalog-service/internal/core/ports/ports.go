package ports

import (
	"context"
	"time"

	"github.com/Holarbtw/homestock-backend/catalog-service/internal/domain"
)

type TxManager interface {
	WithinTransaction(ctx context.Context, f func(ctx context.Context) error) error
}

type ProductRepository interface {
	GetByGTIN(ctx context.Context, gtin string) (*domain.Product, error)

	CreateProduct(ctx context.Context, product domain.Product) (*domain.Product, error)
	UpsertSource(ctx context.Context, source domain.ProductSource) error
}

type NegativeLookupCacheRepository interface {
	IsActive(ctx context.Context, gtin, provider string, at time.Time) (bool, error)

	Upsert(ctx context.Context, gtin, provider string, checkedAt, expiresAt time.Time) error
}
