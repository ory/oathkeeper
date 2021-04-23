package authn_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
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
)

func TestAuthenticatorCookieSession(t *testing.T) {
	conf := internal.NewConfigurationWithDefaults()
	reg := internal.NewRegistry(conf)
	session := new(AuthenticationSession)

	pipelineAuthenticator, err := reg.PipelineAuthenticator("cookie_session")
	require.NoError(t, err)

	t.Run("method=authenticate", func(t *testing.T) {
		t.Run("description=should fail because session store returned 400", func(t *testing.T) {
			testServer, _ := makeServer(400, `{}`)
			err := pipelineAuthenticator.Authenticate(
				makeRequest("GET", "/", map[string]string{"sessionid": "zyx"}, ""),
				session,
				json.RawMessage(fmt.Sprintf(`{"check_session_url": "%s"}`, testServer.URL)),
				nil,
			)
			require.Error(t, err, "%#v", errors.Cause(err))
		})

		t.Run("description=should pass because session store returned 200", func(t *testing.T) {
			testServer, _ := makeServer(200, `{"subject": "123", "extra": {"foo": "bar"}}`)
			err := pipelineAuthenticator.Authenticate(
				makeRequest("GET", "/", map[string]string{"sessionid": "zyx"}, ""),
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
			testServer, requestRecorder := makeServer(200, `{"subject": "123"}`)
			err := pipelineAuthenticator.Authenticate(
				makeRequest("PUT", "/users/123?query=string", map[string]string{"sessionid": "zyx"}, ""),
				session,
				json.RawMessage(fmt.Sprintf(`{"check_session_url": "%s"}`, testServer.URL)),
				nil,
			)
			require.NoError(t, err, "%#v", errors.Cause(err))
			assert.Len(t, requestRecorder.requests, 1)
			r := requestRecorder.requests[0]
			assert.Equal(t, r.Method, "PUT")
			assert.Equal(t, r.URL.Path, "/users/123?query=string")
			assert.Equal(t, r.Header.Get("Cookie"), "sessionid=zyx")
			assert.Equal(t, &AuthenticationSession{Subject: "123"}, session)
		})

		t.Run("description=should pass through method and headers ONLY to auth server when PreservePath is true", func(t *testing.T) {
			testServer, requestRecorder := makeServer(200, `{"subject": "123"}`)
			err := pipelineAuthenticator.Authenticate(
				makeRequest("PUT", "/users/123?query=string", map[string]string{"sessionid": "zyx"}, ""),
				session,
				json.RawMessage(fmt.Sprintf(`{"check_session_url": "%s", "preserve_path": true}`, testServer.URL)),
				nil,
			)
			require.NoError(t, err, "%#v", errors.Cause(err))
			assert.Len(t, requestRecorder.requests, 1)
			r := requestRecorder.requests[0]
			assert.Equal(t, r.Method, "PUT")
			assert.Equal(t, r.URL.Path, "/")
			assert.Equal(t, r.Header.Get("Cookie"), "sessionid=zyx")
			assert.Equal(t, &AuthenticationSession{Subject: "123"}, session)
		})

		t.Run("description=does not pass request body through to auth server", func(t *testing.T) {
			testServer, requestRecorder := makeServer(200, `{}`)
			pipelineAuthenticator.Authenticate(
				makeRequest("POST", "/", map[string]string{"sessionid": "zyx"}, "Some body..."),
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
			testServer, requestRecorder := makeServer(200, `{}`)
			err := pipelineAuthenticator.Authenticate(
				makeRequest("GET", "/", map[string]string{"sessionid": "zyx"}, ""),
				session,
				json.RawMessage(fmt.Sprintf(`{"only": ["session", "sid"], "check_session_url": "%s"}`, testServer.URL)),
				nil,
			)
			assert.Equal(t, errors.Cause(err), ErrAuthenticatorNotResponsible)
			assert.Empty(t, requestRecorder.requests)
		})

		t.Run("description=should fallthrough if is missing and it has no cookies", func(t *testing.T) {
			testServer, requestRecorder := makeServer(200, `{}`)
			err := pipelineAuthenticator.Authenticate(
				makeRequest("GET", "/", map[string]string{}, ""),
				session,
				json.RawMessage(fmt.Sprintf(`{"check_session_url": "%s"}`, testServer.URL)),
				nil,
			)
			assert.Equal(t, errors.Cause(err), ErrAuthenticatorNotResponsible)
			assert.Empty(t, requestRecorder.requests)
		})

		t.Run("description=should not fallthrough if only is specified and cookie specified is set", func(t *testing.T) {
			testServer, _ := makeServer(200, `{}`)
			err := pipelineAuthenticator.Authenticate(
				makeRequest("GET", "/", map[string]string{"sid": "zyx"}, ""),
				session,
				json.RawMessage(fmt.Sprintf(`{"only": ["session", "sid"], "check_session_url": "%s"}`, testServer.URL)),
				nil,
			)
			require.NoError(t, err, "%#v", errors.Cause(err))
		})

		t.Run("description=should work with nested extra keys", func(t *testing.T) {
			testServer, _ := makeServer(200, `{"subject": "123", "session": {"foo": "bar"}}`)
			err := pipelineAuthenticator.Authenticate(
				makeRequest("GET", "/", map[string]string{"sessionid": "zyx"}, ""),
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
			testServer, _ := makeServer(200, `{"identity": {"id": "123"}, "session": {"foo": "bar"}}`)
			err := pipelineAuthenticator.Authenticate(
				makeRequest("GET", "/", map[string]string{"sessionid": "zyx"}, ""),
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
	})
}

type RequestRecorder struct {
	requests []*http.Request
	bodies   [][]byte
}

func makeServer(statusCode int, responseBody string) (*httptest.Server, *RequestRecorder) {
	requestRecorder := &RequestRecorder{}
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestRecorder.requests = append(requestRecorder.requests, r)
		requestBody, _ := ioutil.ReadAll(r.Body)
		requestRecorder.bodies = append(requestRecorder.bodies, requestBody)
		w.WriteHeader(statusCode)
		w.Write([]byte(responseBody))
	}))
	return testServer, requestRecorder
}

func makeRequest(method string, path string, cookies map[string]string, bodyStr string) *http.Request {
	var body io.ReadCloser
	header := http.Header{}
	if bodyStr != "" {
		body = ioutil.NopCloser(bytes.NewBufferString(bodyStr))
		header.Add("Content-Length", strconv.Itoa(len(bodyStr)))
	}
	req := &http.Request{
		Method: method,
		URL:    &url.URL{Path: path},
		Header: header,
		Body:   body,
	}
	for name, value := range cookies {
		req.AddCookie(&http.Cookie{Name: name, Value: value})
	}
	return req
}
