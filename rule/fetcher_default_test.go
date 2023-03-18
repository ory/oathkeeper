// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package rule_test

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/x/watcherx"

	"github.com/ory/x/configx"
	"github.com/ory/x/logrusx"

	"github.com/ory/x/stringslice"

	"github.com/ory/oathkeeper/driver/configuration"
	"github.com/ory/oathkeeper/internal"
	"github.com/ory/oathkeeper/internal/cloudstorage"
	"github.com/ory/oathkeeper/rule"
)

const testRule = `[{"id":"test-rule-5","upstream":{"preserve_host":true,"strip_path":"/api","url":"https://mybackend.com/api"},"match":{"url":"myproxy.com/api","methods":["GET","POST"]},"authenticators":[{"handler":"noop"},{"handler":"anonymous"}],"authorizer":{"handler":"allow"},"mutators":[{"handler":"noop"}]}]`
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

	err = dst.Sync()
	if err != nil {
		t.Fatal(err)
	}

	// We need to sleep here because the changes need to be picked up by FetcherDefault.watch
	time.Sleep(100 * time.Millisecond)
}

func TestFetcherReload(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	configFile, err := os.CreateTemp(t.TempDir(), "config-*.yaml")
	require.NoError(t, err)
	t.Cleanup(func() { configFile.Close() })

	configChanged := make(chan struct{})

	conf := internal.NewConfigurationWithDefaults(
		configx.WithContext(ctx),
		configx.WithLogger(logrusx.New("", "", logrusx.ForceLevel(logrus.TraceLevel))),
		configx.SkipValidation(),
		configx.WithConfigFiles(configFile.Name()),
		configx.AttachWatcher(func(event watcherx.Event, err error) {
			go func() { configChanged <- struct{}{} }()
		}),
	)
	r := internal.NewRegistry(conf)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(testRule))
	}))
	defer ts.Close()

	go func() { require.NoError(t, r.RuleFetcher().Watch(ctx)) }()

	// initial config without a repo and without a matching strategy
	copyToFile(t, "config_no_repo.yaml", configFile)
	<-configChanged

	rules := eventuallyListRules(ctx, t, r, 0)
	require.Empty(t, rules)

	strategy, err := r.RuleRepository().MatchingStrategy(ctx)
	require.NoError(t, err)
	require.Equal(t, configuration.DefaultMatchingStrategy, strategy)

	// config with a repo and without a matching strategy
	copyToFile(t, "config_default.yaml", configFile)
	<-configChanged

	rules = eventuallyListRules(ctx, t, r, 1)
	require.Equal(t, "test-rule-1-glob", rules[0].ID)

	strategy, err = r.RuleRepository().MatchingStrategy(ctx)
	require.NoError(t, err)
	require.Equal(t, configuration.Regexp, strategy)

	// config with a glob matching strategy
	copyToFile(t, "config_glob.yaml", configFile)
	<-configChanged

	rules = eventuallyListRules(ctx, t, r, 1)
	require.Equal(t, "test-rule-1-glob", rules[0].ID)

	strategy, err = r.RuleRepository().MatchingStrategy(ctx)
	require.NoError(t, err)
	require.Equal(t, configuration.Glob, strategy)

	// config with unknown matching strategy
	copyToFile(t, "config_error.yaml", configFile)
	<-configChanged

	rules = eventuallyListRules(ctx, t, r, 1)
	require.Equal(t, "test-rule-1-glob", rules[0].ID)

	strategy, err = r.RuleRepository().MatchingStrategy(ctx)
	require.NoError(t, err)
	require.Equal(t, "UNKNOWN", string(strategy))

	// config with regexp matching strategy
	copyToFile(t, "config_regexp.yaml", configFile)
	<-configChanged

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
	configChanged := make(chan struct{})
	conf := internal.NewConfigurationWithDefaults(
		configx.WithContext(ctx),
		configx.SkipValidation(),
		configx.WithLogger(logrusx.New("", "", logrusx.ForceLevel(logrus.TraceLevel))),
		configx.WithConfigFiles(configFile.Name()),
		configx.AttachWatcher(func(event watcherx.Event, err error) {
			go func() { configChanged <- struct{}{} }()
		}),
	)
	// set default values for all test cases
	conf.SetForTest(t, configuration.AuthorizerAllowIsEnabled, true)
	conf.SetForTest(t, configuration.AuthorizerDenyIsEnabled, true)
	conf.SetForTest(t, configuration.AuthenticatorNoopIsEnabled, true)
	conf.SetForTest(t, configuration.AuthenticatorAnonymousIsEnabled, true)
	conf.SetForTest(t, configuration.MutatorNoopIsEnabled, true)
	conf.SetForTest(t, configuration.MutatorHeaderIsEnabled, true)
	conf.SetForTest(t, configuration.MutatorIDTokenIsEnabled, true)
	conf.SetForTest(t, configuration.MutatorIDTokenJWKSURL, "https://stub/.well-known/jwks.json")
	conf.SetForTest(t, configuration.MutatorIDTokenIssuerURL, "https://stub")

	r := internal.NewRegistry(conf)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(testRule))
	}))
	t.Cleanup(ts.Close)

	require.NoError(t, os.WriteFile(configFile.Name(), []byte(""), 0666))

	require.NoError(t, r.RuleFetcher().Watch(ctx))

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
			config: fmt.Sprintf(`
access_rules:
  repositories:
  - file://../test/stub/rules.json
  - file://../test/stub/rules.yaml
  - file:///invalid/path
  - inline://%s
  - %s
`, base64.StdEncoding.EncodeToString([]byte(`- id: test-rule-4
  upstream:
    preserve_host: true
    strip_path: "/api"
    url: https://mybackend.com/api
  match:
    url: myproxy.com/api
    methods:
    - GET
    - POST
  authenticators:
  - handler: noop
  - handler: anonymous
  authorizer:
    handler: allow
  mutators:
  - handler: noop
`)), ts.URL),
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
			<-configChanged

			rules := eventuallyListRules(ctx, t, r, len(tc.expectIDs))
			strategy, err := r.RuleRepository().MatchingStrategy(ctx)
			require.NoError(t, err)
			require.Equal(t, tc.expectedStrategy, strategy)

			ids := make([]string, len(rules))
			for k, r := range rules {
				ids[k] = r.ID
			}

			assert.ElementsMatch(t, ids, tc.expectIDs)
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
	require.NoError(t, repoFile.Sync())

	configFile.WriteString(fmt.Sprintf(`
access_rules:
  repositories:
  - file://%s
`, repoFile.Name()))
	require.NoError(t, configFile.Sync())

	conf := internal.NewConfigurationWithDefaults(
		configx.WithContext(ctx),
		configx.SkipValidation(),
		configx.WithLogger(logrusx.New("", "", logrusx.ForceLevel(logrus.TraceLevel))),
		configx.WithConfigFiles(configFile.Name()),
	)
	conf.SetForTest(t, configuration.AuthenticatorNoopIsEnabled, true)
	conf.SetForTest(t, configuration.AuthorizerAllowIsEnabled, true)
	conf.SetForTest(t, configuration.MutatorNoopIsEnabled, true)

	r := internal.NewRegistry(conf)

	go func() {
		require.NoError(t, r.RuleFetcher().Watch(ctx))
	}()

	const rulePattern = `{
  "id": "%s",
  "upstream": {
    "preserve_host": true,
    "strip_path": "/api",
    "url": "https://a"
  },
  "match": {
    "url": "a",
    "methods": ["GET"]
  },
  "authenticators": [{"handler": "noop"}],
  "authorizer": {"handler": "allow"},
  "mutators": [{"handler": "noop"}]
}`
	for k, tc := range []struct {
		ids []string
	}{
		{},
		{ids: []string{"1"}},
		{ids: []string{"1", "2"}},
		{ids: []string{"2", "3", "4"}},
	} {
		t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
			rawRules := make([]string, len(tc.ids))
			for k, id := range tc.ids {
				rawRules[k] = fmt.Sprintf(rulePattern, id)
			}
			content := fmt.Sprintf("[%s]", strings.Join(rawRules, ","))

			repoFile.Truncate(0)
			repoFile.WriteAt([]byte(content), 0)
			repoFile.Sync()

			actualRules := eventuallyListRules(ctx, t, r, len(tc.ids))

			ids := make([]string, len(rawRules))
			for k, r := range actualRules {
				ids[k] = r.ID
			}

			for _, id := range tc.ids {
				assert.True(t, stringslice.Has(ids, id), "\nexpected: %v\nactual: %v", tc.ids, ids)
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
	watchFile := filepath.Join(watchDir, "access-rules.json")

	conf := internal.NewConfigurationWithDefaults(
		configx.SkipValidation(),
		configx.WithContext(ctx),
		configx.WithLogger(logrusx.New("", "", logrusx.ForceLevel(logrus.TraceLevel))),
	)
	r := internal.NewRegistry(conf)

	// Configure watcher
	conf.SetForTest(t, configuration.AccessRuleRepositories, []string{"file://" + watchFile})
	conf.SetForTest(t, configuration.AuthenticatorNoopIsEnabled, true)
	conf.SetForTest(t, configuration.AuthorizerAllowIsEnabled, true)
	conf.SetForTest(t, configuration.MutatorNoopIsEnabled, true)

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
		dir := filepath.Join(watchDir, ".."+uuid.New().String())
		require.NoError(t, os.Mkdir(dir, 0777))

		fp := filepath.Join(dir, "access-rules.json")
		require.NoError(t, os.WriteFile(fp, []byte(data), 0640))

		// this is the symlink: ..data -> ..2019_08_01_07_42_33.068812649
		_ = os.Rename(filepath.Join(watchDir, "..data"), filepath.Join(watchDir, "..data_tmp"))
		require.NoError(t, exec.Command("ln", "-sfn", dir, filepath.Join(watchDir, "..data")).Run())
		if cleanup != nil {
			cleanup()
		}

		// symlink equivalent: access-rules.json -> ..data/access-rules.json
		require.NoError(t, exec.Command("ln", "-sfn", filepath.Join(watchDir, "..data", "access-rules.json"), watchFile).Run())

		t.Logf("Created access rule file at: file://%s", fp)
		t.Logf("Created symbolink link at: file://%s", filepath.Join(watchDir, "..data"))

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
			cleanup = configMapUpdate(t, fmt.Sprintf(`[{
  "id": "%d",
  "upstream": {
    "preserve_host": true,
    "strip_path": "/api",
    "url": "https://a"
  },
  "match": {
    "url": "a",
    "methods": ["GET"]
  },
  "authenticators": [{"handler": "noop"}],
  "authorizer": {"handler": "allow"},
  "mutators": [{"handler": "noop"}]
}]`, i), cleanup)

			rules := eventuallyListRules(ctx, t, r, 1)

			require.Len(t, rules, 1)
			require.Equal(t, fmt.Sprintf("%d", i), rules[0].ID)
		})
	}
}

func TestFetchRulesFromObjectStorage(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	configFile, _ := os.CreateTemp(t.TempDir(), ".oathkeeper-*.yml")
	configFile.WriteString(`
authenticators:
  noop: { enabled: true }
  anonymous: { enabled: true }
authorizers:
  allow: { enabled: true }
  deny: { enabled: true }
mutators:
  noop: { enabled: true }
  header: { enabled: true }
  id_token: 
    enabled: true
    config:
      jwks_url: https://stub/.well-known/jwks.json
      issuer_url: https://stub

access_rules:
  repositories:
  - s3://oathkeeper-test-bucket/path/prefix/rules.json
  - gs://oathkeeper-test-bucket/path/prefix/rules.json
  - azblob://path/prefix/rules.json
`)
	require.NoError(t, configFile.Sync())

	conf := internal.NewConfigurationWithDefaults(
		configx.SkipValidation(),
		configx.WithContext(ctx),
		configx.WithLogger(logrusx.New("", "", logrusx.ForceLevel(logrus.TraceLevel))),
		configx.WithConfigFiles(configFile.Name()),
	)
	r := internal.NewRegistry(conf)
	r.RuleFetcher().(rule.URLMuxSetter).SetURLMux(cloudstorage.NewTestURLMux(t))

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
		if err != nil {
			t.Logf("Error listing rules: %+v", err)
			return false
		}
		return len(rules) == expectedLen
	}, 2*time.Second, 10*time.Millisecond)
	require.Len(t, rules, expectedLen)
	return
}
