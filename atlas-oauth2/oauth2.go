package atlasoauth2

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/craftamap/atlas-gonnect/store"
        "github.com/golang-jwt/jwt"
)

const JWT_CLAIM_PREFIX = "urn:atlassian:connect"
const AUTHORIZATION_SERVER_URL = "https://oauth-2-authorization-server.services.atlassian.com"
const GRANT_TYPE = "urn:ietf:params:oauth:grant-type:jwt-bearer"

func createTokenForAccountId(tenant *store.Tenant, accountId string) (string, error) {
	subject := JWT_CLAIM_PREFIX + ":useraccountid:" + accountId
	issuer := JWT_CLAIM_PREFIX + ":clientid:" + tenant.OauthClientId

	claims := struct {
		Tenant string `json:"tnt"`
		jwt.StandardClaims
	}{
		Tenant: tenant.BaseURL,
		StandardClaims: jwt.StandardClaims{
			Issuer:    issuer,
			Subject:   subject,
			Audience:  AUTHORIZATION_SERVER_URL,
			ExpiresAt: time.Now().Add(59 * time.Second).Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(tenant.SharedSecret))

	if err != nil {
		return "", err
	}

	return signedToken, nil
}

func GetAccessToken(tenant *store.Tenant, userAccountId string, scopes []string) (string, error) {
	// TODO: We should probably use an oauth2 library - for now though, lets keep it "simple"
	// TODO: Add Caching

	jwtToken, err := createTokenForAccountId(tenant, userAccountId)
	if err != nil {
		return "", err
	}

	reader := strings.NewReader(strings.ReplaceAll(url.Values(map[string][]string{
		"grant_type": {GRANT_TYPE},
		"assertion":  {jwtToken},
		"scope":      {strings.ToUpper(strings.Join(scopes, " "))},
	}).Encode(), "+", "%20"))

	log.Print(jwtToken)

	req, err := http.NewRequest("POST", AUTHORIZATION_SERVER_URL+"/oauth2/token", reader)
	if err != nil {
		return "", err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}

	if res.StatusCode < 200 || res.StatusCode > 299 {
		return "", errors.New(res.Status)
	}

	responseBody := make(map[string]interface{})

	err = json.NewDecoder(res.Body).Decode(&responseBody)
	if err != nil {
		return "", err
	}

	log.Printf("%+v\n", responseBody)

	if val, ok := responseBody["token_type"]; !ok || val != "Bearer" {
		return "", errors.New("response body did not contain a bearer token")
	}

	return responseBody["access_token"].(string), nil
}
