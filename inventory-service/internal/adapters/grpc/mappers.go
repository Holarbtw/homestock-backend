package grpc

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/Holarbtw/homestock-backend/inventory-service/internal/core/domain"
	"github.com/Holarbtw/homestock-backend/pkg/apperrors"
	catalogv1 "github.com/Holarbtw/homestock-backend/pkg/pb/catalog/v1"
	"github.com/Holarbtw/homestock-backend/pkg/quantity"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func toDomainError(err error) error {
	if err == nil {
		return nil
	}

	switch status.Code(err){
	case codes.Canceled:
		return domain.ErrCatalogUnavailable.WithInternal(err)
	case codes.DeadlineExceeded:
		return domain.ErrCatalogTimeout.WithInternal(err)
	case codes.InvalidArgument:
		return domain.ErrInvalidGTIN.WithInternal(err)
	case codes.NotFound:
		return domain.ErrCatalogItemNotFound
	case codes.Unavailable:
		return domain.ErrCatalogUnavailable
	default:
		return apperrors.ErrInternalServer.WithInternal(err)
	}
}

func toDomainCatalogProduct(resp *catalogv1.GetProductResponse) (*domain.CatalogProduct, error){
	if resp.GetProduct() == nil{
		return nil, apperrors.ErrInternalServer.WithInternal(errors.New("catalog response is missing product"))
	}

	catalogProduct := &domain.CatalogProduct{
		GTIN: resp.Product.Gtin,
		Name: resp.Product.Name,
		Brand: resp.Product.Brand,
		Category: resp.Product.Category,
	}

	id, err := uuid.Parse(resp.Product.Id)
	if err != nil {
		return nil, apperrors.ErrInternalServer.WithInternal(err)
	}
	catalogProduct.ID = id

	if resp.Product.ImageUrl != nil {
		imageURL, err := url.Parse(*resp.Product.ImageUrl)
		if err != nil {
			return nil, apperrors.ErrInternalServer.WithInternal(err)
		}
		if (imageURL.Scheme != "http" && imageURL.Scheme != "https") || imageURL.Hostname() == "" {
			return nil, apperrors.ErrInternalServer.WithInternal(
				fmt.Errorf("catalog product %s image URL must use HTTP or HTTPS and include a hostname", id),
			)
		}
		catalogProduct.ImageURL = imageURL
	}

	if resp.Product.PackageSize != nil{
		unit, err := toDomainUnit(resp.Product.PackageSize.Unit)
		if err != nil {
			return nil, apperrors.ErrInternalServer.WithInternal(err)
		}
		catalogProduct.Unit = unit
		packageSize, err := quantity.ParseQuantity(resp.Product.PackageSize.Quantity)
		if err != nil {
			return nil, apperrors.ErrInternalServer.WithInternal(err)
		}
		if packageSize.Thousandths() == 0{
			return nil, apperrors.ErrInternalServer.WithInternal(errors.New("catalog package quantity must be positive"))
		}
		catalogProduct.PackageQuantity = packageSize
	}

	if resp.Product.CreatedAt != nil{
		createdAt := resp.Product.CreatedAt.AsTime()
		catalogProduct.CreatedAt = &createdAt
	}

	if resp.Product.UpdatedAt != nil{
		updatedAt := resp.Product.UpdatedAt.AsTime()
		catalogProduct.UpdatedAt = &updatedAt
	}

	return catalogProduct, nil
}

func toDomainUnit(unit catalogv1.Unit) (domain.Unit, error) {
    switch unit {
    case catalogv1.Unit_UNIT_MILLILITER:
        return domain.Milliliter, nil
    case catalogv1.Unit_UNIT_GRAM:
        return domain.Gram, nil
    case catalogv1.Unit_UNIT_PIECE:
        return domain.Pcs, nil
    default:
        return "", fmt.Errorf("unsupported catalog unit: %d", unit)
    }
}
