package gonnect

import (
	"io"
	"log"
	"strings"
	"sync"
	"testing"

	"github.com/craftamap/atlas-gonnect/store"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/jinzhu/gorm"
)

func TestNewAddon(t *testing.T) {
	memoryGorm1, err := gorm.Open("sqlite3", ":memory:")
	if err != nil {
		t.Error(err)
	}
	nameL := "example"
	name := &nameL
	keyL := "com.github.craftamap.atlassian-gonnect.example"
	key := &keyL

	testCases := []struct {
		descriptorReader io.Reader
		configReader     io.Reader
		addon            *Addon
		expectError      bool
	}{
		// It should fail because no descriptor nor config
		{
			descriptorReader: strings.NewReader(""),
			configReader:     strings.NewReader(""),
			addon:            &Addon{},
			expectError:      true,
		},
		// It should fail because no descriptor nor config
		{
			descriptorReader: strings.NewReader("{}"),
			configReader:     strings.NewReader("{}"),
			addon:            &Addon{},
			expectError:      true,
		},
		// It should fail because no valid descriptor
		{
			descriptorReader: strings.NewReader("THIS IS SPARTA!"),
			configReader:     strings.NewReader(`{"currentProfile": "dev", "profiles": {"dev": {"baseUrl": "http://test/","port": 8080,"store": {"type": "sqlite3","databaseUrl": ":memory:"}}}}`),
			addon:            &Addon{},
			expectError:      true,
		},
		// It should fail because an invalid Descriptor is provided
		{
			descriptorReader: strings.NewReader("{}"),
			configReader:     strings.NewReader(`{"currentProfile": "dev", "profiles": {"dev": {"baseUrl": "http://test/","port": 8080,"store": {"type": "sqlite3","databaseUrl": ":memory:"}}}}`),
			addon:            &Addon{},
			expectError:      true,
		},
		// No key
		{
			descriptorReader: strings.NewReader(`{"name":"example"}`),
			configReader:     strings.NewReader(`{"currentProfile": "dev", "profiles": {"dev": {"baseUrl": "http://test/","port": 8080,"store": {"type": "sqlite3","databaseUrl": ":memory:"}}}}`),
			expectError:      true,
		},
		// there wil hopefully never be sqlite99
		{
			descriptorReader: strings.NewReader(`{"name":"example"}`),
			configReader:     strings.NewReader(`{"currentProfile": "dev", "profiles": {"dev": {"baseUrl": "http://test/","port": 8080,"store": {"type": "sqlite99","databaseUrl": ":memory:"}}}}`),
			expectError:      true,
		},
		// Successful test
		{
			descriptorReader: strings.NewReader(`{"name":"example", "key": "com.github.craftamap.atlassian-gonnect.example"}`),
			configReader:     strings.NewReader(`{"currentProfile": "dev", "profiles": {"dev": {"baseUrl": "http://test/","port": 8080,"store": {"type": "sqlite3","databaseUrl": ":memory:"}}}}`),
			addon: &Addon{
				CurrentProfile: "dev",
				Store:          &store.Store{Database: memoryGorm1},
				Config: &Profile{
					Port:    8080,
					BaseUrl: "http://test/",
					Store: StoreConfiguration{
						Type:        "sqlite3",
						DatabaseUrl: ":memory:",
					},
				},
				AddonDescriptor: map[string]interface{}{
					"name": "example",
					"key":  "com.github.craftamap.atlassian-gonnect.example",
				},
				Name: name,
				Key:  key,
			},
			expectError: false,
		},
	}

	for _, testCase := range testCases {
		addon, err := NewAddon(testCase.configReader, testCase.descriptorReader)
		if err != nil {
			if testCase.expectError {
				log.Printf("Expected error: %s", err)
			} else {
				t.Error(err)
			}
			continue
		} else if testCase.expectError {
			t.Error("Expected error, but got no error")
			continue
		}
		// The equal check of an addon will be funny to code...
		if testCase.addon == nil && addon != nil {
			t.Errorf("Expected error to be nil, but addon is %+v", addon)
		}
		if addon != nil {
			if addon.Config == nil || testCase.addon.Config == nil || !cmp.Equal(*addon.Config, *testCase.addon.Config) {
				t.Errorf("Expected addon.Config to be %+v, but got %+v", testCase.addon.Config, addon.Config)
			}

			if addon.Store == nil || testCase.addon.Store == nil || !cmp.Equal(*addon.Store, *testCase.addon.Store, cmpopts.IgnoreUnexported(gorm.DB{}, sync.RWMutex{})) {
				t.Errorf("Expected addon.Config to be %+v, but got %+v", testCase.addon.Store, addon.Store)
			}

			if addon.CurrentProfile != testCase.addon.CurrentProfile {
				t.Errorf("Expected addon.CurrentProfile to be %s, but got %s", testCase.addon.CurrentProfile, addon.CurrentProfile)
			}

			if !cmp.Equal(addon.AddonDescriptor, testCase.addon.AddonDescriptor) {
				t.Errorf("Expected addon.AddonDescriptor to be %+v, but got %+v", testCase.addon.AddonDescriptor, addon.AddonDescriptor)
			}
			if *addon.Key != *testCase.addon.Key || *addon.Name != *testCase.addon.Name {
				t.Errorf("Expected addon.Key to be %s, but got %s. Expected addon.Name to be %s, but got %s.",
					*testCase.addon.Key, *addon.Key, *testCase.addon.Name, *addon.Name)
			}
		}
	}

}

func TestReadAddonDescriptor(t *testing.T) {
	testCases := []struct {
		DescriptorReader io.Reader
		BaseUrl          string
		DescriptorMap    map[string]interface{}
		ExpectError      bool
	}{
		{
			DescriptorReader: strings.NewReader("THIS IS INVALID JSON"),
			BaseUrl:          "",
			DescriptorMap:    map[string]interface{}{},
			ExpectError:      true,
		},
		{
			DescriptorReader: strings.NewReader("{}"),
			BaseUrl:          "",
			DescriptorMap:    map[string]interface{}{},
			ExpectError:      false,
		},
		{
			DescriptorReader: strings.NewReader(`{"abc":"def"}`),
			BaseUrl:          "",
			DescriptorMap:    map[string]interface{}{"abc": "def"},
			ExpectError:      false,
		},
		{
			DescriptorReader: strings.NewReader(`{"abc":"{{BaseUrl}}"}`),
			BaseUrl:          "",
			DescriptorMap:    map[string]interface{}{"abc": "def"},
			ExpectError:      true,
		},
	}

	for _, testCase := range testCases {
		descriptorMap, err := readAddonDescriptor(testCase.DescriptorReader, testCase.BaseUrl)
		if err != nil {
			if !testCase.ExpectError {
				t.Error(err)
			}
			continue
		}

		if !cmp.Equal(descriptorMap, testCase.DescriptorMap) {
			t.Errorf("Expected Descriptor to be %s, but got %s", testCase.DescriptorMap, descriptorMap)
		}

	}

}
