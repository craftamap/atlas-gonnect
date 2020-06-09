package middleware

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	gonnect "github.com/craftamap/atlas-gonnect"
	atlasjwt "github.com/craftamap/atlas-gonnect/atlas-jwt"
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
	//TODO: Refactor to be more compact
	//TODO: scoping

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

	if unverifiedClaims["aud"] != nil && unverifiedClaims["aud"] != "" {
		clientKey = unverifiedClaims["aud"].(string)
	}

	tenant, err := h.addon.Store.Get(clientKey)

	if err != nil {
		w.WriteHeader(500)
		h.addon.Logger.Warn("Could not lookup stored client data for clientKey")
		return
	}

	secret := tenant.SharedSecret
	if secret == "" {
		w.WriteHeader(401)
		h.addon.Logger.Warn("Could not find JQT sharedSecret in tenant clientKey")
		return
	}

	verifiedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		// TODO: We should not check the token header for the method instead of using HMAC by default
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
		expectedHash := atlasjwt.CreateQueryStringHash(r, false, h.addon.Config.BaseUrl)
		if claims["qsh"] != expectedHash {
			// If that didn't verify, it might be a  post/put - check the request body too
			expectedHash := atlasjwt.CreateQueryStringHash(r, false, h.addon.Config.BaseUrl)
			if claims["qsh"] != expectedHash {
				h.addon.Logger.Errorf("Auth failure: Query hash mismatch: Received %s but calculated %s", claims["qsh"], expectedHash)
			}
		}
	}

	h.addon.Logger.Info("Auth successful")

	createSessionToken := func() (string, error) {
		verClaims := verifiedToken.Claims.(jwt.MapClaims)

		claims := &jwt.StandardClaims{
			Issuer: *h.addon.Key,
			// TODO: Check if subject can be asserted
			Audience: clientKey,
		}
		subject, ok := verClaims["subject"].(string)
		if ok {
			claims.Subject = subject
		}

		// TODO: We may have to add the context workaround, but lets ignore it for now

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		signedToken, err := token.SignedString([]byte(tenant.SharedSecret))
		if err != nil {
			return "", err
		}

		w.Header().Set("X-acpt", signedToken) // TODO: Do we really need to do this?

		return signedToken, nil
	}

	tokenString, err := createSessionToken()
	if err != nil {
		//TODO: Do we really want to fail here?
		w.WriteHeader(500)
		h.addon.Logger.Warnf("Could not create new access token %s", err)
		panic(err)
	}

	oldVerClaims := verifiedToken.Claims.(jwt.MapClaims)

	verifiedParams := map[string]string{
		"clientKey":   clientKey,
		"hostBaseUrl": h.addon.Config.BaseUrl,
		"token":       tokenString,
		// TODO: We may have to add the context workaround instead of just using sub as userAccountId, but lets ignore it for now
		"userAccountId": oldVerClaims["sub"].(string),
	}

	requestHandler := NewRequestMiddleware(h.addon, verifiedParams)

	requestHandler(h.h).ServeHTTP(w, r)
}

func NewAuthenticationMiddleware(addon *gonnect.Addon) func(h http.Handler) http.Handler {
	return func(handler http.Handler) http.Handler {
		return AuthenticationMiddleware{handler, addon}
	}
}
