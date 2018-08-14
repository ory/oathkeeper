package proxy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"text/template"

	"github.com/ory/oathkeeper/rule"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCredentialsIssuerCookies(t *testing.T) {
	var testMap = map[string]struct {
		Session *AuthenticationSession
		Rule    *rule.Rule
		Config  json.RawMessage
		Request *http.Request
		Match   []*http.Cookie
		Err     error
	}{
		"Simple Subject": {
			Session: &AuthenticationSession{Subject: "foo"},
			Rule:    &rule.Rule{ID: "test-rule"},
			Config:  json.RawMessage([]byte(`{"cookies": {"user": "{{ print .Subject }}"}}`)),
			Request: &http.Request{Header: http.Header{}},
			Match:   []*http.Cookie{&http.Cookie{Name: "user", Value: "foo"}},
			Err:     nil,
		},
		"Unknown Config Field": {
			Session: &AuthenticationSession{},
			Rule:    &rule.Rule{ID: "test-rule2"},
			Config:  json.RawMessage([]byte(`{"bar": "baz"}`)),
			Request: &http.Request{Header: http.Header{}},
			Match:   []*http.Cookie{},
			Err:     errors.New(`json: unknown field "bar"`),
		},
		"Complex Subject": {
			Session: &AuthenticationSession{Subject: "foo"},
			Rule:    &rule.Rule{ID: "test-rule3"},
			Config:  json.RawMessage([]byte(`{"cookies": {"user": "realm:resources:users:{{ print .Subject }}"}}`)),
			Request: &http.Request{Header: http.Header{}},
			Match:   []*http.Cookie{&http.Cookie{Name: "user", Value: "realm:resources:users:foo"}},
			Err:     nil,
		},
		"Subject & Extras": {
			Session: &AuthenticationSession{Subject: "foo", Extra: map[string]interface{}{"iss": "issuer", "aud": "audience"}},
			Rule:    &rule.Rule{ID: "test-rule4"},
			Config:  json.RawMessage([]byte(`{"cookies":{"user": "{{ print .Subject }}", "issuer": "{{ print .Extra.iss }}", "audience": "{{ print .Extra.aud }}"}}`)),
			Request: &http.Request{Header: http.Header{}},
			Match: []*http.Cookie{
				&http.Cookie{Name: "user", Value: "foo"},
				&http.Cookie{Name: "issuer", Value: "issuer"},
				&http.Cookie{Name: "audience", Value: "audience"},
			},
			Err: nil,
		},
		"All In One Cookie": {
			Session: &AuthenticationSession{Subject: "foo", Extra: map[string]interface{}{"iss": "issuer", "aud": "audience"}},
			Rule:    &rule.Rule{ID: "test-rule5"},
			Config:  json.RawMessage([]byte(`{"cookies":{"kitchensink": "{{ print .Subject }} {{ print .Extra.iss }} {{ print .Extra.aud }}"}}`)),
			Request: &http.Request{Header: http.Header{}},
			Match: []*http.Cookie{
				&http.Cookie{Name: "kitchensink", Value: "foo issuer audience"},
			},
			Err: nil,
		},
		"Scrub Incoming Cookies": {
			Session: &AuthenticationSession{Subject: "anonymous"},
			Rule:    &rule.Rule{ID: "test-rule6"},
			Config:  json.RawMessage([]byte(`{"cookies":{"user": "{{ print .Subject }}", "issuer": "{{ print .Extra.iss }}", "audience": "{{ print .Extra.aud }}"}}`)),
			Request: &http.Request{
				Header: http.Header{"Cookie": []string{"user=admin;issuer=issuer;audience=audience"}},
			},
			Match: []*http.Cookie{
				&http.Cookie{Name: "user", Value: "anonymous"},
				&http.Cookie{Name: "issuer", Value: ""},
				&http.Cookie{Name: "audience", Value: ""},
			},
			Err: nil,
		},
		"Missing Extras": {
			Session: &AuthenticationSession{Subject: "foo", Extra: map[string]interface{}{}},
			Rule:    &rule.Rule{ID: "test-rule7"},
			Config:  json.RawMessage([]byte(`{"cookies":{"issuer": "{{ print .Extra.iss }}"}}`)),
			Request: &http.Request{Header: http.Header{}},
			Match:   []*http.Cookie{&http.Cookie{Name: "issuer", Value: ""}},
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
			Rule: &rule.Rule{ID: "test-rule8"},
			Config: json.RawMessage([]byte(`{"cookies":{
				"nested-int": "{{ print .Extra.nested.int }}",
				"nested-float64": "{{ print .Extra.nested.float64 }}",
				"nested-bool": "{{ print .Extra.nested.bool}}",
				"nested-nonexistent": "{{ print .Extra.nested.nil }}"
			}}`)),
			Request: &http.Request{Header: http.Header{}},
			Match: []*http.Cookie{
				&http.Cookie{Name: "nested-int", Value: "10"},
				&http.Cookie{Name: "nested-float64", Value: "3.14159"},
				&http.Cookie{Name: "nested-bool", Value: "true"},
				&http.Cookie{Name: "nested-nonexistent", Value: ""},
			},
			Err: nil,
		},
	}

	for testName, specs := range testMap {
		t.Run(testName, func(t *testing.T) {
			issuer := NewCredentialsIssuerCookies()

			// Must return non-nil issuer
			assert.NotNil(t, issuer)

			// Issuer must return non-empty ID
			assert.NotEmpty(t, issuer.GetID())

			if specs.Err == nil {
				require.NoError(t, issuer.Issue(specs.Request, specs.Session, specs.Config, specs.Rule))
			} else {
				err := issuer.Issue(specs.Request, specs.Session, specs.Config, specs.Rule)
				assert.Equal(t, specs.Err.Error(), err.Error())
			}

			assert.Equal(t, serializeCookies(specs.Match), serializeCookies(specs.Request.Cookies()))
		})
	}

	t.Run("Caching", func(t *testing.T) {
		for _, specs := range testMap {
			issuer := NewCredentialsIssuerCookies()

			overrideCookies := []*http.Cookie{}

			cache := template.New("rules")

			var cfg CredentialsCookiesConfig
			d := json.NewDecoder(bytes.NewBuffer(specs.Config))
			d.Decode(&cfg)

			for cookie, _ := range cfg.Cookies {
				templateId := fmt.Sprintf("%s:%s", specs.Rule.ID, cookie)
				cache.New(templateId).Parse("override")
				overrideCookies = append(overrideCookies, &http.Cookie{Name: cookie, Value: "override"})
			}

			issuer.RulesCache = cache

			if specs.Err == nil {
				require.NoError(t, issuer.Issue(specs.Request, specs.Session, specs.Config, specs.Rule))
			} else {
				err := issuer.Issue(specs.Request, specs.Session, specs.Config, specs.Rule)
				assert.Equal(t, specs.Err.Error(), err.Error())
			}

			assert.Equal(t, serializeCookies(overrideCookies), serializeCookies(specs.Request.Cookies()))
		}
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
