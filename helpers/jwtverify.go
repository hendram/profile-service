package helpers

import (
        "crypto/rsa"
        "encoding/base64"
        "encoding/json"
        "errors"
        "log"
        "math/big"
        "net/http"

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

        log.Println("Fetching JWKS from auth service")

        resp, err := http.Get("http://localhost:8080/.well-known/jwks.json")
        if err != nil {
                log.Println("JWKS fetch failed:", err)
                return err
        }
        defer resp.Body.Close()

        var data jwksResp
        if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
                log.Println("JWKS decode failed:", err)
                return err
        }

        jwks = map[string]*rsa.PublicKey{}

        for _, k := range data.Keys {

                nBytes, _ := base64.RawURLEncoding.DecodeString(k.N)
                eBytes, _ := base64.RawURLEncoding.DecodeString(k.E)

                n := new(big.Int).SetBytes(nBytes)
                e := int(new(big.Int).SetBytes(eBytes).Int64())

                jwks[k.Kid] = &rsa.PublicKey{N: n, E: e}

                log.Println("Loaded JWKS key:", k.Kid)
        }

        return nil
}

func getKey(token *jwt.Token) (interface{}, error) {

        kid, ok := token.Header["kid"].(string)
        if !ok {
                log.Println("JWT error: no kid in header")
                return nil, errors.New("no kid")
        }

        log.Println("JWT header kid:", kid)

        key := jwks[kid]
        if key == nil {

                log.Println("Key not cached, refetching JWKS")

                if err := fetchJWKS(); err != nil {
                        return nil, err
                }

                key = jwks[kid]
        }

        if key == nil {
                log.Println("JWT error: unknown kid:", kid)
                return nil, errors.New("unknown kid")
        }

        return key, nil
}

func VerifyJWT(tokenStr string) (jwt.MapClaims, error) {

        if jwks == nil {
                if err := fetchJWKS(); err != nil {
                        return nil, err
                }
        }

        token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {

                log.Println("JWT alg:", t.Method.Alg())

                if t.Method.Alg() != "RS256" {
                        log.Println("JWT error: invalid algorithm")
                        return nil, errors.New("invalid alg")
                }

                return getKey(t)
        })

        if err != nil {
                log.Println("JWT parse error:", err)
                return nil, err
        }

        if !token.Valid {
                log.Println("JWT invalid")
                return nil, errors.New("invalid token")
        }

        claims := token.Claims.(jwt.MapClaims)

        log.Println("JWT claims:", claims)

        if claims["iss"] != "auth-service" {
                log.Println("JWT issuer mismatch:", claims["iss"])
                return nil, errors.New("invalid issuer")
        }

        if claims["aud"] != "buzzer-service" {
                log.Println("JWT audience mismatch:", claims["aud"])
                return nil, errors.New("invalid audience")
        }

        log.Println("JWT verified successfully")

        return claims, nil
}

