// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package proxy_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace"

	"github.com/ory/herodot"
	"github.com/ory/x/configx"
	"github.com/ory/x/logrusx"

	"github.com/ory/oathkeeper/driver"
	"github.com/ory/oathkeeper/driver/configuration"
	"github.com/ory/oathkeeper/internal"
	"github.com/ory/oathkeeper/pipeline/authn"
	"github.com/ory/oathkeeper/rule"
	"github.com/ory/oathkeeper/x"
)

var TestHeader = http.Header{"Test-Header": []string{"Test-Value"}}

func newTestRequest(u string) *http.Request {
	return &http.Request{URL: x.ParseURLOrPanic(u), Method: "GET", Header: TestHeader}
}

func TestHandleError(t *testing.T) {
	for k, tc := range []struct {
		d          string
		inputErr   error
		rule       *rule.Rule
		header     http.Header
		assert     func(t *testing.T, w *httptest.ResponseRecorder)
		setup      func(t *testing.T, config configuration.Configuration)
		configOpts []configx.OptionModifier
	}{
		{
			d:        "should return a JSON error per default and work with nil rules",
			inputErr: herodot.ErrNotFound(),
			assert: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Equal(t, 404, w.Code)
				assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
			},
		},
		{
			d:        "should return a 500 error when no handler is enabled",
			inputErr: herodot.ErrNotFound(),
			setup: func(t *testing.T, config configuration.Configuration) {
				config.SetForTest(t, configuration.ErrorsJSONIsEnabled, false)
			},
			assert: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Equal(t, 500, w.Code)
			},
		},
		{
			d:        "should return the found response",
			inputErr: herodot.ErrUnauthorized(),
			setup: func(t *testing.T, config configuration.Configuration) {
				config.SetForTest(t, configuration.ErrorsRedirectIsEnabled, true)
			},
			rule: &rule.Rule{
				Errors: []rule.ErrorHandler{{
					Handler: "redirect",
					Config:  json.RawMessage(`{"to":"http://test/test","when":[{"error":["unauthorized"]}]}`),
				}},
			},
			assert: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Equal(t, 302, w.Code)
				assert.Equal(t, "http://test/test", w.Header().Get("Location"))
			},
		},
		{
			d:        "should return a JSON error because the error is not unauthorized and JSON is the default",
			inputErr: herodot.ErrNotFound(),
			setup: func(t *testing.T, config configuration.Configuration) {
				config.SetForTest(t, configuration.ErrorsRedirectIsEnabled, true)
				config.SetForTest(t, configuration.ErrorsHandlers+".redirect.config.to", "http://test/test")
			},
			rule: &rule.Rule{
				Errors: []rule.ErrorHandler{{
					Handler: "redirect",
					Config:  json.RawMessage(`{"when":[{"error":["unauthorized"]}]}`),
				}},
			},
			assert: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Equal(t, 404, w.Code)
				assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
			},
		},
		{
			d:        "should pick the appropriate (json) error handler for the request when multiple are configured",
			inputErr: herodot.ErrNotFound(),
			setup: func(t *testing.T, config configuration.Configuration) {
				config.SetForTest(t, configuration.ErrorsRedirectIsEnabled, true)
			},
			header: map[string][]string{"Accept": {"application/json"}},
			rule: &rule.Rule{
				Errors: []rule.ErrorHandler{{
					Handler: "json",
					Config:  json.RawMessage(`{"when":[{"error":["unauthorized"],"request":{"header":{"accept":["application/json"]}}}]}`),
				}, {
					Handler: "redirect",
					Config:  json.RawMessage(`{"to":"http://test/test","when":[{"error":["unauthorized"],"request":{"header":{"accept":["application/xml"]}}}]}`),
				}},
			},
			assert: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Equal(t, 404, w.Code)
				assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
			},
		},
		{
			d:        "should redirect to the specified endpoint by picking the appropriate error handler (redirect)",
			inputErr: herodot.ErrUnauthorized(),
			setup: func(t *testing.T, config configuration.Configuration) {
				config.SetForTest(t, configuration.ErrorsRedirectIsEnabled, true)
			},
			header: map[string][]string{"Accept": {"application/xml"}},
			rule: &rule.Rule{
				Errors: []rule.ErrorHandler{{
					Handler: "json",
					Config:  json.RawMessage(`{"when":[{"error":["unauthorized"],"request":{"header":{"accept":["application/json"]}}}]}`),
				}, {
					Handler: "redirect",
					Config:  json.RawMessage(`{"to":"http://test/test","when":[{"error":["unauthorized"],"request":{"header":{"accept":["application/xml"]}}}]}`),
				}},
			},
			assert: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Equal(t, 302, w.Code)
				assert.Equal(t, "http://test/test", w.Header().Get("Location"))
			},
		},
		{
			d:        "should respond with the appropriate fallback handler (here www_authenticate)",
			inputErr: herodot.ErrUnauthorized(),
			setup: func(t *testing.T, config configuration.Configuration) {
				config.SetForTest(t, configuration.ErrorsRedirectIsEnabled, true)
				config.SetForTest(t, configuration.ErrorsWWWAuthenticateIsEnabled, true)
				config.SetForTest(t, configuration.ErrorsFallback, []string{"www_authenticate", "json"})
			},
			header: map[string][]string{"Accept": {"mime/undefined"}},
			rule: &rule.Rule{
				Errors: []rule.ErrorHandler{{
					Handler: "json",
					Config:  json.RawMessage(`{"when":[{"error":["unauthorized"],"request":{"header":{"accept":["application/json"]}}}]}`),
				}, {
					Handler: "redirect",
					Config:  json.RawMessage(`{"to":"http://test/test","when":[{"error":["unauthorized"],"request":{"header":{"accept":["application/xml"]}}}]}`),
				}},
			},
			assert: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Equal(t, 401, w.Code)
				assert.Equal(t, "Basic realm=Please authenticate.", w.Header().Get("WWW-Authenticate"))
			},
		},
		{
			d:        "should respond with the appropriate fallback handler (here json)",
			inputErr: herodot.ErrForbidden(),
			// We set the fallback to first run www_authenticate. But because the error is not_found, as
			// is defined in the when clause, we should see a json error instead!
			configOpts: []configx.OptionModifier{configx.WithConfigFiles(x.WriteFile(t, `
errors:
  fallback:
    - www_authenticate
    - json
  handlers:
    redirect:
      enabled: true
      config:
        to: http://test/test
    www_authenticate:
      enabled: true
      config:
        when:
          - error:
            - not_found
`))},
			header: map[string][]string{"Accept": {"mime/undefined"}},
			rule: &rule.Rule{
				Errors: []rule.ErrorHandler{{
					Handler: "json",
					Config:  json.RawMessage(`{"when":[{"error":["unauthorized"],"request":{"header":{"accept":["application/json"]}}}]}`),
				}, {
					Handler: "redirect",
					Config:  json.RawMessage(`{"to":"http://test/test","when":[{"error":["unauthorized"],"request":{"header":{"accept":["application/xml"]}}}]}`),
				}},
			},
			assert: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Equal(t, 403, w.Code)
				assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
			},
		},
		{
			d:        "should return a 500 error because no fallback could handle the error",
			inputErr: herodot.ErrForbidden(),
			// We set the fallback to first run www_authenticate. But because the error is not_found, as
			// is defined in the when clause, we should see the 500 misconfigured error
			configOpts: []configx.OptionModifier{configx.WithConfigFiles(x.WriteFile(t, `
errors:
  fallback:
    - www_authenticate
  handlers:
    www_authenticate:
      enabled: true
      config:
        when:
          - error:
            - not_found
`))},
			header: map[string][]string{"Accept": {"mime/undefined"}},
			rule: &rule.Rule{
				Errors: []rule.ErrorHandler{{
					Handler: "json",
					Config:  json.RawMessage(`{"when":[{"error":["unauthorized"],"request":{"header":{"accept":["application/json"]}}}]}`),
				}, {
					Handler: "redirect",
					Config:  json.RawMessage(`{"to":"http://test/test","when":[{"error":["unauthorized"],"request":{"header":{"accept":["application/xml"]}}}]}`),
				}},
			},
			assert: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Equal(t, 500, w.Code)
				assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
				assert.Contains(t, w.Body.String(), "no matching error handling strategy was found")
			},
		},
	} {
		t.Run(fmt.Sprintf("case=%d/description=%s", k, tc.d), func(t *testing.T) {
			reg := internal.NewRegistry(t, append(tc.configOpts, configx.SkipValidation())...)

			if tc.setup != nil {
				tc.setup(t, reg.Config())
			}

			r := httptest.NewRequest("GET", "/test", nil)
			w := httptest.NewRecorder()
			r.Header = tc.header

			reg.ProxyRequestHandler().HandleError(w, r, tc.rule, tc.inputErr)
			tc.assert(t, w)
		})
	}
}

func TestRequestHandler(t *testing.T) {
	for k, tc := range []struct {
		d         string
		setup     func(t *testing.T, config configuration.Configuration)
		rule      rule.Rule
		r         *http.Request
		expectErr bool
	}{
		{
			d:         "should fail because the rule is missing authn, authz, and mutator",
			expectErr: true,
			r:         newTestRequest("http://localhost"),
			rule: rule.Rule{
				Authenticators: []rule.Handler{},
				Authorizer:     rule.Handler{},
				Mutators:       []rule.Handler{},
			},
		},
		{
			d: "should fail because the rule is missing authn, authz, and mutator even when some pipelines are enabled",
			setup: func(t *testing.T, config configuration.Configuration) {
				config.SetForTest(t, configuration.AuthenticatorNoopIsEnabled, true)
				config.SetForTest(t, configuration.AuthorizerAllowIsEnabled, true)
				config.SetForTest(t, configuration.MutatorNoopIsEnabled, true)
			},
			expectErr: true,
			r:         newTestRequest("http://localhost"),
			rule: rule.Rule{
				Authenticators: []rule.Handler{},
				Authorizer:     rule.Handler{},
				Mutators:       []rule.Handler{},
			},
		},
		{
			d: "should pass",
			setup: func(t *testing.T, config configuration.Configuration) {
				config.SetForTest(t, configuration.AuthenticatorNoopIsEnabled, true)
				config.SetForTest(t, configuration.AuthorizerAllowIsEnabled, true)
				config.SetForTest(t, configuration.MutatorNoopIsEnabled, true)
			},
			expectErr: false,
			r:         newTestRequest("http://localhost"),
			rule: rule.Rule{
				Authenticators: []rule.Handler{{Handler: "noop"}},
				Authorizer:     rule.Handler{Handler: "allow"},
				Mutators:       []rule.Handler{{Handler: "noop"}},
			},
		},
		{
			d: "should fail when authn is set but not authz nor mutator",
			setup: func(t *testing.T, config configuration.Configuration) {
				config.SetForTest(t, configuration.AuthenticatorAnonymousIsEnabled, true)
				config.SetForTest(t, configuration.AuthorizerAllowIsEnabled, true)
				config.SetForTest(t, configuration.MutatorNoopIsEnabled, true)
			},
			expectErr: true,
			r:         newTestRequest("http://localhost"),
			rule: rule.Rule{
				Authenticators: []rule.Handler{{Handler: "anonymous"}},
				Authorizer:     rule.Handler{},
				Mutators:       []rule.Handler{},
			},
		},
		{
			d: "should fail when authn, authz is set but not mutator",
			setup: func(t *testing.T, config configuration.Configuration) {
				config.SetForTest(t, configuration.AuthenticatorAnonymousIsEnabled, true)
				config.SetForTest(t, configuration.AuthorizerAllowIsEnabled, true)
				config.SetForTest(t, configuration.MutatorNoopIsEnabled, true)
			},
			expectErr: true,
			r:         newTestRequest("http://localhost"),
			rule: rule.Rule{
				Authenticators: []rule.Handler{{Handler: "anonymous"}},
				Authorizer:     rule.Handler{Handler: "allow"},
				Mutators:       []rule.Handler{},
			},
		},
		{
			d: "should fail when authn is invalid because not enabled",
			setup: func(t *testing.T, config configuration.Configuration) {
				config.SetForTest(t, configuration.AuthenticatorAnonymousIsEnabled, false)
				config.SetForTest(t, configuration.AuthorizerAllowIsEnabled, true)
				config.SetForTest(t, configuration.MutatorNoopIsEnabled, true)
			},
			expectErr: true,
			r:         newTestRequest("http://localhost"),
			rule: rule.Rule{
				Authenticators: []rule.Handler{{Handler: "anonymous"}},
				Authorizer:     rule.Handler{Handler: "allow"},
				Mutators:       []rule.Handler{{Handler: "noop"}},
			},
		},
		{
			d: "should fail when authz is invalid because not enabled",
			setup: func(t *testing.T, config configuration.Configuration) {
				config.SetForTest(t, configuration.AuthenticatorAnonymousIsEnabled, true)
				config.SetForTest(t, configuration.AuthorizerAllowIsEnabled, false)
				config.SetForTest(t, configuration.MutatorNoopIsEnabled, true)
			},
			expectErr: true,
			r:         newTestRequest("http://localhost"),
			rule: rule.Rule{
				Authenticators: []rule.Handler{{Handler: "anonymous"}},
				Authorizer:     rule.Handler{Handler: "allow"},
				Mutators:       []rule.Handler{{Handler: "noop"}},
			},
		},
		{
			d: "should fail when mutator is invalid because not enabled",
			setup: func(t *testing.T, config configuration.Configuration) {
				config.SetForTest(t, configuration.AuthenticatorAnonymousIsEnabled, true)
				config.SetForTest(t, configuration.AuthorizerAllowIsEnabled, true)
				config.SetForTest(t, configuration.MutatorNoopIsEnabled, false)
			},
			expectErr: true,
			r:         newTestRequest("http://localhost"),
			rule: rule.Rule{
				Authenticators: []rule.Handler{{Handler: "anonymous"}},
				Authorizer:     rule.Handler{Handler: "allow"},
				Mutators:       []rule.Handler{{Handler: "noop"}},
			},
		},
		{
			d: "should fail when authn does not exist",
			setup: func(t *testing.T, config configuration.Configuration) {
				config.SetForTest(t, configuration.AuthenticatorAnonymousIsEnabled, true)
				config.SetForTest(t, configuration.AuthorizerAllowIsEnabled, true)
				config.SetForTest(t, configuration.MutatorNoopIsEnabled, true)
			},
			expectErr: true,
			r:         newTestRequest("http://localhost"),
			rule: rule.Rule{
				Authenticators: []rule.Handler{{Handler: "invalid-id"}},
				Authorizer:     rule.Handler{Handler: "allow"},
				Mutators:       []rule.Handler{{Handler: "noop"}},
			},
		},
		{
			d: "should fail when authz does not exist",
			setup: func(t *testing.T, config configuration.Configuration) {
				config.SetForTest(t, configuration.AuthenticatorAnonymousIsEnabled, true)
				config.SetForTest(t, configuration.AuthorizerAllowIsEnabled, true)
				config.SetForTest(t, configuration.MutatorNoopIsEnabled, true)
			},
			expectErr: true,
			r:         newTestRequest("http://localhost"),
			rule: rule.Rule{
				Authenticators: []rule.Handler{{Handler: "anonymous"}},
				Authorizer:     rule.Handler{Handler: "invalid-id"},
				Mutators:       []rule.Handler{{Handler: "noop"}},
			},
		},
		{
			d: "should fail when mutator does not exist",
			setup: func(t *testing.T, config configuration.Configuration) {
				config.SetForTest(t, configuration.AuthenticatorAnonymousIsEnabled, true)
				config.SetForTest(t, configuration.AuthorizerAllowIsEnabled, true)
				config.SetForTest(t, configuration.MutatorNoopIsEnabled, true)
			},
			expectErr: true,
			r:         newTestRequest("http://localhost"),
			rule: rule.Rule{
				Authenticators: []rule.Handler{{Handler: "anonymous"}},
				Authorizer:     rule.Handler{Handler: "allow"},
				Mutators:       []rule.Handler{{Handler: "invalid-id"}},
			},
		},
	} {
		t.Run(fmt.Sprintf("case=%d/description=%s", k, tc.d), func(t *testing.T) {
			reg := internal.NewRegistry(t)

			if tc.setup != nil {
				tc.setup(t, reg.Config())
			}

			_, err := reg.ProxyRequestHandler().HandleRequest(tc.r, &tc.rule)
			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestInitializeSession(t *testing.T) {
	for _, tc := range []struct {
		desc          string
		match         rule.Match
		strategy      configuration.MatchingStrategy
		url           string
		expectContext authn.MatchContext
	}{
		{
			desc:     "Rule without capture",
			url:      "http://localhost",
			strategy: configuration.Regexp,
			match: rule.Match{
				URL: "http://localhost",
			},
			expectContext: authn.MatchContext{
				RegexpCaptureGroups: []string{},
				URL:                 x.ParseURLOrPanic("http://localhost"),
				Method:              "GET",
				Header:              TestHeader,
			},
		},
		{
			desc:     "Rule with one capture",
			url:      "http://localhost/user",
			strategy: configuration.Regexp,
			match: rule.Match{
				URL: "http://localhost/<.*>",
			},
			expectContext: authn.MatchContext{
				RegexpCaptureGroups: []string{"user"},
				URL:                 x.ParseURLOrPanic("http://localhost/user"),
				Method:              "GET",
				Header:              TestHeader,
			},
		},
		{
			desc:     "Request with query params",
			url:      "http://localhost/user?param=test",
			strategy: configuration.Regexp,
			match: rule.Match{
				URL: "http://localhost/<.*>",
			},
			expectContext: authn.MatchContext{
				RegexpCaptureGroups: []string{"user"},
				URL:                 x.ParseURLOrPanic("http://localhost/user?param=test"),
				Method:              "GET",
				Header:              TestHeader,
			},
		},
		{
			desc:     "Rule with 2 captures",
			url:      "http://localhost/user?param=test",
			strategy: configuration.Regexp,
			match: rule.Match{
				URL: "<http|https>://localhost/<.*>",
			},
			expectContext: authn.MatchContext{
				RegexpCaptureGroups: []string{"http", "user"},
				URL:                 x.ParseURLOrPanic("http://localhost/user?param=test"),
				Method:              "GET",
				Header:              TestHeader,
			},
		},
		{
			desc:     "Rule with 1 capture and path traversal",
			url:      "http://localhost/user/../admin/secrets?param=test",
			strategy: configuration.Regexp,
			match: rule.Match{
				URL: "http://localhost/admin/<.*>",
			},
			expectContext: authn.MatchContext{
				RegexpCaptureGroups: []string{"secrets"},
				URL:                 x.ParseURLOrPanic("http://localhost/admin/secrets?param=test"),
				Method:              "GET",
				Header:              TestHeader,
			},
		},
		{
			desc:     "Rule with Glob matching strategy",
			url:      "http://localhost/user?param=test",
			strategy: configuration.Glob,
			match: rule.Match{
				URL: "<http|https>://localhost/<*>",
			},
			expectContext: authn.MatchContext{
				RegexpCaptureGroups: []string{},
				URL:                 x.ParseURLOrPanic("http://localhost/user?param=test"),
				Method:              "GET",
				Header:              TestHeader,
			},
		},
	} {
		t.Run(fmt.Sprintf("description=%s", tc.desc), func(t *testing.T) {
			reg := internal.NewRegistry(t, configx.WithValue(configuration.AccessRuleMatchingStrategy, tc.strategy))

			session := reg.ProxyRequestHandler().InitializeAuthnSession(
				newTestRequest(tc.url),
				&rule.Rule{Match: &tc.match},
			)

			require.NotNil(t, session)
			require.NotNil(t, session.MatchContext.Header)
			assert.EqualValues(t, tc.expectContext, session.MatchContext)
		})
	}
}

func TestHandleRequestLoggingFields(t *testing.T) {
	conf := internal.NewConfigurationWithDefaults(t)

	var buf bytes.Buffer
	l := logrusx.NewT(t, logrusx.ForceFormat("json"), logrusx.ForceLevel(logrus.DebugLevel))
	l.Logrus().Out = &buf

	reg := driver.NewRegistryMemory(conf, l)

	traceID, err := trace.TraceIDFromHex("4bf92f3577b34da6a3ce929d0e0e4736")
	require.NoError(t, err)
	spanID, err := trace.SpanIDFromHex("00f067aa0ba902b7")
	require.NoError(t, err)
	spanCtx := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    traceID,
		SpanID:     spanID,
		TraceFlags: trace.FlagsSampled,
	})
	ctx := trace.ContextWithSpanContext(context.Background(), spanCtx)

	// Trigger the "no authentication handler was set" warning to produce a log line
	// containing the rule_id and otel trace fields.
	r := newTestRequest("http://localhost").WithContext(ctx)
	rl := &rule.Rule{ID: "my-rule-id"}
	_, err = reg.ProxyRequestHandler().HandleRequest(r, rl)
	require.Error(t, err)

	out := buf.String()
	assert.Contains(t, out, `"rule_id":"my-rule-id"`, out)
	assert.Contains(t, out, `"trace_id":"4bf92f3577b34da6a3ce929d0e0e4736"`, out)
	assert.Contains(t, out, `"span_id":"00f067aa0ba902b7"`, out)
}
