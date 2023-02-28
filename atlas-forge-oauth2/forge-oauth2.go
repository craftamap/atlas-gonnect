package atlasforgeoauth2

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/craftamap/atlas-gonnect/store"
	"github.com/patrickmn/go-cache"
)

const TOKEN_ENDPOINT_PRODUCTION = "https://auth.atlassian.com/oauth/token"
const IDENTITY_AUDIENCE_PRODUCTION = "api.atlassian.com"

type BearerTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	Scope       string `json:"scope"`
	ExpiresIn   int    `json:"expires_in"`
}

var bearerTokenCache = cache.New(1*time.Hour, 10*time.Minute)

func FetchBearerToken(tenant *store.Tenant) (BearerTokenResponse, error) {
	// TODO: implement null-checks here

	payload := map[string]string{
		"grant_type":    "client_credentials",
		"client_id":     tenant.OauthClientId,
		"client_secret": tenant.SharedSecret,
		"audience":      IDENTITY_AUDIENCE_PRODUCTION,
	}

	var body bytes.Buffer
	json.NewEncoder(&body).Encode(payload)

	res, err := http.Post(TOKEN_ENDPOINT_PRODUCTION, "application/json", &body)
	if err != nil {
		return BearerTokenResponse{}, fmt.Errorf("failed to get bearer token for tenant %+v, %w", tenant.ClientKey, err)
	}
	defer res.Body.Close()
	var response BearerTokenResponse

	err = json.NewDecoder(res.Body).Decode(&response)
	if err != nil {
		return BearerTokenResponse{}, fmt.Errorf("failed to get bearer token out of response, %w", err)
	}

	return response, nil
}

func GetBearerToken(tenant *store.Tenant) (string, error) {
	// TODO: Figure out a reasonable cache key here.
	//       For some reason, atlassian is caching using a hash of the OauthClientId. For now, let's do the same.
	hashBytes := sha256.Sum256([]byte(tenant.OauthClientId))
	key := fmt.Sprintf("%x", hashBytes)

	potentialTokenResponse, ok := bearerTokenCache.Get(key)
	if ok {
		return potentialTokenResponse.(BearerTokenResponse).AccessToken, nil
	}

	tokenResponse, err := FetchBearerToken(tenant)
	if err != nil {
		return "", fmt.Errorf("failed to fetch bearer token in GetBearerToken, %w", err)
	}

	// Use 99% of expires_in to account for timing inaccuracies. Assuming 3600 seconds, we would lose 36 seconds
	bearerTokenCache.Set(key, tokenResponse, time.Duration(float64(tokenResponse.ExpiresIn)*0.99)*time.Second)

	return tokenResponse.AccessToken, nil
}
