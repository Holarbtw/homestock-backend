package postgres

import (
	"fmt"
	"net/url"

	"github.com/Holarbtw/homestock-backend/catalog-service/internal/domain"
	"github.com/Holarbtw/homestock-backend/pkg/apperrors"
	pkgQuantity "github.com/Holarbtw/homestock-backend/pkg/quantity"
)

func toDomainProduct(p *productDB) (*domain.Product, error) {
	if p == nil {
		return nil, nil
	}

	product := domain.Product{
		ID:        p.ID,
		GTIN:      p.GTIN,
		Name:      p.Name,
		Brand:     p.Brand,
		Category:  p.Category,
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	}

	var parsedURL *url.URL
	var err error
	if p.ImageURL != nil {
		parsedURL, err = url.Parse(*p.ImageURL)
		if err != nil {
			return nil, apperrors.ErrInternalServer.WithInternal(err)
		}
		if (parsedURL.Scheme != "http" && parsedURL.Scheme != "https") || parsedURL.Hostname() == "" {
			return nil, apperrors.ErrInternalServer.WithInternal(
				fmt.Errorf("%w: product %s has an invalid image URL", errInconsistentProductData, p.ID),
			)
		}
	}

	product.ImageURL = parsedURL

	if p.PackageQuantity == nil && p.PackageUnit == nil {
		return &product, nil
	}

	if p.PackageQuantity != nil && p.PackageUnit != nil {
		var unit domain.Unit
		switch domain.Unit(*p.PackageUnit) {
		case domain.Milliliter:
			unit = domain.Milliliter
		case domain.Gram:
			unit = domain.Gram
		case domain.Pcs:
			unit = domain.Pcs
		default:
			return nil, apperrors.ErrInternalServer.WithInternal(errInconsistentProductData)
		}

		var quantity *pkgQuantity.Quantity
		quantity, err = pkgQuantity.ParseQuantity(*p.PackageQuantity)
		if err != nil {
			return nil, apperrors.ErrInternalServer.WithInternal(err)
		}

		product.PackageSize = &domain.PackageSize{
			Quantity: *quantity,
			Unit:     unit,
		}

		return &product, nil
	}

	return nil, apperrors.ErrInternalServer.WithInternal(errInconsistentProductData)
}
