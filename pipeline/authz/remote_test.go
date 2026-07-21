// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package authz_test

import (
	"encoding/json"
	"errors"
	fmt "fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/sjson"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	"github.com/ory/oathkeeper/internal"

	"github.com/ory/oathkeeper/driver/configuration"
	"github.com/ory/oathkeeper/helper"
	"github.com/ory/oathkeeper/pipeline/authn"
	. "github.com/ory/oathkeeper/pipeline/authz"
	"github.com/ory/oathkeeper/rule"
	"github.com/ory/x/configx"
)

func TestAuthorizerRemoteAuthorize(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name               string
		setup              func(t *testing.T) *httptest.Server
		session            *authn.AuthenticationSession
		sessionHeaderMatch *http.Header
		body               string
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
			config:  json.RawMessage(`{"remote":"http://unresolvable-host/path"}`),
			assertErr: func(t *testing.T, err error) {
				dnsErr, ok := errors.AsType[*net.DNSError](err)
				require.Truef(t, ok, "%#v", err)
				assert.Equal(t, "unresolvable-host", dnsErr.Name)
			},
		},
		{
			name:    "invalid headers type",
			session: &authn.AuthenticationSession{},
			config:  json.RawMessage(`{"remote":"http://host/path","headers":"string"}`),
			assertErr: func(t *testing.T, err error) {
				assert.ErrorIs(t, err, ErrAuthorizerNotEnabled())
			},
		},
		{
			name:    "invalid headers template",
			session: &authn.AuthenticationSession{},
			config:  json.RawMessage(`{"remote":"http://host/path","headers":{"Subject":"{{ Invalid Template }}"}}`),
			assertErr: func(t *testing.T, err error) {
				assert.ErrorContains(t, err, `function "Invalid" not defined`)
			},
		},
		{
			name:    "headers template with unknown field",
			session: &authn.AuthenticationSession{},
			config:  json.RawMessage(`{"remote":"http://host/path","headers":{"Subject":"{{ .UnknownField }}"}}`),
			assertErr: func(t *testing.T, err error) {
				assert.ErrorContains(t, err, `can't evaluate field UnknownField`)
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
			config:  json.RawMessage(`{}`),
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
			config:  json.RawMessage(`{}`),
			assertErr: func(t *testing.T, err error) {
				assert.ErrorContains(t, err, "unexpected status code")
			},
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
	} {
		t.Run(tc.name, func(t *testing.T) {
			if tc.setup != nil {
				server := tc.setup(t)
				t.Cleanup(server.Close)
				var err error
				tc.config, err = sjson.SetBytes(tc.config, "remote", server.URL)
				require.NoError(t, err)
			}

			reg := internal.NewRegistry(t)

			a := NewAuthorizerRemote(reg)
			r := &http.Request{
				Header: map[string][]string{
					"Content-Type": {"text/plain"},
					"User-Agent":   {"Fancy Browser 5.1"},
				},
			}
			if tc.body != "" {
				r.Body = io.NopCloser(strings.NewReader(tc.body))
			}

			err := a.Authorize(r, tc.session, tc.config, &rule.Rule{})
			if tc.assertErr != nil {
				tc.assertErr(t, err)
			} else {
				assert.NoError(t, err)
			}
			if tc.body != "" {
				body, err := io.ReadAll(r.Body)
				require.NoError(t, err)
				assert.Equal(t, tc.body, string(body), "body must stay intact")
			}

			if tc.sessionHeaderMatch != nil {
				assert.Equal(t, tc.sessionHeaderMatch, &tc.session.Header)
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
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			reg := internal.NewRegistry(t,
				configx.WithValue(configuration.AuthorizerRemoteIsEnabled, tc.enabled),
				configx.SkipValidation(),
			)
			a := NewAuthorizerRemote(reg)
			if err := a.Validate(tc.config); (err != nil) != tc.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}

// This test must NOT use t.Parallel() because it mutates global OTEL state.
func TestAuthorizerRemoteTracePropagation(t *testing.T) {
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

	config := json.RawMessage(fmt.Sprintf(`{"remote":%q}`, server.URL))

	reg := internal.NewRegistry(t)

	a := NewAuthorizerRemote(reg)
	r, err := http.NewRequestWithContext(t.Context(), "POST", "", nil)
	require.NoError(t, err)
	r.Header.Set("Content-Type", "text/plain")
	err = a.Authorize(r, &authn.AuthenticationSession{}, config, &rule.Rule{})
	require.NoError(t, err)
	assert.NotEmpty(t, gotTraceparent, "expected traceparent header to be propagated to remote authorizer endpoint")
}

// TestAuthorizerRemoteHonorsMaxDelayTimeout is a regression test for a bug where
// the parsed retry.max_delay duration was multiplied by an extra factor of
// time.Millisecond, inflating the outbound HTTP timeout by 1e6 (a 50ms setting
// became ~13.9 hours, effectively disabling the timeout). With the bug the call
// blocks until the slow server responds; with the fix it times out promptly.
func TestAuthorizerRemoteHonorsMaxDelayTimeout(t *testing.T) {
	t.Parallel()

	slow := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(slow.Close)

	config, err := sjson.SetBytes(
		[]byte(`{"retry":{"max_delay":"50ms","give_up_after":"10ms"}}`), "remote", slow.URL)
	require.NoError(t, err)

	reg := internal.NewRegistry(t)
	a := NewAuthorizerRemote(reg)
	r, err := http.NewRequestWithContext(t.Context(), "POST", "", nil)
	require.NoError(t, err)
	r.Header.Set("Content-Type", "text/plain")

	start := time.Now()
	err = a.Authorize(r, &authn.AuthenticationSession{}, config, &rule.Rule{})
	elapsed := time.Since(start)

	require.Error(t, err)
	assert.Less(t, elapsed, time.Second,
		"the configured 50ms max_delay must time the request out; the inflated value would block ~2s on the slow server")
}
