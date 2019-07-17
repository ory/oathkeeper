package rule

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"

	"github.com/ory/x/urlx"

	"github.com/ory/viper"
	"github.com/ory/x/httpx"
	"github.com/ory/x/viperx"

	"github.com/ory/oathkeeper/driver/configuration"
	"github.com/ory/oathkeeper/x"

	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
)

type event struct {
	et     eventType
	path   url.URL
	source string
}

type eventType int

const (
	eventRepositoryConfigChange eventType = iota
	eventFileChanged
)

var _ Fetcher = new(FetcherDefault)

type fetcherRegistry interface {
	x.RegistryLogger
	RuleRepository() Repository
}

type FetcherDefault struct {
	c  configuration.Provider
	r  fetcherRegistry
	hc *http.Client

	cache map[string][]Rule

	watching []url.URL

	lock sync.Mutex
}

func NewFetcherDefault(
	c configuration.Provider,
	r fetcherRegistry,
) *FetcherDefault {
	return &FetcherDefault{
		r:     r,
		c:     c,
		hc:    httpx.NewResilientClientLatencyToleranceHigh(nil),
		cache: map[string][]Rule{},
	}
}

func (f *FetcherDefault) configUpdate(ctx context.Context, watcher *fsnotify.Watcher, replace []url.URL, events chan event) error {
	var updateWatcher = func(sources []url.URL, cb func(source string) error) error {
		for _, source := range sources {
			if source.Scheme == "file" {
				if err := cb(strings.Replace(source.String(), "file://", "", 1)); err != nil {
					if os.IsNotExist(err) {
						f.r.Logger().WithError(err).WithField("file", source.String()).Errorf("Not watching config file for changes because it does not exist. Check the configuration or restart the service if the issue persists.")
					} else if os.IsPermission(err) {
						f.r.Logger().WithError(err).WithField("file", source.String()).Errorf("Not watching config file for changes because permission is denied. Check the configuration or restart the service if the issue persists.")
					} else if strings.Contains(err.Error(), "non-existent kevent") {
						// ignore this error because it is fired when removing a source that does not have a watcher which can happen and is negligible
					} else {
						return errors.Wrapf(err, "unable to modify file watcher for file: %s", source.String())
					}
				}
			}
		}
		return nil
	}

	f.lock.Lock()
	oldSources := make([]url.URL, 0, len(f.cache))
	for k := range f.cache {
		oldSources = append(oldSources,
			*urlx.ParseOrPanic(k), // This is always valid  because only we set the cache.
		)
	}
	f.cache = make(map[string][]Rule)
	f.lock.Unlock()

	if err := updateWatcher(oldSources, watcher.Remove); err != nil {
		return err
	}

	if err := updateWatcher(replace, watcher.Add); err != nil {
		return err
	}

	if len(replace) == 0 {
		if err := f.r.RuleRepository().Set(ctx, []Rule{}); err != nil {
			return err
		}
	}
	for _, source := range replace {
		go func(s url.URL) {
			events <- event{et: eventFileChanged, path: s, source: "config_update"}
		}(source)
	}
	return nil
}

func (f *FetcherDefault) sourceUpdate(e event) ([]Rule, error) {
	if e.path.Scheme == "file" {
		u, err := url.Parse("file://" + filepath.Clean(strings.TrimPrefix(e.path.String(), "file://")))
		if err != nil {
			return nil, err
		}

		e.path = *u
	}

	rules, err := f.fetch(e.path)
	if err != nil {
		return nil, err
	}

	f.lock.Lock()
	defer f.lock.Unlock()

	f.cache[e.path.String()] = rules

	var total []Rule
	for _, items := range f.cache {
		total = append(total, items...)
	}

	return total, nil
}

func (f *FetcherDefault) Watch(ctx context.Context) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()

	events := make(chan event)
	defer close(events)
	return f.watch(ctx, watcher, events)
}

func (f *FetcherDefault) watch(ctx context.Context, watcher *fsnotify.Watcher, events chan event) error {
	viperx.AddWatcher(func(e fsnotify.Event) error {
		if !viper.HasChanged(configuration.ViperKeyAccessRuleRepositories) {
			return nil
		}

		go func() {
			events <- event{et: eventRepositoryConfigChange, source: "viper_watcher"}
		}()

		return nil
	})

	go func() {
		events <- event{et: eventRepositoryConfigChange, source: "entrypoint"}
	}()

	for {
		select {
		case e, ok := <-watcher.Events:
			if !ok {
				// channel was closed
				return nil
			}

			if e.Op&fsnotify.Remove == fsnotify.Remove {
				f.r.Logger().
					Debugf("Detected that a access rule repository file has been removed, reloading config.")
				// If a file was removed it's likely that the config changed as well - reload!
				go func() {
					events <- event{et: eventRepositoryConfigChange, source: "fsnotify_remove"}
				}()
				continue
			}

			source, err := url.Parse("file://" + e.Name)
			if err != nil {
				return errors.Wrapf(err, "unable to parse file: %s", e.Name)
			}

			f.r.Logger().
				WithField("event", "fsnotify").
				WithField("file", source.String()).
				WithField("op", e.Op.String()).
				Debugf("Detected access rule repository file change.")

			go func() {
				events <- event{et: eventFileChanged, path: *source, source: "fsnotify_update"}
			}()
		case e, ok := <-events:
			if !ok {
				// channel was closed
				return nil
			}

			switch e.et {
			case eventRepositoryConfigChange:
				f.r.Logger().
					WithField("event", "config_change").
					WithField("source", e.source).
					Debugf("Access rule watcher received an update.")
				if err := f.configUpdate(ctx, watcher, f.c.AccessRuleRepositories(), events); err != nil {
					return err
				}
			case eventFileChanged:
				f.r.Logger().
					WithField("event", "repository_change").
					WithField("source", e.source).
					WithField("file", e.path.String()).
					Debugf("Access rule watcher received an update.")

				rules, err := f.sourceUpdate(e)
				if err != nil {
					f.r.Logger().WithError(err).WithField("file", e.path.String()).Error("Unable to update access rules from given location, changes will be ignored. Check the configuration or restart the service if the issue persists.")
					continue
				}

				if err := f.r.RuleRepository().Set(ctx, rules); err != nil {
					return errors.Wrapf(err, "unable to reset access rule repository")
				}
			}
		}
	}
}

func (f *FetcherDefault) fetch(source url.URL) ([]Rule, error) {
	switch source.Scheme {
	case "http":
		fallthrough
	case "https":
		return f.fetchRemote(source.String())
	case "file":
		p := strings.Replace(source.String(), "file://", "", 1)
		if path.Ext(p) == ".json" || path.Ext(p) == ".yaml" || path.Ext(p) == ".yml" {
			return f.fetchFile(p)
		}
		return f.fetchDir(p)
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

func (f *FetcherDefault) fetchDir(source string) ([]Rule, error) {
	var rules []Rule
	if err := filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return errors.Wrapf(err, "rule: %s", source)
		}
		if info.IsDir() {
			return nil
		}

		interim, err := f.fetchFile(path)
		if err != nil {
			return err
		}

		rules = append(rules, interim...)

		return nil
	}); err != nil {
		return nil, err
	}
	return rules, nil
}

func (f *FetcherDefault) fetchFile(source string) ([]Rule, error) {
	fp, err := os.Open(source)
	if err != nil {
		return nil, errors.Wrapf(err, "rule: %s", source)
	}
	defer fp.Close()

	return f.decode(fp)
}

func (f *FetcherDefault) decode(r io.Reader) ([]Rule, error) {
	b, err := ioutil.ReadAll(r)
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
