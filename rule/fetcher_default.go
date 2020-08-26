package rule

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strings"
	"sync"

	"github.com/ory/x/watcherx"

	"github.com/fsnotify/fsnotify"

	"github.com/ory/x/stringslice"

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
	eventRepositoryConfigChanged eventType = iota
	eventMatchingStrategyChanged
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

	wChans map[string]watcherx.EventChannel

	lock sync.Mutex
	wg   sync.WaitGroup
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

func (f *FetcherDefault) configUpdate(ctx context.Context, replace []url.URL, events chan event) error {
	var directoriesToWatch []string
	var filesBeingWatched []string
	for _, fileToWatch := range replace {
		if fileToWatch.Scheme == "file" {
			p := filepath.Clean(strings.Replace(fileToWatch.String(), "file://", "", 1))
			filesBeingWatched = append(filesBeingWatched, p)
			directoryToWatch, _ := filepath.Split(p)
			directoriesToWatch = append(directoriesToWatch, directoryToWatch)
		}
	}
	directoriesToWatch = stringslice.Unique(directoriesToWatch)

	var updateWatcher = func(sources []string, cb func(source string) error) error {
		for _, source := range sources {
			if err := cb(source); err != nil {
				if os.IsNotExist(err) {
					f.r.Logger().WithError(err).WithField("file", source).Error("Not watching config file for changes because it does not exist. Check the configuration or restart the service if the issue persists.")
				} else if os.IsPermission(err) {
					f.r.Logger().WithError(err).WithField("file", source).Error("Not watching config file for changes because permission is denied. Check the configuration or restart the service if the issue persists.")
				} else if strings.Contains(err.Error(), "non-existent kevent") {
					// ignore this error because it is fired when removing a source that does not have a watcher which can happen and is negligible
				} else {
					return errors.Wrapf(err, "unable to modify file watcher for file: %s", source)
				}
			}
		}
		return nil
	}

	f.lock.Lock()

	// First we remove all the directories being watched
	if err := updateWatcher(f.directoriesBeingWatched, watcher.Remove); err != nil {
		f.r.Logger().WithError(err).Error("Unable to modify (remove) file watcher. If the issue persists, restart the service.")
	}

	f.directoriesBeingWatched = directoriesToWatch
	f.filesBeingWatched = filesBeingWatched

	// Next we (re-) add all the directories to watch
	if err := updateWatcher(directoriesToWatch, watcher.Add); err != nil {
		f.r.Logger().WithError(err).Error("Unable to modify (add) file watcher. If the issue persists, restart the service.")
	}

	// And we need to reset the rule cache
	f.cache = make(map[string][]Rule)
	f.lock.Unlock()

	// If there are no more sources to watch we reset the rule repository as a whole
	if len(replace) == 0 {
		f.r.Logger().WithField("repos", viper.AllSettings()).Warn("No access rule repositories have been defined in the updated config.")
		if err := f.r.RuleRepository().Set(ctx, []Rule{}); err != nil {
			return err
		}
	}

	// Let's fetch all of the repos
	for _, source := range replace {
		f.enqueueEvent(events, event{et: eventFileChanged, path: source, source: "config_update"})
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
	contentEvents := make(watcherx.EventChannel)
	if err := f.watchRepositories(ctx); err != nil {
		return err
	}

	configEvents := make(chan event)
	f.watchConfig(configEvents)

	go f.handleEvents(ctx, contentEvents, configEvents)

	return nil
}

func (f *FetcherDefault) watchRepositories(ctx context.Context) error {
	f.lock.Lock()
	defer f.lock.Unlock()

	for _, u := range f.c.AccessRuleRepositories() {
		c := make(watcherx.EventChannel)
		if err := watcherx.Watch(ctx, &u, c); err != nil {
			return err
		}
		f.wChans[u.String()] = c
	}
	return nil
}

func (f *FetcherDefault) watchConfig(c chan event) {
	var pc map[string]interface{}

	viperx.AddWatcher(func(e fsnotify.Event) error {
		if reflect.DeepEqual(pc, viper.Get(configuration.ViperKeyAccessRuleRepositories)) {
			f.r.Logger().
				Debug("Not reloading access rule repositories because configuration value has not changed.")
			return nil
		}

		f.enqueueEvent(c, event{et: eventRepositoryConfigChanged, source: "viper_watcher"})
		return nil
	})
	f.enqueueEvent(c, event{et: eventRepositoryConfigChanged, source: "entrypoint"})

	var strategy map[string]interface{}
	viperx.AddWatcher(func(e fsnotify.Event) error {
		if reflect.DeepEqual(strategy, viper.Get(configuration.ViperKeyAccessRuleMatchingStrategy)) {
			f.r.Logger().
				Debug("Not reloading access rule matching strategy because configuration value has not changed.")
			return nil
		}

		f.enqueueEvent(c, event{et: eventMatchingStrategyChanged, source: "viper_watcher"})
		return nil
	})
	f.enqueueEvent(c, event{et: eventMatchingStrategyChanged, source: "entrypoint"})
}

func (f *FetcherDefault) handleEvents(ctx context.Context, configEvents chan event) {
	for {
		cases := make([]reflect.SelectCase, 2, len(f.wChans)+2)
		cases[0] = reflect.SelectCase{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(ctx.Done()),
		}
		cases[1] = reflect.SelectCase{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(configEvents),
		}
		for _, c := range f.wChans {
			cases = append(cases, reflect.SelectCase{
				Dir:  reflect.SelectRecv,
				Chan: reflect.ValueOf(c),
			})
		}

		chosen, recv, ok := reflect.Select(cases)
		if chosen == 0 && !ok {
			// context got canceled
			return
		}

		if chosen == 1 {
			if !ok {
				// configEvents channel was closed
				f.r.Logger().Debug("The config event channel was closed")
				return
			}

			e, ok := recv.Interface().(event)
			if !ok {
				f.r.Logger().Debugf("Received unexpected event %+v of type %s on configEvents channel", recv.Interface(), recv.Type().String())
			}

			switch e.et {
			case eventRepositoryConfigChanged:
				f.r.Logger().
					WithField("event", "config_change").
					WithField("source", e.source).
					Debugf("Viper detected a configuration change, reloading config.")
				if err := f.configUpdate(ctx, f.c.AccessRuleRepositories()); err != nil {
					f.r.Logger().WithError(err).Error()
					continue
				}
			case eventMatchingStrategyChanged:
				f.r.Logger().
					WithField("event", "matching_strategy_config_change").
					WithField("source", e.source).
					Debugf("Viper detected a configuration change, updating matching strategy")
				if err := f.r.RuleRepository().SetMatchingStrategy(ctx, f.c.AccessRuleMatchingStrategy()); err != nil {
					return errors.Wrapf(err, "unable to update matching strategy")
				}
			case eventFileChanged:
				f.r.Logger().
					WithField("event", "repository_change").
					WithField("source", e.source).
					WithField("file", e.path.String()).
					Debugf("One or more access rule repositories changed, reloading access rules.")

				rules, err := f.sourceUpdate(e)
				if err != nil {
					f.r.Logger().WithError(err).
						WithField("file", e.path.String()).
						Error("Unable to update access rules from given location, changes will be ignored. Check the configuration or restart the service if the issue persists.")
					continue
				}

				if err := f.r.RuleRepository().Set(ctx, rules); err != nil {
					return errors.Wrapf(err, "unable to reset access rule repository")
				}
			}
		}

		e, ok := recv.Interface().(watcherx.Event)
		if !ok {
			// received unknow event type
		}
		f.r.Logger().
			WithField("source", "watcherx").
			WithField("file", e.Source()).
			WithField("type", fmt.Sprintf("%T", e)).
			Debugf("Detected file change in directory containing access rules. Triggering a reload.")

		switch te := e.(type) {
		case *watcherx.ChangeEvent:
			rules, err := f.decode(e.Reader())
			if err != nil {
				f.r.Logger().WithError(err).
					WithField("file", e.Source()).
					Error("Unable to update access rules from given location, changes will be ignored. Check the configuration or restart the service if the issue persists.")
				continue
			}

			if err := f.r.RuleRepository().Set(ctx, rules); err != nil {
				f.r.Logger().WithError(err).
					WithField("file", e.Source()).
					Error("Unable to reset access rule repository.")
				return
			}
		case *watcherx.RemoveEvent:
		// TODO implement rule removal in repository
		case *watcherx.ErrorEvent:
			f.r.Logger().
				WithField("file", e.Source()).
				Error(te.Error())
			continue
		}

		select {
		case <-ctx.Done():
			return
		case e, ok := <-contentEvents:
			if !ok {
				// channel was closed
				f.r.Logger().Debug("The content channel was closed")
				return
			}

			f.r.Logger().
				WithField("source", "watcherx").
				WithField("file", e.Source()).
				WithField("type", fmt.Sprintf("%T", e)).
				Debugf("Detected file change in directory containing access rules. Triggering a reload.")

			switch te := e.(type) {
			case *watcherx.ChangeEvent:
				rules, err := f.decode(e.Reader())
				if err != nil {
					f.r.Logger().WithError(err).
						WithField("file", e.Source()).
						Error("Unable to update access rules from given location, changes will be ignored. Check the configuration or restart the service if the issue persists.")
					continue
				}

				if err := f.r.RuleRepository().Set(ctx, rules); err != nil {
					f.r.Logger().WithError(err).
						WithField("file", e.Source()).
						Error("Unable to reset access rule repository.")
					return
				}
			case *watcherx.RemoveEvent:
			// TODO implement rule removal in repository
			case *watcherx.ErrorEvent:
				f.r.Logger().
					WithField("file", e.Source()).
					Error(te.Error())
				continue
			}

		case e, ok := <-configEvents:

		}
	}
}

func (f *FetcherDefault) enqueueEvent(events chan event, evt event) {
	f.wg.Add(1)
	go func() {
		defer f.wg.Done()

		events <- evt
	}()
}

func (f *FetcherDefault) fetch(source url.URL) ([]Rule, error) {
	f.r.Logger().
		WithField("location", source.String()).
		Debugf("Fetching access rules from given location because something changed.")

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
