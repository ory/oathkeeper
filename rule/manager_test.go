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
	"regexp"
	"testing"

	"net/url"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	"github.com/ory/ladon/compiler"
	"github.com/ory/oathkeeper/pkg"
	"github.com/ory/sqlcon/dockertest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mustCompileRegex(t *testing.T, pattern string) *regexp.Regexp {
	exp, err := compiler.CompileRegex(pattern, '<', '>')
	require.NoError(t, err)
	return exp
}

func mustParseURL(t *testing.T, u string) *url.URL {
	exp, err := url.Parse(u)
	require.NoError(t, err)
	return exp
}

func TestMain(m *testing.M) {
	ex := dockertest.Register()
	code := m.Run()
	ex.Exit(code)
}

func TestManagers(t *testing.T) {
	managers := map[string]Manager{
		"memory": NewMemoryManager(),
	}

	if !testing.Short() {
		connectToPostgres(t, managers)
		connectToMySQL(t, managers)
	}

	for k, manager := range managers {
		r1 := Rule{
			ID:                 "foo1",
			Description:        "Create users rule",
			MatchesURLCompiled: mustCompileRegex(t, "/users/([0-9]+)"),
			MatchesURL:         "/users/([0-9]+)",
			MatchesMethods:     []string{"POST"},
			RequiredResource:   "users:$1",
			RequiredAction:     "create:$1",
			RequiredScopes:     []string{"users.create"},
			Upstream: &Upstream{
				URLParsed:    mustParseURL(t, "http://localhost:1235/"),
				URL:          "http://localhost:1235/",
				StripPath:    "/bar",
				PreserveHost: true,
			},
		}
		r2 := Rule{
			ID:                 "foo2",
			Description:        "Get users rule",
			MatchesURLCompiled: mustCompileRegex(t, "/users/([0-9]+)"),
			MatchesURL:         "/users/([0-9]+)",
			MatchesMethods:     []string{"GET"},
			Mode:               "abc",
			RequiredScopes:     []string{},
			Upstream: &Upstream{
				URLParsed:    mustParseURL(t, "http://localhost:333/"),
				URL:          "http://localhost:333/",
				StripPath:    "/foo",
				PreserveHost: false,
			},
		}

		t.Run("case="+k, func(t *testing.T) {
			_, err := manager.GetRule("1")
			require.Error(t, err)

			require.NoError(t, manager.CreateRule(&r1))
			require.NoError(t, manager.CreateRule(&r2))

			result, err := manager.GetRule(r1.ID)
			require.NoError(t, err)
			assert.EqualValues(t, &r1, result)

			result, err = manager.GetRule(r2.ID)
			require.NoError(t, err)
			assert.EqualValues(t, &r2, result)

			results, err := manager.ListRules(pkg.RulesUpperLimit, 0)
			require.NoError(t, err)
			assert.Len(t, results, 2)
			assert.True(t, results[0].ID != results[1].ID)

			r1.RequiredResource = r1.RequiredResource + "abc"
			r1.RequiredAction = r1.RequiredAction + "abc"
			r1.Description = r1.Description + "abc"
			require.NoError(t, manager.UpdateRule(&r1))

			result, err = manager.GetRule(r1.ID)
			require.NoError(t, err)
			assert.EqualValues(t, &r1, result)

			require.NoError(t, manager.DeleteRule(r1.ID))

			results, err = manager.ListRules(pkg.RulesUpperLimit, 0)
			require.NoError(t, err)
			assert.Len(t, results, 1)
			assert.True(t, results[0].ID != r1.ID)
		})
	}
}

func connectToPostgres(t *testing.T, managers map[string]Manager) {
	db, err := dockertest.ConnectToTestPostgreSQL()
	if err != nil {
		t.Logf("Could not connect to database: %v", err)
		t.FailNow()
		return
	}

	s := NewSQLManager(db)
	if _, err := s.CreateSchemas(); err != nil {
		t.Logf("Could not create postgres schema: %v", err)
		t.FailNow()
		return
	}

	managers["postgres"] = s
}

func connectToMySQL(t *testing.T, managers map[string]Manager) {
	db, err := dockertest.ConnectToTestMySQL()
	if err != nil {
		t.Logf("Could not connect to database: %v", err)
		t.FailNow()
		return
	}

	s := NewSQLManager(db)
	if _, err := s.CreateSchemas(); err != nil {
		t.Logf("Could not create postgres schema: %v", err)
		t.FailNow()
		return
	}

	managers["mysql"] = s
}
