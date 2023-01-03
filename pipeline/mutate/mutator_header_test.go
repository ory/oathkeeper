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

	"github.com/ory/x/configx"

	"github.com/ory/oathkeeper/driver/configuration"
	"github.com/ory/oathkeeper/internal"

	"github.com/ory/oathkeeper/pipeline/authn"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	. "github.com/ory/oathkeeper/pipeline/mutate"
	"github.com/ory/oathkeeper/rule"
)

func TestCredentialsIssuerHeaders(t *testing.T) {
	t.Parallel()
	conf := internal.NewConfigurationWithDefaults(configx.SkipValidation())
	reg := internal.NewRegistry(conf)

	a, err := reg.PipelineMutator("header")
	require.NoError(t, err)
	assert.Equal(t, "header", a.GetID())

	t.Run("method=mutate", func(t *testing.T) {
		var testMap = map[string]struct {
			Session *authn.AuthenticationSession
			Rule    *rule.Rule
			Config  json.RawMessage
			Request *http.Request
			Match   http.Header
			Err     error
		}{
			"Simple Subject": {
				Session: &authn.AuthenticationSession{Subject: "foo"},
				Rule:    &rule.Rule{ID: "test-rule"},
				Config:  json.RawMessage([]byte(`{"headers":{"X-User": "{{ print .Subject }}"}}`)),
				Request: &http.Request{Header: http.Header{}},
				Match:   http.Header{"X-User": []string{"foo"}},
				Err:     nil,
			},
			"Complex Subject": {
				Session: &authn.AuthenticationSession{Subject: "foo"},
				Rule:    &rule.Rule{ID: "test-rule2"},
				Config:  json.RawMessage([]byte(`{"headers":{"X-User": "realm:resources:users:{{ print .Subject }}"}}`)),
				Request: &http.Request{Header: http.Header{}},
				Match:   http.Header{"X-User": []string{"realm:resources:users:foo"}},
				Err:     nil,
			},
			"Subject & Extras": {
				Session: &authn.AuthenticationSession{Subject: "foo", Extra: map[string]interface{}{"iss": "issuer", "aud": "audience"}},
				Rule:    &rule.Rule{ID: "test-rule3"},
				Config:  json.RawMessage([]byte(`{"headers":{"X-User": "{{ print .Subject }}", "X-Issuer": "{{ print .Extra.iss }}", "X-Audience": "{{ print .Extra.aud }}"}}`)),
				Request: &http.Request{Header: http.Header{}},
				Match:   http.Header{"X-User": []string{"foo"}, "X-Issuer": []string{"issuer"}, "X-Audience": []string{"audience"}},
				Err:     nil,
			},
			"All In One Header": {
				Session: &authn.AuthenticationSession{Subject: "foo", Extra: map[string]interface{}{"iss": "issuer", "aud": "audience"}},
				Rule:    &rule.Rule{ID: "test-rule4"},
				Config:  json.RawMessage([]byte(`{"headers":{"X-Kitchen-Sink": "{{ print .Subject }} {{ print .Extra.iss }} {{ print .Extra.aud }}"}}`)),
				Request: &http.Request{Header: http.Header{}},
				Match:   http.Header{"X-Kitchen-Sink": []string{"foo issuer audience"}},
				Err:     nil,
			},
			"Scrub Incoming Headers": {
				Session: &authn.AuthenticationSession{Subject: "anonymous"},
				Rule:    &rule.Rule{ID: "test-rule5"},
				Config:  json.RawMessage([]byte(`{"headers":{"X-User": "{{ print .Subject }}", "X-Issuer": "{{ print .Extra.iss }}", "X-Audience": "{{ print .Extra.aud }}"}}`)),
				Request: &http.Request{Header: http.Header{"X-User": []string{"admin"}, "X-Issuer": []string{"issuer"}, "X-Audience": []string{"audience"}}},
				Match:   http.Header{"X-User": []string{"anonymous"}, "X-Issuer": []string{""}, "X-Audience": []string{""}},
				Err:     nil,
			},
			"Missing Extras": {
				Session: &authn.AuthenticationSession{Subject: "foo", Extra: map[string]interface{}{}},
				Rule:    &rule.Rule{ID: "test-rule6"},
				Config:  json.RawMessage([]byte(`{"headers":{"X-Issuer": "{{ print .Extra.iss }}"}}`)),
				Request: &http.Request{Header: http.Header{}},
				Match:   http.Header{"X-Issuer": []string{""}},
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
				Rule: &rule.Rule{ID: "test-rule7"},
				Config: json.RawMessage([]byte(`{"headers":{
				"X-Nested-Int": "{{ print .Extra.nested.int }}",
				"X-Nested-Float64": "{{ print .Extra.nested.float64 }}",
				"X-Nested-Bool": "{{ print .Extra.nested.bool}}",
				"X-Nested-Nonexistent": "{{ print .Extra.nested.nil }}"
			}}`)),
				Request: &http.Request{Header: http.Header{}},
				Match: http.Header{
					"X-Nested-Int":         []string{"10"},
					"X-Nested-Float64":     []string{"3.14159"},
					"X-Nested-Bool":        []string{"true"},
					"X-Nested-Nonexistent": []string{""},
				},
				Err: nil,
			},
			"Unknown Config Field": {
				Session: &authn.AuthenticationSession{Subject: "foo", Extra: map[string]interface{}{}},
				Rule:    &rule.Rule{ID: "test-rule8"},
				Config:  json.RawMessage(`{"bar":"baz"}`),
				Request: &http.Request{Header: http.Header{}},
				Match:   http.Header{},
				Err:     errors.New(`json: unknown field "bar"`),
			},
			"Advanced template with sprig function": {
				Session: &authn.AuthenticationSession{
					Subject: "foo",
					Extra: map[string]interface{}{
						"example.com/some-claims": []string{"Foo", "Bar"},
					},
				},
				Rule: &rule.Rule{ID: "test-rule9"},
				Config: json.RawMessage([]byte(`{"headers":{
					"Example-Claims": "{{- (index .Extra \"example.com/some-claims\") | join \",\" -}}"
				}}`)),
				Request: &http.Request{Header: http.Header{}},
				Match: http.Header{
					"Example-Claims": []string{"Foo,Bar"},
				},
				Err: nil,
			},
			"Use request captures to header": {
				Session: &authn.AuthenticationSession{
					Subject: "foo",
					MatchContext: authn.MatchContext{
						RegexpCaptureGroups: []string{"Foo", "Bar"},
					},
				},
				Rule: &rule.Rule{ID: "test-rule10"},
				Config: json.RawMessage([]byte(`{"headers":{
					"Example-Claims": "{{ index .MatchContext.RegexpCaptureGroups 0}}"
				}}`)),
				Request: &http.Request{Header: http.Header{}},
				Match: http.Header{
					"Example-Claims": []string{"Foo"},
				},
				Err: nil,
			},
		}

		t.Run("cache=disabled", func(t *testing.T) {
			for testName, specs := range testMap {
				t.Run(testName, func(t *testing.T) {
					err := a.Mutate(specs.Request, specs.Session, specs.Config, specs.Rule)
					if specs.Err == nil {
						// Issuer must run without error
						require.NoError(t, err)
					} else {
						assert.Error(t, err, specs.Err.Error())
					}

					specs.Request.Header = specs.Session.Header
					if specs.Session.Header == nil {
						specs.Request.Header = http.Header{}
					}

					// Output request headers must match test specs
					assert.Equal(t, specs.Match, specs.Request.Header)
				})
			}
		})

		t.Run("cache=enabled", func(t *testing.T) {
			for _, specs := range testMap {
				overrideHeaders := http.Header{}

				cache := template.New("rules")

				var cfg MutatorHeaderConfig
				require.NoError(t, json.NewDecoder(bytes.NewBuffer(specs.Config)).Decode(&cfg))

				for hdr := range cfg.Headers {
					templateId := fmt.Sprintf("%s:%s", specs.Rule.ID, hdr)
					_, err := cache.New(templateId).Parse("override")
					require.NoError(t, err)
					overrideHeaders.Add(hdr, "override")
				}

				a.(*MutatorHeader).WithCache(cache)

				err := a.Mutate(specs.Request, specs.Session, specs.Config, specs.Rule)
				if specs.Err == nil {
					// Issuer must run without error
					require.NoError(t, err)
				} else {
					assert.Error(t, err, specs.Err.Error())
				}

				specs.Request.Header = specs.Session.Header
				if specs.Session.Header == nil {
					specs.Request.Header = http.Header{}
				}

				assert.Equal(t, overrideHeaders, specs.Request.Header)
			}
		})
	})

	t.Run("method=validate", func(t *testing.T) {
		conf.SetForTest(t, configuration.MutatorHeaderIsEnabled, true)
		require.NoError(t, a.Validate(json.RawMessage(`{"headers":{}}`)))

		conf.SetForTest(t, configuration.MutatorHeaderIsEnabled, false)
		require.Error(t, a.Validate(json.RawMessage(`{"headers":{}}`)))
	})
}
