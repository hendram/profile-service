package middleware

import (
	"context"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const (
	ContextEmail  contextKey = "email"
	ContextRoles  contextKey = "roles"
	ContextUserID contextKey = "user_id"
)

func JWTAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("JWT_SECRET")), nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		claims := token.Claims.(jwt.MapClaims)

		// email
		email, ok := claims["email"].(string)
		if !ok {
			http.Error(w, "Invalid token email", http.StatusUnauthorized)
			return
		}

		// user id (sub)
		sub, ok := claims["sub"].(float64)
		if !ok {
			http.Error(w, "Invalid token sub", http.StatusUnauthorized)
			return
		}
		userID := int(sub)

		// roles support both formats
		var roles []string

		if roleStr, ok := claims["role"].(string); ok {
			roles = []string{roleStr}
		} else if arr, ok := claims["roles"].([]interface{}); ok {
			for _, r := range arr {
				roles = append(roles, r.(string))
			}
		} else {
			http.Error(w, "Invalid token role", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), ContextEmail, email)
		ctx = context.WithValue(ctx, ContextRoles, roles)
		ctx = context.WithValue(ctx, ContextUserID, userID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
