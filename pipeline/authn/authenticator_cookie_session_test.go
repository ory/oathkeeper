// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package authn_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/oathkeeper/internal"
	. "github.com/ory/oathkeeper/pipeline/authn"
	"github.com/ory/oathkeeper/x/header"
)

func TestAuthenticatorCookieSession(t *testing.T) {
	t.Parallel()
	conf := internal.NewConfigurationWithDefaults()
	reg := internal.NewRegistry(conf)
	session := new(AuthenticationSession)

	pipelineAuthenticator, err := reg.PipelineAuthenticator("cookie_session")
	require.NoError(t, err)

	t.Run("method=authenticate", func(t *testing.T) {
		t.Run("description=should fail because session store returned 400", func(t *testing.T) {
			testServer, _ := makeServer(t, 400, `{}`)
			err := pipelineAuthenticator.Authenticate(
				makeRequest("GET", "/", "", map[string]string{"sessionid": "zyx"}, ""),
				session,
				json.RawMessage(fmt.Sprintf(`{"check_session_url": "%s"}`, testServer.URL)),
				nil,
			)
			require.Error(t, err, "%#v", errors.Cause(err))
		})

		t.Run("description=should pass because session store returned 200", func(t *testing.T) {
			testServer, _ := makeServer(t, 200, `{"subject": "123", "extra": {"foo": "bar"}}`)
			err := pipelineAuthenticator.Authenticate(
				makeRequest("GET", "/", "", map[string]string{"sessionid": "zyx"}, ""),
				session,
				json.RawMessage(fmt.Sprintf(`{"check_session_url": "%s"}`, testServer.URL)),
				nil,
			)
			require.NoError(t, err, "%#v", errors.Cause(err))
			assert.Equal(t, &AuthenticationSession{
				Subject: "123",
				Extra:   map[string]interface{}{"foo": "bar"},
			}, session)
		})

		t.Run("description=should pass through method, path, and headers to auth server", func(t *testing.T) {

			testServer, requestRecorder := makeServer(t, 200, `{"subject": "123"}`)
			err := pipelineAuthenticator.Authenticate(
				makeRequest("PUT", "/users/123", "query=string", map[string]string{"sessionid": "zyx"}, ""),
				session,
				json.RawMessage(fmt.Sprintf(`{"check_session_url": "%s", "preserve_query": false}`, testServer.URL)),
				nil,
			)
			require.NoError(t, err, "%#v", errors.Cause(err))
			assert.Len(t, requestRecorder.requests, 1)
			r := requestRecorder.requests[0]
			assert.Equal(t, "PUT", r.Method)
			assert.Equal(t, "/users/123", r.URL.Path)
			assert.Equal(t, "query=string", r.URL.RawQuery)
			assert.Equal(t, "sessionid=zyx", r.Header.Get("Cookie"))
			assert.Equal(t, &AuthenticationSession{Subject: "123"}, session)
		})

		t.Run("description=should pass through method and headers ONLY to auth server when PreservePath and PreserveQuery are true", func(t *testing.T) {
			testServer, requestRecorder := makeServer(t, 200, `{"subject": "123"}`)
			err := pipelineAuthenticator.Authenticate(
				makeRequest("PUT", "/users/123", "query=string", map[string]string{"sessionid": "zyx"}, ""),
				session,
				json.RawMessage(fmt.Sprintf(`{"check_session_url": "%s", "preserve_path": true, "preserve_query": true}`, testServer.URL)),
				nil,
			)
			require.NoError(t, err, "%#v", errors.Cause(err))
			assert.Len(t, requestRecorder.requests, 1)
			r := requestRecorder.requests[0]
			assert.Equal(t, r.Method, "PUT")
			assert.Equal(t, r.URL.Path, "/")
			assert.Equal(t, r.URL.RawQuery, "")
			assert.Equal(t, r.Header.Get("Cookie"), "sessionid=zyx")
			assert.Equal(t, &AuthenticationSession{Subject: "123"}, session)
		})

		t.Run("should preserve path, query in check_session_url when preserve_path, preserve_query are true", func(t *testing.T) {
			testServer, requestRecorder := makeServer(t, 200, `{"subject": "123"}`)
			err := pipelineAuthenticator.Authenticate(
				makeRequest("PUT", "/client/request/path", "q=client-request-query", map[string]string{"sessionid": "zyx"}, ""),
				session,
				json.RawMessage(fmt.Sprintf(`{"check_session_url": "%s/configured/path?q=configured-query", "preserve_path": true, "preserve_query": true}`, testServer.URL)),
				nil,
			)
			require.NoError(t, err, "%#v", errors.Cause(err))
			assert.Len(t, requestRecorder.requests, 1)
			r := requestRecorder.requests[0]
			assert.Equal(t, r.Method, "PUT")
			assert.Equal(t, r.URL.Path, "/configured/path")
			assert.Equal(t, r.URL.RawQuery, "q=configured-query")
			assert.Equal(t, r.Header.Get("Cookie"), "sessionid=zyx")
			assert.Equal(t, &AuthenticationSession{Subject: "123"}, session)
		})

		t.Run("should override method", func(t *testing.T) {
			testServer, requestRecorder := makeServer(t, 200, `{"subject": "123"}`)
			err := pipelineAuthenticator.Authenticate(
				makeRequest("PUT", "/client/request/path", "q=client-request-query", map[string]string{"sessionid": "zyx"}, ""),
				session,
				json.RawMessage(fmt.Sprintf(`{"check_session_url": "%s/configured/path?q=configured-query", "force_method": "GET"}`, testServer.URL)),
				nil,
			)
			require.NoError(t, err, "%#v", errors.Cause(err))
			assert.Len(t, requestRecorder.requests, 1)
			r := requestRecorder.requests[0]
			assert.Equal(t, r.Method, "GET")
		})

		t.Run("description=should pass through x-forwarded-host if preserve_host is set to true", func(t *testing.T) {
			testServer, requestRecorder := makeServer(t, 200, `{"subject": "123"}`)
			req := makeRequest("PUT", "/users/123", "query=string", map[string]string{"sessionid": "zyx"}, "")
			expectedHost := "some-host"
			req.Host = expectedHost
			err := pipelineAuthenticator.Authenticate(
				req,
				session,
				json.RawMessage(fmt.Sprintf(`{"check_session_url": "%s", "preserve_host": true}`, testServer.URL)),
				nil,
			)
			require.NoError(t, err, "%#v", errors.Cause(err))
			assert.Len(t, requestRecorder.requests, 1)
			r := requestRecorder.requests[0]
			assert.Equal(t, r.Method, "PUT")
			assert.Equal(t, expectedHost, r.Header.Get("X-Forwarded-Host"))
			assert.Empty(t, req.Header.Get("X-Forwarded-Host"), "The original header must NOT be modified")
			assert.Equal(t, r.Header.Get("Cookie"), "sessionid=zyx")
			assert.Equal(t, &AuthenticationSession{Subject: "123"}, session)
		})

		t.Run("description=should pass not override x-forwarded-host in set_headers", func(t *testing.T) {
			testServer, requestRecorder := makeServer(t, 200, `{"subject": "123"}`)
			req := makeRequest("PUT", "/users/123", "query=string", map[string]string{"sessionid": "zyx"}, "")
			expectedHost := "some-host"
			req.Host = expectedHost
			err := pipelineAuthenticator.Authenticate(
				req,
				session,
				json.RawMessage(fmt.Sprintf(`{"check_session_url": "%s", "additional_headers": {"X-Forwarded-Host": "not-some-host", "X-Foo": "bar"}, "preserve_host": true}`, testServer.URL)),
				nil,
			)
			require.NoError(t, err, "%#v", errors.Cause(err))
			assert.Len(t, requestRecorder.requests, 1)
			r := requestRecorder.requests[0]
			assert.Equal(t, r.Method, "PUT")
			assert.Equal(t, expectedHost, r.Header.Get("X-Forwarded-Host"))
			assert.Equal(t, "bar", r.Header.Get("X-Foo"))
			assert.Empty(t, req.Header.Get("X-Forwarded-Host"), "The original header must NOT be modified")
			assert.Empty(t, req.Header.Get("X-Foo"), "The original header must NOT be modified")
			assert.Equal(t, r.Header.Get("Cookie"), "sessionid=zyx")
			assert.Equal(t, &AuthenticationSession{Subject: "123"}, session)
		})

		t.Run("description=does not pass request body through to auth server", func(t *testing.T) {
			testServer, requestRecorder := makeServer(t, 200, `{}`)
			pipelineAuthenticator.Authenticate(
				makeRequest("POST", "/", "", map[string]string{"sessionid": "zyx"}, "Some body..."),
				session,
				json.RawMessage(fmt.Sprintf(`{"check_session_url": "%s"}`, testServer.URL)),
				nil,
			)
			assert.Len(t, requestRecorder.requests, 1)
			assert.Len(t, requestRecorder.bodies, 1)
			r := requestRecorder.requests[0]
			assert.Equal(t, r.ContentLength, int64(0))
			assert.Equal(t, requestRecorder.bodies[0], []byte{})
		})

		t.Run("description=should fallthrough if only is specified and no cookie specified is set", func(t *testing.T) {
			testServer, requestRecorder := makeServer(t, 200, `{}`)
			err := pipelineAuthenticator.Authenticate(
				makeRequest("GET", "/", "", map[string]string{"sessionid": "zyx"}, ""),
				session,
				json.RawMessage(fmt.Sprintf(`{"only": ["session", "sid"], "check_session_url": "%s"}`, testServer.URL)),
				nil,
			)
			assert.Equal(t, errors.Cause(err), ErrAuthenticatorNotResponsible)
			assert.Empty(t, requestRecorder.requests)
		})

		t.Run("description=should fallthrough if is missing and it has no cookies", func(t *testing.T) {
			testServer, requestRecorder := makeServer(t, 200, `{}`)
			err := pipelineAuthenticator.Authenticate(
				makeRequest("GET", "/", "", map[string]string{}, ""),
				session,
				json.RawMessage(fmt.Sprintf(`{"check_session_url": "%s"}`, testServer.URL)),
				nil,
			)
			assert.Equal(t, errors.Cause(err), ErrAuthenticatorNotResponsible)
			assert.Empty(t, requestRecorder.requests)
		})

		t.Run("description=should not fallthrough if only is specified and cookie specified is set", func(t *testing.T) {
			testServer, _ := makeServer(t, 200, `{}`)
			err := pipelineAuthenticator.Authenticate(
				makeRequest("GET", "/", "", map[string]string{"sid": "zyx"}, ""),
				session,
				json.RawMessage(fmt.Sprintf(`{"only": ["session", "sid"], "check_session_url": "%s"}`, testServer.URL)),
				nil,
			)
			require.NoError(t, err, "%#v", errors.Cause(err))
		})

		t.Run("description=should work with nested extra keys", func(t *testing.T) {
			testServer, _ := makeServer(t, 200, `{"subject": "123", "session": {"foo": "bar"}}`)
			err := pipelineAuthenticator.Authenticate(
				makeRequest("GET", "/", "", map[string]string{"sessionid": "zyx"}, ""),
				session,
				json.RawMessage(fmt.Sprintf(`{"check_session_url": "%s", "extra_from": "session"}`, testServer.URL)),
				nil,
			)
			require.NoError(t, err, "%#v", errors.Cause(err))
			assert.Equal(t, &AuthenticationSession{
				Subject: "123",
				Extra:   map[string]interface{}{"foo": "bar"},
			}, session)
		})

		t.Run("description=should work with the root key for extra and a custom subject key", func(t *testing.T) {
			testServer, _ := makeServer(t, 200, `{"identity": {"id": "123"}, "session": {"foo": "bar"}}`)
			err := pipelineAuthenticator.Authenticate(
				makeRequest("GET", "/", "", map[string]string{"sessionid": "zyx"}, ""),
				session,
				json.RawMessage(fmt.Sprintf(`{"check_session_url": "%s", "subject_from": "identity.id", "extra_from": "@this"}`, testServer.URL)),
				nil,
			)
			require.NoError(t, err, "%#v", errors.Cause(err))
			assert.Equal(t, &AuthenticationSession{
				Subject: "123",
				Extra:   map[string]interface{}{"session": map[string]interface{}{"foo": "bar"}, "identity": map[string]interface{}{"id": "123"}},
			}, session)
		})
		t.Run("description=should work with custom header forwarded", func(t *testing.T) {
			requestRecorder := &RequestRecorder{}
			testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requestRecorder.requests = append(requestRecorder.requests, r)
				requestBody, _ := io.ReadAll(r.Body)
				requestRecorder.bodies = append(requestRecorder.bodies, requestBody)
				if r.Header.Get("X-User") == "" {
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"identity": {"id": "123"}, "session": {"foo": "bar"}}`))
			}))
			req := makeRequest("GET", "/", "", map[string]string{"sessionid": "zyx"}, "")
			req.Header.Add("X-UsEr", "123")
			err := pipelineAuthenticator.Authenticate(
				req,
				session,
				json.RawMessage(fmt.Sprintf(`{"check_session_url": "%s", "subject_from": "identity.id", "extra_from": "@this", "forward_http_headers": ["X-User"]}`, testServer.URL)),
				nil,
			)
			require.NoError(t, err, "%#v", errors.Cause(err))
			assert.Equal(t, &AuthenticationSession{
				Subject: "123",
				Extra:   map[string]interface{}{"session": map[string]interface{}{"foo": "bar"}, "identity": map[string]interface{}{"id": "123"}},
			}, session)
		})
	})
}

func TestPrepareRequest(t *testing.T) {
	t.Run("prepare request should return only configured headers", func(t *testing.T) {
		testCases := []struct {
			requestHeaders  []string
			expectedHeaders []string
			conf            *AuthenticatorCookieSessionConfiguration
		}{
			{
				requestHeaders:  []string{header.Authorization, header.AcceptEncoding},
				expectedHeaders: []string{},
				conf:            &AuthenticatorCookieSessionConfiguration{},
			},
			{
				requestHeaders:  []string{header.Authorization, header.AcceptEncoding},
				expectedHeaders: []string{header.AcceptEncoding},
				conf: &AuthenticatorCookieSessionConfiguration{
					// This value is coming from the configuration and may use incorrect casing.
					ForwardHTTPHeaders: []string{
						"acCept-enCodinG",
					},
				},
			},
			{
				requestHeaders:  []string{header.Authorization, header.AcceptEncoding},
				expectedHeaders: []string{header.Authorization},
				conf: &AuthenticatorCookieSessionConfiguration{
					ForwardHTTPHeaders: []string{
						header.Authorization,
					},
				},
			},
		}

		for _, testCase := range testCases {
			r := makeRequest("GET", "/", "", map[string]string{"sessionID": "zyx"}, "")
			for _, h := range testCase.requestHeaders {
				r.Header.Add(h, h)
			}
			expected := http.Header{}
			for _, h := range testCase.expectedHeaders {
				expected.Add(h, h)
			}
			req, err := PrepareRequest(r, testCase.conf)
			assert.NoError(t, err)
			assert.Equal(t, expected, req.Header)
		}
	})

}

type RequestRecorder struct {
	requests []*http.Request
	bodies   [][]byte
}

func makeServer(t *testing.T, statusCode int, responseBody string) (*httptest.Server, *RequestRecorder) {
	requestRecorder := &RequestRecorder{}
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestRecorder.requests = append(requestRecorder.requests, r)
		requestBody, _ := io.ReadAll(r.Body)
		requestRecorder.bodies = append(requestRecorder.bodies, requestBody)
		w.WriteHeader(statusCode)
		w.Write([]byte(responseBody))
	}))
	t.Cleanup(testServer.Close)
	return testServer, requestRecorder
}

func makeRequest(method string, path string, rawQuery string, cookies map[string]string, bodyStr string) *http.Request {
	var body io.ReadCloser
	header := http.Header{}
	if bodyStr != "" {
		body = io.NopCloser(bytes.NewBufferString(bodyStr))
		header.Add("Content-Length", strconv.Itoa(len(bodyStr)))
	}
	req := &http.Request{
		Method: method,
		URL:    &url.URL{Path: path, RawQuery: rawQuery},
		Header: header,
		Body:   body,
	}
	for name, value := range cookies {
		req.AddCookie(&http.Cookie{Name: name, Value: value})
	}
	return req
}
