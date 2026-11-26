package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	pkgJWT "github.com/Holarbtw/homestock-backend/pkg/jwt"
	pkgLogger "github.com/Holarbtw/homestock-backend/pkg/logger"
	httpSwagger "github.com/swaggo/http-swagger"
)

func NewRouter(itemHandler *ItemHandler, jwtVerifier *pkgJWT.TokenVerifier, catalogSearchHandler *CatalogSearchHandler) http.Handler{
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(pkgLogger.LoggingMiddleware)
	r.Use(middleware.Recoverer)

	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	r.Get("/swagger/*", httpSwagger.WrapHandler)

	r.Group(func(r chi.Router) {
		r.Use(jwtVerifier.AuthMiddleware)

		r.Post("/items", itemHandler.CreateItem)
		r.Get("/items", itemHandler.ListItems)
		r.Get("/items/{item_id}", itemHandler.GetItem)
		r.Patch("/items/{item_id}", itemHandler.UpdateItem)
		r.Delete("/items/{item_id}", itemHandler.DeleteItem)
		r.Get("/catalog/product/{gtin}", catalogSearchHandler.FindByGTIN)
	})

	return r
}