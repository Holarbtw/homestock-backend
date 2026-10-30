package grpc

import (
	"context"

	"github.com/Holarbtw/homestock-backend/catalog-service/internal/core/ports"
	catalogv1 "github.com/Holarbtw/homestock-backend/pkg/pb/catalog/v1"
)

type CatalogServer struct {
	catalogv1.UnimplementedCatalogServiceServer
	catalogService ports.CatalogService
}

func NewCatalogServer(catalogService ports.CatalogService) *CatalogServer{
	return &CatalogServer{
		catalogService: catalogService,
	}
}

func (s *CatalogServer) GetProduct(ctx context.Context, req *catalogv1.GetProductRequest) (*catalogv1.GetProductResponse, error){
	gtin := req.GetGtin()

	product, err := s.catalogService.GetProduct(ctx, gtin)
	if err != nil {
		return nil, err
	}

	return &catalogv1.GetProductResponse{
		Product: toProtoProduct(*product),
	}, nil
}
