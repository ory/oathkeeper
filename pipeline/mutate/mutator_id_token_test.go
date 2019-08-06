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

package mutate_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"

	"github.com/ory/viper"

	"github.com/ory/oathkeeper/credentials"
	"github.com/ory/oathkeeper/driver/configuration"
	"github.com/ory/oathkeeper/internal"
	"github.com/ory/oathkeeper/pipeline/authn"
	"github.com/ory/x/urlx"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMutatorIDToken(t *testing.T) {
	conf := internal.NewConfigurationWithDefaults()
	reg := internal.NewRegistry(conf)

	a, err := reg.PipelineMutator("id_token")
	require.NoError(t, err)
	assert.Equal(t, "id_token", a.GetID())

	viper.Set(configuration.ViperKeyMutatorIDTokenIssuerURL, "/foo/bar")

	t.Run("method=mutate", func(t *testing.T) {
		for k, tc := range []struct {
			k   string
			ttl time.Duration
		}{
			{k: "file://../../test/stub/jwks-hs.json"},
			{k: "file://../../test/stub/jwks-rsa-multiple.json"},
			{k: "file://../../test/stub/jwks-ecdsa.json"},
		} {
			t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
				viper.Set(configuration.ViperKeyMutatorIDTokenJWKSURL, tc.k)
				viper.Set(configuration.ViperKeyMutatorIDTokenTTL, tc.ttl)

				r := &http.Request{Header: http.Header{}}
				s := &authn.AuthenticationSession{Subject: "foo"}

				err := a.Mutate(r, s, json.RawMessage([]byte(`{ "aud": ["foo", "bar"] }`)), nil)
				require.NoError(t, err)
				token := strings.Replace(s.Header.Get("Authorization"), "Bearer ", "", 1)

				result, err := reg.CredentialsVerifier().Verify(context.Background(), token, &credentials.ValidationContext{
					Algorithms: []string{"RS256", "HS256", "ES256"},
					Audiences:  []string{"foo", "bar"},
					KeyURLs:    []url.URL{*urlx.ParseOrPanic(tc.k)},
				})
				require.NoError(t, err)

				ttl := time.Minute // default from config is time.Minute
				if tc.ttl > 0 {
					ttl = tc.ttl
				}
				assert.Equal(t, "foo", fmt.Sprintf("%s", result.Claims.(jwt.MapClaims)["sub"]))
				assert.True(t, time.Now().Add(ttl).Unix() >= int64(result.Claims.(jwt.MapClaims)["exp"].(float64)))
			})
		}
	})

	t.Run("method=validate", func(t *testing.T) {
		for k, tc := range []struct {
			e    bool
			i    string
			j    string
			pass bool
		}{
			{e: false, pass: false},
			{e: true, pass: false},
			{e: true, i: "/foo", pass: false},
			{e: true, j: "/foo", pass: false},
			{e: true, i: "/foo", j: "/foo", pass: true},
		} {
			t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
				viper.Set(configuration.ViperKeyMutatorIDTokenIsEnabled, tc.e)
				viper.Set(configuration.ViperKeyMutatorIDTokenIssuerURL, tc.i)
				viper.Set(configuration.ViperKeyMutatorIDTokenJWKSURL, tc.j)
				if tc.pass {
					require.NoError(t, a.Validate())
				} else {
					require.Error(t, a.Validate())
				}
			})
		}
	})
}
