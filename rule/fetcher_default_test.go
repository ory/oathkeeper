package rule_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/oathkeeper/driver/configuration"
	"github.com/ory/oathkeeper/internal"
	"github.com/ory/oathkeeper/rule"
)

const testRule = `[{"id":"test-rule-5","upstream":{"preserve_host":true,"strip_path":"/api","url":"mybackend.com/api"},"match":{"url":"myproxy.com/api","methods":["GET","POST"]},"authenticators":[{"handler":"noop"},{"handler":"anonymous"}],"authorizer":{"handler":"allow"},"mutator":{"handler":"noop"}}]`

func TestFetcher(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(testRule))
	}))
	defer ts.Close()

	conf := internal.NewConfigurationWithDefaults()
	r := internal.NewRegistry(conf)

	for k, tc := range []struct {
		sources   []string
		expectIDs []string
		expectErr bool
	}{
		{},
		{sources: []string{}},
		{sources: []string{""}},
		{
			sources:   []string{"ftp://not-valid"},
			expectErr: true,
		},
		{
			sources: []string{
				"file://../test/stub/rules.json",
				// test-rule-4
				"inline://W3siaWQiOiJ0ZXN0LXJ1bGUtNCIsInVwc3RyZWFtIjp7InByZXNlcnZlX2hvc3QiOnRydWUsInN0cmlwX3BhdGgiOiIvYXBpIiwidXJsIjoibXliYWNrZW5kLmNvbS9hcGkifSwibWF0Y2giOnsidXJsIjoibXlwcm94eS5jb20vYXBpIiwibWV0aG9kcyI6WyJHRVQiLCJQT1NUIl19LCJhdXRoZW50aWNhdG9ycyI6W3siaGFuZGxlciI6Im5vb3AifSx7ImhhbmRsZXIiOiJhbm9ueW1vdXMifV0sImF1dGhvcml6ZXIiOnsiaGFuZGxlciI6ImFsbG93In0sIm11dGF0b3IiOnsiaGFuZGxlciI6Im5vb3AifX1d",
				ts.URL,
			},
			expectIDs: []string{"test-rule-1", "test-rule-2", "test-rule-3", "test-rule-4", "test-rule-5"},
		},
	} {
		t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
			viper.Set(configuration.ViperKeyAccessRuleRepositories, tc.sources)
			rules, err := r.RuleFetcher().Fetch()
			if tc.expectErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			ids := make([]string, len(rules))
			for k, r := range rules {
				ids[k] = r.ID
			}
			if tc.expectIDs == nil {
				tc.expectIDs = []string{}
			}
			assert.EqualValues(t, ids, tc.expectIDs)
		})
	}
}

func TestFetcherDefaultFetchFormats(t *testing.T) {
	expected := []rule.Rule{
		{
			ID: "test-rule-1",
			Match: rule.RuleMatch{
				Methods: []string{"GET", "POST"},
				URL:     "myproxy.com/api",
			},
			Authenticators: []rule.RuleHandler{
				{
					Handler: "noop",
				},
				{
					Handler: "anonymous",
				},
			},
			Authorizer: rule.RuleHandler{
				Handler: "allow",
			},
			Mutator: rule.RuleHandler{
				Handler: "noop",
			},
			Upstream: rule.Upstream{
				PreserveHost: true,
				StripPath:    "/api",
				URL:          "mybackend.com/api",
			},
		},
	}

	testCases := map[string]struct {
		fpath string
	}{
		"json file": {
			fpath: "testdata/rules.json",
		},
		"yaml file": {
			fpath: "testdata/rules.yaml",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			conf := internal.NewConfigurationWithDefaults()
			viper.Set(configuration.ViperKeyAccessRuleRepositories, []string{"file://" + tc.fpath})

			r := internal.NewRegistry(conf)
			rules, err := r.RuleFetcher().Fetch()
			require.NoError(t, err)

			assert.Equal(t, expected, rules)
		})
	}
}
