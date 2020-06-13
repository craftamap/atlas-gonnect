package gonnect

import (
	"io"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

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
