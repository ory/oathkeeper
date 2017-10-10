package rule

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/ory/dockertest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var resources []*dockertest.Resource
var pool *dockertest.Pool

func kjillAll() {
	for _, resource := range resources {
		if err := pool.Purge(resource); err != nil {
			log.Printf("Got an error while trying to purge resource: %s", err)
		}
	}
	resources = []*dockertest.Resource{}
}

func mustCompileRegex(t *testing.T, pattern string) *regexp.Regexp {
	exp, err := regexp.Compile(pattern)
	require.NoError(t, err)
	return exp
}

func TestMain(m *testing.M) {
	code := m.Run()
	kjillAll()
	os.Exit(code)
}

func TestManagers(t *testing.T) {
	managers := map[string]Manager{
		"memory": NewMemoryManager(),
	}

	if !testing.Short() {
		connectToPostgres(t, managers)
	}

	for k, manager := range managers {

		r1 := Rule{
			ID:               "foo1",
			Description:      "Create users rule",
			MatchesPath:      mustCompileRegex(t, "/users/([0-9]+)"),
			MatchesMethods:   []string{"POST"},
			RequiredResource: "users:$1",
			RequiredAction:   "create:$1",
			RequiredScopes:   []string{"users.create"},
		}
		r2 := Rule{
			ID:                  "foo2",
			Description:         "Get users rule",
			MatchesPath:         mustCompileRegex(t, "/users/([0-9]+)"),
			MatchesMethods:      []string{"GET"},
			AllowAnonymous:      true,
			RequiredScopes:      []string{},
			BypassAuthorization: true,
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

			results, err := manager.ListRules()
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

			results, err = manager.ListRules()
			require.NoError(t, err)
			assert.Len(t, results, 1)
			assert.True(t, results[0].ID != r1.ID)
		})
	}
}

func connectToPostgres(t *testing.T, managers map[string]Manager) {
	s := NewSQLManager(connectToPostgresDB(t))
	if _, err := s.CreateSchemas(); err != nil {
		t.Logf("Could not create postgres schema: %v", err)
		t.FailNow()
		return
	}

	managers["postgres"] = s
}

func connectToPostgresDB(t *testing.T) *sqlx.DB {
	var db *sqlx.DB
	var err error
	pool, err = dockertest.NewPool("")
	if err != nil {
		t.Fatalf("Could not connect to docker: %s", err)
	}

	resource, err := pool.Run("postgres", "9.6", []string{"POSTGRES_PASSWORD=secret", "POSTGRES_DB=oathkeeper"})
	if err != nil {
		t.Fatalf("Could not start resource: %s", err)
	}

	if err = pool.Retry(func() error {
		var err error
		db, err = sqlx.Open("postgres", fmt.Sprintf("postgres://postgres:secret@localhost:%s/oathkeeper?sslmode=disable", resource.GetPort("5432/tcp")))
		if err != nil {
			return err
		}
		return db.Ping()
	}); err != nil {
		pool.Purge(resource)
		t.Fatalf("Could not connect to docker: %s", err)
	}

	resources = append(resources, resource)
	return db
}
