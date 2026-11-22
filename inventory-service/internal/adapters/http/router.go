package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	pkgJWT "github.com/Holarbtw/homestock-backend/pkg/jwt"
	pkgLogger "github.com/Holarbtw/homestock-backend/pkg/logger"
	httpSwagger "github.com/swaggo/http-swagger"
)

func NewRouter(handler *ItemHandler, jwtVerifier *pkgJWT.TokenVerifier) http.Handler{
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

		r.Post("/items", handler.CreateItem)
		r.Get("/items", handler.ListItems)
		r.Get("/items/{item_id}", handler.GetItem)
		r.Patch("/items/{item_id}", handler.UpdateItem)
		r.Delete("/items/{item_id}", handler.DeleteItem)
	})

	return r
}