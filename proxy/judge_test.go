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

package proxy

import (
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"testing"

	"github.com/ory/ladon/compiler"
	"github.com/ory/oathkeeper/rule"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mustCompileRegex(t *testing.T, pattern string) *regexp.Regexp {
	exp, err := compiler.CompileRegex(pattern, '<', '>')
	require.NoError(t, err)
	return exp
}

func mustGenerateURL(t *testing.T, u string) *url.URL {
	up, err := url.Parse(u)
	require.NoError(t, err)
	return up
}

func TestJudge(t *testing.T) {
	rules := map[string]rule.Rule{
		"1": {
			MatchesMethods:     []string{"GET"},
			MatchesURLCompiled: mustCompileRegex(t, "http://localhost/users/<[0-9]+>"),
			RequiredResource:   "users:$1",
			RequiredAction:     "get:$1",
			RequiredScopes:     []string{"users.get"},
			Mode:               "a",
		}, "2": {
			MatchesMethods:     []string{"GET"},
			MatchesURLCompiled: mustCompileRegex(t, "http://localhost/articles/<[0-9]+>"),
			RequiredResource:   "users:$1",
			RequiredAction:     "get:$1",
			RequiredScopes:     []string{"users.get"},
			Mode:               "c",
		},
	}
	j := &Judge{
		Logger:  logrus.New(),
		Jury:    map[string]Juror{"a": &JurorPassThrough{L: logrus.New()}},
		Matcher: &rule.CachedMatcher{Rules: rules},
	}

	for k, tc := range []struct {
		r *http.Request
		e func(*testing.T, *Session, error)
	}{
		{
			r: &http.Request{Method: "GET", Host: "localhost", URL: mustGenerateURL(t, "http://localhost/users/1234")},
			e: func(t *testing.T, s *Session, err error) {
				require.NoError(t, err)
				assert.Empty(t, s.ClientID)
				assert.Empty(t, s.Subject)
				assert.True(t, s.Anonymous)
			},
		},
		{
			r: &http.Request{Method: "GET", Host: "localhost", URL: mustGenerateURL(t, "http://localhost/articles/1234")},
			e: func(t *testing.T, s *Session, err error) {
				require.Error(t, err)
			},
		},
		{
			r: &http.Request{Method: "GET", Host: "localhost", URL: mustGenerateURL(t, "http://localhost/foo/1234")},
			e: func(t *testing.T, s *Session, err error) {
				require.Error(t, err)
			},
		},
	} {
		t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
			s, err := j.EvaluateAccessRequest(tc.r)
			tc.e(t, s, err)
		})
	}
}
