package http

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog/log"

	"pipo-edu-project/internal/auth"
	"pipo-edu-project/internal/metrics"
)

type Handler struct {
	Auth    *auth.TokenManager
	Service Service
	Metrics *metrics.Metrics
	CORS    []string
}

type Service interface {
	AuthService
	UserService
	PassService
	GuestService
	EntryService
}

func NewRouter(handler *Handler) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)
	r.Use(CORSMiddleware(handler.CORS))
	r.Use(LoggingMiddleware)
	if handler.Metrics != nil {
		r.Use(handler.Metrics.Middleware)
	}

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	if handler.Metrics != nil {
		r.Get("/metrics", handler.Metrics.Handler().ServeHTTP)
	}

	r.Get("/openapi.yaml", func(w http.ResponseWriter, r *http.Request) {
		root, _ := os.Getwd()
		path := filepath.Join(root, "api", "openapi.yaml")
		http.ServeFile(w, r, path)
	})
	r.Get("/docs", swaggerUIHandler)

	r.Route("/auth", func(r chi.Router) {
		r.Post("/login", handler.HandleLogin)
		r.Post("/refresh", handler.HandleRefresh)
	})

	r.Group(func(r chi.Router) {
		r.Use(auth.AuthMiddleware(handler.Auth))

		r.Route("/users", func(r chi.Router) {
			r.Use(auth.RequireRoles(auth.RoleAdmin))
			r.Post("/", handler.HandleCreateUser)
			r.Get("/", handler.HandleListUsers)
			r.Get("/{id}", handler.HandleGetUser)
			r.Patch("/{id}", handler.HandleUpdateUser)
			r.Delete("/{id}", handler.HandleDeleteUser)
			r.Post("/{id}/restore", handler.HandleRestoreUser)
		})

		r.Route("/passes", func(r chi.Router) {
			r.Get("/search", handler.HandleSearchPasses)
			r.Post("/", handler.HandleCreatePass)
			r.Get("/", handler.HandleListPasses)
			r.Get("/{id}", handler.HandleGetPass)
			r.Patch("/{id}", handler.HandleUpdatePass)
			r.Delete("/{id}", handler.HandleDeletePass)
			r.Post("/{id}/restore", handler.HandleRestorePass)
			r.Post("/{id}/entry", handler.HandleEntry)
			r.Post("/{id}/exit", handler.HandleExit)
		})

		r.Route("/guest-requests", func(r chi.Router) {
			r.Post("/", handler.HandleCreateGuest)
			r.Get("/", handler.HandleListGuest)
			r.Get("/{id}", handler.HandleGetGuest)
			r.Patch("/{id}", handler.HandleUpdateGuest)
			r.Delete("/{id}", handler.HandleDeleteGuest)
			r.Post("/{id}/restore", handler.HandleRestoreGuest)
		})
	})

	log.Info().Msg("router initialized")
	return r
}

func swaggerUIHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`<!doctype html>
<html>
<head>
  <meta charset="utf-8" />
  <title>Swagger UI</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css" />
</head>
<body>
<div id="swagger-ui"></div>
<script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
<script>
window.onload = () => {
  window.ui = SwaggerUIBundle({
    url: '/openapi.yaml',
    dom_id: '#swagger-ui',
  });
};
</script>
</body>
</html>`))
}
