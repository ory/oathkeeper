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
	"testing"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/oathkeeper/helper"
	"github.com/ory/oathkeeper/pkg"
	"github.com/ory/sqlcon/dockertest"
)

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
		r1 := testRules[0]
		r2 := testRules[1]
		r3 := testRules[2]

		t.Run("case="+k, func(t *testing.T) {
			_, err := manager.GetRule("1")
			require.Error(t, err)

			// Updating of a non-existent rule should throw 409
			require.EqualError(t, manager.UpdateRule(&r3), helper.ErrResourceConflict.Error())

			require.NoError(t, manager.CreateRule(&r1))
			require.NoError(t, manager.CreateRule(&r2))
			require.NoError(t, manager.CreateRule(&r3))

			result, err := manager.GetRule(r1.ID)
			require.NoError(t, err)
			assert.EqualValues(t, &r1, result)

			result, err = manager.GetRule(r2.ID)
			require.NoError(t, err)
			assert.EqualValues(t, &r2, result)

			result, err = manager.GetRule(r3.ID)
			require.NoError(t, err)
			// this makes sure that the conversion worked properly
			if string(r3.Authorizer.Config) == "{}" {
				r3.Authorizer.Config = nil
			}
			if string(r3.CredentialsIssuer.Config) == "{}" {
				r3.CredentialsIssuer.Config = nil
			}
			for k, an := range r3.Authenticators {
				if string(an.Config) == "{}" {
					r3.Authenticators[k].Config = nil
				}
			}
			assert.EqualValues(t, &r3, result)

			results, err := manager.ListRules(pkg.RulesUpperLimit, 0)
			require.NoError(t, err)
			assert.Len(t, results, 3)
			assert.True(t, results[0].ID != results[1].ID)

			r1.Authorizer = RuleHandler{Handler: "allow", Config: []byte(`{ "type": "some" }`)}
			r1.Authenticators = []RuleHandler{{Handler: "auth_none", Config: []byte(`{ "name": "foo" }`)}}
			r1.CredentialsIssuer = RuleHandler{Handler: "plain", Config: []byte(`{ "text": "anything" }`)}
			r1.Description = r1.Description + "abc"
			r1.Match.Methods = []string{"HEAD"}
			require.NoError(t, manager.UpdateRule(&r1))

			result, err = manager.GetRule(r1.ID)
			require.NoError(t, err)
			assert.EqualValues(t, &r1, result)

			require.NoError(t, manager.DeleteRule(r1.ID))

			results, err = manager.ListRules(pkg.RulesUpperLimit, 0)
			require.NoError(t, err)
			assert.Len(t, results, 2)
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
