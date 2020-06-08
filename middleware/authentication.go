package middleware

import (
	"log"
	"net/http"
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
		w.Write([]byte("Could not find auth data on request"))
		return
	}

	unverifiedClaims, ok := extractUnverifiedClaims(token, nil)

	if !ok {
		w.WriteHeader(401)
		w.Write([]byte("Invalid JWT"))
		return
	}

	if (unverifiedClaims["iss"] == "") {
		w.WriteHeader(401)
		w.Write([]byte("JWT claim did not contain the issuer (iss) claim"))
		return	
	}

	//clientKey := unverifiedClaims["iss"]
	
	//TODO: aud-stuff

	

	h.h.ServeHTTP(w, r)
}

func NewAuthenticationMiddleware(addon *gonnect.Addon) func(h http.Handler) http.Handler {
	return func(handler http.Handler) http.Handler {
		return AuthenticationMiddleware{handler, addon}
	}
}
