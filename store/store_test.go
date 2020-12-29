package store

import (
	"testing"
)

func TestNew(t *testing.T) {
	// ToDo: better tests
	testCases := []struct {
		DBType      string
		DatabaseUrl string
		expectError bool
	}{
		{
			DBType:      "sqlite3",
			DatabaseUrl: ":memory:",
			expectError: false,
		},
		{
			DBType:      "postgres",
			DatabaseUrl: "postgres://postgres:root@localhost/gonnect?sslmode=disable",
			expectError: false,
		},
	}

	for _, testCase := range testCases {
		_, err := New(testCase.DBType, testCase.DatabaseUrl)
		if err != nil {
			t.Error(err)
		}
	}
}

func TestSet(t *testing.T) {
	testDatabases := []struct {
		DBType      string
		DatabaseUrl string
	}{
		{
			DBType:      "sqlite3",
			DatabaseUrl: ":memory:",
		},
		{
			DBType:      "postgres",
			DatabaseUrl: "postgres://postgres:root@localhost/gonnect?sslmode=disable",
		},
	}
	testCases := []struct {
		Tenant *Tenant
	}{
		{
			Tenant: &Tenant{ClientKey: "unique-client-identifier",
				SharedSecret:   "a-secret-key-not-to-be-lost",
				BaseURL:        "https://example.atlassian.net",
				ProductType:    "jira",
				Description:    "AtlassianJiraathttps://example.atlassian.net",
				AddonInstalled: true,
				EventType:      "installed",
			},
		},
		{
			Tenant: &Tenant{ClientKey: "unique-client-identifier",
				BaseURL:        "https://example.atlassian.net",
				ProductType:    "jira",
				Description:    "AtlassianJiraathttps://example.atlassian.net",
				AddonInstalled: false,
				EventType:      "uninstalled",
			},
		},
	}

	for _, testCase := range testCases {
		for _, testDatabase := range testDatabases {
			store, err := New(testDatabase.DBType, testDatabase.DatabaseUrl)
			if err != nil {
				t.Error(err)
			}
			_, err = store.Set(testCase.Tenant)
			if err != nil {
				t.Error(err)
			}
		}
	}
}
