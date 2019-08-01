package rule_test

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/viper"
	"github.com/ory/x/stringslice"
	"github.com/ory/x/viperx"

	"github.com/ory/oathkeeper/driver/configuration"
	"github.com/ory/oathkeeper/internal"
)

const testRule = `[{"id":"test-rule-5","upstream":{"preserve_host":true,"strip_path":"/api","url":"mybackend.com/api"},"match":{"url":"myproxy.com/api","methods":["GET","POST"]},"authenticators":[{"handler":"noop"},{"handler":"anonymous"}],"authorizer":{"handler":"allow"},"mutator":{"handler":"noop"}}]`

func TestFetcherWatchConfig(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(testRule))
	}))
	defer ts.Close()

	id := uuid.New().String()
	require.NoError(t, ioutil.WriteFile(filepath.Join(os.TempDir(), ".oathkeeper-"+id+".yml"), []byte(""), 0666))

	viper.Reset()
	viperx.InitializeConfig("oathkeeper-"+id, os.TempDir(), nil)
	viperx.WatchConfig(nil, nil)
	conf := internal.NewConfigurationWithDefaults()
	r := internal.NewRegistry(conf)

	go func() {
		require.NoError(t, r.RuleFetcher().Watch(context.TODO()))
	}()

	for k, tc := range []struct {
		config     string
		tmpContent string
		expectIDs  []string
		expectNone bool
	}{
		{config: ""},
		{
			config: `
access_rules:
  repositories:
  - ftp://not-valid
`,
			expectNone: true,
		},
		{
			config: `
access_rules:
  repositories:
  - file://../test/stub/rules.json
  - file://../test/stub/rules.yaml
  - invalid
  - file:///invalid/path
  - inline://W3siaWQiOiJ0ZXN0LXJ1bGUtNCIsInVwc3RyZWFtIjp7InByZXNlcnZlX2hvc3QiOnRydWUsInN0cmlwX3BhdGgiOiIvYXBpIiwidXJsIjoibXliYWNrZW5kLmNvbS9hcGkifSwibWF0Y2giOnsidXJsIjoibXlwcm94eS5jb20vYXBpIiwibWV0aG9kcyI6WyJHRVQiLCJQT1NUIl19LCJhdXRoZW50aWNhdG9ycyI6W3siaGFuZGxlciI6Im5vb3AifSx7ImhhbmRsZXIiOiJhbm9ueW1vdXMifV0sImF1dGhvcml6ZXIiOnsiaGFuZGxlciI6ImFsbG93In0sIm11dGF0b3IiOnsiaGFuZGxlciI6Im5vb3AifX1d
  - ` + ts.URL + `
`,
			expectIDs: []string{"test-rule-1", "test-rule-2", "test-rule-3", "test-rule-4", "test-rule-5", "test-rule-1-yaml"},
		},
		{
			config: `
access_rules:
  repositories:
  - file://../test/stub/rules.yaml
`,
			expectIDs: []string{"test-rule-1-yaml"},
		},
		{
			config: `
access_rules:
  repositories:
`,
		},
	} {
		t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
			require.NoError(t, ioutil.WriteFile(filepath.Join(os.TempDir(), ".oathkeeper-"+id+".yml"), []byte(tc.config), 0666))
			time.Sleep(time.Millisecond * 500)

			rules, err := r.RuleRepository().List(context.Background(), 500, 0)
			require.NoError(t, err)
			require.Len(t, rules, len(tc.expectIDs))

			ids := make([]string, len(rules))
			for k, r := range rules {
				ids[k] = r.ID
			}

			for _, id := range tc.expectIDs {
				assert.True(t, stringslice.Has(ids, id), "\nexpected: %v\nactual: %v", tc.expectIDs, ids)
			}
		})
	}
}

func TestFetcherWatchRepositoryFromFS(t *testing.T) {
	id := uuid.New().String()
	repository := path.Join(os.TempDir(), "access-rules-"+id+".json")
	require.NoError(t, ioutil.WriteFile(repository, []byte("[]"), 0666))

	require.NoError(t, ioutil.WriteFile(filepath.Join(os.TempDir(), ".oathkeeper-"+id+".yml"), []byte(`
access_rules:
  repositories:
  - file://`+repository+`
`), 0666))

	viper.Reset()
	viperx.InitializeConfig("oathkeeper-"+id, os.TempDir(), nil)
	viperx.WatchConfig(nil, nil)

	conf := internal.NewConfigurationWithDefaults()
	r := internal.NewRegistry(conf)

	go func() {
		require.NoError(t, r.RuleFetcher().Watch(context.TODO()))
	}()

	for k, tc := range []struct {
		content   string
		expectIDs []string
	}{
		{content: "[]"},
		{content: `[{"id":"1"}]`, expectIDs: []string{"1"}},
		{content: `[{"id":"1"},{"id":"2"}]`, expectIDs: []string{"1", "2"}},
		{content: `[{"id":"2"},{"id":"3"}]`, expectIDs: []string{"2", "3"}},
	} {
		t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
			require.NoError(t, ioutil.WriteFile(repository, []byte(tc.content), 0666))
			time.Sleep(time.Millisecond * 500)

			rules, err := r.RuleRepository().List(context.Background(), 500, 0)
			require.NoError(t, err)

			ids := make([]string, len(rules))
			for k, r := range rules {
				ids[k] = r.ID
			}

			require.Len(t, ids, len(tc.expectIDs))
			for _, id := range tc.expectIDs {
				assert.True(t, stringslice.Has(ids, id), "\nexpected: %v\nactual: %v", tc.expectIDs, ids)
			}
		})
	}
}

func TestFetcherWatchRepositoryFromKubernetesConfigMap(t *testing.T) {
	viper.Reset()

	// Set up temp dir and file to watch
	watchDir, err := ioutil.TempDir("", "")
	require.NoError(t, err)
	watchFile := path.Join(watchDir, "access-rules.json")

	// Configure watcher
	viper.Set(configuration.ViperKeyAccessRuleRepositories, []string{"file://"+watchFile})
	conf := internal.NewConfigurationWithDefaults()
	r := internal.NewRegistry(conf)

	// This emulates a config map update
	var configMapUpdate = func(t *testing.T, data string, cleanup func(t *testing.T)) func(t *testing.T) {
		dir, err := ioutil.TempDir("", "")
		require.NoError(t, err)

		location := path.Join(dir, uuid.New().String()+".json")
		require.NoError(t, ioutil.WriteFile(location, []byte(data), 0640))

		if cleanup != nil {
			cleanup(t)
		}
		require.NoError(t, os.Symlink(location, watchFile))

		return func(t *testing.T) {
			require.NoError(t, os.Remove(location))
			require.NoError(t, os.RemoveAll(dir))
			require.NoError(t, os.Remove(watchFile))
		}
	}

	go func() {
		require.NoError(t, r.RuleFetcher().Watch(context.TODO()))
	}()

	var cleanup func(t *testing.T)

	for k, tc := range []struct {
		content   string
		expectIDs []string
	}{
		{content: "[]"},
		{content: `[{"id":"1"}]`, expectIDs: []string{"1"}},
		{content: `[{"id":"1"},{"id":"2"}]`, expectIDs: []string{"1", "2"}},
		{content: `[{"id":"2"},{"id":"3"}]`, expectIDs: []string{"2", "3"}},
	} {
		t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
			cleanup = configMapUpdate(t, tc.content, cleanup)

			time.Sleep(time.Millisecond * 100)

			rules, err := r.RuleRepository().List(context.Background(), 500, 0)
			require.NoError(t, err)

			ids := make([]string, len(rules))
			for k, r := range rules {
				ids[k] = r.ID
			}

			require.Len(t, ids, len(tc.expectIDs))
			for _, id := range tc.expectIDs {
				assert.True(t, stringslice.Has(ids, id), "\nexpected: %v\nactual: %v", tc.expectIDs, ids)
			}
		})
	}
}
