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
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"regexp"
	"testing"

	"github.com/ory/oathkeeper/helper"
	"github.com/ory/oathkeeper/rsakey"
	"github.com/ory/oathkeeper/rule"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type jurorDenyAll struct{}

func (j *jurorDenyAll) GetID() string {
	return "pass_through_deny"
}

func (j jurorDenyAll) Try(r *http.Request, rl *rule.Rule, u *url.URL) (*Session, error) {
	return nil, errors.WithStack(helper.ErrUnauthorized)
}

type jurorAcceptAll struct{}

func (j *jurorAcceptAll) GetID() string {
	return "pass_through_accept"
}

func (j jurorAcceptAll) Try(r *http.Request, rl *rule.Rule, u *url.URL) (*Session, error) {
	return &Session{
		Subject:   "",
		Anonymous: true,
		ClientID:  "",
		Disabled:  true,
	}, nil
}

func TestProxy(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.NotEmpty(t, helper.BearerTokenFromRequest(r))
		fmt.Fprint(w, r.Header.Get("Authorization"))
	}))
	defer backend.Close()

	u, _ := url.Parse(backend.URL)
	d := NewProxy(u, nil, nil, &rsakey.LocalManager{KeyStrength: 512})

	proxy := httptest.NewServer(&httputil.ReverseProxy{Director: d.Director, Transport: d})
	defer proxy.Close()

	acceptRule := rule.Rule{MatchesMethods: []string{"GET"}, MatchesURLCompiled: mustCompileRegex(t, proxy.URL+"/users/<[0-9]+>"), Mode: "pass_through_accept"}
	denyRule := rule.Rule{MatchesMethods: []string{"GET"}, MatchesURLCompiled: mustCompileRegex(t, proxy.URL+"/users/<[0-9]+>"), Mode: "pass_through_deny"}

	for k, tc := range []struct {
		url       string
		code      int
		message   string
		rules     map[string]rule.Rule
		transform func(r *http.Request)
		d         string
	}{
		{
			d:     "should fail because url does not exist in rule set",
			url:   proxy.URL + "/invalid",
			rules: map[string]rule.Rule{},
			code:  http.StatusNotFound,
		},
		{
			d:     "should fail because url does exist but is matched by two rules",
			url:   proxy.URL + "/users/1234",
			rules: map[string]rule.Rule{"1": acceptRule, "2": acceptRule},
			code:  http.StatusInternalServerError,
		},
		{
			d:     "should pass",
			url:   proxy.URL + "/users/1234",
			rules: map[string]rule.Rule{"1": acceptRule},
			code:  http.StatusOK,
			transform: func(r *http.Request) {
				r.Header.Add("Authorization", "bearer token")
			},
			message: "bearer token",
		},
		{
			d:     "should fail because invalid credentials",
			url:   proxy.URL + "/users/1234",
			rules: map[string]rule.Rule{"A": denyRule},
			code:  http.StatusUnauthorized,
		},
	} {
		t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
			matcher := &rule.CachedMatcher{Rules: tc.rules}
			d.Judge = NewJudge(logrus.New(), matcher, "", []Juror{new(jurorAcceptAll), new(jurorDenyAll)})

			req, err := http.NewRequest("GET", tc.url, nil)
			require.NoError(t, err)
			if tc.transform != nil {
				tc.transform(req)
			}

			res, err := http.DefaultClient.Do(req)
			require.NoError(t, err)

			greeting, err := ioutil.ReadAll(res.Body)
			res.Body.Close()
			require.NoError(t, err)

			assert.Equal(t, tc.code, res.StatusCode)
			if tc.message != "" {
				assert.Equal(t, tc.message, fmt.Sprintf("%s", greeting))
			}
		})
	}
}

func panicCompileRegex(pattern string) *regexp.Regexp {
	exp, err := regexp.Compile(pattern)
	if err != nil {
		panic(err.Error())
	}
	return exp
}

func BenchmarkDirector(b *testing.B) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, r.Header.Get("Authorization"))
	}))
	defer backend.Close()

	logger := logrus.New()
	logger.Level = logrus.WarnLevel
	u, _ := url.Parse(backend.URL)
	d := NewProxy(u, nil, logger, &rsakey.LocalManager{KeyStrength: 512})

	p := httptest.NewServer(&httputil.ReverseProxy{Director: d.Director, Transport: d})
	defer p.Close()

	jt := &JurorPassThrough{L: logrus.New()}
	matcher := &rule.CachedMatcher{Rules: map[string]rule.Rule{
		"A": {MatchesMethods: []string{"GET"}, MatchesURLCompiled: panicCompileRegex(p.URL + "/users"), Mode: jt.GetID()},
		"B": {MatchesMethods: []string{"GET"}, MatchesURLCompiled: panicCompileRegex(p.URL + "/users/<[0-9]+>"), Mode: jt.GetID()},
		"C": {MatchesMethods: []string{"GET"}, MatchesURLCompiled: panicCompileRegex(p.URL + "/<[0-9]+>"), Mode: jt.GetID()},
		"D": {MatchesMethods: []string{"GET"}, MatchesURLCompiled: panicCompileRegex(p.URL + "/other/<.+>"), Mode: jt.GetID()},
	}}
	d.Judge = NewJudge(logger, matcher, "", []Juror{jt})

	req, _ := http.NewRequest("GET", p.URL+"/users", nil)

	b.Run("case=fetch_user_endpoint", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			res, err := http.DefaultClient.Do(req)
			if err != nil {
				b.FailNow()
			}

			if res.StatusCode != http.StatusOK {
				b.FailNow()
			}
		}
	})
}
