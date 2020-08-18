package shoreline

import (
	"errors"
	"fmt"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
)

// TokenClaims is the token  content
type TokenClaims struct {
	IsServer  string   `json:"svr"`
	UserID    string   `json:"usr"`
	UserRoles []string `json:"roles,omitempty"`
	Extended  bool     `json:"ext,omitempty"`
	jwt.StandardClaims
}

const tokenSignMethod = "HS256"

var errSessionTokenInvalid = errors.New("SessionToken: is invalid")

// UnpackAndVerifyToken validate a shoreline token
func UnpackAndVerifyToken(packedToken string, secret string) (*TokenClaims, error) {
	if packedToken == "" {
		return nil, fmt.Errorf("Invalid token")
	}

	keyFunc := func(t *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	}

	jwtToken, err := jwt.ParseWithClaims(packedToken, &TokenClaims{}, keyFunc)
	if err != nil {
		return nil, err
	}
	if !jwtToken.Valid {
		return nil, errSessionTokenInvalid
	}
	if jwtToken.Method.Alg() != tokenSignMethod {
		return nil, errSessionTokenInvalid
	}

	claims := jwtToken.Claims.(*TokenClaims)

	if !claims.VerifyExpiresAt(time.Now().UTC().Unix(), true) {
		return nil, errSessionTokenInvalid
	}

	return claims, nil
}
