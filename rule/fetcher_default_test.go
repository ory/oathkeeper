package rule_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/x/configx"
	"github.com/ory/x/logrusx"

	"github.com/ory/x/stringslice"

	"github.com/ory/oathkeeper/driver/configuration"
	"github.com/ory/oathkeeper/internal"
	"github.com/ory/oathkeeper/internal/cloudstorage"
	"github.com/ory/oathkeeper/rule"
)

const testRule = `[{"id":"test-rule-5","upstream":{"preserve_host":true,"strip_path":"/api","url":"mybackend.com/api"},"match":{"url":"myproxy.com/api","methods":["GET","POST"]},"authenticators":[{"handler":"noop"},{"handler":"anonymous"}],"authorizer":{"handler":"allow"},"mutators":[{"handler":"noop"}]}]`
const testConfigPath = "../test/update"

func copyToFile(t *testing.T, src string, dst *os.File) {
	t.Helper()

	source, err := os.Open(filepath.Join(testConfigPath, src))
	if err != nil {
		t.Fatal(err)
	}
	defer source.Close()

	_, err = dst.Seek(0, 0)
	if err != nil {
		t.Fatal(err)
	}

	err = dst.Truncate(0)
	if err != nil {
		t.Fatal(err)
	}

	_, err = io.Copy(dst, source)
	if err != nil {
		t.Fatal(err)
	}

	// sleep some time to let the watcher pick up the changes.
	time.Sleep(100 * time.Millisecond)
}

func TestFetcherReload(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	configFile, err := os.CreateTemp(t.TempDir(), "config-*.yaml")
	require.NoError(t, err)
	t.Cleanup(func() { configFile.Close() })
	conf := internal.NewConfigurationWithDefaults(
		configx.WithContext(ctx),
		configx.WithLogger(logrusx.New("", "", logrusx.ForceLevel(logrus.TraceLevel))),
		configx.SkipValidation(),
		configx.WithConfigFiles(configFile.Name()),
	)
	r := internal.NewRegistry(conf)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(testRule))
	}))
	defer ts.Close()

	go func() { require.NoError(t, r.RuleFetcher().Watch(ctx)) }()

	// initial config without a repo and without a matching strategy
	copyToFile(t, "config_no_repo.yaml", configFile)

	rules := eventuallyListRules(ctx, t, r, 0)
	require.Empty(t, rules)

	strategy, err := r.RuleRepository().MatchingStrategy(ctx)
	require.NoError(t, err)
	require.Equal(t, configuration.Regexp, strategy)

	// config with a repo and without a matching strategy
	copyToFile(t, "config_default.yaml", configFile)

	rules = eventuallyListRules(ctx, t, r, 1)
	require.Equal(t, "test-rule-1-glob", rules[0].ID)

	strategy, err = r.RuleRepository().MatchingStrategy(ctx)
	require.NoError(t, err)
	require.Equal(t, configuration.Regexp, strategy)

	// config with a glob matching strategy
	copyToFile(t, "config_glob.yaml", configFile)

	rules = eventuallyListRules(ctx, t, r, 1)
	require.Equal(t, "test-rule-1-glob", rules[0].ID)

	strategy, err = r.RuleRepository().MatchingStrategy(ctx)
	require.NoError(t, err)
	require.Equal(t, configuration.Glob, strategy)

	// config with unknown matching strategy
	copyToFile(t, "config_error.yaml", configFile)

	rules = eventuallyListRules(ctx, t, r, 1)
	require.Equal(t, "test-rule-1-glob", rules[0].ID)

	strategy, err = r.RuleRepository().MatchingStrategy(ctx)
	require.NoError(t, err)
	require.Equal(t, "UNKNOWN", string(strategy))

	// config with regexp matching strategy
	copyToFile(t, "config_regexp.yaml", configFile)

	rules = eventuallyListRules(ctx, t, r, 1)
	require.Equal(t, "test-rule-1-glob", rules[0].ID)

	strategy, err = r.RuleRepository().MatchingStrategy(ctx)
	require.NoError(t, err)
	require.Equal(t, configuration.Regexp, strategy)
}

func TestFetcherWatchConfig(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	configFile, err := os.CreateTemp(t.TempDir(), "config-*.yaml")
	require.NoError(t, err)
	configFile.Close()
	conf := internal.NewConfigurationWithDefaults(
		configx.WithContext(ctx),
		configx.SkipValidation(),
		configx.WithLogger(logrusx.New("", "", logrusx.ForceLevel(logrus.TraceLevel))),
		configx.WithConfigFiles(configFile.Name()),
	)
	r := internal.NewRegistry(conf)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(testRule))
	}))
	t.Cleanup(ts.Close)

	require.NoError(t, os.WriteFile(configFile.Name(), []byte(""), 0666))

	go func() { require.NoError(t, r.RuleFetcher().Watch(ctx)) }()

	for k, tc := range []struct {
		config           string
		tmpContent       string
		expectIDs        []string
		expectedStrategy configuration.MatchingStrategy
	}{
		{
			config:           "",
			expectedStrategy: configuration.DefaultMatchingStrategy,
		},
		{
			config: `
access_rules:
  repositories:
  - ftp://not-valid
`,
			expectedStrategy: configuration.DefaultMatchingStrategy,
		},
		{
			config: `
access_rules:
  repositories:
  - file://../test/stub/rules.json
  - file://../test/stub/rules.yaml
  - file:///invalid/path
  - inline://W3siaWQiOiJ0ZXN0LXJ1bGUtNCIsInVwc3RyZWFtIjp7InByZXNlcnZlX2hvc3QiOnRydWUsInN0cmlwX3BhdGgiOiIvYXBpIiwidXJsIjoibXliYWNrZW5kLmNvbS9hcGkifSwibWF0Y2giOnsidXJsIjoibXlwcm94eS5jb20vYXBpIiwibWV0aG9kcyI6WyJHRVQiLCJQT1NUIl19LCJhdXRoZW50aWNhdG9ycyI6W3siaGFuZGxlciI6Im5vb3AifSx7ImhhbmRsZXIiOiJhbm9ueW1vdXMifV0sImF1dGhvcml6ZXIiOnsiaGFuZGxlciI6ImFsbG93In0sIm11dGF0b3JzIjpbeyJoYW5kbGVyIjoibm9vcCJ9XX1d
  - ` + ts.URL + "\n",
			expectedStrategy: configuration.DefaultMatchingStrategy,
			expectIDs:        []string{"test-rule-1", "test-rule-2", "test-rule-3", "test-rule-4", "test-rule-5", "test-rule-1-yaml"},
		},
		{
			config: `
access_rules:
  repositories:
    - file://../test/stub/rules.yaml
  matching_strategy: glob
`,
			expectIDs:        []string{"test-rule-1-yaml"},
			expectedStrategy: configuration.Glob,
		},
		{
			config: `
access_rules:
  repositories: []
  matching_strategy: regexp
`,
			expectedStrategy: configuration.Regexp,
		},
	} {
		t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
			require.NoError(t, os.WriteFile(configFile.Name(), []byte(tc.config), 0666))

			rules := eventuallyListRules(ctx, t, r, len(tc.expectIDs))
			strategy, err := r.RuleRepository().MatchingStrategy(ctx)
			require.NoError(t, err)
			require.Equal(t, tc.expectedStrategy, strategy)

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
	if runtime.GOOS == "windows" {
		t.Skip("Skipping watcher tests on windows")
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tempDir := t.TempDir()
	configFile, err := os.CreateTemp(tempDir, "config-*.yaml")
	require.NoError(t, err)
	t.Cleanup(func() { configFile.Close() })

	repoFile, err := os.CreateTemp(tempDir, "access-rules-*.json")
	require.NoError(t, err)
	t.Cleanup(func() { repoFile.Close() })
	repoFile.WriteString("[]")

	configFile.WriteString(fmt.Sprintf(`
access_rules:
  repositories:
  - file://%s
`, repoFile.Name()))

	conf := internal.NewConfigurationWithDefaults(
		configx.WithContext(ctx),
		configx.SkipValidation(),
		configx.WithLogger(logrusx.New("", "", logrusx.ForceLevel(logrus.TraceLevel))),
		configx.WithConfigFiles(configFile.Name()),
	)
	r := internal.NewRegistry(conf)

	go func() {
		require.NoError(t, r.RuleFetcher().Watch(ctx))
	}()

	for k, tc := range []struct {
		content   string
		expectIDs []string
	}{
		{content: "[]"},
		{content: `[{"id":"1"}]`, expectIDs: []string{"1"}},
		{content: `[{"id":"1"},{"id":"2"}]`, expectIDs: []string{"1", "2"}},
		{content: `[{"id":"2"},{"id":"3"},{"id":"4"}]`, expectIDs: []string{"2", "3", "4"}},
	} {
		t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
			repoFile.Truncate(0)
			repoFile.WriteAt([]byte(tc.content), 0)

			rules := eventuallyListRules(ctx, t, r, len(tc.expectIDs))

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

func TestFetcherWatchRepositoryFromKubernetesConfigMap(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping watcher tests on windows")
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up temp dir and file to watch
	watchDir := t.TempDir()
	watchFile := path.Join(watchDir, "access-rules.json")

	conf := internal.NewConfigurationWithDefaults(
		configx.SkipValidation(),
		configx.WithContext(ctx),
		configx.WithLogger(logrusx.New("", "", logrusx.ForceLevel(logrus.TraceLevel))),
	)
	r := internal.NewRegistry(conf)

	// Configure watcher
	conf.SetForTest(t, configuration.AccessRuleRepositories, []string{"file://" + watchFile})

	// This emulates a config map update
	// drwxr-xr-x    2 root     root          4096 Aug  1 07:42 ..2019_08_01_07_42_33.068812649
	// lrwxrwxrwx    1 root     root            31 Aug  1 07:42 ..data -> ..2019_08_01_07_42_33.068812649
	// lrwxrwxrwx    1 root     root            24 Aug  1 07:42 access-rules.json -> ..data/access-rules.json

	// time="2019-08-02T14:32:28Z" level=debug msg="Access rule watcher received an update." event=config_change source=entrypoint
	// time="2019-08-02T14:32:28Z" level=debug msg="Access rule watcher received an update." event=repository_change file="file:///etc/rules/access-rules.json" source=config_update
	// time="2019-08-02T14:32:28Z" level=debug msg="Fetching access rules from given location because something changed." location="file:///etc/rules/access-rules.json"

	// time="2019-08-02T14:33:33Z" level=debug msg="Detected file change in a watching directory." event=fsnotify file=/etc/rules/..2019_08_02_14_33_33.108628482 op=CREATE
	// time="2019-08-02T14:33:33Z" level=debug msg="Detected file change in a watching directory." event=fsnotify file=/etc/rules/..2019_08_02_14_33_33.108628482 op=CHMOD
	// time="2019-08-02T14:33:33Z" level=debug msg="Detected file change in a watching directory." event=fsnotify file=/etc/rules/..data_tmp op=RENAME
	// time="2019-08-02T14:33:33Z" level=debug msg="Detected file change in a watching directory." event=fsnotify file=/etc/rules/..data op=CREATE
	// time="2019-08-02T14:33:33Z" level=debug msg="Detected file change in a watching directory." event=fsnotify file=/etc/rules/..2019_08_02_14_32_23.285779182 op=REMOVE

	var configMapUpdate = func(t *testing.T, data string, cleanup func()) func() {

		// this is the equivalent of /etc/rules/..2019_08_01_07_42_33.068812649
		dir := path.Join(watchDir, ".."+uuid.New().String())
		require.NoError(t, os.Mkdir(dir, 0777))

		fp := path.Join(dir, "access-rules.json")
		require.NoError(t, os.WriteFile(fp, []byte(data), 0640))

		// this is the symlink: ..data -> ..2019_08_01_07_42_33.068812649
		_ = os.Rename(path.Join(watchDir, "..data"), path.Join(watchDir, "..data_tmp"))
		require.NoError(t, exec.Command("ln", "-sfn", dir, path.Join(watchDir, "..data")).Run())
		if cleanup != nil {
			cleanup()
		}

		// symlink equivalent: access-rules.json -> ..data/access-rules.json
		require.NoError(t, exec.Command("ln", "-sfn", path.Join(watchDir, "..data", "access-rules.json"), watchFile).Run())

		t.Logf("Created access rule file at: file://%s", fp)
		t.Logf("Created symbolink link at: file://%s", path.Join(watchDir, "..data"))

		return func() {
			if err := os.RemoveAll(dir); err != nil {
				panic(err)
			}
		}
	}

	var cleanup func()

	go func() {
		require.NoError(t, r.RuleFetcher().Watch(ctx))
	}()

	for i := 0; i < 10; i++ {
		t.Run(fmt.Sprintf("case=%d", i), func(t *testing.T) {
			cleanup = configMapUpdate(t, fmt.Sprintf(`[{"id":"%d"}]`, i), cleanup)

			rules := eventuallyListRules(ctx, t, r, 1)

			require.Len(t, rules, 1)
			require.Equal(t, fmt.Sprintf("%d", i), rules[0].ID)
		})
	}
}

func TestFetchRulesFromObjectStorage(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	t.Cleanup(func() {
		cloudstorage.SetCurrentTest(nil)
	})

	cloudstorage.SetCurrentTest(t)

	configFile, _ := os.CreateTemp(t.TempDir(), ".oathkeeper-*.yml")
	configFile.WriteString(`
authenticators:
  noop: { enabled: true }

access_rules:
  repositories:
  - s3://oathkeeper-test-bucket/path/prefix/rules.json
  - gs://oathkeeper-test-bucket/path/prefix/rules.json
  - azblob://path/prefix/rules.json
`)

	conf := internal.NewConfigurationWithDefaults(
		configx.SkipValidation(),
		configx.WithContext(ctx),
		configx.WithLogger(logrusx.New("", "", logrusx.ForceLevel(logrus.TraceLevel))),
		configx.WithConfigFiles(configFile.Name()),
	)
	r := internal.NewRegistry(conf)

	go func() {
		require.NoError(t, r.RuleFetcher().Watch(ctx))
	}()

	eventuallyListRules(ctx, t, r, 9)
}

func eventuallyListRules(ctx context.Context, t *testing.T, r rule.Registry, expectedLen int) (rules []rule.Rule) {
	t.Helper()
	var err error
	assert.Eventually(t, func() bool {
		rules, err = r.RuleRepository().List(ctx, 500, 0)
		require.NoError(t, err)
		return len(rules) == expectedLen
	}, 2*time.Second, 10*time.Millisecond)
	return
}
