package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"

	gonnect "github.com/craftamap/atlas-gonnect"
	"github.com/craftamap/atlas-gonnect/util"
	"github.com/dgrijalva/jwt-go"
)

const (
	CONNECT_INSTALL_KEYS_CDN_URL = "https://connect-install-keys.atlassian.com"
)

func isJwtAsymmetric(r *http.Request) bool {
	tokenStr, ok := ExtractJwt(r)
	if !ok {
		return ok
	}

	token, err := jwt.Parse(tokenStr, nil)
	if err != nil {
		return false
	}
	return token.Method == jwt.SigningMethodRS256
}

func fetchKeyWithKeyId(keyId string) (string, error) {
	keyCdnUrl, err := url.Parse(CONNECT_INSTALL_KEYS_CDN_URL)
	if err != nil {
		return "", err
	}

	keyCdnUrl.Path = path.Join(keyCdnUrl.Path, keyId)

	response, err := http.Get(keyCdnUrl.String())
	if err != nil {
		return "", err
	}
	// TODO: somehow return a 404 here
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", err
	}
	bodyString := string(body)

	return bodyString, nil
}

func decodeAsymmetric(tokenStr string, publicKey string, signedAlgorithm jwt.SigningMethod, noVerify bool) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenStr, nil)
	if err != nil {
		return nil, err
	}
	if token.Method.Alg() != signedAlgorithm.Alg() {
		return nil, fmt.Errorf("Unexpected signing method: %v", token.Method.Alg())
	}

	claims := token.Claims

	if !noVerify {
		token, err = jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			return []byte(publicKey), nil
		})
		if err != nil {
			return nil, err
		}
		claims = token.Claims
	}

	return claims.(jwt.MapClaims), nil
}

func decodeAsymmetricToken(tokenStr string, noVerify bool) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenStr, nil)
	if err != nil {
		return nil, err
	}

	keyIdI, ok := token.Header["kid"]
	if !ok {
		return nil, fmt.Errorf("keyId is missing")
	}
	keyId, ok := keyIdI.(string)
	if !ok || keyId == "" {
		return nil, fmt.Errorf("keyId is missing")
	}

	publicKey, err := fetchKeyWithKeyId(keyId)
	if err != nil {
		return nil, err
	}

	return decodeAsymmetric(tokenStr, publicKey, jwt.SigningMethodRS256, noVerify)
}

type signedInstallMiddleware struct {
	next  http.Handler
	addon *gonnect.Addon
}

func (h signedInstallMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	clientKey, err := h.verifyAsymmetricJwtAndGetClaims(r)
	if err != nil {
		util.SendError(w, h.addon, 401, err.Error())
		return
	}

	ctx := context.WithValue(r.Context(), "clientKey", clientKey)
	r = r.WithContext(ctx)

	h.ServeHTTP(w, r)
}

func (h signedInstallMiddleware) verifyAsymmetricJwtAndGetClaims(r *http.Request) (string, error) {
	tokenStr, ok := ExtractJwt(r)
	if !ok {
		return "", fmt.Errorf("Could not find authentication data on request")
	}

	unverifiedClaims, err := decodeAsymmetricToken(tokenStr, true)
	if err != nil {
		return "", err
	}

	if unverifiedClaims["iss"] == "" {
		return "", fmt.Errorf("JWT claim did not contain the issuer (iss) claim")
	}

	clientKey := unverifiedClaims["iss"].(string)

	if unverifiedClaims.VerifyAudience(h.addon.Config.BaseUrl, true) {
		return "", fmt.Errorf("JWT claim did not contain the correct audience (aud) claim")
	}

	if unverifiedClaims["qsh"] != "" {
		return "", fmt.Errorf("JWT claim did not contain the query string hash (qsh) claim")
	}

	verifiedClaims, err := decodeAsymmetricToken(tokenStr, false)
	if err != nil {
		return "", err
	}

	if err := verifiedClaims.Valid(); err != nil {
		return "", fmt.Errorf("Authentication request has expired.")
	}

	ok = ValidateQshFromRequest(verifiedClaims, r, h.addon, false)
	if !ok {
		return "", fmt.Errorf("Auth failure: Query hash mismatch")
	}

	return clientKey, nil
}

type VerifyInstallationMiddleware struct {
	next  http.Handler
	addon *gonnect.Addon
}

func (h VerifyInstallationMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Body == http.NoBody {
		util.SendError(w, h.addon, 401, "No registration info provided")
		return
	}

	b := bytes.NewBuffer(make([]byte, 0))
	reader := io.TeeReader(r.Body, b)
	defer r.Body.Close()

	responseData := map[string]interface{}{}
	json.NewDecoder(reader).Decode(&responseData)

	r.Body = ioutil.NopCloser(b)

	// TODO: Add whitelist feature

	baseUrl, ok := responseData["baseUrl"]
	if !ok {
		util.SendError(w, h.addon, 401, "No baseUrl provided for registration info")
		return
	}

	clientKey, ok := responseData["clientKey"]
	if !ok {
		h.addon.Logger.Warnf("No clientKey provided for host %s", baseUrl)
		return
	}

	if h.addon.Config.SignedInstall && isJwtAsymmetric(r) {
		signedInstallMiddleware{
			addon: h.addon,
			next: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// TODO: add check
				if r.Context().Value("clientKey") == clientKey {
					h.next.ServeHTTP(w, r)
				} else {
					util.SendError(w, h.addon, 401, "clientKey in install payload did not match authenticated client")
					return
				}
			}),
		}.ServeHTTP(w, r)
	} else {
		_, err := h.addon.Store.Get(clientKey.(string))
		if err != nil {
			// If err is set here, we serve the normal installation
			h.next.ServeHTTP(w, r)
		} else {
			authHandler := NewAuthenticationMiddleware(h.addon, false)
			authHandler(http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
				if req.Context().Value("clientKey") == clientKey {
					h.next.ServeHTTP(writer, req)
				} else {
					util.SendError(w, h.addon, 401, fmt.Sprintf("clientKey in install payload did not match authenticated client; payload: %s, auth: %s", clientKey, r.Context().Value("clientKey")))
					return
				}
			})).ServeHTTP(w, r)
		}
	}

}

func NewVerifyInstallationMiddleware(addon *gonnect.Addon) func(h http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return VerifyInstallationMiddleware{next, addon}
	}
}
