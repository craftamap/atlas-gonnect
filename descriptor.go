package gonnect

import (
	"errors"
	"strings"

	"github.com/xeipuuv/gojsonschema"
)

const jiraschema = "https://bitbucket.org/atlassian/connect-schemas/raw/master/jira-global-schema.json"
const confschema = "https://bitbucket.org/atlassian/connect-schemas/raw/master/confluence-global-schema.json"

func ValidateAppDescriptor(product string, appDescriptor string) (bool, error, []gojsonschema.ResultError) {
	var schemaLoader gojsonschema.JSONLoader
	if strings.ToLower(product) == "jira" {
		schemaLoader = gojsonschema.NewReferenceLoader(jiraschema)
	} else if strings.ToLower(product) == "confluence" {
		schemaLoader = gojsonschema.NewReferenceLoader(confschema)
	} else {
		return false, errors.New("Product not supported"), []gojsonschema.ResultError{}
	}
	documentLoader := gojsonschema.NewStringLoader(appDescriptor)

	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		return false, err, []gojsonschema.ResultError{}
	}

	if result.Valid() {
		return true, nil, []gojsonschema.ResultError{}
	} else {
		return false, nil, result.Errors()
	}

}
