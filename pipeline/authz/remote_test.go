package authz_test

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/sjson"

	"github.com/ory/x/logrusx"

	"github.com/ory/viper"

	"github.com/ory/oathkeeper/driver/configuration"
	"github.com/ory/oathkeeper/pipeline/authn"
	. "github.com/ory/oathkeeper/pipeline/authz"
	"github.com/ory/oathkeeper/rule"
)

func TestAuthorizerRemoteAuthorize(t *testing.T) {
	tests := []struct {
		name               string
		setup              func(t *testing.T) *httptest.Server
		session            *authn.AuthenticationSession
		sessionHeaderMatch *http.Header
		body               string
		config             json.RawMessage
		wantErr            bool
	}{
		{
			name:    "invalid configuration",
			session: &authn.AuthenticationSession{},
			config:  json.RawMessage(`{}`),
			wantErr: true,
		},
		{
			name:    "unresolvable host",
			session: &authn.AuthenticationSession{},
			config:  json.RawMessage(`{"remote":"http://unresolvable-host/path",}`),
			wantErr: true,
		},
		{
			name:    "invalid json",
			session: &authn.AuthenticationSession{},
			config:  json.RawMessage(`{"remote":"http://host/path","headers":"{"}`),
			wantErr: true,
		},
		{
			name: "forbidden",
			setup: func(t *testing.T) *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusForbidden)
				}))
			},
			session: &authn.AuthenticationSession{},
			config:  json.RawMessage(`{}`),
			wantErr: true,
		},
		{
			name: "unexpected status code",
			setup: func(t *testing.T) *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusBadRequest)
				}))
			},
			session: &authn.AuthenticationSession{},
			config:  json.RawMessage(`{}`),
			wantErr: true,
		},
		{
			name: "nobody",
			setup: func(t *testing.T) *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					assert.Contains(t, r.Header, "Content-Type")
					assert.Contains(t, r.Header["Content-Type"], "text/plain")
					body, err := ioutil.ReadAll(r.Body)
					require.NoError(t, err)
					assert.Equal(t, "", string(body))
					w.WriteHeader(http.StatusOK)
				}))
			},
			body:   "",
			config: json.RawMessage(`{}`),
		},
		{
			name: "ok",
			setup: func(t *testing.T) *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					assert.Contains(t, r.Header, "Content-Type")
					assert.Contains(t, r.Header["Content-Type"], "text/plain")
					assert.Nil(t, r.Header["Authorization"])
					body, err := ioutil.ReadAll(r.Body)
					require.NoError(t, err)
					assert.Equal(t, "testtest", string(body))
					w.WriteHeader(http.StatusOK)
				}))
			},
			body:   "testtest",
			config: json.RawMessage(`{}`),
		},
		{
			name: "ok with large body",
			setup: func(t *testing.T) *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					body, err := ioutil.ReadAll(r.Body)
					require.NoError(t, err)
					assert.True(t, strings.Repeat("1", 1024*1024) == string(body))
					w.WriteHeader(http.StatusOK)
				}))
			},
			body:   strings.Repeat("1", 1024*1024),
			config: json.RawMessage(`{}`),
		},
		{
			name: "ok with allowed headers",
			setup: func(t *testing.T) *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("X-Foo", "bar")
					w.WriteHeader(http.StatusOK)
				}))
			},
			session:            new(authn.AuthenticationSession),
			sessionHeaderMatch: &http.Header{"X-Foo": []string{"bar"}},
			config:             json.RawMessage(`{"forward_response_headers_to_upstream":["X-Foo"]}`),
		},
		{
			name: "ok with not allowed headers",
			setup: func(t *testing.T) *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("X-Bar", "foo")
					w.WriteHeader(http.StatusOK)
				}))
			},
			session:            new(authn.AuthenticationSession),
			sessionHeaderMatch: &http.Header{"X-Foo": []string{""}},
			config:             json.RawMessage(`{"forward_response_headers_to_upstream":["X-Foo"]}`),
		},
		{
			name: "authentication session",
			setup: func(t *testing.T) *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					assert.Contains(t, r.Header, "Subject")
					assert.Contains(t, r.Header["Subject"], "alice")
					w.WriteHeader(http.StatusOK)
				}))
			},
			session: &authn.AuthenticationSession{
				Subject: "alice",
				Extra:   map[string]interface{}{"foo": "bar"},
				MatchContext: authn.MatchContext{
					RegexpCaptureGroups: []string{"baz"},
				},
			},
			config: json.RawMessage(`{"headers":{"Subject": "{{ .Subject }}"}}`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				server := tt.setup(t)
				defer server.Close()
				tt.config, _ = sjson.SetBytes(tt.config, "remote", server.URL)
			}

			p := configuration.NewViperProvider(logrusx.New("", ""))
			a := NewAuthorizerRemote(p)
			r := &http.Request{
				Header: map[string][]string{
					"Content-Type": {"text/plain"},
					"User-Agent":   {"Fancy Browser 5.1"},
				},
			}
			if tt.body != "" {
				r.Body = ioutil.NopCloser(strings.NewReader(tt.body))
			}
			if err := a.Authorize(r, tt.session, tt.config, &rule.Rule{}); (err != nil) != tt.wantErr {
				t.Errorf("Authorize() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.sessionHeaderMatch != nil {
				assert.Equal(t, tt.sessionHeaderMatch, &tt.session.Header)
			}
		})
	}
}

func TestAuthorizerRemoteValidate(t *testing.T) {
	tests := []struct {
		name    string
		enabled bool
		config  json.RawMessage
		wantErr bool
	}{
		{
			name:    "disabled",
			config:  json.RawMessage(`{}`),
			wantErr: true,
		},
		{
			name:    "empty configuration",
			enabled: true,
			config:  json.RawMessage(`{}`),
			wantErr: true,
		},
		{
			name:    "missing remote",
			enabled: true,
			config:  json.RawMessage(`{}`),
			wantErr: true,
		},
		{
			name:    "invalid url",
			enabled: true,
			config:  json.RawMessage(`{"remote":"invalid-url",}`),
			wantErr: true,
		},
		{
			name:    "valid configuration",
			enabled: true,
			config:  json.RawMessage(`{"remote":"http://host/path"}`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := configuration.NewViperProvider(logrusx.New("", ""))
			a := NewAuthorizerRemote(p)
			viper.Set(configuration.ViperKeyAuthorizerRemoteIsEnabled, tt.enabled)
			if err := a.Validate(tt.config); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
