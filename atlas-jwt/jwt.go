package atlasjwt

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/url"
	"sort"
	"strings"
)

const CANONICAL_QUERY_SEPARATOR = "&"

func canonicalizeUri(req *http.Request, baseUrlString string) string {
	path := req.URL.Path
	//TODO: Handle error here
	baseUrl, _ := url.Parse(baseUrlString)
	baseUrlPath := baseUrl.Path

	// If path and baseUrlPath have the same path, trim the path
	if strings.HasPrefix(path, baseUrlPath) {
		path = strings.TrimPrefix(path, baseUrlPath)
	}

	// if the path is an empty string now, this means the baseUrlPath AND the path were "/".
	// We therefore return "/"
	// Otherwise, the path is currently the path after the baseUrlPath
	if len(path) == 0 {
		return "/"
	}

	// If the separator is not URL encoded then the following URLs have the same query-string-hash:
	//   https://djtest9.jira-dev.com/rest/api/2/project&a=b?x=y
	//   https://djtest9.jira-dev.com/rest/api/2/project?a=b&x=y
	strings.ReplaceAll(path, CANONICAL_QUERY_SEPARATOR, url.QueryEscape(CANONICAL_QUERY_SEPARATOR))

	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	if len(path) > 1 && strings.HasSuffix(path, "/") {
		path = strings.TrimSuffix(path, "/")
	}

	return path
}

func canonicalizeQueryString(req *http.Request, checkBodyForParam bool) string {
	queryParams := req.URL.Query()

	if checkBodyForParam && len(queryParams) == 0 && (strings.ToUpper(req.Method) == "POST" || strings.ToUpper(req.Method) == "PUT") {
		// TODO: This could return errors...
		req.ParseForm()
		queryParams = req.PostForm
	}

	sortedQueryStrings := make([]string, 0)
	query := make([]string, 0)
	for key := range queryParams {
		if key != "jwt" {
			query = append(query, key)
		}
	}
	sort.Strings(query)
	for _, key := range query {
		if key == "__proto__" {
			continue
		}

		param := queryParams[key]
		sort.Strings(param)
		for idx, value := range param {
			param[idx] = strings.Replace(url.QueryEscape(value), "+", "%20", -1)
		}
		paramValue := strings.Join(param, ",")
		sortedQueryStrings = append(sortedQueryStrings, strings.Replace(url.QueryEscape(key), "+", "%20", -1)+"="+paramValue)
	}
	return strings.Join(sortedQueryStrings, "&")
}

func CreateQueryStringHash(req *http.Request, checkBodyForParam bool, baseUrlString string) string {
	createCanonicalRequest := func() string {
		return strings.ToUpper(req.Method) +
			CANONICAL_QUERY_SEPARATOR +
			canonicalizeUri(req, baseUrlString) +
			CANONICAL_QUERY_SEPARATOR +
			canonicalizeQueryString(req, checkBodyForParam)
	}

	s := createCanonicalRequest()
	h := sha256.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}
