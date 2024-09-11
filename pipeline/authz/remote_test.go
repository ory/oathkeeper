// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package authz_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

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

func TestAuthorizerRemoteAuthorize(t *testing.T) {
	t.Parallel()
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
			name:    "invalid headers type",
			session: &authn.AuthenticationSession{},
			config:  json.RawMessage(`{"remote":"http://host/path","headers":"string"}`),
			wantErr: true,
		},
		{
			name:    "invalid headers template",
			session: &authn.AuthenticationSession{},
			config:  json.RawMessage(`{"remote":"http://host/path","headers":{"Subject":"{{ Invalid Template }}"}}`),
			wantErr: true,
		},
		{
			name:    "headers template with unknown field",
			session: &authn.AuthenticationSession{},
			config:  json.RawMessage(`{"remote":"http://host/path","headers":{"Subject":"{{ .UnknownField }}"}}`),
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
			config:  json.RawMessage(`{}`),
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
			config:  json.RawMessage(`{}`),
			wantErr: true,
		},
		{
			name: "nobody",
			setup: func(t *testing.T) *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					assert.Contains(t, r.Header, "Content-Type")
					assert.Contains(t, r.Header["Content-Type"], "text/plain")
					body, err := io.ReadAll(r.Body)
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
					body, err := io.ReadAll(r.Body)
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
					body, err := io.ReadAll(r.Body)
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
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
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
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
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

			l := logrusx.New("", "")
			p, err := configuration.NewKoanfProvider(
				context.Background(), nil, l)
			if err != nil {
				l.WithError(err).Fatal("Failed to initialize configuration")
			}
			a := NewAuthorizerRemote(p, otelx.NewNoop(l, p.TracingConfig()))
			r := &http.Request{
				Header: map[string][]string{
					"Content-Type": {"text/plain"},
					"User-Agent":   {"Fancy Browser 5.1"},
				},
			}
			if tt.body != "" {
				r.Body = io.NopCloser(strings.NewReader(tt.body))
			}
			if err := a.Authorize(r, tt.session, tt.config, &rule.Rule{}); (err != nil) != tt.wantErr {
				t.Errorf("Authorize() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.body != "" {
				body, err := io.ReadAll(r.Body)
				require.NoError(t, err)
				require.Equal(t, tt.body, string(body), "body must stay intact")
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
		{
			name:    "valid configuration with partial retry 1",
			enabled: true,
			config:  json.RawMessage(`{"remote":"http://host/path","retry":{"max_delay":"100ms"}}`),
		},
		{
			name:    "valid configuration with partial retry 2",
			enabled: true,
			config:  json.RawMessage(`{"remote":"http://host/path","retry":{"give_up_after":"3s"}}`),
		},
		{
			name:    "valid configuration with retry",
			enabled: true,
			config:  json.RawMessage(`{"remote":"http://host/path","retry":{"give_up_after":"3s", "max_delay":"100ms"}}`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := configuration.NewKoanfProvider(
				context.Background(), nil, logrusx.New("", ""),
				configx.SkipValidation())
			require.NoError(t, err)
			l := logrusx.New("", "")
			a := NewAuthorizerRemote(p, otelx.NewNoop(l, p.TracingConfig()))
			p.SetForTest(t, configuration.AuthorizerRemoteIsEnabled, tt.enabled)
			if err := a.Validate(tt.config); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
