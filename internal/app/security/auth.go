package security

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
)

// LoadBearerToken reads the token from disk.
func LoadBearerToken(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read bearer token: %w", err)
	}
	token := strings.TrimSpace(string(data))
	if token == "" {
		return "", errors.New("bearer token file empty")
	}
	return token, nil
}

// BearerAuthMiddleware enforces a static bearer token.
func BearerAuthMiddleware(token string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			auth := r.Header.Get("Authorization")
			if auth == "" || !strings.HasPrefix(auth, "Bearer ") {
				http.Error(w, "missing bearer token", http.StatusUnauthorized)
				return
			}
			if strings.TrimPrefix(auth, "Bearer ") != token {
				http.Error(w, "invalid bearer token", http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
