package middleware

import "net/http"

func RequireRoles(allowed ...string) func(http.Handler) http.Handler {

    allowedMap := map[string]bool{}
    for _, r := range allowed {
        allowedMap[r] = true
    }

    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

            roles := r.Context().Value(ContextRoles).([]string)

            for _, role := range roles {
                if allowedMap[role] {
                    next.ServeHTTP(w, r)
                    return
                }
            }

            http.Error(w, "Forbidden", http.StatusForbidden)
        })
    }
}
