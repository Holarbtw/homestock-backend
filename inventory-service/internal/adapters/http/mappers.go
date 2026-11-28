package http

import (
	"encoding/json"
	"errors"
	"net/url"

	"github.com/Holarbtw/homestock-backend/inventory-service/internal/core/domain"
	"github.com/Holarbtw/homestock-backend/inventory-service/internal/core/ports"
	"github.com/Holarbtw/homestock-backend/pkg/apperrors"
	pkgQuantity "github.com/Holarbtw/homestock-backend/pkg/quantity"
)

func toCatalogProductResponse(product *domain.CatalogProduct) CatalogProductResponse {
	return CatalogProductResponse{
		CatalogProductID: product.ID,
		GTIN:             product.GTIN,
		Name:             product.Name,
		Brand:            product.Brand,
		Unit:             string(product.Unit),
		PackageQuantity:  optionalQuantityString(product.PackageQuantity),
		ExternalImageURL: optionalURLString(product.ImageURL),
	}
}

func toItemDetailsResponse(stock *domain.ItemStock, batches []domain.Batch) *ItemDetailsResponse {
	return &ItemDetailsResponse{
		Item:    toItemResponse(stock),
		Batches: toBatchResponses(batches),
	}
}

func toCreateItemResponse(stock *domain.ItemStock, batches []domain.Batch, operation *domain.StockOperation) *CreateItemResponse {
	response := &CreateItemResponse{
		Item:    toItemResponse(stock),
		Batches: toBatchResponses(batches),
	}
	if operation != nil {
		operationResponse := toStockOperationResponse(*operation)
		response.Operation = &operationResponse
	}
	return response
}

func toStockChangeResponse(stock *domain.ItemStock, batches []domain.Batch, operation domain.StockOperation) StockChangeResponse {
	return StockChangeResponse{
		Item:      toItemResponse(stock),
		Batches:   toBatchResponses(batches),
		Operation: toStockOperationResponse(operation),
	}
}

func toItemResponse(stock *domain.ItemStock) ItemResponse {
	item := stock.Item
	return ItemResponse{
		ID:               item.ID,
		SpaceID:          item.SpaceID,
		CatalogProductID: item.CatalogProductID,
		GTIN:             item.GTIN,
		Name:             item.Name,
		Brand:            item.Brand,
		Category:         string(item.Category),
		Unit:             string(item.Unit),
		MinimumQuantity:  optionalQuantityString(item.MinimumQuantity),
		PackageQuantity:  optionalQuantityString(item.PackageQuantity),
		ImageURL:         optionalURLString(item.ExternalImageURL),
		TotalQuantity:    stock.TotalQuantity.String(),
		IsLowStock:       stock.IsLowStock(),
		CreatedAt:        item.CreatedAt.UTC(),
		UpdatedAt:        item.UpdatedAt.UTC(),
	}
}

func toBatchResponses(batches []domain.Batch) []BatchResponse {
	responses := make([]BatchResponse, len(batches))
	for i, batch := range batches {
		responses[i] = toBatchResponse(batch)
	}
	return responses
}

func toBatchResponse(batch domain.Batch) BatchResponse {
	return BatchResponse{
		ID:                batch.ID,
		SpaceID:           batch.SpaceID,
		ItemID:            batch.ItemID,
		RemainingQuantity: batch.RemainingQuantity.String(),
		PurchasedOn:       optionalDateString(batch.PurchasedOn),
		ExpiresOn:         optionalDateString(batch.ExpiresOn),
		AddedAt:           batch.AddedAt.UTC(),
	}
}

func toStockOperationResponse(operation domain.StockOperation) StockOperationResponse {
	lines := make([]StockOperationLineResponse, len(operation.Lines))
	for i, line := range operation.Lines {
		lines[i] = StockOperationLineResponse{
			BatchID:  line.BatchID,
			Quantity: line.Quantity.String(),
		}
	}
	return StockOperationResponse{
		ID:          operation.ID,
		SpaceID:     operation.SpaceID,
		ItemID:      operation.ItemID,
		ActorUserID: operation.ActorUserID,
		Type:        string(operation.Type),
		Quantity:    operation.Quantity.String(),
		OccurredAt:  operation.OccurredAt.UTC(),
		Lines:       lines,
	}
}

func optionalQuantityString(value *pkgQuantity.Quantity) *string {
	if value == nil {
		return nil
	}
	text := value.String()
	return &text
}

func optionalURLString(value *url.URL) *string {
	if value == nil {
		return nil
	}
	text := value.String()
	return &text
}

func optionalDateString(value *domain.Date) *string {
	if value == nil {
		return nil
	}
	text := value.String()
	return &text
}

type fieldCheck struct {
	field string
	err   error
}

func parseBatchRequest(req AddBatchRequest, fieldPrefix string) (*ports.InitialBatchInput, []fieldCheck) {
	amount, amountErr := pkgQuantity.ParseQuantity(req.Quantity)
	purchasedOn, purchaseErr := parseOptionalValue(req.PurchasedOn, domain.ParseDate)
	expiresOn, expiryErr := parseOptionalValue(req.ExpiresOn, domain.ParseDate)
	if amountErr != nil || purchaseErr != nil || expiryErr != nil {
		return nil, []fieldCheck{
			{fieldPrefix + "quantity", amountErr},
			{fieldPrefix + "purchased_on", purchaseErr},
			{fieldPrefix + "expires_on", expiryErr},
		}
	}
	return &ports.InitialBatchInput{
		Quantity:    *amount,
		PurchasedOn: purchasedOn,
		ExpiresOn:   expiresOn,
	}, nil
}

func collectFieldErrors(checks []fieldCheck) ([]apperrors.FieldError, error) {
	var fields []apperrors.FieldError
	for _, check := range checks {
		if check.err == nil {
			continue
		}
		var fieldErr *apperrors.CustomError
		if !errors.As(check.err, &fieldErr) {
			return nil, check.err
		}
		fields = append(fields, apperrors.FieldError{Field: check.field, Code: fieldErr.Code})
	}
	return fields, nil
}

func toAddBatchInput(req AddBatchRequest) (*ports.AddBatchInput, error) {
	batch, checks := parseBatchRequest(req, "")
	fields, err := collectFieldErrors(checks)
	if err != nil {
		return nil, err
	}
	if len(fields) > 0 {
		return nil, domain.ErrInvalidAddBatch.WithFields(fields)
	}
	return &ports.AddBatchInput{
		Quantity:    batch.Quantity,
		PurchasedOn: batch.PurchasedOn,
		ExpiresOn:   batch.ExpiresOn,
	}, nil
}

func toCreateItemInput(req CreateItemRequest) (*ports.CreateItemInput, error) {
	minimum, minimumErr := parseOptionalValue(req.MinimumQuantity, pkgQuantity.ParseQuantity)
	packageQuantity, packageErr := parseOptionalValue(req.PackageQuantity, pkgQuantity.ParseQuantity)
	imageURL, imageErr := parseOptionalValue(req.ExternalImageURL, url.Parse)
	if imageErr != nil {
		imageErr = domain.ErrInvalidExternalImageURL.WithInternal(imageErr)
	}

	checks := []fieldCheck{
		{"minimum_quantity", minimumErr},
		{"package_quantity", packageErr},
		{"external_image_url", imageErr},
	}

	var firstBatch *ports.InitialBatchInput
	if req.FirstBatch != nil {
		var batchChecks []fieldCheck
		firstBatch, batchChecks = parseBatchRequest(*req.FirstBatch, "first_batch.")
		checks = append(checks, batchChecks...)
	}

	fields, err := collectFieldErrors(checks)
	if err != nil {
		return nil, err
	}
	if len(fields) > 0 {
		return nil, domain.ErrInvalidItemCreate.WithFields(fields)
	}

	return &ports.CreateItemInput{
		Name:             req.Name,
		Brand:            req.Brand,
		CatalogProductID: req.CatalogProductID,
		GTIN:             req.GTIN,
		Category:         domain.Category(req.Category),
		Unit:             domain.Unit(req.Unit),
		MinimumQuantity:  minimum,
		PackageQuantity:  packageQuantity,
		ExternalImageURL: imageURL,
		FirstBatch:       firstBatch,
	}, nil
}

func parseOptionalValue[T any](value *string, parse func(string) (*T, error)) (*T, error) {
	if value == nil {
		return nil, nil
	}
	return parse(*value)
}

func toUpdateItemInput(updateRequest UpdateItemRequest) (*ports.UpdateItemInput, error){
	name, err := decodeOptional[string](updateRequest.Name)
	if err != nil {
		return nil, err
	}

	brand, err := decodeOptional[string](updateRequest.Brand)
	if err != nil {
		return nil, err
	}

	category, err := decodeOptional[domain.Category](updateRequest.Category)
	if err != nil {
		return nil, err
	}

	quantityString, err := decodeOptional[string](updateRequest.MinimumQuantity)
	if err != nil {
		return nil, err
	}
	var quantity ports.Optional[pkgQuantity.Quantity]
	quantity.Set = quantityString.Set
	if quantityString.Set && quantityString.Value != nil{
		parsed, err:= pkgQuantity.ParseQuantity(*quantityString.Value)
		if err != nil {
			return nil, err
		}
		quantity.Value = parsed
	}

	return &ports.UpdateItemInput{
		Name: name,
		Brand: brand,
		Category: category,
		MinimumQuantity: quantity,
	}, nil
}

func decodeOptional[T any](raw json.RawMessage) (ports.Optional[T], error){
	if raw == nil{
		return ports.Optional[T]{}, nil
	}

	var value *T
	if err := json.Unmarshal(raw, &value); err != nil{
		return ports.Optional[T]{}, apperrors.ErrInvalidRequest.WithInternal(err)
	}

	return ports.Optional[T]{Set: true, Value: value}, nil
}
