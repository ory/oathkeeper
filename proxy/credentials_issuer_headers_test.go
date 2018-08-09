package proxy

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/ory/oathkeeper/rule"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCredentialsIssuerHeaders(t *testing.T) {

	var testMap = map[string]struct {
		Session *AuthenticationSession
		Rule    *rule.Rule
		Config  json.RawMessage
		Request *http.Request
		Match   http.Header
	}{
		"Simple Subject": {
			Session: &AuthenticationSession{Subject: "foo"},
			Rule:    &rule.Rule{ID: "test-rule"},
			Config:  json.RawMessage([]byte(`{"X-User": "{{ .Subject }}"}`)),
			Request: &http.Request{Header: http.Header{}},
			Match:   http.Header{"X-User": []string{"foo"}},
		},
		"Complex Subject": {
			Session: &AuthenticationSession{Subject: "foo"},
			Rule:    &rule.Rule{ID: "test-rule2"},
			Config:  json.RawMessage([]byte(`{"X-User": "realm:resources:users:{{ .Subject }}"}`)),
			Request: &http.Request{Header: http.Header{}},
			Match:   http.Header{"X-User": []string{"realm:resources:users:foo"}},
		},
		"Subject & Extras": {
			Session: &AuthenticationSession{Subject: "foo", Extra: map[string]interface{}{"iss": "issuer", "aud": "audience"}},
			Rule:    &rule.Rule{ID: "test-rule3"},
			Config:  json.RawMessage([]byte(`{"X-User": "{{ .Subject }}", "X-Issuer": "{{ .Extra.iss }}", "X-Audience": "{{ .Extra.aud }}"}`)),
			Request: &http.Request{Header: http.Header{}},
			Match:   http.Header{"X-User": []string{"foo"}, "X-Issuer": []string{"issuer"}, "X-Audience": []string{"audience"}},
		},
		"All In One Header": {
			Session: &AuthenticationSession{Subject: "foo", Extra: map[string]interface{}{"iss": "issuer", "aud": "audience"}},
			Rule:    &rule.Rule{ID: "test-rule4"},
			Config:  json.RawMessage([]byte(`{"X-Kitchen-Sink": "{{ .Subject }} {{ .Extra.iss }} {{ .Extra.aud }}"}`)),
			Request: &http.Request{Header: http.Header{}},
			Match:   http.Header{"X-Kitchen-Sink": []string{"foo issuer audience"}},
		},
		"Scrub Incoming Headers": {
			Session: &AuthenticationSession{Subject: "anonymous"},
			Rule:    &rule.Rule{ID: "test-rule5"},
			Config:  json.RawMessage([]byte(`{"X-User": "{{ .Subject }}", "X-Issuer": "{{ .Extra.iss }}", "X-Audience": "{{ .Extra.aud }}"}`)),
			Request: &http.Request{Header: http.Header{"X-User": []string{"admin"}, "X-Issuer": []string{"issuer"}, "X-Audience": []string{"audience"}}},
			Match:   http.Header{"X-User": []string{"anonymous"}, "X-Issuer": []string{""}, "X-Audience": []string{""}},
		},
	}

	for testName, specs := range testMap {
		t.Run(testName, func(t *testing.T) {
			issuer := NewCredentialsIssuerHeaders()

			// Must return non-nil issuer
			assert.NotNil(t, issuer)

			// Issuer must return non-empty ID
			assert.NotEmpty(t, issuer.GetID())

			// Issuer must run without error
			require.NoError(t, issuer.Issue(specs.Request, specs.Session, specs.Config, specs.Rule))

			// Output request headers must match test specs
			assert.Equal(t, specs.Match, specs.Request.Header)
		})
	}
}
