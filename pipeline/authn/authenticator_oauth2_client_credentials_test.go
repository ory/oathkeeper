// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package authn_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ory/oathkeeper/driver/configuration"
	"github.com/ory/oathkeeper/internal"
	"github.com/ory/oathkeeper/pipeline/authn"
	"github.com/ory/x/configx"

	"github.com/julienschmidt/httprouter"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/sjson"

	"github.com/ory/herodot"
	"github.com/ory/oathkeeper/helper"
)

func authOkDynamic(u string) *http.Request {
	authOk := &http.Request{Header: http.Header{}}
	authOk.SetBasicAuth(u, "secret")

	return authOk
}

func TestAuthenticatorOAuth2ClientCredentials(t *testing.T) {
	t.Parallel()
	conf := internal.NewConfigurationWithDefaults(configx.SkipValidation())
	reg := internal.NewRegistry(conf)

	a, err := reg.PipelineAuthenticator("oauth2_client_credentials")
	require.NoError(t, err)
	assert.Equal(t, "oauth2_client_credentials", a.GetID())

	authOk := &http.Request{Header: http.Header{}}
	authOk.SetBasicAuth("client", "secret")

	authInvalid := &http.Request{Header: http.Header{}}
	authInvalid.SetBasicAuth("foo", "bar")

	upstreamFailure := &http.Request{Header: http.Header{}}
	upstreamFailure.SetBasicAuth("client", "secret")

	calls := 0
	var logger herodot.ErrorReporter = nil
	for k, tc := range []struct {
		d             string
		r             *http.Request
		config        json.RawMessage
		token_url     string
		setup         func(*testing.T, *httprouter.Router, json.RawMessage)
		check         func(*testing.T, *httprouter.Router, json.RawMessage)
		expectErr     error
		expectSession *authn.AuthenticationSession
	}{
		{
			d:         "fails due to invalid token url",
			r:         &http.Request{Header: http.Header{}},
			expectErr: authn.ErrAuthenticatorNotResponsible,
			config:    json.RawMessage(`{}`),
			token_url: "http://foo",
		},
		{
			d:         "fails due to invalid client credentials",
			r:         authInvalid,
			expectErr: helper.ErrUnauthorized,
			config:    json.RawMessage(`{}`),
			token_url: "",
			setup: func(t *testing.T, h *httprouter.Router, _ json.RawMessage) {
				h.POST("/oauth2/token", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
					h := herodot.NewJSONWriter(logger)
					h.WriteError(w, r, helper.ErrUnauthorized)
				})
			},
		},
		{
			d:             "passes due to valid client credentials and returns access token",
			r:             authOk,
			expectErr:     nil,
			expectSession: &authn.AuthenticationSession{Subject: "client"},
			config:        json.RawMessage(`{}`),
			token_url:     "",
			setup: func(t *testing.T, h *httprouter.Router, _ json.RawMessage) {
				h.POST("/oauth2/token", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
					h := herodot.NewJSONWriter(logger)
					h.Write(w, r, map[string]interface{}{"access_token": "foo-token", "expires_in": 3600})
				})
			},
		},
		{
			d:             "passes due to enabled cache",
			r:             authOkDynamic("cache-case-1"),
			expectErr:     nil,
			expectSession: &authn.AuthenticationSession{Subject: "cache-case-1"},
			config:        json.RawMessage(`{ "cache": { "enabled": true } }`),
			token_url:     "",
			setup: func(t *testing.T, h *httprouter.Router, c json.RawMessage) {
				calls := 0
				h.POST("/oauth2/token", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
					calls++
					if calls == 2 {
						h := herodot.NewJSONWriter(logger)
						h.WriteError(w, r, helper.ErrUpstreamServiceNotAvailable)
						return
					}

					h := herodot.NewJSONWriter(logger)
					h.Write(w, r, map[string]interface{}{"access_token": "foo-token", "expires_in": 3600})
				})

				session := new(authn.AuthenticationSession)
				err := a.Authenticate(authOkDynamic("cache-case-1"), session, c, nil)

				require.NoError(t, err)

				// wait cache to save value
				time.Sleep(time.Millisecond * 10)
			},
		},
		{
			d:             "passes due to enabled cache with expired cache",
			r:             authOkDynamic("cache-case-2"),
			expectErr:     nil,
			expectSession: &authn.AuthenticationSession{Subject: "cache-case-2"},
			config:        json.RawMessage(`{ "cache": { "enabled": true } }`),
			token_url:     "",
			setup: func(t *testing.T, h *httprouter.Router, c json.RawMessage) {
				h.POST("/oauth2/token", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
					h := herodot.NewJSONWriter(logger)
					h.Write(w, r, map[string]interface{}{"access_token": "foo-token", "expires_in": 1})
				})

				session := new(authn.AuthenticationSession)
				err := a.Authenticate(authOkDynamic("cache-case-2"), session, c, nil)

				require.NoError(t, err)

				// wait for cache to expire
				time.Sleep(time.Second * 1)
			},
		},
		{
			d:             "passes due to enabled cache with no expiry",
			r:             authOkDynamic("cache-case-3"),
			expectErr:     nil,
			expectSession: &authn.AuthenticationSession{Subject: "cache-case-3"},
			config:        json.RawMessage(`{ "cache": { "enabled": true } }`),
			token_url:     "",
			setup: func(t *testing.T, h *httprouter.Router, c json.RawMessage) {
				calls := 0
				h.POST("/oauth2/token", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
					calls++
					if calls == 2 {
						h := herodot.NewJSONWriter(logger)
						h.WriteError(w, r, helper.ErrUpstreamServiceNotAvailable)
						return
					}

					h := herodot.NewJSONWriter(logger)
					h.Write(w, r, map[string]interface{}{"access_token": "foo-token", "expires_in": 3600})
				})

				session := new(authn.AuthenticationSession)
				err := a.Authenticate(authOkDynamic("cache-case-3"), session, c, nil)

				require.NoError(t, err)

				// wait cache to save value
				time.Sleep(time.Millisecond * 10)
			},
		},
		{
			d:             "passes with no shared cache between different token URLs",
			r:             authOkDynamic("cache-case-4"),
			expectErr:     nil,
			expectSession: &authn.AuthenticationSession{Subject: "cache-case-4"},
			config:        json.RawMessage(`{ "cache": { "enabled": true } }`),
			token_url:     "",
			setup: func(t *testing.T, h *httprouter.Router, c json.RawMessage) {
				calls = 0
				h.POST("/oauth2/token", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
					calls++
					if calls == 3 {
						h := herodot.NewJSONWriter(logger)
						t.Errorf("expected only 2 calls to token endpoint this is number %d", calls)
						h.WriteError(w, r, helper.ErrUpstreamServiceNotAvailable)
						return
					}

					h := herodot.NewJSONWriter(logger)
					h.Write(w, r, map[string]interface{}{"access_token": "foo-token", "expires_in": 3600})
				})

				ts := httptest.NewServer(h)

				session := new(authn.AuthenticationSession)
				// First request
				err = a.Authenticate(authOkDynamic("cache-case-4"), session, json.RawMessage(`{ "token_url": "`+ts.URL+`/oauth2/token", "cache": { "enabled": true } }`), nil)

				// wait cache to save value
				time.Sleep(time.Millisecond * 10)

				// Second request to test caching
				err = a.Authenticate(authOkDynamic("cache-case-4"), session, json.RawMessage(`{ "token_url": "`+ts.URL+`/oauth2/token", "cache": { "enabled": true } }`), nil)

				require.NoError(t, err)

				// wait cache to save value
				time.Sleep(time.Millisecond * 10)
			},
			check: func(t *testing.T, router *httprouter.Router, message json.RawMessage) {
				require.Equal(t, 2, calls, "expected a call to the token endpoint per token URL config")
			},
		},
		{
			d:             "passes with no shared cache between different token URLs",
			r:             authOkDynamic("cache-case-5"),
			expectErr:     nil,
			expectSession: &authn.AuthenticationSession{Subject: "cache-case-5"},
			config:        json.RawMessage(`{ "cache": { "enabled": true } }`),
			token_url:     "",
			setup: func(t *testing.T, h *httprouter.Router, c json.RawMessage) {
				calls = 0
				h.POST("/oauth2/token", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
					calls++
					if calls == 3 {
						h := herodot.NewJSONWriter(logger)
						t.Errorf("expected only 2 calls to token endpoint this is number %d", calls)
						h.WriteError(w, r, helper.ErrUpstreamServiceNotAvailable)
						return
					}

					h := herodot.NewJSONWriter(logger)
					h.Write(w, r, map[string]interface{}{"access_token": "foo-token", "expires_in": 3600})
				})

				var authnConfig authn.AuthenticatorOAuth2Configuration
				json.Unmarshal(c, &authnConfig) //nolint:errcheck,gosec // test overrides config

				authnConfig.Scopes = []string{"some-scope"}
				authnConfig.Cache.TTL = "6h"
				scopeConfig, _ := json.Marshal(authnConfig)

				session := new(authn.AuthenticationSession)
				// First request
				err = a.Authenticate(authOkDynamic("cache-case-5"), session, scopeConfig, nil)

				// wait cache to save value
				time.Sleep(time.Millisecond * 10)

				// Second request to check caching
				err = a.Authenticate(authOkDynamic("cache-case-5"), session, scopeConfig, nil)

				require.NoError(t, err)

				// wait cache to save value
				time.Sleep(time.Millisecond * 10)
			},
			check: func(t *testing.T, router *httprouter.Router, message json.RawMessage) {
				require.Equal(t, 2, calls, "expected a call to the token endpoint per scope config")
			},
		},
		{
			d:         "fails and returns 503 Service Unavailable error due to the unavailability of the upstream service",
			r:         upstreamFailure,
			expectErr: helper.ErrUpstreamServiceNotAvailable,
			config:    json.RawMessage(`{}`),
			token_url: "",
			setup: func(t *testing.T, h *httprouter.Router, _ json.RawMessage) {
				h.POST("/oauth2/token", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
					h := herodot.NewJSONWriter(logger)
					h.WriteError(w, r, helper.ErrUpstreamServiceNotAvailable)
				})
			},
		},
		{
			d:         "fails and returns 504 Gateway Timeout error due to upstream service timeout",
			r:         upstreamFailure,
			expectErr: helper.ErrUpstreamServiceTimeout,
			config:    json.RawMessage(`{}`),
			token_url: "",
			setup: func(t *testing.T, h *httprouter.Router, _ json.RawMessage) {
				h.POST("/oauth2/token", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
					h := herodot.NewJSONWriter(logger)
					h.WriteError(w, r, helper.ErrUpstreamServiceTimeout)
				})
			},
		},
		{
			d:         "fails and returns 500 Internal Server Error error due to an unexpected error in the upstream service",
			r:         upstreamFailure,
			expectErr: helper.ErrUpstreamServiceInternalServerError,
			config:    json.RawMessage(`{}`),
			token_url: "",
			setup: func(t *testing.T, h *httprouter.Router, _ json.RawMessage) {
				h.POST("/oauth2/token", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
					h := herodot.NewJSONWriter(logger)
					h.WriteError(w, r, helper.ErrUpstreamServiceInternalServerError)
				})
			},
		},
		{
			d:         "fails and returns 404 Not Found error as the upstream service could not find the requested resource ",
			r:         upstreamFailure,
			expectErr: helper.ErrUpstreamServiceNotFound,
			config:    json.RawMessage(`{}`),
			token_url: "",
			setup: func(t *testing.T, h *httprouter.Router, _ json.RawMessage) {
				h.POST("/oauth2/v1/token", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
					h := herodot.NewJSONWriter(logger)
					h.Write(w, r, map[string]interface{}{"access_token": "foo-token"})
				})
			},
		},
		{
			d:         "fails and returns 401 Unauthorized error as the upstream service returns 403 Forbidden",
			r:         upstreamFailure,
			expectErr: helper.ErrUnauthorized,
			config:    json.RawMessage(`{}`),
			token_url: "",
			setup: func(t *testing.T, h *httprouter.Router, _ json.RawMessage) {
				h.POST("/oauth2/token", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
					h := herodot.NewJSONWriter(logger)
					h.WriteError(w, r, helper.ErrForbidden)
				})
			},
		},
	} {
		t.Run(fmt.Sprintf("method=authenticate/case=%d", k), func(t *testing.T) {
			router := httprouter.New()

			ts := httptest.NewServer(router)

			if tc.token_url != "" {
				tc.config, _ = sjson.SetBytes(tc.config, "token_url", tc.token_url)
			} else {
				tc.config, _ = sjson.SetBytes(tc.config, "token_url", ts.URL+"/oauth2/token")
			}

			if tc.setup != nil {
				tc.setup(t, router, tc.config)
			}

			session := new(authn.AuthenticationSession)
			err := a.Authenticate(tc.r, session, tc.config, nil)

			if tc.expectErr != nil {
				require.EqualError(t, errors.Cause(err), tc.expectErr.Error())
			} else {
				require.NoError(t, err)
			}

			if tc.expectSession != nil {
				assert.EqualValues(t, tc.expectSession, session)
			}

			if tc.check != nil {
				tc.check(t, router, tc.config)
			}
		})
	}

	h := httprouter.New()
	h.POST("/oauth2/token", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		h := herodot.NewJSONWriter(logger)
		u, p, ok := r.BasicAuth()
		if !ok || u != "client" || p != "secret" {
			h.WriteError(w, r, helper.ErrUnauthorized)
			return
		}
		h.Write(w, r, map[string]interface{}{"access_token": "foo-token"})
	})

	ts := httptest.NewServer(h)

	t.Run("method=validate", func(t *testing.T) {
		conf.SetForTest(t, configuration.AuthenticatorOAuth2ClientCredentialsIsEnabled, false)
		require.Error(t, a.Validate(json.RawMessage(`{"token_url":""}`)))

		conf.SetForTest(t, configuration.AuthenticatorOAuth2ClientCredentialsIsEnabled, false)
		require.Error(t, a.Validate(json.RawMessage(`{"token_url":"`+ts.URL+"/oauth2/token"+`"}`)))

		conf.SetForTest(t, configuration.AuthenticatorOAuth2ClientCredentialsIsEnabled, true)
		require.Error(t, a.Validate(json.RawMessage(`{"token_url":""}`)))

		conf.SetForTest(t, configuration.AuthenticatorOAuth2ClientCredentialsIsEnabled, true)
		require.NoError(t, a.Validate(json.RawMessage(`{"token_url":"`+ts.URL+"/oauth2/token"+`"}`)))

		conf.SetForTest(t, configuration.AuthenticatorOAuth2ClientCredentialsIsEnabled, true)
		require.NoError(t, a.Validate(json.RawMessage(`{"token_url":"`+ts.URL+"/oauth2/token"+`","retry":{"give_up_after":"3s", "max_delay":"100ms"}}`)))
	})
}
