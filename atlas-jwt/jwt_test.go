package atlasjwt

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestCreateQueryStringHash(t *testing.T) {
	cases := []struct {
		method            string
		url               string
		body              io.Reader
		checkForBodyParam bool
		baseUrl           string
		expectedHash      string
	}{
		// it should correctly create qsh without query string
		{method: "GET", url: "/path", body: http.NoBody, checkForBodyParam: false, baseUrl: "", expectedHash: "799be84a7fa35570087163c0cd9af3abff7ac05c2c12ba0bb1d7eebc984b3ac2"},
		// a missing prefix slash should not change the hash
		{method: "GET", url: "path", body: http.NoBody, checkForBodyParam: false, baseUrl: "", expectedHash: "799be84a7fa35570087163c0cd9af3abff7ac05c2c12ba0bb1d7eebc984b3ac2"},
		// an an trailing slash also should not change the hash
		{method: "GET", url: "path/", body: http.NoBody, checkForBodyParam: false, baseUrl: "", expectedHash: "799be84a7fa35570087163c0cd9af3abff7ac05c2c12ba0bb1d7eebc984b3ac2"},
		// it should correctly create qsh without path or query string
		{method: "GET", url: "", body: http.NoBody, checkForBodyParam: false, baseUrl: "", expectedHash: "c88caad15a1c1a900b8ac08aa9686f4e8184539bea1deda36e2f649430df3239"},
		// it should correctly create qsh without path or query string
		{method: "GET", url: "/", body: http.NoBody, checkForBodyParam: false, baseUrl: "", expectedHash: "c88caad15a1c1a900b8ac08aa9686f4e8184539bea1deda36e2f649430df3239"},
		// it should correctly create qsh with query string
		{method: "GET", url: "/hello-world?lic=none&tz=Australia%2FSydney&cp=%2Fjira&user_key=&loc=en-US&user_id=&jwt=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjEzODY5MTEzNTYsImlzcyI6ImppcmE6MTU0ODk1OTUiLCJxc2giOiI4MDYzZmY0Y2ExZTQxZGY3YmM5MGM4YWI2ZDBmNjIwN2Q0OTFjZjZkYWQ3YzY2ZWE3OTdiNDYxNGI3MTkyMmU5IiwiaWF0IjoxMzg2OTExMTc2fQ.rAsxpHv0EvpXkhjnZnSV14EXJgDx3KSQjgYRjfKnFt8&xdm_e=http%3A%2F%2Fstorm%3A2990&xdm_c=channel-servlet-hello-world&xdm_p=1", body: http.NoBody, checkForBodyParam: false, baseUrl: "", expectedHash: "8063ff4ca1e41df7bc90c8ab6d0f6207d491cf6dad7c66ea797b4614b71922e9"},
		// it should correctly create qsh with POST body query string
		{method: "POST", url: "/hello-world", body: strings.NewReader("lic=none&tz=Australia%2FSydney&cp=%2Fjira&user_key=&loc=en-US&user_id=&jwt=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjEzODY5MTEzNTYsImlzcyI6ImppcmE6MTU0ODk1OTUiLCJxc2giOiI4MDYzZmY0Y2ExZTQxZGY3YmM5MGM4YWI2ZDBmNjIwN2Q0OTFjZjZkYWQ3YzY2ZWE3OTdiNDYxNGI3MTkyMmU5IiwiaWF0IjoxMzg2OTExMTc2fQ.rAsxpHv0EvpXkhjnZnSV14EXJgDx3KSQjgYRjfKnFt8&xdm_e=http%3A%2F%2Fstorm%3A2990&xdm_c=channel-servlet-hello-world&xdm_p=1"), checkForBodyParam: true, baseUrl: "", expectedHash: "d7e7f00660965fc15745b2c423a89b85d0853c4463faca362e0371d008eb0927"},
		{method: "POST", url: "/hello-world", body: strings.NewReader("lic=none&tz=Australia%2FSydney&cp=%2Fjira&user_key=&loc=en-US&user_id=&jwt=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjEzODY5MTEzNTYsImlzcyI6ImppcmE6MTU0ODk1OTUiLCJxc2giOiI4MDYzZmY0Y2ExZTQxZGY3YmM5MGM4YWI2ZDBmNjIwN2Q0OTFjZjZkYWQ3YzY2ZWE3OTdiNDYxNGI3MTkyMmU5IiwiaWF0IjoxMzg2OTExMTc2fQ.rAsxpHv0EvpXkhjnZnSV14EXJgDx3KSQjgYRjfKnFt8&xdm_e=http%3A%2F%2Fstorm%3A2990&xdm_c=channel-servlet-hello-world&xdm_p=1"), checkForBodyParam: false, baseUrl: "", expectedHash: "6f95f3738e1b037a3bebbe0ad237d80fdbc1d5ae452e98ce03a9c004c178ebb4"},
	}

	for _, tCase := range cases {
		req, err := http.NewRequest(tCase.method, tCase.url, tCase.body)
		if err != nil {
			t.Error(err)
		}
		qsh := CreateQueryStringHash(req, tCase.checkForBodyParam, tCase.baseUrl)
		if tCase.expectedHash != qsh {
			t.Errorf("did not match expected value: expected %s, got %s ;\n  Testcase: %+v", tCase.expectedHash, qsh, tCase)
		}
	}

}
