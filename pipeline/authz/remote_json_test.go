// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package authz_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/sjson"

	"github.com/ory/x/configx"
	"github.com/ory/x/logrusx"

	"github.com/ory/oathkeeper/driver/configuration"
	"github.com/ory/oathkeeper/pipeline/authn"
	. "github.com/ory/oathkeeper/pipeline/authz"
	"github.com/ory/oathkeeper/rule"
	"github.com/ory/x/otelx"
)

func TestAuthorizerRemoteJSONAuthorize(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name               string
		setup              func(t *testing.T) *httptest.Server
		session            *authn.AuthenticationSession
		sessionHeaderMatch *http.Header
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
			name:    "invalid headers type",
			session: &authn.AuthenticationSession{},
			config:  json.RawMessage(`{"remote":"http://host/path","payload":"{\"match\":\"baz\"}","headers":"string"}`),
			wantErr: true,
		},
		{
			name:    "invalid headers template",
			session: &authn.AuthenticationSession{},
			config:  json.RawMessage(`{"remote":"http://host/path","payload":"{\"match\":\"baz\"}","headers":{"Subject":"{{ Invalid Template }}"}}`),
			wantErr: true,
		},
		{
			name:    "headers template with unknown field",
			session: &authn.AuthenticationSession{},
			config:  json.RawMessage(`{"remote":"http://host/path","payload":"{\"match\":\"baz\"}","headers":{"Subject":"{{ .UnknownField }}"}}`),
			wantErr: true,
		},
		{
			name: "forbidden",
			setup: func(t *testing.T) *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
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
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
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
					assert.Contains(t, r.Header, "Authorization")
					assert.Contains(t, r.Header["Authorization"], "Bearer token")
					body, err := io.ReadAll(r.Body)
					require.NoError(t, err)
					assert.Equal(t, string(body), "{}")
					w.WriteHeader(http.StatusOK)
				}))
			},
			session: &authn.AuthenticationSession{},
			config:  json.RawMessage(`{"payload":"{}"}`),
		},
		{
			name: "ok with allowed response headers",
			setup: func(t *testing.T) *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					w.Header().Set("X-Foo", "bar")
					w.WriteHeader(http.StatusOK)
				}))
			},
			session:            new(authn.AuthenticationSession),
			sessionHeaderMatch: &http.Header{"X-Foo": []string{"bar"}},
			config:             json.RawMessage(`{"payload":"{}","forward_response_headers_to_upstream":["X-Foo"]}`),
		},
		{
			name: "ok with not allowed response headers",
			setup: func(t *testing.T) *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					w.Header().Set("X-Bar", "foo")
					w.WriteHeader(http.StatusOK)
				}))
			},
			session:            new(authn.AuthenticationSession),
			sessionHeaderMatch: &http.Header{"X-Foo": []string{""}},
			config:             json.RawMessage(`{"payload":"{}","forward_response_headers_to_upstream":["X-Foo"]}`),
		},
		{
			name: "authentication session",
			setup: func(t *testing.T) *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					body, err := io.ReadAll(r.Body)
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
			name: "authentication session with extra request headers",
			setup: func(t *testing.T) *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					body, err := io.ReadAll(r.Body)

					require.NoError(t, err)
					assert.Equal(t, string(body), `{"match":"baz"}`)
					assert.Equal(t, r.Header.Get("Subject"), "alice")
					w.WriteHeader(http.StatusOK)
				}))
			},
			session: &authn.AuthenticationSession{
				Subject: "alice",
			},
			config: json.RawMessage(`{"payload":"{\"match\":\"baz\"}","headers":{"Subject":"{{ .Subject }}","Empty-Header":""}}`),
		},
		{
			name: "json array",
			setup: func(t *testing.T) *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					body, err := io.ReadAll(r.Body)
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
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if tt.setup != nil {
				server := tt.setup(t)
				defer server.Close()
				tt.config, _ = sjson.SetBytes(tt.config, "remote", server.URL)
			}

			l := logrusx.New("", "")
			p, err := configuration.NewKoanfProvider(context.Background(), nil, l)
			if err != nil {
				l.WithError(err).Fatal("Failed to initialize configuration")
			}
			a := NewAuthorizerRemoteJSON(p, otelx.NewNoop(l, p.TracingConfig()))
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()
			r, err := http.NewRequestWithContext(ctx, "", "", nil)
			require.NoError(t, err)
			r.Header = map[string][]string{"Authorization": {"Bearer token"}}
			if err := a.Authorize(r, tt.session, tt.config, &rule.Rule{}); (err != nil) != tt.wantErr {
				t.Errorf("Authorize() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.sessionHeaderMatch != nil {
				assert.Equal(t, tt.sessionHeaderMatch, &tt.session.Header)
			}
		})
	}
}

func TestAuthorizerRemoteJSONValidate(t *testing.T) {
	t.Parallel()
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
		{
			name:    "valid configuration with headers",
			enabled: true,
			config:  json.RawMessage(`{"remote":"http://host/path","payload":"{}","headers":{"Authorization":"Bearer token"}}`),
		},
		{
			name:    "valid configuration with partial retry 1",
			enabled: true,
			config:  json.RawMessage(`{"remote":"http://host/path","payload":"{}","retry":{"max_delay":"100ms"}}`),
		},
		{
			name:    "valid configuration with partial retry 2",
			enabled: true,
			config:  json.RawMessage(`{"remote":"http://host/path","payload":"{}","retry":{"give_up_after":"3s"}}`),
		},
		{
			name:    "valid configuration with retry",
			enabled: true,
			config:  json.RawMessage(`{"remote":"http://host/path","payload":"{}","retry":{"give_up_after":"3s", "max_delay":"100ms"}}`),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			p, err := configuration.NewKoanfProvider(
				context.Background(), nil, logrusx.New("", ""),
				configx.SkipValidation(),
			)
			require.NoError(t, err)
			l := logrusx.New("", "")
			a := NewAuthorizerRemoteJSON(p, otelx.NewNoop(l, p.TracingConfig()))
			p.SetForTest(t, configuration.AuthorizerRemoteJSONIsEnabled, tt.enabled)
			if err := a.Validate(tt.config); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAuthorizerRemoteJSONConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		raw      json.RawMessage
		expected *AuthorizerRemoteJSONConfiguration
	}{
		{
			name: "valid configuration with forward_response_headers_to_upstream",
			raw:  json.RawMessage(`{"remote":"http://host/path","payload":"{}","forward_response_headers_to_upstream":["X-Foo"]}`),
			expected: &AuthorizerRemoteJSONConfiguration{
				Remote:                           "http://host/path",
				Payload:                          "{}",
				ForwardResponseHeadersToUpstream: []string{"X-Foo"},
				Retry: &AuthorizerRemoteJSONRetryConfiguration{
					Timeout: "100ms", // default timeout from schema
					MaxWait: "1s",
				},
			},
		},
		{
			name: "valid configuration without forward_response_headers_to_upstream",
			raw:  json.RawMessage(`{"remote":"http://host/path","payload":"{}"}`),
			expected: &AuthorizerRemoteJSONConfiguration{
				Remote:                           "http://host/path",
				Payload:                          "{}",
				ForwardResponseHeadersToUpstream: []string{},
				Retry: &AuthorizerRemoteJSONRetryConfiguration{
					Timeout: "100ms", // default timeout from schema
					MaxWait: "1s",
				},
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run("case="+tt.name, func(t *testing.T) {
			t.Parallel()
			p, err := configuration.NewKoanfProvider(
				context.Background(), nil, logrusx.New("", ""),
			)
			require.NoError(t, err)
			l := logrusx.New("", "")
			a := NewAuthorizerRemoteJSON(p, otelx.NewNoop(l, p.TracingConfig()))
			actual, err := a.Config(tt.raw)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, actual)
		})
	}
}
