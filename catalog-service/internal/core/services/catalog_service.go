package services

import (
	"context"

	"github.com/Holarbtw/homestock-backend/catalog-service/internal/core/ports"
	"github.com/Holarbtw/homestock-backend/catalog-service/internal/domain"
)

type CatalogService struct {
	productRepo ports.ProductRepository
}

func NewCatalogService(productRepo ports.ProductRepository) *CatalogService {
	return &CatalogService{productRepo: productRepo}
}

func (s *CatalogService) GetProduct(ctx context.Context, gtin string) (*domain.Product, error) {
	normalizedGTIN, err := normalizeGTIN(gtin)
	if err != nil {
		return nil, err
	}

	return s.productRepo.GetByGTIN(ctx, normalizedGTIN)
}
