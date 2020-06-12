package gonnect

import (
	"io"
	"strings"
	"testing"
)

func TestNewConfig(t *testing.T) {
	testCases := []struct {
		Config      io.Reader
		Profile     *Profile
		profilename string
		err         error
	}{
		{
			Config: strings.NewReader(`{"currentProfile": "dev", "profiles": {"dev": {"baseUrl": "http://test/","port": 8080,"store": {"type": "sqlite3",
                "databaseUrl": "db.sqlite"}}}}`),
			err: nil,
			Profile: &Profile{
				Port:    8080,
				BaseUrl: "http://test/",
				Store: StoreConfiguration{
					Type:        "sqlite3",
					DatabaseUrl: "db.sqlite",
				},
			},
			profilename: "dev",
		},
		{
			Config: strings.NewReader(`{"profiles": {"prod": {"baseUrl": "http://test/","port": 8080,"store": {"type": "sqlite3",
                "databaseUrl": "db.sqlite"}}}}`),
			err:         ErrConfigProfileNotFound,
			Profile:     nil,
			profilename: "",
		},
		{
			Config: strings.NewReader(`{"currentProfile":"dev","profiles": {"prod": {"baseUrl": "http://test/","port": 8080,"store": {"type": "sqlite3",
                "databaseUrl": "db.sqlite"}}}}`),
			err:         ErrConfigProfileNotFound,
			Profile:     nil,
			profilename: "",
		},
		{
			Config: strings.NewReader(`{"currentProfile":"","profiles": {"prod": {"baseUrl": "http://test/","port": 8080,"store": {"type": "sqlite3",
                "databaseUrl": "db.sqlite"}}}}`),
			err:         ErrConfigNoProfileSelected,
			Profile:     nil,
			profilename: "",
		},
	}

	for _, testCase := range testCases {
		profile, profilename, err := NewConfig(testCase.Config)
		if err != testCase.err {
			t.Errorf("Expected error to be %s, but got %s", testCase.err, err)
			return
		}

		if profilename != testCase.profilename {
			t.Errorf("Expected profilename to be %s, but got %s", testCase.profilename, profilename)
			return
		}

		if profile != nil && *profile != *testCase.Profile {
			t.Errorf("Expected profile to be %+v, but got %+v", testCase.Profile, profile)
			return
		}

	}
}
