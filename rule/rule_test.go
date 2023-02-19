// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package rule

import (
	"encoding/json"
	"net/url"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/oathkeeper/driver/configuration"
)

func mustParse(t *testing.T, u string) *url.URL {
	p, err := url.Parse(u)
	require.NoError(t, err)
	return p
}

func TestRule(t *testing.T) {
	rules := []Rule{
		{
			Match: &Match{
				Methods: []string{"DELETE"},
				URL:     "https://localhost/users/<[0-9]+>",
			},
		},
		{
			Match: &Match{
				Methods: []string{"DELETE"},
				URL:     "https://localhost/users/<[[:digit:]]*>",
			},
		},
		{
			Match: &Match{
				Methods: []string{"DELETE"},
				URL:     "https://localhost/users/<[0-9]*>",
			},
		},
	}

	var tests = []struct {
		method        string
		url           string
		expectedMatch bool
		expectedErr   error
	}{
		{
			method:        "DELETE",
			url:           "https://localhost/users/1234",
			expectedMatch: true,
			expectedErr:   nil,
		},
		{
			method:        "DELETE",
			url:           "https://localhost/users/1234?key=value&key1=value1",
			expectedMatch: true,
			expectedErr:   nil,
		},
		{
			method:        "DELETE",
			url:           "https://localhost/users/abcd",
			expectedMatch: false,
			expectedErr:   nil,
		},
	}
	for ind, tcase := range tests {
		t.Run(strconv.FormatInt(int64(ind), 10), func(t *testing.T) {
			testFunc := func(rule Rule, strategy configuration.MatchingStrategy) {
				matched, err := rule.IsMatching(strategy, tcase.method, mustParse(t, tcase.url), ProtocolHTTP)
				assert.Equal(t, tcase.expectedMatch, matched)
				assert.Equal(t, tcase.expectedErr, err)
			}
			t.Run("rule0", func(t *testing.T) {
				testFunc(rules[0], configuration.Regexp)
			})
			t.Run("rule1", func(t *testing.T) {
				testFunc(rules[1], configuration.Regexp)
			})
			t.Run("rule2", func(t *testing.T) {
				testFunc(rules[2], configuration.Glob)
			})
		})
	}
}

func TestRule1(t *testing.T) {
	r := &Rule{
		Match: &Match{
			Methods: []string{"DELETE"},
			URL:     "https://localhost/users/<(?!admin).*>",
		},
	}

	var tests = []struct {
		method        string
		url           string
		expectedMatch bool
		expectedErr   error
	}{
		{
			method:        "DELETE",
			url:           "https://localhost/users/manager",
			expectedMatch: true,
			expectedErr:   nil,
		},
		{
			method:        "DELETE",
			url:           "https://localhost/users/1234?key=value&key1=value1",
			expectedMatch: true,
			expectedErr:   nil,
		},
		{
			method:        "DELETE",
			url:           "https://localhost/users/admin",
			expectedMatch: false,
			expectedErr:   nil,
		},
	}
	for ind, tcase := range tests {
		t.Run(strconv.FormatInt(int64(ind), 10), func(t *testing.T) {
			matched, err := r.IsMatching(configuration.Regexp, tcase.method, mustParse(t, tcase.url), ProtocolHTTP)
			assert.Equal(t, tcase.expectedMatch, matched)
			assert.Equal(t, tcase.expectedErr, err)
		})
	}
}

func TestRuleWithCustomMethod(t *testing.T) {
	r := &Rule{
		Match: &Match{
			Methods: []string{"CUSTOM"},
			URL:     "https://localhost/users/<(?!admin).*>",
		},
	}

	var tests = []struct {
		method        string
		url           string
		expectedMatch bool
		expectedErr   error
	}{
		{
			method:        "CUSTOM",
			url:           "https://localhost/users/manager",
			expectedMatch: true,
			expectedErr:   nil,
		},
		{
			method:        "CUSTOM",
			url:           "https://localhost/users/1234?key=value&key1=value1",
			expectedMatch: true,
			expectedErr:   nil,
		},
		{
			method:        "DELETE",
			url:           "https://localhost/users/admin",
			expectedMatch: false,
			expectedErr:   nil,
		},
	}
	for ind, tcase := range tests {
		t.Run(strconv.FormatInt(int64(ind), 10), func(t *testing.T) {
			matched, err := r.IsMatching(configuration.Regexp, tcase.method, mustParse(t, tcase.url), ProtocolHTTP)
			assert.Equal(t, tcase.expectedMatch, matched)
			assert.Equal(t, tcase.expectedErr, err)
		})
	}
}

func TestRule_UnmarshalJSON(t *testing.T) {
	var tests = []struct {
		name     string
		json     string
		expected Rule
		err      assert.ErrorAssertionFunc
	}{

		{name: "unmarshal gRPC match",
			json: `
{
	"id": "123",
	"description": "description",
	"authorizers": "nil",
	"match": { "authority": "example.com", "full_method": "/full/method" }
}
`,
			expected: Rule{
				ID:          "123",
				Description: "description",
				Match:       &MatchGRPC{Authority: "example.com", FullMethod: "/full/method"},
			},
			err: assert.NoError,
		},

		{name: "unmarshal HTTP match",
			json: `
{
	"id": "123",
	"description": "description",
	"authorizers": "nil",
	"match": { "url": "example.com/some/method", "methods": ["GET", "PUT"] }
}
`,
			expected: Rule{
				ID:          "123",
				Description: "description",
				Match:       &Match{Methods: []string{"GET", "PUT"}, URL: "example.com/some/method"},
			},
			err: assert.NoError,
		},

		{name: "err on invalid version",
			json: `
{
	"id": "123",
	"version": "42"
}
`,
			err: assert.Error,
		},

		{name: "err on invalid match",
			json: `
{
	"id": "123",
	"description": "description",
	"authorizers": "nil",
	"match": { foo }
}
`,
			err: assert.Error,
		},
	}

	for _, tc := range tests {
		t.Run("case="+tc.name, func(t *testing.T) {
			var (
				actual Rule
				err    error
			)
			err = json.Unmarshal([]byte(tc.json), &actual)
			assert.Equal(t, tc.expected, actual)
			tc.err(t, err)
		})
	}
}
