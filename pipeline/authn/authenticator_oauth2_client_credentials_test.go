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

	"github.com/ory/herodot"
	"github.com/ory/oathkeeper/helper"
)

func TestAuthenticatorOAuth2ClientCredentials(t *testing.T) {
	conf := internal.NewConfigurationWithDefaults()
	reg := internal.NewRegistry(conf)

	a, err := reg.PipelineAuthenticator("oauth2_client_credentials")
	require.NoError(t, err)
	assert.Equal(t, "oauth2_client_credentials", a.GetID())

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

	authOk := &http.Request{Header: http.Header{}}
	authOk.SetBasicAuth("client", "secret")

	authInvalid := &http.Request{Header: http.Header{}}
	authInvalid.SetBasicAuth("foo", "bar")

	for k, tc := range []struct {
		r             *http.Request
		config        json.RawMessage
		expectErr     error
		expectSession *authn.AuthenticationSession
	}{
		{
			r:         &http.Request{Header: http.Header{}},
			expectErr: authn.ErrAuthenticatorNotResponsible,
			config:    json.RawMessage(`{"token_url":"http://foo"}`),
		},
		{
			r:         authInvalid,
			expectErr: helper.ErrUnauthorized,
			config:    json.RawMessage(`{"token_url":"` + ts.URL + "/oauth2/token" + `"}`),
		},
		{
			r:             authOk,
			expectErr:     nil,
			expectSession: &authn.AuthenticationSession{Subject: "client"},
			config:        json.RawMessage(`{"token_url":"` + ts.URL + "/oauth2/token" + `"}`),
		},
	} {
		t.Run(fmt.Sprintf("method=authenticate/case=%d", k), func(t *testing.T) {
			session, err := a.Authenticate(tc.r, tc.config, nil)

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

	t.Run("method=validate", func(t *testing.T) {
		viper.Set(configuration.ViperKeyAuthenticatorOAuth2ClientCredentialsIsEnabled, false)
		require.Error(t, a.Validate(json.RawMessage(`{"token_url":""}`)))

		viper.Set(configuration.ViperKeyAuthenticatorOAuth2ClientCredentialsIsEnabled, false)
		require.Error(t, a.Validate(json.RawMessage(`{"token_url":"`+ts.URL+"/oauth2/token"+`"}`)))

		viper.Set(configuration.ViperKeyAuthenticatorOAuth2ClientCredentialsIsEnabled, true)
		require.Error(t, a.Validate(json.RawMessage(`{"token_url":""}`)))

		viper.Set(configuration.ViperKeyAuthenticatorOAuth2ClientCredentialsIsEnabled, true)
		require.NoError(t, a.Validate(json.RawMessage(`{"token_url":"`+ts.URL+"/oauth2/token"+`"}`)))
	})
}
