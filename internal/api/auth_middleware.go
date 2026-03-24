package api

import (
	"net/http"
	"strings"

	"github.com/yandex-development-2-team/Go/internal/auth"
)

func RequireAuth(jwtSecret string, next http.HandlerFunc) http.HandlerFunc {
	secret := []byte(jwtSecret)

	return func(w http.ResponseWriter, r *http.Request) {
		h := r.Header.Get("Authorization")
		if h == "" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		
		const prefix = "Bearer "
		if !strings.HasPrefix(h, prefix) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		token := strings.TrimSpace(strings.TrimPrefix(h, prefix))
		if token == "" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		_, err := auth.ParseAccessToken(secret, token)
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next(w, r)
	}
}
