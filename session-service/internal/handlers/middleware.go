package handlers

import (
	"database/sql"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jellyfinhanced/shared/auth"
)

// AuthMiddleware validates tokens and injects user context (reuses shared auth)
func AuthMiddleware(db *sql.DB) chi.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := r.Header.Get("X-Emby-Authorization")
			if token == "" {
				// Some endpoints may allow anonymous access
				next.ServeHTTP(w, r)
				return
			}

			// Validate token using shared auth middleware
			isValid, userID, _ := auth.ValidateToken(token, db)
			if !isValid {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Inject user into context
			r = r.WithContext(auth.SetUserInContext(r.Context(), userID))
			next.ServeHTTP(w, r)
		})
	}
}
