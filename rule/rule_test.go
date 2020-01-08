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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mustParse(t *testing.T, u string) *url.URL {
	p, err := url.Parse(u)
	require.NoError(t, err)
	return p
}

func TestRule(t *testing.T) {
	r := &Rule{
		Match: &Match{
			Methods: []string{"DELETE"},
			URL:     "https://localhost/users/<[0-9]+>",
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
		t.Run(string(ind), func(t *testing.T) {
			matched, err := r.IsMatching(tcase.method, mustParse(t, tcase.url))
			assert.Equal(t, tcase.expectedMatch, matched)
			assert.Equal(t, tcase.expectedErr, err)
		})
	}
}

func TestRule1(t *testing.T) {
	r := &Rule{
		Match: &Match{
			Methods: []string{"DELETE"},
			URL:     "https://localhost/users/<[[:digit:]]*>",
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
		t.Run(string(ind), func(t *testing.T) {
			matched, err := r.IsMatching(tcase.method, mustParse(t, tcase.url))
			assert.Equal(t, tcase.expectedMatch, matched)
			assert.Equal(t, tcase.expectedErr, err)
		})
	}
}

func TestRule2(t *testing.T) {
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
		t.Run(string(ind), func(t *testing.T) {
			matched, err := r.IsMatching(tcase.method, mustParse(t, tcase.url))
			assert.Equal(t, tcase.expectedMatch, matched)
			assert.Equal(t, tcase.expectedErr, err)
		})
	}
}
