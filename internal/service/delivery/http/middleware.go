package deliveryhttp

import (
	"net/http"
	"strings"

	"github.com/plastm0oo/sop_contest/internal/auth"
	"github.com/plastm0oo/sop_contest/internal/service"
)

func (h *handler) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			writeError(w, http.StatusUnauthorized, "нет access token")
			return
		}

		parts := strings.Fields(authHeader)
		if len(parts) != 2 || parts[0] != "Bearer" {
			writeError(w, http.StatusUnauthorized, "неверный формат Authorization header")
			return
		}

		claims, err := auth.ParseAccessToken(parts[1], h.jwtSecret)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "недействительный access token")
			return
		}

		ctx := service.WithAuthContext(r.Context(), service.AuthContext{
			UserID: claims.UserID,
			Email:  claims.Email,
			Role:   claims.Role,
		})

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (h *handler) adminMiddleware(next http.Handler) http.Handler {
	return h.authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		role, ok := service.RoleFromContext(r.Context())
		if !ok || role != "admin" {
			writeError(w, http.StatusForbidden, "недостаточно прав")
			return
		}

		next.ServeHTTP(w, r)
	}))
}
