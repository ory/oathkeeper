// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package authz_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/sjson"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	"github.com/ory/oathkeeper/helper"
	"github.com/ory/oathkeeper/internal"

	"github.com/ory/oathkeeper/driver/configuration"
	"github.com/ory/oathkeeper/pipeline/authn"
	. "github.com/ory/oathkeeper/pipeline/authz"
	"github.com/ory/oathkeeper/rule"
	"github.com/ory/x/configx"
)

func TestAuthorizerRemoteJSONAuthorize(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name               string
		setup              func(t *testing.T) *httptest.Server
		session            *authn.AuthenticationSession
		sessionHeaderMatch *http.Header
		config             json.RawMessage
		assertErr          func(t *testing.T, err error)
	}{
		{
			name:    "invalid configuration",
			session: &authn.AuthenticationSession{},
			config:  json.RawMessage(`{}`),
			assertErr: func(t *testing.T, err error) {
				assert.ErrorIs(t, err, ErrAuthorizerNotEnabled())
			},
		},
		{
			name:    "unresolvable host",
			session: &authn.AuthenticationSession{},
			config:  json.RawMessage(`{"remote":"http://unresolvable-host/path","payload":"{}"}`),
			assertErr: func(t *testing.T, err error) {
				dnsErr, ok := errors.AsType[*net.DNSError](err)
				require.Truef(t, ok, "%#v", err)
				assert.Equal(t, "unresolvable-host", dnsErr.Name)
			},
		},
		{
			name:    "invalid template",
			session: &authn.AuthenticationSession{},
			config:  json.RawMessage(`{"remote":"http://host/path","payload":"{{"}`),
			assertErr: func(t *testing.T, err error) {
				assert.ErrorContains(t, err, "unclosed action")
			},
		},
		{
			name:    "unknown field",
			session: &authn.AuthenticationSession{},
			config:  json.RawMessage(`{"remote":"http://host/path","payload":"{{ .foo }}"}`),
			assertErr: func(t *testing.T, err error) {
				assert.ErrorContains(t, err, "can't evaluate field foo")
			},
		},
		{
			name:    "invalid json",
			session: &authn.AuthenticationSession{},
			config:  json.RawMessage(`{"remote":"http://host/path","payload":"{"}`),
			assertErr: func(t *testing.T, err error) {
				assert.ErrorContains(t, err, "unexpected end of JSON input")
			},
		},
		{
			name:    "invalid headers type",
			session: &authn.AuthenticationSession{},
			config:  json.RawMessage(`{"remote":"http://host/path","payload":"{\"match\":\"baz\"}","headers":"string"}`),
			assertErr: func(t *testing.T, err error) {
				assert.ErrorIs(t, err, ErrAuthorizerNotEnabled())
			},
		},
		{
			name:    "invalid headers template",
			session: &authn.AuthenticationSession{},
			config:  json.RawMessage(`{"remote":"http://host/path","payload":"{\"match\":\"baz\"}","headers":{"Subject":"{{ Invalid Template }}"}}`),
			assertErr: func(t *testing.T, err error) {
				assert.ErrorContains(t, err, `function "Invalid" not defined`)
			},
		},
		{
			name:    "headers template with unknown field",
			session: &authn.AuthenticationSession{},
			config:  json.RawMessage(`{"remote":"http://host/path","payload":"{\"match\":\"baz\"}","headers":{"Subject":"{{ .UnknownField }}"}}`),
			assertErr: func(t *testing.T, err error) {
				assert.ErrorContains(t, err, "can't evaluate field UnknownField")
			},
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
			assertErr: func(t *testing.T, err error) {
				assert.ErrorIs(t, err, helper.ErrForbidden())
			},
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
			assertErr: func(t *testing.T, err error) {
				assert.ErrorContains(t, err, "expected status code 200 but got 400")
			},
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
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.setup != nil {
				server := tc.setup(t)
				t.Cleanup(server.Close)
				tc.config, _ = sjson.SetBytes(tc.config, "remote", server.URL)
			}

			reg := internal.NewRegistry(t)
			a := NewAuthorizerRemoteJSON(reg)
			r, err := http.NewRequestWithContext(t.Context(), "", "", nil)
			require.NoError(t, err)
			r.Header = map[string][]string{"Authorization": {"Bearer token"}}

			err = a.Authorize(r, tc.session, tc.config, &rule.Rule{})
			if tc.assertErr != nil {
				tc.assertErr(t, err)
			} else {
				assert.NoError(t, err)
			}

			if tc.sessionHeaderMatch != nil {
				assert.Equal(t, tc.sessionHeaderMatch, &tc.session.Header)
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
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			reg := internal.NewRegistry(t,
				configx.WithValue(configuration.AuthorizerRemoteJSONIsEnabled, tc.enabled),
				configx.SkipValidation(),
			)
			a := NewAuthorizerRemoteJSON(reg)
			if err := a.Validate(tc.config); (err != nil) != tc.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tc.wantErr)
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
	for _, tc := range tests {
		t.Run("case="+tc.name, func(t *testing.T) {
			reg := internal.NewRegistry(t)
			a := NewAuthorizerRemoteJSON(reg)
			actual, err := a.Config(tc.raw)
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestAuthorizerRemoteJSONTracePropagation(t *testing.T) {
	// This test must NOT use t.Parallel() because it mutates global OTEL state.
	// t.Parallel()

	// Set up a real tracer provider so otelhttp.NewTransport creates sampled spans.
	tp := sdktrace.NewTracerProvider(sdktrace.WithSampler(sdktrace.AlwaysSample()))

	prevTP := otel.GetTracerProvider()
	prevProp := otel.GetTextMapPropagator()
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})
	t.Cleanup(func() {
		otel.SetTracerProvider(prevTP)
		otel.SetTextMapPropagator(prevProp)
		_ = tp.Shutdown(t.Context())
	})

	var gotTraceparent string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotTraceparent = r.Header.Get("Traceparent")
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(server.Close)

	config := json.RawMessage(fmt.Sprintf(`{"remote":%q,"payload":"{}"}`, server.URL))

	reg := internal.NewRegistry(t)
	a := NewAuthorizerRemoteJSON(reg)
	r, err := http.NewRequestWithContext(t.Context(), "", "", nil)
	require.NoError(t, err)
	err = a.Authorize(r, &authn.AuthenticationSession{}, config, &rule.Rule{})
	require.NoError(t, err)
	assert.NotEmpty(t, gotTraceparent, "expected traceparent header to be propagated to remote_json authorizer endpoint")
}
