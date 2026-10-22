package http

import (
	"net/http"

	pkgJWT "github.com/Holarbtw/homestock-backend/pkg/jwt"
	pkgLogger "github.com/Holarbtw/homestock-backend/pkg/logger"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger"
)

const maxRequestBodySize = 1 << 20 // 1 MB

func maxBodySizeMiddleware(limit int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, limit)
			next.ServeHTTP(w, r)
		})
	}
}

func NewRouter(handler *AuthHandler, jwtVerifier *pkgJWT.TokenVerifier) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(maxBodySizeMiddleware(maxRequestBodySize))
	r.Use(pkgLogger.LoggingMiddleware)
	r.Use(middleware.Recoverer)

	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	r.Get("/swagger/*", httpSwagger.WrapHandler)

	r.Group(func(r chi.Router) {
		r.Post("/register", handler.Register)
		r.Post("/login", handler.Login)
		r.Post("/guest-login", handler.LoginGuest)
		r.Post("/refresh", handler.RefreshToken)
		r.Post("/logout", handler.Logout)
	})

	r.Group(func(r chi.Router) {
		r.Use(jwtVerifier.AuthMiddleware)

		r.Get("/profile", handler.GetProfile)
		r.Post("/link-guest", handler.LinkGuest)
		r.Patch("/password", handler.ChangePassword)
		r.Patch("/email", handler.ChangeEmail)
		r.Patch("/profile", handler.UpdateProfile)
	})

	return r
}
