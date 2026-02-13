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
    ContextEmail contextKey = "email"
    ContextRoles contextKey = "roles"
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

        email := claims["email"].(string)

        rolesInterface := claims["roles"].([]interface{})
        roles := make([]string, len(rolesInterface))
        for i, r := range rolesInterface {
            roles[i] = r.(string)
        }

        ctx := context.WithValue(r.Context(), ContextEmail, email)
        ctx = context.WithValue(ctx, ContextRoles, roles)

        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
