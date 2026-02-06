package http

import (
	"net/http"
	"time"

	"github.com/rs/zerolog/log"

	"pipo-edu-project/internal/identity"
)

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rw, r)
		duration := time.Since(start)
		userID, _ := identity.UserIDFrom(r.Context())
		role, _ := identity.RoleFrom(r.Context())

		log.Info().
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Int("status", rw.status).
			Dur("duration", duration).
			Str("user_id", userID).
			Str("role", role).
			Msg("request")
	})
}
