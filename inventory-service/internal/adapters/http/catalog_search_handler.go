package http

import (
	"net/http"

	"github.com/Holarbtw/homestock-backend/inventory-service/internal/core/ports"
	"github.com/Holarbtw/homestock-backend/pkg/apperrors"
	"github.com/Holarbtw/homestock-backend/pkg/httpresponse"
	"github.com/go-chi/chi/v5"
)

type CatalogSearchHandler struct {
	catalogSearchService ports.CatalogSearchService
}

func NewCatalogSearchHandler(catalogSearchService ports.CatalogSearchService) *CatalogSearchHandler {
	return &CatalogSearchHandler{
		catalogSearchService: catalogSearchService,
	}
}

// @Summary Find a catalog product by GTIN
// @Description Returns product details for prefilling a new inventory item. Choose the inventory category when creating the item.
// @Description The catalog accepts 8, 12, 13 or 14 digits and returns a 14-digit GTIN padded with leading zeros.
// @Tags Catalog
// @Security ApiKeyAuth
// @Produce json
// @Param gtin path string true "Product GTIN" example(04600634855338)
// @Success 200 {object} DataResponse[CatalogProductResponse]
// @Failure 400 {object} apperrors.CustomError "INVALID_GTIN"
// @Failure 401 {object} apperrors.CustomError "UNAUTHORIZED, INVALID_TOKEN"
// @Failure 404 {object} apperrors.CustomError "CATALOG_ITEM_NOT_FOUND"
// @Failure 500 {object} apperrors.CustomError "INTERNAL_SERVER_ERROR"
// @Failure 503 {object} apperrors.CustomError "CATALOG_SERVICE_UNAVAILABLE"
// @Failure 504 {object} apperrors.CustomError "CATALOG_SERVICE_TIMEOUT"
// @Router /catalog/product/{gtin} [get]
func (h *CatalogSearchHandler) FindByGTIN(w http.ResponseWriter, r *http.Request) {
	gtin := chi.URLParam(r, "gtin")

	product, err := h.catalogSearchService.FindByGTIN(r.Context(), gtin)
	if err != nil {
		apperrors.RespondWithError(w, r, err)
		return
	}

	httpresponse.WriteJSON(w, http.StatusOK, DataResponse[CatalogProductResponse]{Data: toCatalogProductResponse(product)})
}
