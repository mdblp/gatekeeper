package shoreline

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path"
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

// XTidepoolSessionToken in the HTTP header
const XTidepoolSessionToken = "x-tidepool-session-token"

// XTidepoolTraceSession in the HTTP header
const XTidepoolTraceSession = "x-tidepool-trace-session"

// XTidepoolServerName for server login
const XTidepoolServerName = "x-tidepool-server-name"

// XTidepoolServerSecret for server login
const XTidepoolServerSecret = "x-tidepool-server-secret"

var tokenSignMethod = jwt.SigningMethodHS256.Name
var errSessionTokenInvalid = errors.New("SessionToken is invalid")

// UnpackAndVerifyToken validate a shoreline token
func UnpackAndVerifyToken(packedToken string, secret string) (*TokenClaims, error) {
	if packedToken == "" {
		return nil, fmt.Errorf("Invalid token")
	}

	keyFunc := func(t *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	}

	parser := new(jwt.Parser)
	parser.ValidMethods = []string{tokenSignMethod}
	parser.SkipClaimsValidation = false
	parser.UseJSONNumber = true
	jwtToken, err := parser.ParseWithClaims(packedToken, &TokenClaims{}, keyFunc)
	if err != nil {
		return nil, err
	}
	if !jwtToken.Valid {
		return nil, errSessionTokenInvalid
	}

	claims := jwtToken.Claims.(*TokenClaims)

	if !claims.VerifyExpiresAt(time.Now().UTC().Unix(), true) {
		return nil, errSessionTokenInvalid
	}

	return claims, nil
}

// ServerLogin with shoreline
func ServerLogin() (string, error) {
	shorelineHost := os.Getenv("SHORELINE_HOST")
	if shorelineHost == "" {
		return "", fmt.Errorf("Missing env var SHORELINE_HOST")
	}
	serverSecret := os.Getenv("SERVER_SECRET")
	if serverSecret == "" {
		return "", fmt.Errorf("Missing env var SERVER_SECRET")
	}

	shorelineURL, err := url.Parse(shorelineHost)
	if err != nil {
		return "", nil
	}

	shorelineURL.Path = path.Join(shorelineURL.Path, "/serverlogin")
	request, err := http.NewRequest(http.MethodPost, shorelineURL.String(), nil)
	if err != nil {
		return "", err
	}
	request.Header.Add(XTidepoolServerName, "gatekeeper")
	request.Header.Add(XTidepoolServerSecret, serverSecret)

	hc := http.Client{}
	response, err := hc.Do(request)
	if err != nil {
		return "", err
	}

	defer response.Body.Close()

	if response.StatusCode != 200 {
		return "", fmt.Errorf("Invalid response from %s: %d", shorelineURL.String(), response.StatusCode)
	}

	return response.Header.Get(XTidepoolSessionToken), nil
}
