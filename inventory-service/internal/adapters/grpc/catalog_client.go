package grpc

import (
	"context"

	"github.com/Holarbtw/homestock-backend/inventory-service/internal/core/domain"
	catalogv1 "github.com/Holarbtw/homestock-backend/pkg/pb/catalog/v1"
)

type CatalogClient struct{
	client catalogv1.CatalogServiceClient
}

func NewCatalogClient(client catalogv1.CatalogServiceClient) *CatalogClient{
	return &CatalogClient{
		client: client,
	}
}

func (c *CatalogClient) FindByGTIN(ctx context.Context, gtin string) (*domain.CatalogProduct, error){
	resp, err := c.client.GetProduct(ctx, &catalogv1.GetProductRequest{
		Gtin: gtin,
	})
	
	if err != nil {
		return nil, toDomainError(err)
	}

	return toDomainCatalogProduct(resp)
}