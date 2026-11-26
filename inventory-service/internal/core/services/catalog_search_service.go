package services

import (
	"context"
	"time"

	"github.com/Holarbtw/homestock-backend/inventory-service/internal/core/domain"
	"github.com/Holarbtw/homestock-backend/inventory-service/internal/core/ports"
)

const catalogSearchTimeoutSeconds = 2

type CatalogSearchService struct {
	catalogSearcher ports.CatalogLookup
}

func NewCatalogSearchService(catalogSearcher ports.CatalogLookup) *CatalogSearchService{
	return &CatalogSearchService{
		catalogSearcher: catalogSearcher,
	}
}

func (s *CatalogSearchService) FindByGTIN(ctx context.Context, gtin string) (*domain.CatalogProduct, error){
	timeoutCtx, cancel := context.WithTimeout(ctx, time.Second * catalogSearchTimeoutSeconds)
	defer cancel()

	product, err := s.catalogSearcher.FindByGTIN(timeoutCtx, gtin)
	if err != nil {
		return nil, err
	}

	return product, nil
}