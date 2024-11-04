// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package mutate_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"text/template"

	"github.com/ory/oathkeeper/internal"

	"github.com/ory/oathkeeper/pipeline/authn"
	. "github.com/ory/oathkeeper/pipeline/mutate"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/oathkeeper/rule"
)

func TestCredentialsIssuerCookies(t *testing.T) {
	t.Parallel()
	conf := internal.NewConfigurationWithDefaults()
	reg := internal.NewRegistry(conf)

	a, err := reg.PipelineMutator("cookie")
	require.NoError(t, err)
	assert.Equal(t, "cookie", a.GetID())

	t.Run("method=mutate", func(t *testing.T) {
		var testMap = map[string]struct {
			Session *authn.AuthenticationSession
			Rule    *rule.Rule
			Config  json.RawMessage
			Request *http.Request
			Match   []*http.Cookie
			Err     error
		}{
			"Simple Subject": {
				Session: &authn.AuthenticationSession{Subject: "foo"},
				Rule:    &rule.Rule{ID: "test-rule"},
				Config:  json.RawMessage([]byte(`{"cookies": {"user": "{{ print .Subject }}"}}`)),
				Request: &http.Request{Header: http.Header{}},
				Match:   []*http.Cookie{{Name: "user", Value: "foo"}},
				Err:     nil,
			},
			"Unknown Config Field": {
				Session: &authn.AuthenticationSession{},
				Rule:    &rule.Rule{ID: "test-rule2"},
				Config:  json.RawMessage([]byte(`{"bar": "baz"}`)),
				Request: &http.Request{Header: http.Header{}},
				Match:   []*http.Cookie{},
				Err:     errors.New(`mutator matching this route is misconfigured or disabled`),
			},
			"Complex Subject": {
				Session: &authn.AuthenticationSession{Subject: "foo"},
				Rule:    &rule.Rule{ID: "test-rule3"},
				Config:  json.RawMessage([]byte(`{"cookies": {"user": "realm:resources:users:{{ print .Subject }}"}}`)),
				Request: &http.Request{Header: http.Header{}},
				Match:   []*http.Cookie{{Name: "user", Value: "realm:resources:users:foo"}},
				Err:     nil,
			},
			"Subject & Extras": {
				Session: &authn.AuthenticationSession{Subject: "foo", Extra: map[string]interface{}{"iss": "issuer", "aud": "audience"}},
				Rule:    &rule.Rule{ID: "test-rule4"},
				Config:  json.RawMessage([]byte(`{"cookies":{"user": "{{ print .Subject }}", "issuer": "{{ print .Extra.iss }}", "audience": "{{ print .Extra.aud }}"}}`)),
				Request: &http.Request{Header: http.Header{}},
				Match: []*http.Cookie{
					{Name: "user", Value: "foo"},
					{Name: "issuer", Value: "issuer"},
					{Name: "audience", Value: "audience"},
				},
				Err: nil,
			},
			"All In One Cookie": {
				Session: &authn.AuthenticationSession{Subject: "foo", Extra: map[string]interface{}{"iss": "issuer", "aud": "audience"}},
				Rule:    &rule.Rule{ID: "test-rule5"},
				Config:  json.RawMessage([]byte(`{"cookies":{"kitchensink": "{{ print .Subject }} {{ print .Extra.iss }} {{ print .Extra.aud }}"}}`)),
				Request: &http.Request{Header: http.Header{}},
				Match: []*http.Cookie{
					{Name: "kitchensink", Value: "foo issuer audience"},
				},
				Err: nil,
			},
			"Scrub Incoming Cookies": {
				Session: &authn.AuthenticationSession{Subject: "anonymous"},
				Rule:    &rule.Rule{ID: "test-rule6"},
				Config:  json.RawMessage([]byte(`{"cookies":{"user": "{{ print .Subject }}", "issuer": "{{ print .Extra.iss }}", "audience": "{{ print .Extra.aud }}"}}`)),
				Request: &http.Request{
					Header: http.Header{"Cookie": []string{"user=admin;issuer=issuer;audience=audience"}},
				},
				Match: []*http.Cookie{
					{Name: "user", Value: "anonymous"},
					{Name: "issuer", Value: ""},
					{Name: "audience", Value: ""},
				},
				Err: nil,
			},
			"Missing Extras": {
				Session: &authn.AuthenticationSession{Subject: "foo", Extra: map[string]interface{}{}},
				Rule:    &rule.Rule{ID: "test-rule7"},
				Config:  json.RawMessage([]byte(`{"cookies":{"issuer": "{{ print .Extra.iss }}"}}`)),
				Request: &http.Request{Header: http.Header{}},
				Match:   []*http.Cookie{{Name: "issuer", Value: ""}},
				Err:     nil,
			},
			"Nested Extras": {
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
				Rule: &rule.Rule{ID: "test-rule8"},
				Config: json.RawMessage([]byte(`{"cookies":{
				"nested-int": "{{ print .Extra.nested.int }}",
				"nested-float64": "{{ print .Extra.nested.float64 }}",
				"nested-bool": "{{ print .Extra.nested.bool}}",
				"nested-nonexistent": "{{ print .Extra.nested.nil }}"
			}}`)),
				Request: &http.Request{Header: http.Header{}},
				Match: []*http.Cookie{
					{Name: "nested-int", Value: "10"},
					{Name: "nested-float64", Value: "3.14159"},
					{Name: "nested-bool", Value: "true"},
					{Name: "nested-nonexistent", Value: ""},
				},
				Err: nil,
			},
			"Advanced template with sprig function": {
				Session: &authn.AuthenticationSession{
					Subject: "foo",
					Extra: map[string]interface{}{
						"example.com/some-claims": []string{"Foo", "Bar"},
					},
				},
				Rule: &rule.Rule{ID: "test-rule9"},
				Config: json.RawMessage([]byte(`{"cookies":{
					"example-claims": "{{- (index .Extra \"example.com/some-claims\") | join \",\" -}}"
				}}`)),
				Request: &http.Request{Header: http.Header{}},
				Match: []*http.Cookie{
					{
						Name:  "example-claims",
						Value: "Foo,Bar",
					},
				},
				Err: nil,
			},
		}

		t.Run("caching=off", func(t *testing.T) {
			for testName, specs := range testMap {
				t.Run(testName, func(t *testing.T) {
					err := a.Mutate(specs.Request, specs.Session, specs.Config, specs.Rule)
					if specs.Err == nil {
						// Issuer must run without error
						require.NoError(t, err)
					} else {
						assert.EqualError(t, err, specs.Err.Error())
					}

					specs.Request.Header = specs.Session.Header
					assert.Equal(t, serializeCookies(specs.Match), serializeCookies(specs.Request.Cookies()))
				})
			}
		})

		t.Run("caching=on", func(t *testing.T) {
			for _, specs := range testMap {
				var overrideCookies []*http.Cookie

				cache := template.New("rules")

				var cfg CredentialsCookiesConfig
				require.NoError(t, json.NewDecoder(bytes.NewBuffer(specs.Config)).Decode(&cfg))

				for cookie := range cfg.Cookies {
					templateId := fmt.Sprintf("%s:%s", specs.Rule.ID, cookie)
					_, err := cache.New(templateId).Parse("override")
					require.NoError(t, err)
					overrideCookies = append(overrideCookies, &http.Cookie{Name: cookie, Value: "override"})
				}

				a.(*MutatorCookie).WithCache(cache)

				err := a.Mutate(specs.Request, specs.Session, specs.Config, specs.Rule)
				if specs.Err == nil {
					// Issuer must run without error
					require.NoError(t, err)
				} else {
					assert.EqualError(t, err, specs.Err.Error())
				}

				specs.Request.Header = specs.Session.Header
				assert.Equal(t, serializeCookies(overrideCookies), serializeCookies(specs.Request.Cookies()))
			}
		})
	})
}

// assert.Equal doesn't handle []*http.Cookie comparisons very well, so
// converting them to a simple map[string]string makes testing easier
func serializeCookies(cookies []*http.Cookie) map[string]string {
	out := map[string]string{}

	for _, cookie := range cookies {
		out[cookie.Name] = cookie.Value
	}

	return out
}
