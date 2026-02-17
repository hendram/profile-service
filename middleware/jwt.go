package middleware

import (
        "context"
        "log"
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

                log.Println("---- JWTAuth middleware hit ----")
                log.Println("Request path:", r.URL.Path)

                authHeader := r.Header.Get("Authorization")
                log.Println("Authorization header:", authHeader)

                if authHeader == "" {
                        log.Println("No Authorization header -> 401")
                        http.Error(w, "Unauthorized", http.StatusUnauthorized)
                        return
                }

                tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
                log.Println("Token extracted:", tokenStr)

                token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
                        log.Println("Parsing token using secret")
                        return []byte(os.Getenv("JWT_SECRET")), nil
                })

                if err != nil {
                        log.Println("Token parse error:", err)
                }

                if err != nil || !token.Valid {
                        log.Println("Invalid token -> 401")
                        http.Error(w, "Unauthorized", http.StatusUnauthorized)
                        return
                }

                log.Println("Token valid")

                claims := token.Claims.(jwt.MapClaims)
                log.Println("Claims:", claims)

                // email
                email, ok := claims["email"].(string)
                if !ok {
                        log.Println("Invalid email claim")
                        http.Error(w, "Invalid token email", http.StatusUnauthorized)
                        return
                }
                log.Println("Email:", email)

                // user id (sub)
                sub, ok := claims["sub"].(float64)
                if !ok {
                        log.Println("Invalid sub claim")
                        http.Error(w, "Invalid token sub", http.StatusUnauthorized)
                        return
                }
                userID := int(sub)
                log.Println("UserID:", userID)

                // roles support both formats
                var roles []string

                if roleStr, ok := claims["role"].(string); ok {
                        roles = []string{roleStr}
                        log.Println("Single role:", roles)
                } else if arr, ok := claims["roles"].([]interface{}); ok {
                        for _, r := range arr {
                                roles = append(roles, r.(string))
                        }
                        log.Println("Roles array:", roles)
                } else {
                        log.Println("No valid roles in token")
                        http.Error(w, "Invalid token role", http.StatusUnauthorized)
                        return
                }

                log.Println("JWT middleware success, passing to next handler")

                ctx := context.WithValue(r.Context(), ContextEmail, email)
                ctx = context.WithValue(ctx, ContextRoles, roles)
                ctx = context.WithValue(ctx, ContextUserID, userID)

                next.ServeHTTP(w, r.WithContext(ctx))
        })
}
