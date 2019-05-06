package proxy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"text/template"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/oathkeeper/rule"
)

func TestCredentialsIssuerHeaders(t *testing.T) {

	var testMap = map[string]struct {
		Session *AuthenticationSession
		Rule    *rule.Rule
		Config  json.RawMessage
		Request *http.Request
		Match   http.Header
		Err     error
	}{
		"Simple Subject": {
			Session: &AuthenticationSession{Subject: "foo"},
			Rule:    &rule.Rule{ID: "test-rule"},
			Config:  json.RawMessage([]byte(`{"headers":{"X-User": "{{ print .Subject }}"}}`)),
			Request: &http.Request{Header: http.Header{}},
			Match:   http.Header{"X-User": []string{"foo"}},
			Err:     nil,
		},
		"Complex Subject": {
			Session: &AuthenticationSession{Subject: "foo"},
			Rule:    &rule.Rule{ID: "test-rule2"},
			Config:  json.RawMessage([]byte(`{"headers":{"X-User": "realm:resources:users:{{ print .Subject }}"}}`)),
			Request: &http.Request{Header: http.Header{}},
			Match:   http.Header{"X-User": []string{"realm:resources:users:foo"}},
			Err:     nil,
		},
		"Subject & Extras": {
			Session: &AuthenticationSession{Subject: "foo", Extra: map[string]interface{}{"iss": "issuer", "aud": "audience"}},
			Rule:    &rule.Rule{ID: "test-rule3"},
			Config:  json.RawMessage([]byte(`{"headers":{"X-User": "{{ print .Subject }}", "X-Issuer": "{{ print .Extra.iss }}", "X-Audience": "{{ print .Extra.aud }}"}}`)),
			Request: &http.Request{Header: http.Header{}},
			Match:   http.Header{"X-User": []string{"foo"}, "X-Issuer": []string{"issuer"}, "X-Audience": []string{"audience"}},
			Err:     nil,
		},
		"All In One Header": {
			Session: &AuthenticationSession{Subject: "foo", Extra: map[string]interface{}{"iss": "issuer", "aud": "audience"}},
			Rule:    &rule.Rule{ID: "test-rule4"},
			Config:  json.RawMessage([]byte(`{"headers":{"X-Kitchen-Sink": "{{ print .Subject }} {{ print .Extra.iss }} {{ print .Extra.aud }}"}}`)),
			Request: &http.Request{Header: http.Header{}},
			Match:   http.Header{"X-Kitchen-Sink": []string{"foo issuer audience"}},
			Err:     nil,
		},
		"Scrub Incoming Headers": {
			Session: &AuthenticationSession{Subject: "anonymous"},
			Rule:    &rule.Rule{ID: "test-rule5"},
			Config:  json.RawMessage([]byte(`{"headers":{"X-User": "{{ print .Subject }}", "X-Issuer": "{{ print .Extra.iss }}", "X-Audience": "{{ print .Extra.aud }}"}}`)),
			Request: &http.Request{Header: http.Header{"X-User": []string{"admin"}, "X-Issuer": []string{"issuer"}, "X-Audience": []string{"audience"}}},
			Match:   http.Header{"X-User": []string{"anonymous"}, "X-Issuer": []string{""}, "X-Audience": []string{""}},
			Err:     nil,
		},
		"Missing Extras": {
			Session: &AuthenticationSession{Subject: "foo", Extra: map[string]interface{}{}},
			Rule:    &rule.Rule{ID: "test-rule6"},
			Config:  json.RawMessage([]byte(`{"headers":{"X-Issuer": "{{ print .Extra.iss }}"}}`)),
			Request: &http.Request{Header: http.Header{}},
			Match:   http.Header{"X-Issuer": []string{""}},
			Err:     nil,
		},
		"Nested Extras": {
			Session: &AuthenticationSession{
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
			Session: &AuthenticationSession{Subject: "foo", Extra: map[string]interface{}{}},
			Rule:    &rule.Rule{ID: "test-rule8"},
			Config:  json.RawMessage(`{"bar":"baz"}`),
			Request: &http.Request{Header: http.Header{}},
			Match:   http.Header{},
			Err:     errors.New(`json: unknown field "bar"`),
		},
	}

	for testName, specs := range testMap {
		t.Run(testName, func(t *testing.T) {
			issuer := NewCredentialsIssuerHeaders()

			// Must return non-nil issuer
			assert.NotNil(t, issuer)

			// Issuer must return non-empty ID
			assert.NotEmpty(t, issuer.GetID())

			header, err := issuer.Transform(specs.Request, specs.Session, specs.Config, specs.Rule)
			if specs.Err == nil {
				// Issuer must run without error
				require.NoError(t, err)
			} else {
				assert.Equal(t, specs.Err.Error(), err.Error())
			}

			specs.Request.Header = header
			if header == nil {
				specs.Request.Header = http.Header{}
			}

			// Output request headers must match test specs
			assert.Equal(t, specs.Match, specs.Request.Header)
		})
	}

	t.Run("Caching", func(t *testing.T) {
		for _, specs := range testMap {
			issuer := NewCredentialsIssuerHeaders()

			overrideHeaders := http.Header{}

			cache := template.New("rules")

			var cfg CredentialsHeadersConfig
			d := json.NewDecoder(bytes.NewBuffer(specs.Config))
			d.Decode(&cfg)

			for hdr := range cfg.Headers {
				templateId := fmt.Sprintf("%s:%s", specs.Rule.ID, hdr)
				cache.New(templateId).Parse("override")
				overrideHeaders.Add(hdr, "override")
			}

			issuer.t = cache

			header, err := issuer.Transform(specs.Request, specs.Session, specs.Config, specs.Rule)
			if specs.Err == nil {
				// Issuer must run without error
				require.NoError(t, err)
			} else {
				assert.Equal(t, specs.Err.Error(), err.Error())
			}

			specs.Request.Header = header
			if header == nil {
				specs.Request.Header = http.Header{}
			}

			assert.Equal(t, overrideHeaders, specs.Request.Header)
		}
	})
}
