// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

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
	"github.com/ory/oathkeeper/x"
	"github.com/ory/x/configx"

	"github.com/golang-jwt/jwt/v5"

	"github.com/ory/oathkeeper/credentials"
	"github.com/ory/oathkeeper/driver/configuration"
	"github.com/ory/oathkeeper/internal"
	"github.com/ory/oathkeeper/pipeline/authn"
	. "github.com/ory/oathkeeper/pipeline/mutate"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type idTokenTestCase struct {
	Rule    *rule.Rule
	Session *authn.AuthenticationSession
	Config  json.RawMessage
	Match   jwt.MapClaims
	K       string
	Ttl     time.Duration
	Err     error
}

var idTokenTestCases = []idTokenTestCase{
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
	{
		Rule: &rule.Rule{ID: "test-rule10"},
		Session: &authn.AuthenticationSession{
			Subject: "foo",
			Extra:   map[string]interface{}{"abc": "value1", "def": "value2"},
			MatchContext: authn.MatchContext{
				RegexpCaptureGroups: []string{"user", "pass"},
			}},
		Config: json.RawMessage([]byte(`{"claims": "{\"custom-claim\": \"{{ print .Extra.abc }}\", \"custom-claim2\": \"{{ printIndex .MatchContext.RegexpCaptureGroups 1}}\", \"aud\": [\"foo\", \"bar\"]}"}`)),
		Match: jwt.MapClaims{
			"custom-claim":  "value1",
			"custom-claim2": "pass",
		},
		K: "file://../../test/stub/jwks-ecdsa.json",
	},
	{
		Rule: &rule.Rule{ID: "test-rule11"},
		Session: &authn.AuthenticationSession{
			Subject: "foo",
			Extra:   map[string]interface{}{"abc": "value1", "def": "value2"},
			MatchContext: authn.MatchContext{
				RegexpCaptureGroups: []string{"user"},
			}},
		Config: json.RawMessage([]byte(`{"claims": "{\"custom-claim\": \"{{ print .Extra.abc }}\", \"custom-claim2\": \"{{ printIndex .MatchContext.RegexpCaptureGroups 1}}\", \"aud\": [\"foo\", \"bar\"]}"}`)),
		Match: jwt.MapClaims{
			"custom-claim":  "value1",
			"custom-claim2": "",
		},
		K: "file://../../test/stub/jwks-ecdsa.json",
	},
	{
		Rule: &rule.Rule{ID: "test-rule12"},
		Session: &authn.AuthenticationSession{
			Subject: "foo",
			Extra:   map[string]interface{}{"abc": "value1", "def": "value2"},
			MatchContext: authn.MatchContext{
				RegexpCaptureGroups: []string{},
			}},
		Config: json.RawMessage([]byte(`{"claims": "{\"custom-claim\": \"{{ print .Extra.abc }}\", \"custom-claim2\": \"{{ printIndex .MatchContext.RegexpCaptureGroups 0}}\", \"aud\": [\"foo\", \"bar\"]}"}`)),
		Match: jwt.MapClaims{
			"custom-claim":  "value1",
			"custom-claim2": "",
		},
		K: "file://../../test/stub/jwks-ecdsa.json",
	},
	{
		Rule:    &rule.Rule{ID: "test-rule13"},
		Session: &authn.AuthenticationSession{Subject: "foo"},
		Config:  json.RawMessage([]byte(`{"claims": "{\"custom-claim\": \"{{ print .Subject }}\", \"aud\": [\"foo\", \"bar\"]}"}`)),
		Match:   jwt.MapClaims{"custom-claim": "foo"},
		Ttl:     30 * time.Second,
		K:       "file://../../test/stub/jwks-hs.json",
	},
}

func parseToken(h http.Header) string {
	return strings.Replace(h.Get("Authorization"), "Bearer ", "", 1)
}

func TestMutatorIDToken(t *testing.T) {
	t.Parallel()
	conf := internal.NewConfigurationWithDefaults(configx.SkipValidation())
	reg := internal.NewRegistry(conf)

	a, err := reg.PipelineMutator("id_token")
	require.NoError(t, err)
	assert.Equal(t, "id_token", a.GetID())

	conf.SetForTest(t, "mutators.id_token.config.issuer_url", "/foo/bar")

	t.Run("method=mutate", func(t *testing.T) {
		r := &http.Request{}
		t.Run("case=token generation and validation", func(t *testing.T) {
			for i, tc := range idTokenTestCases {
				t.Run(fmt.Sprintf("case=%d", i), func(t *testing.T) {
					tc.Config, _ = sjson.SetBytes(tc.Config, "jwks_url", tc.K)
					if tc.Ttl > 0 {
						tc.Config, _ = sjson.SetBytes(tc.Config, "ttl", tc.Ttl.String())
					}
					err := a.Mutate(r, tc.Session, tc.Config, tc.Rule)
					if tc.Err != nil {
						assert.EqualError(t, err, tc.Err.Error())
						return
					}
					require.NoError(t, err)

					token := parseToken(tc.Session.Header)
					result, err := reg.CredentialsVerifier().Verify(context.Background(), token, &credentials.ValidationContext{
						Algorithms: []string{"RS256", "HS256", "ES256"},
						Audiences:  []string{"foo", "bar"},
						KeyURLs:    []url.URL{*x.ParseURLOrPanic(tc.K)},
					})
					require.NoError(t, err, "token: %s", token)

					ttl := 15 * time.Minute // default from config is 15 minutes
					if tc.Ttl > 0 {
						ttl = tc.Ttl
					}
					assert.Equal(t, "foo", fmt.Sprintf("%s", result.Claims.(jwt.MapClaims)["sub"]))
					assert.Equal(t, "/foo/bar", fmt.Sprintf("%s", result.Claims.(jwt.MapClaims)["iss"]))
					assert.True(t, time.Now().Add(ttl).Unix() >= int64(result.Claims.(jwt.MapClaims)["exp"].(float64)))

					for key, val := range tc.Match {
						assert.Equal(t, val, result.Claims.(jwt.MapClaims)[key])
					}
				})
			}
		})

		t.Run("case=test token cache", func(t *testing.T) {
			mutate := func(t *testing.T, session authn.AuthenticationSession, config json.RawMessage) string {
				require.NoError(t, a.Mutate(new(http.Request), &session, config, &rule.Rule{ID: "1"}))
				return parseToken(session.Header)
			}

			session := &authn.AuthenticationSession{Subject: "foo", Extra: map[string]interface{}{"bar": "baz"}}
			config := json.RawMessage(`{"ttl": "100ms", "claims": "{\"foo\": \"{{ print .Extra.bar }}\", \"aud\": [\"foo\"]}", "jwks_url": "file://../../test/stub/jwks-ecdsa.json"}`)

			t.Run("subcase=different tokens because expired", func(t *testing.T) {
				config, _ := sjson.SetBytes(config, "ttl", "1ms")
				prev := mutate(t, *session, config)
				time.Sleep(time.Millisecond)
				assert.NotEqual(t, prev, mutate(t, *session, config))
			})

			t.Run("subcase=same tokens because expired is long enough", func(t *testing.T) {
				prev := mutate(t, *session, config)
				time.Sleep(10 * time.Millisecond) // give the cache buffers some time
				assert.Equal(t, prev, mutate(t, *session, config))
			})

			t.Run("subcase=different tokens because expired is long but was reached", func(t *testing.T) {
				prev := mutate(t, *session, config)
				time.Sleep(150 * time.Millisecond) // give the cache buffers some time
				assert.NotEqual(t, prev, mutate(t, *session, config))
			})

			t.Run("subcase=different tokens because different subjects", func(t *testing.T) {
				prev := mutate(t, *session, config)
				s := *session
				s.Subject = "not-foo"
				assert.NotEqual(t, prev, mutate(t, s, config))
			})

			t.Run("subcase=different tokens because session extra changed", func(t *testing.T) {
				prev := mutate(t, *session, config)
				s := *session
				s.Extra = map[string]interface{}{"bar": "not-baz"}
				assert.NotEqual(t, prev, mutate(t, s, config))
			})

			t.Run("subcase=different tokens because claim options changed", func(t *testing.T) {
				prev := mutate(t, *session, config)
				config := json.RawMessage(`{"ttl": "3s", "claims": "{\"foo\": \"{{ print .Extra.bar }}\", \"aud\": [\"not-foo\"]}", "jwks_url": "file://../../test/stub/jwks-ecdsa.json"}`)
				assert.NotEqual(t, prev, mutate(t, *session, config))
			})

			t.Run("subcase=same tokens because session extra changed but claims ignore the extra claims", func(t *testing.T) {
				t.Skip("Skipped because cache hit rate is too low, see: https://github.com/ory/oathkeeper/issues/371")

				prev := mutate(t, *session, config)
				time.Sleep(time.Second)
				s := *session
				s.Extra = map[string]interface{}{"bar": "baz", "not-bar": "whatever"}
				assert.Equal(t, prev, mutate(t, s, config))
			})

			t.Run("subcase=different tokens because issuer changed", func(t *testing.T) {
				prev := mutate(t, *session, config)
				config, _ := sjson.SetBytes(config, "issuer_url", "/not-baz/not-bar")
				assert.NotEqual(t, prev, mutate(t, *session, config))
			})

			t.Run("subcase=different tokens because JWKS source changed", func(t *testing.T) {
				prev := mutate(t, *session, config)
				config, _ := sjson.SetBytes(config, "jwks_url", "file://../../test/stub/jwks-hs.json")
				assert.NotEqual(t, prev, mutate(t, *session, config))
			})
		})

		t.Run("case=ensure template cache", func(t *testing.T) {
			tc := idTokenTestCase{
				Rule:    &rule.Rule{ID: "test-rule"},
				Session: &authn.AuthenticationSession{Subject: "foo", Extra: map[string]interface{}{"abc": "value1", "def": "value2"}},
				Config:  json.RawMessage([]byte(`{"claims": "{\"custom-claim\": \"{{ print .Extra.abc }}/{{ print .Extra.def }}\", \"aud\": [\"foo\", \"bar\"]}", "jwks_url": "file://../../test/stub/jwks-ecdsa.json"}`)),
				K:       "file://../../test/stub/jwks-ecdsa.json",
			}
			cache := template.New("rules")

			var cfg CredentialsIDTokenConfig
			require.NoError(t, json.NewDecoder(bytes.NewBuffer(tc.Config)).Decode(&cfg))

			_, err := cache.New(cfg.ClaimsTemplateID()).Parse(`{"custom-claim": "override", "aud": ["override"]}`)
			require.NoError(t, err)

			a.(*MutatorIDToken).WithCache(cache)
			require.NoError(t, a.Mutate(r, tc.Session, tc.Config, tc.Rule))

			token := strings.Replace(tc.Session.Header.Get("Authorization"), "Bearer ", "", 1)
			result, err := reg.CredentialsVerifier().Verify(context.Background(), token, &credentials.ValidationContext{
				Algorithms: []string{"RS256", "HS256", "ES256"},
				Audiences:  []string{"override"},
				KeyURLs:    []url.URL{*x.ParseURLOrPanic(tc.K)},
			})
			require.NoError(t, err, "token: %s (%s) %v", token, cfg.ClaimsTemplateID(), cache.Lookup(cfg.ClaimsTemplateID()))

			assert.Equal(t, "foo", fmt.Sprintf("%s", result.Claims.(jwt.MapClaims)["sub"]))
			assert.Equal(t, "/foo/bar", fmt.Sprintf("%s", result.Claims.(jwt.MapClaims)["iss"]))
			assert.Equal(t, "override", fmt.Sprintf("%s", result.Claims.(jwt.MapClaims)["custom-claim"]))
			assert.Equal(t, "override", fmt.Sprintf("%s", result.Claims.(jwt.MapClaims)["aud"].([]interface{})[0]))
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
				conf.SetForTest(t, configuration.MutatorIDTokenIsEnabled, tc.e)
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

func BenchmarkMutatorIDToken(b *testing.B) {
	issuers := []string{"foo", "bar", "baz", "zab"}

	conf := internal.NewConfigurationWithDefaults()
	reg := internal.NewRegistry(conf)
	rl := &rule.Rule{ID: "test-rule"}
	r := &http.Request{}

	var tcs []idTokenTestCase
	for _, tc := range idTokenTestCases {
		if tc.Err == nil {
			tcs = append(tcs, tc)
		}
	}

	a, err := reg.PipelineMutator("id_token")
	require.NoError(b, err)

	for alg, key := range map[string]string{
		"RS256": "file://../../test/stub/jwks-rsa-multiple.json",
		"HS256": "file://../../test/stub/jwks-hs.json",
		"ES256": "file://../../test/stub/jwks-ecdsa.json",
	} {
		b.Run("alg="+alg, func(b *testing.B) {
			for _, enableCache := range []bool{true, false} {
				a.(*MutatorIDToken).SetCaching(enableCache)
				b.Run(fmt.Sprintf("cache=%v", enableCache), func(b *testing.B) {
					var tc idTokenTestCase
					var config []byte

					b.ResetTimer()
					for i := 0; i < b.N; i++ {
						tc = tcs[i%len(tcs)]
						config, _ = sjson.SetBytes(tc.Config, "jwks_url", key)
						conf.SetForTest(b, "mutators.id_token.config.issuer_url", "/"+issuers[i%len(issuers)])

						b.StartTimer()
						err := a.Mutate(r, tc.Session, config, rl)
						b.StopTimer()

						require.NoError(b, err)
					}
				})
			}
		})
	}
}
