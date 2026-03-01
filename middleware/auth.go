package middleware

import (
    "context"
    "crypto/rsa"
    "encoding/base64"
    "encoding/json"
    "errors"
    "math/big"
    "net/http"
    "strings"

    "github.com/golang-jwt/jwt/v5"
)

var jwks map[string]*rsa.PublicKey

type jwk struct {
    Kid string `json:"kid"`
    N   string `json:"n"`
    E   string `json:"e"`
}

type jwksResp struct {
    Keys []jwk `json:"keys"`
}

func fetchJWKS() error {

    resp, err := http.Get("http://auth-service:8080/.well-known/jwks.json")
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    var data jwksResp
    if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
        return err
    }

    jwks = map[string]*rsa.PublicKey{}

    for _, k := range data.Keys {

        nBytes, _ := base64.RawURLEncoding.DecodeString(k.N)
        eBytes, _ := base64.RawURLEncoding.DecodeString(k.E)

        n := new(big.Int).SetBytes(nBytes)
        e := int(new(big.Int).SetBytes(eBytes).Int64())

        jwks[k.Kid] = &rsa.PublicKey{N: n, E: e}
    }

    return nil
}

func getKey(token *jwt.Token) (interface{}, error) {

    kid, ok := token.Header["kid"].(string)
    if !ok {
        return nil, errors.New("no kid in token")
    }

    key := jwks[kid]
    if key == nil {

        if err := fetchJWKS(); err != nil {
            return nil, err
        }

        key = jwks[kid]
    }

    if key == nil {
        return nil, errors.New("unknown kid")
    }

    return key, nil
}

func JWTAuth(next http.Handler) http.Handler {

    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

        auth := r.Header.Get("Authorization")
        if auth == "" {
            http.Error(w, "missing token", http.StatusUnauthorized)
            return
        }

        tokenStr := strings.TrimPrefix(auth, "Bearer ")

        if jwks == nil {
            if err := fetchJWKS(); err != nil {
                http.Error(w, "auth unavailable", 500)
                return
            }
        }

        token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {

            if t.Method.Alg() != "RS256" {
                return nil, errors.New("invalid alg")
            }

            return getKey(t)
        })

        if err != nil || !token.Valid {
            http.Error(w, "invalid token", 401)
            return
        }

        claims := token.Claims.(jwt.MapClaims)

        // convert roles → []string
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
