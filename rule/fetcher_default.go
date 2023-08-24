// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package rule

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/trace"
	"gocloud.dev/blob"
	_ "gocloud.dev/blob/azureblob"
	_ "gocloud.dev/blob/gcsblob"
	_ "gocloud.dev/blob/s3blob"

	"github.com/ory/x/httpx"
	"github.com/ory/x/urlx"
	"github.com/ory/x/watcherx"

	"github.com/ory/oathkeeper/driver/configuration"
	"github.com/ory/oathkeeper/internal/cloudstorage"
	"github.com/ory/oathkeeper/x"
)

var _ Fetcher = new(FetcherDefault)

type fetcherRegistry interface {
	x.RegistryLogger
	RuleRepository() Repository
	Tracer() trace.Tracer
}

type FetcherDefault struct {
	config   configuration.Provider
	registry fetcherRegistry
	hc       *http.Client
	mux      *blob.URLMux

	cache          map[string][]Rule
	cancelWatchers map[string]context.CancelFunc
	events         chan watcherx.Event

	lock sync.Mutex
}

func NewFetcherDefault(
	config configuration.Provider,
	registry fetcherRegistry,
) *FetcherDefault {
	return &FetcherDefault{
		registry: registry,
		config:   config,
		mux:      cloudstorage.NewURLMux(),
		hc: httpx.NewResilientClient(
			httpx.ResilientClientWithConnectionTimeout(15*time.Second),
			httpx.ResilientClientWithTracer(registry.Tracer()),
		).StandardClient(),
		cache:          make(map[string][]Rule),
		cancelWatchers: make(map[string]context.CancelFunc),
		events:         make(chan watcherx.Event),
	}
}

func (f *FetcherDefault) SetURLMux(mux *blob.URLMux) {
	f.mux = mux
}

func splitLocalRemoteRepos(ruleRepos []url.URL) (files []string, nonFiles []url.URL) {
	files = make([]string, 0, len(ruleRepos))
	nonFiles = make([]url.URL, 0, len(ruleRepos))
	for _, repo := range ruleRepos {
		if repo.Scheme == "file" || repo.Scheme == "" {
			files = append(files,
				filepath.Clean(
					urlx.GetURLFilePath(&repo)))
		} else {
			nonFiles = append(nonFiles, repo)
		}
	}
	return files, nonFiles
}

// watchLocalFiles watches all files that are configured in the config and are not watched already.
// It also cancels watchers for files that are no longer configured. This function is idempotent.
func (f *FetcherDefault) watchLocalFiles(ctx context.Context) {
	f.lock.Lock()

	repoChanged := false
	cancelWatchers := make(map[string]context.CancelFunc, len(f.cancelWatchers))

	localFiles, _ := splitLocalRemoteRepos(f.config.AccessRuleRepositories())
	for _, fp := range localFiles {
		if cancel, ok := f.cancelWatchers[fp]; !ok {
			// watch all files we are not yet watching
			repoChanged = true
			ctx, cancelWatchers[fp] = context.WithCancel(ctx)
			w, err := watcherx.WatchFile(ctx, fp, f.events)
			if err != nil {
				f.registry.Logger().WithError(err).WithField("file", fp).Error("Unable to watch file, ignoring it.")
				continue
			}
			// we force reading the files
			done, err := w.DispatchNow()
			if err != nil {
				f.registry.Logger().WithError(err).WithField("file", fp).Error("Unable to read file, ignoring it.")
				continue
			}
			go func() { <-done }() // we do not need to wait here, but we need to clear the channel
		} else {
			// keep watching files we are already watching
			cancelWatchers[fp] = cancel
		}
	}

	// cancel watchers for files we are no longer watching
	for fp, cancel := range f.cancelWatchers {
		if _, ok := cancelWatchers[fp]; !ok {
			f.registry.Logger().WithField("file", fp).Info("Stopped watching access rule file.")
			repoChanged = true
			cancel()

			delete(f.cache, fp)
		}
	}
	f.cancelWatchers = cancelWatchers

	f.lock.Unlock() // release lock before processing events

	if repoChanged {
		f.registry.Logger().WithField("repos", f.config.Get(configuration.AccessRuleRepositories)).Info("Detected access rule repository change, processing updates.")
		if err := f.updateRulesFromCache(ctx); err != nil {
			f.registry.Logger().WithError(err).WithField("event_source", "local repo change").Error("Unable to update access rules.")
		}
	}
}

func (f *FetcherDefault) Watch(ctx context.Context) error {
	f.watchLocalFiles(ctx)

	getRemoteRepos := func() map[url.URL]struct{} {
		_, remoteRepos := splitLocalRemoteRepos(f.config.AccessRuleRepositories())
		repos := make(map[url.URL]struct{}, len(remoteRepos))
		for _, repo := range remoteRepos {
			repos[repo] = struct{}{}
		}
		return repos
	}

	// capture the previous config values to detect changes, and trigger initial processing
	strategy := f.config.AccessRuleMatchingStrategy()
	if err := f.processStrategyUpdate(ctx, strategy); err != nil {
		return err
	}

	remoteRepos := getRemoteRepos()
	if err := f.processRemoteRepoUpdate(ctx, nil, remoteRepos); err != nil {
		return err
	}

	f.config.AddWatcher(func(_ watcherx.Event, err error) {
		if err != nil {
			return
		}
		// watch files that need to be watched
		f.watchLocalFiles(ctx)

		// update the matching strategy if it changed
		if newStrategy := f.config.AccessRuleMatchingStrategy(); newStrategy != strategy {
			f.registry.Logger().WithField("strategy", newStrategy).Info("Detected access rule matching strategy change, processing updates.")
			if err := f.processStrategyUpdate(ctx, newStrategy); err != nil {
				f.registry.Logger().WithError(err).Error("Unable to update access rule matching strategy.")
			} else {
				strategy = newStrategy
			}
		}

		// update & fetch the remote repos if they changed
		newRemoteRepos := getRemoteRepos()
		if err := f.processRemoteRepoUpdate(ctx, remoteRepos, newRemoteRepos); err != nil {
			f.registry.Logger().WithError(err).Error("Unable to update remote access rule repository config.")
		}
		remoteRepos = newRemoteRepos
	})

	go f.processLocalUpdates(ctx)
	return nil
}

func (f *FetcherDefault) processStrategyUpdate(ctx context.Context, newValue configuration.MatchingStrategy) error {
	if err := f.registry.RuleRepository().SetMatchingStrategy(ctx, newValue); err != nil {
		return err
	}
	return nil
}

func (f *FetcherDefault) processRemoteRepoUpdate(ctx context.Context, oldRepos, newRepos map[url.URL]struct{}) error {
	repoChanged := false
	for repo := range newRepos {
		if _, ok := f.cache[repo.String()]; !ok {
			repoChanged = true
			f.registry.Logger().WithField("repo", repo.String()).Info("New repo detected, fetching access rules.")

			rules, err := f.fetch(repo)
			if err != nil {
				f.registry.Logger().WithError(err).WithField("repo", repo.String()).Error("Unable to fetch access rules.")
				return err
			}
			f.cacheRules(repo.String(), rules)
		}
	}
	for repo := range oldRepos {
		if _, ok := newRepos[repo]; !ok {
			repoChanged = true
			f.registry.Logger().WithField("repo", repo.String()).Info("Repo was removed, removing access rules.")

			f.lock.Lock()
			delete(f.cache, repo.String())
			f.lock.Unlock()
		}
	}
	if repoChanged {
		if err := f.updateRulesFromCache(ctx); err != nil {
			f.registry.Logger().WithError(err).WithField("event_source", "remote change").Error("Unable to update access rules.")
			return err
		}
	}
	return nil
}

func (f *FetcherDefault) processLocalUpdates(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case e, ok := <-f.events:
			if !ok {
				// channel was closed
				return
			}

			f.registry.Logger().
				WithField("event", "fsnotify").
				WithField("file", e.Source()).
				Info("Detected file change for access rules. Triggering a reload.")

			if e.Reader() == nil {
				f.registry.Logger().WithField("file", e.Source()).Error("Unable to read access rules probably because they were deleted, skipping those.")
				continue
			}
			rules, err := f.decode(e.Reader())
			if err != nil {
				f.registry.Logger().WithField("file", e.Source()).WithError(err).Error("Unable to decode access rules, skipping those.")
				continue
			}

			f.cacheRules(e.Source(), rules)

			if err := f.updateRulesFromCache(ctx); err != nil {
				f.registry.Logger().WithError(err).WithField("event_source", "local change").Error("Unable to update access rules.")
			}
		}
	}
}

func (f *FetcherDefault) cacheRules(source string, rules []Rule) {
	f.lock.Lock()
	defer f.lock.Unlock()

	f.cache[source] = rules
}

func (f *FetcherDefault) updateRulesFromCache(ctx context.Context) error {
	f.lock.Lock()
	defer f.lock.Unlock()

	allRules := make([]Rule, 0)
	for _, rules := range f.cache {
		allRules = append(allRules, rules...)
	}
	return f.registry.RuleRepository().Set(ctx, allRules)
}

func (f *FetcherDefault) fetch(source url.URL) ([]Rule, error) {
	f.registry.Logger().
		WithField("location", source.String()).
		Debugf("Fetching access rules from given location because something changed.")

	switch source.Scheme {
	case "azblob", "gs", "s3":
		return f.fetchFromStorage(source)
	case "http", "https":
		return f.fetchRemote(source.String())
	case "inline":
		src, err := base64.StdEncoding.DecodeString(strings.Replace(source.String(), "inline://", "", 1))
		if err != nil {
			return nil, errors.Wrapf(err, "rule: %s", source.String())
		}
		return f.decode(bytes.NewBuffer(src))
	}
	return nil, errors.Errorf("rule: source url uses an unknown scheme: %s", source.String())
}

func (f *FetcherDefault) fetchRemote(source string) ([]Rule, error) {
	res, err := f.hc.Get(source)
	if err != nil {
		return nil, errors.Wrapf(err, "rule: %s", source)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, errors.Errorf("rule: expected http response status code 200 but got %d when fetching: %s", res.StatusCode, source)
	}

	return f.decode(res.Body)
}

func (f *FetcherDefault) decode(r io.Reader) ([]Rule, error) {
	b, err := io.ReadAll(r)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	var ks []Rule

	if json.Valid(b) {
		d := json.NewDecoder(bytes.NewReader(b))
		d.DisallowUnknownFields()
		if err := d.Decode(&ks); err != nil {
			return nil, errors.WithStack(err)
		}
		return ks, nil
	}

	if err := yaml.Unmarshal(b, &ks); err != nil {
		return nil, errors.WithStack(err)
	}

	return ks, nil
}

func (f *FetcherDefault) fetchFromStorage(source url.URL) ([]Rule, error) {
	ctx := context.Background()
	bucket, err := f.mux.OpenBucket(ctx, source.Scheme+"://"+source.Host)
	if err != nil {
		return nil, err
	}
	defer bucket.Close()

	r, err := bucket.NewReader(ctx, source.Path[1:], nil)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	return f.decode(r)
}
