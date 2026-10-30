package grpc

import (
	"github.com/Holarbtw/homestock-backend/catalog-service/internal/domain"
	catalogv1 "github.com/Holarbtw/homestock-backend/pkg/pb/catalog/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func toProtoUnit(u domain.Unit) catalogv1.Unit{
	unit := catalogv1.Unit_UNIT_UNSPECIFIED
	switch u{
	case domain.Milliliter:
		unit = catalogv1.Unit_UNIT_MILLILITER
	case domain.Gram:
		unit = catalogv1.Unit_UNIT_GRAM
	case domain.Pcs:
		unit = catalogv1.Unit_UNIT_PIECE
	} 

	return unit
}

func toProtoProduct(p domain.Product) *catalogv1.Product{
	product := &catalogv1.Product{
		Id: p.ID.String(),
		Gtin: p.GTIN,
		Name: p.Name,
		Brand: p.Brand,
		Category: p.Category,
		CreatedAt: timestamppb.New(p.CreatedAt),
		UpdatedAt: timestamppb.New(p.UpdatedAt),
	}

	var imageUrl *string
	if p.ImageURL != nil{
		s := p.ImageURL.String()
		imageUrl = &s
	}

	product.ImageUrl = imageUrl

	var packageSize *catalogv1.PackageSize
	if p.PackageSize != nil{
		quantity := p.PackageSize.Quantity.String()
		unit := toProtoUnit(p.PackageSize.Unit)

		packageSize = &catalogv1.PackageSize{
			Quantity: quantity,
			Unit: unit,
		}
	}

	product.PackageSize = packageSize

	return product
}