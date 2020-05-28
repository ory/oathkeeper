package authz_test

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
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

func TestAuthorizerRemoteJSONAuthorize(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(t *testing.T) *httptest.Server
		session *authn.AuthenticationSession
		config  json.RawMessage
		wantErr bool
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
			config:  json.RawMessage(`{"remote":"http://unresolvable-host/path","payload":"{}"}`),
			wantErr: true,
		},
		{
			name:    "invalid template",
			session: &authn.AuthenticationSession{},
			config:  json.RawMessage(`{"remote":"http://host/path","payload":"{{"}`),
			wantErr: true,
		},
		{
			name:    "unknown field",
			session: &authn.AuthenticationSession{},
			config:  json.RawMessage(`{"remote":"http://host/path","payload":"{{ .foo }}"}`),
			wantErr: true,
		},
		{
			name:    "invalid json",
			session: &authn.AuthenticationSession{},
			config:  json.RawMessage(`{"remote":"http://host/path","payload":"{"}`),
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
			config:  json.RawMessage(`{"payload":"{}"}`),
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
			config:  json.RawMessage(`{"payload":"{}"}`),
			wantErr: true,
		},
		{
			name: "ok",
			setup: func(t *testing.T) *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					assert.Contains(t, r.Header, "Content-Type")
					assert.Contains(t, r.Header["Content-Type"], "application/json")
					body, err := ioutil.ReadAll(r.Body)
					require.NoError(t, err)
					assert.Equal(t, string(body), "{}")
					w.WriteHeader(http.StatusOK)
				}))
			},
			session: &authn.AuthenticationSession{},
			config:  json.RawMessage(`{"payload":"{}"}`),
		},
		{
			name: "authentication session",
			setup: func(t *testing.T) *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					body, err := ioutil.ReadAll(r.Body)
					require.NoError(t, err)
					assert.Equal(t, string(body), `{"subject":"alice","extra":"bar","match":"baz"}`)
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
			config: json.RawMessage(`{"payload":"{\"subject\":\"{{ .Subject }}\",\"extra\":\"{{ .Extra.foo }}\",\"match\":\"{{ index .MatchContext.RegexpCaptureGroups 0 }}\"}"}`),
		},
		{
			name: "json array",
			setup: func(t *testing.T) *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					body, err := ioutil.ReadAll(r.Body)
					require.NoError(t, err)
					assert.Equal(t, string(body), `["foo","bar"]`)
					w.WriteHeader(http.StatusOK)
				}))
			},
			session: &authn.AuthenticationSession{},
			config:  json.RawMessage(`{"payload":"[\"foo\",\"bar\"]"}`),
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
			a := NewAuthorizerRemoteJSON(p)
			if err := a.Authorize(&http.Request{}, tt.session, tt.config, &rule.Rule{}); (err != nil) != tt.wantErr {
				t.Errorf("Authorize() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAuthorizerRemoteJSONValidate(t *testing.T) {
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
			name:    "missing payload",
			enabled: true,
			config:  json.RawMessage(`{"remote":"http://host/path"}`),
			wantErr: true,
		},
		{
			name:    "missing remote",
			enabled: true,
			config:  json.RawMessage(`{"payload":"{}"}`),
			wantErr: true,
		},
		{
			name:    "invalid url",
			enabled: true,
			config:  json.RawMessage(`{"remote":"invalid-url","payload":"{}"}`),
			wantErr: true,
		},
		{
			name:    "valid configuration",
			enabled: true,
			config:  json.RawMessage(`{"remote":"http://host/path","payload":"{}"}`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := configuration.NewViperProvider(logrusx.New("", ""))
			a := NewAuthorizerRemoteJSON(p)
			viper.Set(configuration.ViperKeyAuthorizerRemoteJSONIsEnabled, tt.enabled)
			if err := a.Validate(tt.config); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
