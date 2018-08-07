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
		"Subject": {
			Session: &AuthenticationSession{Subject: "foo"},
			Rule:    &rule.Rule{ID: "test-rule"},
			Config:  json.RawMessage([]byte(`{"X-User": "{{ .Subject }}"}`)),
			Request: &http.Request{Header: http.Header{}},
			Match:   http.Header{"X-User": []string{"foo"}},
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
			assert.Equal(t, specs.Request.Header, specs.Match)
		})
	}
}
