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

func TestAuthenticatorForward(t *testing.T) {
	conf := internal.NewConfigurationWithDefaults()
	reg := internal.NewRegistry(conf)
	session := new(AuthenticationSession)

	pipelineAuthenticator, err := reg.PipelineAuthenticator("forward")
	require.NoError(t, err)

	t.Run("method=authenticate", func(t *testing.T) {
		t.Run("description=should fail because remote returned 400", func(t *testing.T) {
			testServer, _ := makeServiceServer(400, `{}`)
			defer testServer.Close()
			err := pipelineAuthenticator.Authenticate(
				makeForwardRequest("GET", "/", map[string]string{"sessionid": "zyx"}, ""),
				session,
				json.RawMessage(fmt.Sprintf(`{"check_session_url": "%s"}`, testServer.URL)),
				nil,
			)
			require.Error(t, err, "%#v", errors.Cause(err))
		})

		t.Run("description=should pass because remote returned 200", func(t *testing.T) {
			testServer, _ := makeServiceServer(200, `{"subject": "123", "extra": {"foo": "bar"}}`)
			defer testServer.Close()
			err := pipelineAuthenticator.Authenticate(
				makeForwardRequest("GET", "/", map[string]string{"sessionid": "zyx"}, ""),
				session,
				json.RawMessage(fmt.Sprintf(`{"service_url": "%s"}`, testServer.URL)),
				nil,
			)
			require.NoError(t, err, "%#v", errors.Cause(err))
			assert.Equal(t, &AuthenticationSession{
				Subject: "123",
				Extra:   map[string]interface{}{"foo": "bar"},
			}, session)
		})

		t.Run("description=should pass through path, headers, method, and body to auth server", func(t *testing.T) {
			testServer, requestRecorder := makeServiceServer(200, `{"subject": "123"}`)
			defer testServer.Close()
			err := pipelineAuthenticator.Authenticate(
				makeForwardRequest("PUT", "/users/123?query=string", map[string]string{"sessionid": "zyx"}, "Test body"),
				session,
				json.RawMessage(fmt.Sprintf(`{"service_url": "%s"}`, testServer.URL)),
				nil,
			)
			require.NoError(t, err, "%#v", errors.Cause(err))
			assert.Len(t, requestRecorder.requests, 1)

			r := requestRecorder.requests[0]

			assert.Equal(t, r.Method, "PUT")
			assert.Equal(t, r.URL.Path, "/users/123?query=string")
			assert.Equal(t, r.Header.Get("sessionid"), "zyx")
			assert.Equal(t, requestRecorder.bodies[0], []byte("Test body"))
			assert.Equal(t, &AuthenticationSession{Subject: "123"}, session)
		})

		t.Run("description=should pass through path, headers, and body and use custom method to auth server", func(t *testing.T) {
			testServer, requestRecorder := makeServiceServer(200, `{"subject": "123"}`)
			defer testServer.Close()
			err := pipelineAuthenticator.Authenticate(
				makeForwardRequest("PUT", "/users/123?query=string", map[string]string{"sessionid": "zyx"}, "Test body"),
				session,
				json.RawMessage(fmt.Sprintf(`{"service_url": "%s", "method": "POST"}`, testServer.URL)),
				nil,
			)
			require.NoError(t, err, "%#v", errors.Cause(err))
			assert.Len(t, requestRecorder.requests, 1)
			r := requestRecorder.requests[0]
			assert.Equal(t, r.Method, "POST")
			assert.Equal(t, r.URL.Path, "/users/123?query=string")
			assert.Equal(t, r.Header.Get("sessionid"), "zyx")
			assert.Equal(t, requestRecorder.bodies[0], []byte("Test body"))
			assert.Equal(t, &AuthenticationSession{Subject: "123"}, session)
		})

		t.Run("description=should pass through method, headers and body to auth server when PreservePath is true (preserve the path from config remote URL)", func(t *testing.T) {
			testServer, requestRecorder := makeServiceServer(200, `{"subject": "123"}`)
			defer testServer.Close()
			err := pipelineAuthenticator.Authenticate(
				makeForwardRequest("PUT", "/users/123?query=string", map[string]string{"sessionid": "zyx"}, ""),
				session,
				json.RawMessage(fmt.Sprintf(`{"service_url": "%s", "preserve_path": true}`, testServer.URL)),
				nil,
			)
			require.NoError(t, err, "%#v", errors.Cause(err))
			assert.Len(t, requestRecorder.requests, 1)
			r := requestRecorder.requests[0]
			assert.Equal(t, r.Method, "PUT")
			assert.Equal(t, r.URL.Path, "/")
			assert.Equal(t, r.Header.Get("sessionid"), "zyx")
			assert.Equal(t, &AuthenticationSession{Subject: "123"}, session)
		})

		t.Run("description=should work with nested extra keys", func(t *testing.T) {
			testServer, _ := makeServiceServer(200, `{"subject": "123", "session": {"foo": "bar"}}`)
			defer testServer.Close()
			err := pipelineAuthenticator.Authenticate(
				makeForwardRequest("GET", "/", map[string]string{"sessionid": "zyx"}, ""),
				session,
				json.RawMessage(fmt.Sprintf(`{"service_url": "%s", "extra_from": "session"}`, testServer.URL)),
				nil,
			)
			require.NoError(t, err, "%#v", errors.Cause(err))
			assert.Equal(t, &AuthenticationSession{
				Subject: "123",
				Extra:   map[string]interface{}{"foo": "bar"},
			}, session)
		})

		t.Run("description=should work with the root key for extra and a custom subject key", func(t *testing.T) {
			testServer, _ := makeServiceServer(200, `{"identity": {"id": "123"}, "session": {"foo": "bar"}}`)
			defer testServer.Close()
			err := pipelineAuthenticator.Authenticate(
				makeForwardRequest("GET", "/", map[string]string{"sessionid": "zyx"}, ""),
				session,
				json.RawMessage(fmt.Sprintf(`{"service_url": "%s", "subject_from": "identity.id", "extra_from": "@this"}`, testServer.URL)),
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

type ForwardRequestRecorder struct {
	requests []*http.Request
	bodies   [][]byte
}

func makeServiceServer(statusCode int, responseBody string) (*httptest.Server, *ForwardRequestRecorder) {
	requestRecorder := &ForwardRequestRecorder{}
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestRecorder.requests = append(requestRecorder.requests, r)
		requestBody, _ := ioutil.ReadAll(r.Body)
		requestRecorder.bodies = append(requestRecorder.bodies, requestBody)
		w.WriteHeader(statusCode)
		w.Write([]byte(responseBody))
	}))
	return testServer, requestRecorder
}

func makeForwardRequest(method string, path string, headers map[string]string, bodyStr string) *http.Request {
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
	for name, value := range headers {
		req.Header.Add(name, value)
	}
	return req
}
