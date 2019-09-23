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
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"text/template"
	"time"

	"github.com/tidwall/sjson"

	"github.com/ory/oathkeeper/rule"

	"github.com/dgrijalva/jwt-go"

	"github.com/ory/viper"

	"github.com/ory/x/urlx"

	"github.com/ory/oathkeeper/credentials"
	"github.com/ory/oathkeeper/driver/configuration"
	"github.com/ory/oathkeeper/internal"
	"github.com/ory/oathkeeper/pipeline/authn"
	. "github.com/ory/oathkeeper/pipeline/mutate"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testCase struct {
	Rule    *rule.Rule
	Session *authn.AuthenticationSession
	Config  json.RawMessage
	Match   jwt.MapClaims
	K       string
	Ttl     time.Duration
	Err     error
}

func TestMutatorIDToken(t *testing.T) {
	conf := internal.NewConfigurationWithDefaults()
	reg := internal.NewRegistry(conf)

	a, err := reg.PipelineMutator("id_token")
	require.NoError(t, err)
	assert.Equal(t, "id_token", a.GetID())

	viper.Set("mutators.id_token.config.issuer_url", "/foo/bar")

	t.Run("method=mutate", func(t *testing.T) {

		r := &http.Request{}

		t.Run("caching=off", func(t *testing.T) {
			var testCases = []testCase{
				{
					Rule:    &rule.Rule{ID: "test-rule1"},
					Session: &authn.AuthenticationSession{Subject: "foo"},
					Config:  json.RawMessage([]byte(`{"claims": "{\"custom-claim\": \"{{ print .Subject }}\", \"aud\": [\"foo\", \"bar\"]}"}`)),
					Match:   jwt.MapClaims{"custom-claim": "foo"},
					K:       "file://../../test/stub/jwks-hs.json",
				},
				{
					Rule:    &rule.Rule{ID: "test-rule2"},
					Session: &authn.AuthenticationSession{Subject: "foo", Extra: map[string]interface{}{"abc": "value1"}},
					Config:  json.RawMessage([]byte(`{"claims": "{\"custom-claim\": \"{{ print .Extra.abc }}\", \"aud\": [\"foo\", \"bar\"]}"}`)),
					Match:   jwt.MapClaims{"custom-claim": "value1"},
					K:       "file://../../test/stub/jwks-rsa-multiple.json",
				},
				{
					Rule:    &rule.Rule{ID: "test-rule3"},
					Session: &authn.AuthenticationSession{Subject: "foo", Extra: map[string]interface{}{"abc": "value1", "def": "value2"}},
					Config:  json.RawMessage([]byte(`{"claims": "{\"custom-claim\": \"{{ print .Extra.abc }}/{{ print .Extra.def }}\", \"aud\": [\"foo\", \"bar\"]}"}`)),
					Match:   jwt.MapClaims{"custom-claim": "value1/value2"},
					K:       "file://../../test/stub/jwks-ecdsa.json",
				},
				{
					Rule:    &rule.Rule{ID: "test-rule4"},
					Session: &authn.AuthenticationSession{Subject: "foo", Extra: map[string]interface{}{"abc": "value1", "def": "value2"}},
					Config:  json.RawMessage([]byte(`{"claims": "{\"custom-claim\": \"{{ print .Extra.abc }}\", \"custom-claim2\": \"{{ print .Extra.def }}\", \"aud\": [\"foo\", \"bar\"]}"}`)),
					Match:   jwt.MapClaims{"custom-claim": "value1", "custom-claim2": "value2"},
					K:       "file://../../test/stub/jwks-ecdsa.json",
				},
				{
					Rule:    &rule.Rule{ID: "test-rule5"},
					Session: &authn.AuthenticationSession{},
					Config:  json.RawMessage([]byte(`{"bad": "key"}`)),
					Match:   jwt.MapClaims{},
					K:       "file://../../test/stub/jwks-hs.json",
					Err:     errors.New(`mutator matching this route is misconfigured or disabled`),
				},
				{
					Rule:    &rule.Rule{ID: "test-rule6"},
					Session: &authn.AuthenticationSession{Subject: "foo", Extra: map[string]interface{}{}},
					Config:  json.RawMessage([]byte(`{"claims": "{\"custom-claim\": \"{{ print .Extra.def }}\", \"aud\": [\"foo\", \"bar\"]}"}`)),
					Match:   jwt.MapClaims{"custom-claim": ""},
					K:       "file://../../test/stub/jwks-rsa-multiple.json",
				},
				{
					Rule: &rule.Rule{ID: "test-rule7"},
					Session: &authn.AuthenticationSession{
						Subject: "foo",
						Extra: map[string]interface{}{
							"nested": map[string]interface{}{
								"int":     int(10),
								"float64": float64(3.14159),
								"bool":    true,
							},
						},
					},
					Config: json.RawMessage([]byte(`{"claims": "{\"custom-claim\": {{ print .Extra.nested.int }},\"custom-claim2\": {{ print .Extra.nested.float64 }},\"custom-claim3\": {{ print .Extra.nested.bool }},\"aud\": [\"foo\", \"bar\"]}"}`)),
					Match: jwt.MapClaims{
						"custom-claim":  float64(10), // the json decoder always converts to float64
						"custom-claim2": 3.14159,
						"custom-claim3": true,
					},
					K: "file://../../test/stub/jwks-ecdsa.json",
				},
				{
					Rule:    &rule.Rule{ID: "test-rule8"},
					Session: &authn.AuthenticationSession{Subject: "foo", Extra: map[string]interface{}{"example.com/some-claims": []string{"Foo", "Bar"}}},
					Config:  json.RawMessage([]byte(`{"claims":"{\"custom-claim\": \"{{- (index .Extra \"example.com/some-claims\") | join \",\" -}}\", \"aud\": [\"foo\", \"bar\"]}"}`)),
					Match:   jwt.MapClaims{"custom-claim": "Foo,Bar"},
					K:       "file://../../test/stub/jwks-hs.json",
				},
				{
					Rule:    &rule.Rule{ID: "test-rule9"},
					Session: &authn.AuthenticationSession{Subject: "foo", Extra: map[string]interface{}{"malicious": "evil"}},
					Config:  json.RawMessage([]byte(`{"claims": "{\"iss\": \"{{ print .Extra.malicious }}\", \"aud\": [\"foo\", \"bar\"]}"}`)),
					Match:   jwt.MapClaims{},
					K:       "file://../../test/stub/jwks-ecdsa.json",
				},
			}

			for i, tc := range testCases {
				t.Run(fmt.Sprintf("case=%d", i), func(t *testing.T) {
					tc.Config, _ = sjson.SetBytes(tc.Config, "jwks_url", tc.K)
					tc.Config, _ = sjson.SetBytes(tc.Config, "ttl", tc.Ttl.String())
					err := a.Mutate(r, tc.Session, tc.Config, tc.Rule)
					if tc.Err == nil {
						require.NoError(t, err)

						token := strings.Replace(tc.Session.Header.Get("Authorization"), "Bearer ", "", 1)

						result, err := reg.CredentialsVerifier().Verify(context.Background(), token, &credentials.ValidationContext{
							Algorithms: []string{"RS256", "HS256", "ES256"},
							Audiences:  []string{"foo", "bar"},
							KeyURLs:    []url.URL{*urlx.ParseOrPanic(tc.K)},
						})
						require.NoError(t, err, "token: %s", token)

						ttl := time.Minute // default from config is time.Minute
						if tc.Ttl > 0 {
							ttl = tc.Ttl
						}
						assert.Equal(t, "foo", fmt.Sprintf("%s", result.Claims.(jwt.MapClaims)["sub"]))
						assert.Equal(t, "/foo/bar", fmt.Sprintf("%s", result.Claims.(jwt.MapClaims)["iss"]))
						assert.True(t, time.Now().Add(ttl).Unix() >= int64(result.Claims.(jwt.MapClaims)["exp"].(float64)))

						for key, val := range tc.Match {
							assert.Equal(t, val, result.Claims.(jwt.MapClaims)[key])
						}

					} else {
						assert.EqualError(t, err, tc.Err.Error())
					}
				})
			}
		})

		t.Run("caching=on", func(t *testing.T) {

			tc := testCase{
				Rule:    &rule.Rule{ID: "test-rule"},
				Session: &authn.AuthenticationSession{Subject: "foo", Extra: map[string]interface{}{"abc": "value1", "def": "value2"}},
				Config:  json.RawMessage([]byte(`{"claims": "{\"custom-claim\": \"{{ print .Extra.abc }}/{{ print .Extra.def }}\", \"aud\": [\"foo\", \"bar\"]}", "jwks_url": "file://../../test/stub/jwks-ecdsa.json"}`)),
				K:       "file://../../test/stub/jwks-ecdsa.json",
			}

			// viper.Set(configuration.ViperKeyMutatorIDTokenJWKSURL, tc.K)
			// viper.Set(configuration.ViperKeyMutatorIDTokenTTL, tc.Ttl)

			cache := template.New("rules")

			var cfg CredentialsIDTokenConfig
			require.NoError(t, json.NewDecoder(bytes.NewBuffer(tc.Config)).Decode(&cfg))

			_, err := cache.New(cfg.ClaimsTemplateID()).Parse(`{"custom-claim": "override", "aud": ["override"]}`)
			require.NoError(t, err)

			a.(*MutatorIDToken).WithCache(cache)

			err = a.Mutate(r, tc.Session, tc.Config, tc.Rule)
			require.NoError(t, err)

			token := strings.Replace(tc.Session.Header.Get("Authorization"), "Bearer ", "", 1)

			result, err := reg.CredentialsVerifier().Verify(context.Background(), token, &credentials.ValidationContext{
				Algorithms: []string{"RS256", "HS256", "ES256"},
				Audiences:  []string{"override"},
				KeyURLs:    []url.URL{*urlx.ParseOrPanic(tc.K)},
			})
			require.NoError(t, err, "token: %s (%s) %v", token, cfg.ClaimsTemplateID(), cache.Lookup(cfg.ClaimsTemplateID()))

			ttl := time.Minute // default from config is time.Minute
			if tc.Ttl > 0 {
				ttl = tc.Ttl
			}
			assert.Equal(t, "foo", fmt.Sprintf("%s", result.Claims.(jwt.MapClaims)["sub"]))
			assert.Equal(t, "/foo/bar", fmt.Sprintf("%s", result.Claims.(jwt.MapClaims)["iss"]))
			assert.Equal(t, "override", fmt.Sprintf("%s", result.Claims.(jwt.MapClaims)["custom-claim"]))
			assert.Equal(t, "override", fmt.Sprintf("%s", result.Claims.(jwt.MapClaims)["aud"].([]interface{})[0]))
			assert.True(t, time.Now().Add(ttl).Unix() >= int64(result.Claims.(jwt.MapClaims)["exp"].(float64)))
		})
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
			{e: true, i: "http://baz/foo", pass: false},
			{e: true, j: "", pass: false},
			{e: true, i: "http://baz/foo", j: "http://baz/foo", pass: true},
		} {
			t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
				viper.Set(configuration.ViperKeyMutatorIDTokenIsEnabled, tc.e)
				// viper.Set(configuration.ViperKeyMutatorIDTokenIssuerURL, tc.i)
				// viper.Set(configuration.ViperKeyMutatorIDTokenJWKSURL, tc.j)
				err := a.Validate(json.RawMessage(`{"issuer_url":"` + tc.i + `", "jwks_url": "` + tc.j + `"}`))
				if tc.pass {
					require.NoError(t, err)
				} else {
					require.Error(t, err)
				}
			})
		}
	})
}
