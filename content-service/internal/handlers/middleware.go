package handlers

import (
	"database/sql"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jellyfinhanced/shared/auth"
)

// AuthMiddleware validates tokens and injects user context
func AuthMiddleware(db *sql.DB) chi.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := r.Header.Get("X-Emby-Authorization")
			if token == "" {
				next.ServeHTTP(w, r)
				return
			}

			isValid, userID, _ := auth.ValidateToken(token, db)
			if !isValid {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			r = r.WithContext(auth.SetUserInContext(r.Context(), userID))
			next.ServeHTTP(w, r)
		})
	}
}
