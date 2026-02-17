package middleware

import (
    "log"
    "net/http"
)

func RequireRoles(allowed ...string) func(http.Handler) http.Handler {

    allowedMap := map[string]bool{}
    for _, r := range allowed {
        allowedMap[r] = true
    }

    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

            val := r.Context().Value(ContextRoles)
            if val == nil {
                log.Println("Access denied: no roles found in context")
                http.Error(w, "Forbidden", http.StatusForbidden)
                return
            }

            roles := val.([]string)
            log.Println("Roles from context:", roles)

            for _, role := range roles {
                if allowedMap[role] {
                    log.Println("Access granted for role:", role)
                    next.ServeHTTP(w, r)
                    return
                }
            }

            log.Println("Access denied: no allowed roles matched")
            http.Error(w, "Forbidden", http.StatusForbidden)
        })
    }
}
