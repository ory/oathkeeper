package authn_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"

	"github.com/tidwall/sjson"

	"net/http/httptest"

	"github.com/julienschmidt/httprouter"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/oathkeeper/internal"
	. "github.com/ory/oathkeeper/pipeline/authn"
)

func TestAuthenticatorBearerToken(t *testing.T) {
	conf := internal.NewConfigurationWithDefaults()
	reg := internal.NewRegistry(conf)

	pipelineAuthenticator, err := reg.PipelineAuthenticator("bearer_token")
	require.NoError(t, err)

	t.Run("method=authenticate", func(t *testing.T) {
		for k, tc := range []struct {
			d              string
			r              *http.Request
			setup          func(*testing.T, *httprouter.Router)
			router         func(http.ResponseWriter, *http.Request)
			config         json.RawMessage
			expectErr      bool
			expectExactErr error
			expectSess     *AuthenticationSession
		}{
			{
				d:         "should fail because no payloads",
				r:         &http.Request{Header: http.Header{}},
				expectErr: true,
			},
			{
				d:              "should return error saying that authenticator is not responsible for validating the request, as the token was not provided in a proper location (default)",
				r:              &http.Request{Header: http.Header{"Foobar": {"bearer token"}}},
				expectErr:      true,
				expectExactErr: ErrAuthenticatorNotResponsible,
			},
			{
				d:              "should return error saying that authenticator is not responsible for validating the request, as the token was not provided in a proper location (custom header)",
				r:              &http.Request{Header: http.Header{"Authorization": {"bearer token"}}},
				config:         []byte(`{"token_from": {"header": "X-Custom-Header"}}`),
				expectErr:      true,
				expectExactErr: ErrAuthenticatorNotResponsible,
			},
			{
				d: "should fail because session store returned 400",
				r: &http.Request{Header: http.Header{"Authorization": {"bearer token"}}, URL: &url.URL{Path: ""}},
				setup: func(t *testing.T, m *httprouter.Router) {
					m.GET("/", func(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
						w.WriteHeader(400)
					})
				},
				expectErr: true,
			},
			{
				d: "should pass because session store returned 200",
				r: &http.Request{Header: http.Header{"Authorization": {"bearer token"}}, URL: &url.URL{Path: ""}},
				setup: func(t *testing.T, m *httprouter.Router) {
					m.GET("/", func(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
						w.WriteHeader(200)
						w.Write([]byte(`{"sub": "123", "extra": {"foo": "bar"}}`))
					})
				},
				expectErr: false,
				expectSess: &AuthenticationSession{
					Subject: "123",
					Extra:   map[string]interface{}{"foo": "bar"},
				},
			},
			{
				d: "should pass through method, path, and headers to auth server",
				r: &http.Request{Header: http.Header{"Authorization": {"bearer zyx"}}, URL: &url.URL{Path: "/users/123?query=string"}, Method: "PUT"},
				router: func(w http.ResponseWriter, r *http.Request) {
					assert.Equal(t, r.Method, "PUT")
					assert.Equal(t, r.URL.Path, "/users/123?query=string")
					assert.Equal(t, r.Header.Get("Authorization"), "bearer zyx")
					w.WriteHeader(200)
					w.Write([]byte(`{"sub": "123"}`))
				},
				expectErr: false,
				expectSess: &AuthenticationSession{
					Subject: "123",
				},
			},
			{
				d: "should pass through method and headers ONLY to auth server when PreservePath is true",
				r: &http.Request{Header: http.Header{"Authorization": {"bearer zyx"}}, URL: &url.URL{Path: "/users/123?query=string"}, Method: "PUT"},
				router: func(w http.ResponseWriter, r *http.Request) {
					assert.Equal(t, r.Method, "PUT")
					assert.Equal(t, r.URL.Path, "/")
					assert.Equal(t, r.Header.Get("Authorization"), "bearer zyx")
					w.WriteHeader(200)
					w.Write([]byte(`{"sub": "123"}`))
				},
				config:    []byte(`{"preserve_path": true}`),
				expectErr: false,
				expectSess: &AuthenticationSession{
					Subject: "123",
				},
			},
			{
				d: "does not pass request body through to auth server",
				r: &http.Request{
					Header: http.Header{
						"Authorization":  {"bearer zyx"},
						"Content-Length": {"4"},
					},
					URL:    &url.URL{Path: "/users/123?query=string"},
					Method: "PUT",
					Body:   ioutil.NopCloser(bytes.NewBufferString("body")),
				},
				router: func(w http.ResponseWriter, r *http.Request) {
					requestBody, _ := ioutil.ReadAll(r.Body)
					assert.Equal(t, r.ContentLength, int64(0))
					assert.Equal(t, requestBody, []byte{})
					w.WriteHeader(200)
					w.Write([]byte(`{"sub": "123"}`))
				},
				expectErr: false,
				expectSess: &AuthenticationSession{
					Subject: "123",
				},
			},
			{
				d: "should work with nested extra keys",
				r: &http.Request{Header: http.Header{"Authorization": {"bearer token"}}, URL: &url.URL{Path: ""}},
				setup: func(t *testing.T, m *httprouter.Router) {
					m.GET("/", func(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
						w.WriteHeader(200)
						w.Write([]byte(`{"sub": "123", "session": {"foo": "bar"}}`))
					})
				},
				config:    []byte(`{"extra_from": "session"}`),
				expectErr: false,
				expectSess: &AuthenticationSession{
					Subject: "123",
					Extra:   map[string]interface{}{"foo": "bar"},
				},
			},
			{
				d: "should work with the root key for extra and a custom subject key",
				r: &http.Request{Header: http.Header{"Authorization": {"bearer token"}}, URL: &url.URL{Path: ""}},
				setup: func(t *testing.T, m *httprouter.Router) {
					m.GET("/", func(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
						w.WriteHeader(200)
						w.Write([]byte(`{"identity": {"id": "123"}, "session": {"foo": "bar"}}`))
					})
				},
				config:    []byte(`{"subject_from": "identity.id", "extra_from": "@this"}`),
				expectErr: false,
				expectSess: &AuthenticationSession{
					Subject: "123",
					Extra:   map[string]interface{}{"session": map[string]interface{}{"foo": "bar"}, "identity": map[string]interface{}{"id": "123"}},
				},
			},
		} {
			t.Run(fmt.Sprintf("case=%d/description=%s", k, tc.d), func(t *testing.T) {

				var ts *httptest.Server
				if tc.router != nil {
					ts = httptest.NewServer(http.HandlerFunc(tc.router))
				} else {
					router := httprouter.New()
					if tc.setup != nil {
						tc.setup(t, router)
					}
					ts = httptest.NewServer(router)
				}
				defer ts.Close()

				tc.config, _ = sjson.SetBytes(tc.config, "check_session_url", ts.URL)
				sess := new(AuthenticationSession)
				err := pipelineAuthenticator.Authenticate(tc.r, sess, tc.config, nil)
				if tc.expectErr {
					require.Error(t, err)
					if tc.expectExactErr != nil {
						assert.EqualError(t, err, tc.expectExactErr.Error(), "%+v", err)
					}
				} else {
					require.NoError(t, err)
				}

				if tc.expectSess != nil {
					assert.Equal(t, tc.expectSess, sess)
				}
			})
		}
	})
}
