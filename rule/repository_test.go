// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package rule

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"reflect"
	"testing"

	"github.com/go-faker/faker/v4"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/x/logrusx"

	"github.com/ory/x/sqlcon/dockertest"

	"github.com/ory/oathkeeper/driver/configuration"
)

func TestMain(m *testing.M) {
	ex := dockertest.Register()
	code := m.Run()
	ex.Exit(code)
}

type validatorNoop struct {
	ret error
}

func (v *validatorNoop) Validate(*Rule) error {
	return v.ret
}

type mockRepositoryRegistry struct {
	v            validatorNoop
	loggerCalled int
}

func (r *mockRepositoryRegistry) RuleValidator() Validator {
	return &r.v
}
func (r *mockRepositoryRegistry) Logger() *logrusx.Logger {
	r.loggerCalled++
	return logrusx.New("", "")
}

func init() {
	err := faker.AddProvider("urlProvider", func(v reflect.Value) (interface{}, error) {
		var m any
		if rand.Intn(2) == 0 {
			m = new(Match)
		} else {
			m = new(MatchGRPC)
		}
		err := faker.FakeData(m)
		return m, err
	})
	if err != nil {
		panic(err)
	}
}

func TestRepository(t *testing.T) {
	for name, repo := range map[string]Repository{
		"memory": NewRepositoryMemory(new(mockRepositoryRegistry)),
	} {
		t.Run(fmt.Sprintf("repository=%s/case=valid rule", name), func(t *testing.T) {
			assert.Error(t, repo.ReadyChecker(new(http.Request)))

			var rules []Rule
			for i := 0; i < 4; i++ {
				var rule Rule
				require.NoError(t, faker.FakeData(&rule))
				rules = append(rules, rule)
			}

			for _, expect := range rules {
				_, err := repo.Get(context.Background(), expect.ID)
				require.Error(t, err)
			}

			inserted := make([]Rule, len(rules))
			copy(inserted, rules)
			inserted = inserted[:len(inserted)-1] // insert all elements but the last
			repo.Set(context.Background(), inserted)
			assert.NoError(t, repo.ReadyChecker(new(http.Request)))

			for _, expect := range inserted {
				got, err := repo.Get(context.Background(), expect.ID)
				require.NoError(t, err)
				assert.EqualValues(t, expect, *got)
			}

			count, err := repo.Count(context.Background())
			require.NoError(t, err)
			assert.Equal(t, len(inserted), count)

			updated := make([]Rule, len(rules))
			copy(updated, rules)
			updated = append(updated[:len(updated)-2], updated[len(updated)-1]) // insert all elements (including last) except before last
			repo.Set(context.Background(), updated)

			count, err = repo.Count(context.Background())
			require.NoError(t, err)
			assert.Equal(t, len(updated), count)

			for _, expect := range updated {
				got, err := repo.Get(context.Background(), expect.ID)
				require.NoError(t, err)
				assert.EqualValues(t, expect, *got)
			}

			_, err = repo.Get(context.Background(), rules[len(rules)-2].ID) // check if before last still exists
			require.Error(t, err)

			count, err = repo.Count(context.Background())
			require.NoError(t, err)
			assert.Equal(t, len(rules)-1, count)

			strategy, err := repo.MatchingStrategy(context.Background())
			require.NoError(t, err)
			require.Equal(t, configuration.MatchingStrategy(""), strategy)

			err = repo.SetMatchingStrategy(context.Background(), configuration.Glob)
			require.NoError(t, err)

			strategy, err = repo.MatchingStrategy(context.Background())
			require.NoError(t, err)
			require.Equal(t, configuration.Glob, strategy)

		})
	}

	expectedErr := errors.New("this is a forced test error and can be ignored")
	mr := &mockRepositoryRegistry{v: validatorNoop{ret: expectedErr}}
	for name, repo := range map[string]Repository{
		"memory": NewRepositoryMemory(mr),
	} {
		t.Run(fmt.Sprintf("repository=%s/case=invalid rule", name), func(t *testing.T) {
			var rule Rule
			require.NoError(t, faker.FakeData(&rule))
			require.NoError(t, repo.Set(context.Background(), []Rule{rule}))
			assert.Equal(t, 1, mr.loggerCalled)
			assert.Error(t, repo.ReadyChecker(new(http.Request)))
		})
	}
}
