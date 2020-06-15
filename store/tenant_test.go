package store

import (
	"io"
	"log"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestNewTenantFromReader(t *testing.T) {
	testCases := []struct {
		Reader      io.Reader
		Tenant      *Tenant
		expectError bool
	}{
		{
			Reader: strings.NewReader(`{"key":"installed-addon-key","clientKey":"unique-client-identifier","sharedSecret":"a-secret-key-not-to-be-lost","serverVersion":"server-version","pluginsVersion":"version-of-connect","baseUrl":"https://example.atlassian.net","displayUrl":"https://docs.example.com","productType":"jira","description":"AtlassianJiraathttps://example.atlassian.net","serviceEntitlementNumber":"SEN-number","eventType":"installed"}`),
			Tenant: &Tenant{ClientKey: "unique-client-identifier",
				SharedSecret:   "a-secret-key-not-to-be-lost",
				BaseURL:        "https://example.atlassian.net",
				ProductType:    "jira",
				Description:    "AtlassianJiraathttps://example.atlassian.net",
				AddonInstalled: true,
				EventType:      "installed",
			},
			expectError: false,
		},
		{
			Reader: strings.NewReader(`{"key":"installed-addon-key","clientKey":"unique-client-identifier","serverVersion":"server-version","pluginsVersion":"version-of-connect","baseUrl":"https://example.atlassian.net","displayUrl":"https://docs.example.com","productType":"jira","description":"AtlassianJiraathttps://example.atlassian.net","serviceEntitlementNumber":"SEN-number","eventType":"uninstalled"}`),
			Tenant: &Tenant{ClientKey: "unique-client-identifier",
				BaseURL:        "https://example.atlassian.net",
				ProductType:    "jira",
				Description:    "AtlassianJiraathttps://example.atlassian.net",
				AddonInstalled: false,
				EventType:      "uninstalled",
			},
			expectError: false,
		},
		{
			Reader:      strings.NewReader(`INVALID JSON`),
			Tenant:      &Tenant{},
			expectError: true,
		},
	}

	for _, testCase := range testCases {
		tenant, err := NewTenantFromReader(testCase.Reader)
		if err != nil {
			if testCase.expectError {
				log.Printf("Expected error: %s\n", err)
			} else {
				t.Error(err)
			}
			return
		} else if testCase.expectError {
			t.Error("Expected error, but no error occured")
			return
		}

		if !cmp.Equal(*tenant, *testCase.Tenant) {
			t.Errorf("Expected tenant to be %+v, but got %+v\f", testCase.Tenant, tenant)
		}
	}
}
