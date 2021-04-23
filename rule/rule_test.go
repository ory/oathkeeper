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

package rule

import (
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
				matched, err := rule.IsMatching(strategy, tcase.method, mustParse(t, tcase.url))
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
			matched, err := r.IsMatching(configuration.Regexp, tcase.method, mustParse(t, tcase.url))
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
			matched, err := r.IsMatching(configuration.Regexp, tcase.method, mustParse(t, tcase.url))
			assert.Equal(t, tcase.expectedMatch, matched)
			assert.Equal(t, tcase.expectedErr, err)
		})
	}
}
