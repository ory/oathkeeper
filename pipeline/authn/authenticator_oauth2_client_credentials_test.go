/*
 * Copyright © 2017-2018 Aeneas Rekkas <aeneas+oss@aeneas.io>
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * @author       Aeneas Rekkas <aeneas+oss@aeneas.io>
 * @copyright  2017-2018 Aeneas Rekkas <aeneas+oss@aeneas.io>
 * @license  	   Apache-2.0
 */

package authn_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ory/viper"

	"github.com/ory/oathkeeper/driver/configuration"
	"github.com/ory/oathkeeper/internal"
	"github.com/ory/oathkeeper/pipeline/authn"

	"github.com/julienschmidt/httprouter"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tidwall/sjson"

	"github.com/ory/herodot"
	"github.com/ory/oathkeeper/helper"
)

func TestAuthenticatorOAuth2ClientCredentials(t *testing.T) {
	conf := internal.NewConfigurationWithDefaults()
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

	for k, tc := range []struct {
		d             string
		r             *http.Request
		config        json.RawMessage
		token_url     string
		setup         func(*testing.T, *httprouter.Router)
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
			setup: func(t *testing.T, h *httprouter.Router) {
				h.POST("/oauth2/token", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
					h := herodot.NewJSONWriter(logrus.New())
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
			setup: func(t *testing.T, h *httprouter.Router) {
				h.POST("/oauth2/token", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
					h := herodot.NewJSONWriter(logrus.New())
					h.Write(w, r, map[string]interface{}{"access_token": "foo-token"})
				})
			},
		},
		{
			d:         "fails and returns 503 Service Unavailable error due to the unavailability of the upstream service",
			r:         upstreamFailure,
			expectErr: helper.ErrUpstreamServiceNotAvailable,
			config:    json.RawMessage(`{}`),
			token_url: "",
			setup: func(t *testing.T, h *httprouter.Router) {
				h.POST("/oauth2/token", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
					h := herodot.NewJSONWriter(logrus.New())
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
			setup: func(t *testing.T, h *httprouter.Router) {
				h.POST("/oauth2/token", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
					h := herodot.NewJSONWriter(logrus.New())
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
			setup: func(t *testing.T, h *httprouter.Router) {
				h.POST("/oauth2/token", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
					h := herodot.NewJSONWriter(logrus.New())
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
			setup: func(t *testing.T, h *httprouter.Router) {
				h.POST("/oauth2/v1/token", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
					h := herodot.NewJSONWriter(logrus.New())
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
			setup: func(t *testing.T, h *httprouter.Router) {
				h.POST("/oauth2/token", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
					h := herodot.NewJSONWriter(logrus.New())
					h.WriteError(w, r, helper.ErrForbidden)
				})
			},
		},
	} {
		t.Run(fmt.Sprintf("method=authenticate/case=%d", k), func(t *testing.T) {
			router := httprouter.New()

			if tc.setup != nil {
				tc.setup(t, router)
			}

			ts := httptest.NewServer(router)

			if tc.token_url != "" {
				tc.config, _ = sjson.SetBytes(tc.config, "token_url", tc.token_url)
			} else {
				tc.config, _ = sjson.SetBytes(tc.config, "token_url", ts.URL+"/oauth2/token")
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
		})
	}

	h := httprouter.New()
	h.POST("/oauth2/token", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		h := herodot.NewJSONWriter(logrus.New())
		u, p, ok := r.BasicAuth()
		if !ok || u != "client" || p != "secret" {
			h.WriteError(w, r, helper.ErrUnauthorized)
			return
		}
		h.Write(w, r, map[string]interface{}{"access_token": "foo-token"})
	})

	ts := httptest.NewServer(h)

	t.Run("method=validate", func(t *testing.T) {
		viper.Set(configuration.ViperKeyAuthenticatorOAuth2ClientCredentialsIsEnabled, false)
		require.Error(t, a.Validate(json.RawMessage(`{"token_url":""}`)))

		viper.Reset()
		viper.Set(configuration.ViperKeyAuthenticatorOAuth2ClientCredentialsIsEnabled, false)
		require.Error(t, a.Validate(json.RawMessage(`{"token_url":"`+ts.URL+"/oauth2/token"+`"}`)))

		viper.Reset()
		viper.Set(configuration.ViperKeyAuthenticatorOAuth2ClientCredentialsIsEnabled, true)
		require.Error(t, a.Validate(json.RawMessage(`{"token_url":""}`)))

		viper.Reset()
		viper.Set(configuration.ViperKeyAuthenticatorOAuth2ClientCredentialsIsEnabled, true)
		require.NoError(t, a.Validate(json.RawMessage(`{"token_url":"`+ts.URL+"/oauth2/token"+`"}`)))

		viper.Reset()
		viper.Set(configuration.ViperKeyAuthenticatorOAuth2ClientCredentialsIsEnabled, true)
		require.NoError(t, a.Validate(json.RawMessage(`{"token_url":"`+ts.URL+"/oauth2/token"+`","retry":{"give_up_after":"3s", "max_delay":"100ms"}}`)))
	})
}
