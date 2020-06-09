package middleware

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strings"

	gonnect "github.com/craftamap/atlas-gonnect"
	"github.com/dgrijalva/jwt-go"
)

const JWT_PARAM = "jwt"
const AUTH_HEADER = "authorization"

type AuthenticationMiddleware struct {
	h     http.Handler
	addon *gonnect.Addon
}

func extractUnverifiedClaims(tokenStr string, validator jwt.Keyfunc) (jwt.MapClaims, bool) {
	token, _ := jwt.Parse(tokenStr, validator)
	if token == nil {
		return nil, false
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		return claims, true
	} else {
		log.Printf("Invalid JWT Token")
		return nil, false
	}
}

func CreateQueryStringHash(req *http.Request, checkBodyForParam bool, baseUrlString string) string {
	const CANONICAL_QUERY_SEPARATOR = "&"
	canonicalizeUri := func() string {
		path := req.URL.Path
		//TODO: Handle error here
		baseUrl, _ := url.Parse(baseUrlString)
		baseUrlPath := baseUrl.Path

		// If path and baseUrlPath have the same path, trim the path
		if strings.HasPrefix(path, baseUrlPath) {
			path = strings.TrimPrefix(path, baseUrlPath)
		}

		// if the path is an empty string now, this means the baseUrlPath AND the path were "/".
		// We therefore return "/"
		// Otherwise, the path is currently the path after the baseUrlPath
		if len(path) == 0 {
			return "/"
		}

		// If the separator is not URL encoded then the following URLs have the same query-string-hash:
		//   https://djtest9.jira-dev.com/rest/api/2/project&a=b?x=y
		//   https://djtest9.jira-dev.com/rest/api/2/project?a=b&x=y
		strings.ReplaceAll(path, CANONICAL_QUERY_SEPARATOR, url.QueryEscape(CANONICAL_QUERY_SEPARATOR))

		if !strings.HasPrefix(path, "/") {
			path = "/" + path
		}

		if len(path) > 1 && strings.HasSuffix(path, "/") {
			path = strings.TrimSuffix(path, "/")
		}

		return path
	}

	canonicalizeQueryString := func() string {
		queryParams := req.URL.Query()

		if checkBodyForParam && len(queryParams) == 0 && (strings.ToUpper(req.Method) == "POST" || strings.ToUpper(req.Method) == "PUT") {
			bodyReader, err := req.GetBody()
			if err == nil {
				bodyContent, err := ioutil.ReadAll(bodyReader)
				if err == nil {
					queryParams, _ = url.ParseQuery(string(bodyContent))
				}
			}
		}

		sortedQueryStrings := make([]string, 0)
		query := make([]string, 0)
		for key := range queryParams {
			if key != "jwt" {
				query = append(query, key)
			}
		}
		sort.Strings(query)
		for _, key := range query {
			if key == "__proto__" {
				continue
			}

			param := queryParams[key]
			sort.Strings(param)
			for idx, value := range param {
				param[idx] = url.QueryEscape(value)
			}
			paramValue := strings.Join(param, ",")
			sortedQueryStrings = append(sortedQueryStrings, url.QueryEscape(key)+"="+paramValue)
		}
		return strings.Join(sortedQueryStrings, "&")
	}

	createCanonicalRequest := func() string {
		return strings.ToUpper(req.Method) +
			CANONICAL_QUERY_SEPARATOR +
			canonicalizeUri() +
			CANONICAL_QUERY_SEPARATOR +
			canonicalizeQueryString()
	}

	s := createCanonicalRequest()
	h := sha256.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}

func (h AuthenticationMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	//TODO: Add better logging here
	//TODO: Add AC_OPTS no-auth

	extractJwt := func() (string, bool) {
		var tokenInQuery = r.URL.Query().Get(JWT_PARAM)
		if tokenInQuery == "" && r.Body == http.NoBody {
			return "", false
		}

		tokenInBody := r.PostFormValue(JWT_PARAM)

		if tokenInQuery != "" && tokenInBody != "" {
			return "", false
		}

		var token string
		if tokenInBody != "" {
			token = tokenInBody
		} else {
			token = tokenInQuery
		}

		authHeader := r.Header.Get(AUTH_HEADER)
		if authHeader != "" && strings.HasPrefix(authHeader, "JWT ") {
			if token != "" {

			} else {
				token = strings.TrimPrefix(authHeader, "JWT ")
			}
		}

		// TODO: JS implements r.Query().Get(TOKEN_KEY_PARAM) and r.Query().Get(TOKEN_KEY_HEADER) as possible
		// Headers. However, it is marked as deprecated - we should follow the development of the js library
		// and see if it gets removed. For now, this should work

		if token == "" {
			return "", false
		}
		return token, true
	}

	token, ok := extractJwt()
	if !ok {
		w.WriteHeader(401)
		h.addon.Logger.Warn("Could not find auth data on request")
		return
	}

	unverifiedClaims, ok := extractUnverifiedClaims(token, nil)

	if !ok {
		w.WriteHeader(401)
		h.addon.Logger.Warn("Invalid JWT")
		return
	}

	if unverifiedClaims["iss"] == "" {
		w.WriteHeader(401)
		h.addon.Logger.Warn("JWT claim did not contain the issuer (iss) claim")
		return
	}

	clientKey := unverifiedClaims["iss"].(string)

	//TODO: aud-stuff -

	tenant, err := h.addon.Store.Get(clientKey)

	if err != nil {
		w.WriteHeader(500)
		h.addon.Logger.Warn("Could not lookup stored client data for clientKey")
		return
	}

	{
		secret := tenant.SharedSecret
		if secret == "" {
			w.WriteHeader(401)
			h.addon.Logger.Warn("Could not find JQT sharedSecret in tenant clientKey")
			return
		}

		verifiedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				w.WriteHeader(401)
				return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
			}

			return []byte(secret), nil
		})

		if err != nil {
			w.WriteHeader(500)
			h.addon.Logger.Warn("Could not verify JWT Token")
			return
		}

		err = verifiedToken.Claims.Valid()
		if err != nil {
			w.WriteHeader(500)
			h.addon.Logger.Warn("Could not find verify JWT Claims; Auth request has expired")
			return
		}

		claims, ok := verifiedToken.Claims.(jwt.MapClaims)
		if !ok {
			w.WriteHeader(500)
			h.addon.Logger.Warn("Could not cast Claims")
			return
		}
		//TODO: Replace true with skip QshVerification
		if true && claims["qsh"] != "" {

		}

	}

	h.h.ServeHTTP(w, r)
}

func NewAuthenticationMiddleware(addon *gonnect.Addon) func(h http.Handler) http.Handler {
	return func(handler http.Handler) http.Handler {
		return AuthenticationMiddleware{handler, addon}
	}
}
