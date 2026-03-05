package middleware

import (
    "context"
    "log"
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
            log.Println("JWTAuth: missing Authorization header")
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }

        tokenStr := strings.TrimPrefix(auth, "Bearer ")

        claims, err := helpers.VerifyJWT(tokenStr)
        if err != nil {
            log.Println("JWTAuth: token verification failed:", err)
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }

        log.Println("JWTAuth: token verified")

        rawRoles, ok := claims["roles"].([]interface{})
        if !ok {
            log.Println("JWTAuth: invalid roles format in JWT")
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }

        roles := make([]string, len(rawRoles))
        for i, r := range rawRoles {
            roles[i] = r.(string)
        }

        log.Println("JWTAuth: roles =", roles)

userID := claims["sub"]

        ctx := context.WithValue(r.Context(), ContextRoles, roles)
ctx = context.WithValue(ctx, ContextUserID, userID)

        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

