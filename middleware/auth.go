package middleware

import (
	"context"
	"net/http"
	"strings"

	"profile-service/helpers"
)

type contextKey string

const (
    ContextRoles  contextKey = "roles"
    ContextUserID contextKey = "user_id"
    ContextEmail  contextKey = "email"
)


func JWTAuth(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		auth := r.Header.Get("Authorization")
		if auth == "" {
			http.Error(w, "missing token", http.StatusUnauthorized)
			return
		}

		tokenStr := strings.TrimPrefix(auth, "Bearer ")

		claims, err := helpers.VerifyJWT(tokenStr)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		rawRoles := claims["roles"].([]interface{})
		roles := make([]string, len(rawRoles))
		for i, r := range rawRoles {
			roles[i] = r.(string)
		}

		ctx := context.WithValue(r.Context(), ContextRoles, roles)
		ctx = context.WithValue(ctx, "user_id", claims["user_id"])

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

