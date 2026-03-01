package helpers

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
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
		return nil, errors.New("no kid")
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

func VerifyJWT(tokenStr string) (jwt.MapClaims, error) {

	if jwks == nil {
		if err := fetchJWKS(); err != nil {
			return nil, err
		}
	}

	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {

		if t.Method.Alg() != "RS256" {
			return nil, errors.New("invalid alg")
		}

		return getKey(t)
	})

	if err != nil || !token.Valid {
		return nil, errors.New("invalid token")
	}

	claims := token.Claims.(jwt.MapClaims)

	// issuer check
	if claims["iss"] != "auth-service" {
		return nil, errors.New("invalid issuer")
	}

	// audience check
	if claims["aud"] != "buzzer-service" {
		return nil, errors.New("invalid audience")
	}

	return claims, nil
}
