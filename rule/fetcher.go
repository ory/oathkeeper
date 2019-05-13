package rule

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/ory/x/httpx"

	"github.com/ory/oathkeeper/driver/configuration"
	"github.com/ory/oathkeeper/x"

	"github.com/pkg/errors"
)

type fetcherRegistry interface {
	x.RegistryLogger
	RuleRepository() Repository
	RuleValidator() Validator
}

type Fetcher struct {
	c  configuration.Provider
	r  fetcherRegistry
	hc *http.Client
}

func NewFetcher(
	c configuration.Provider,
	r fetcherRegistry,
) *Fetcher {
	return &Fetcher{
		r: r, c: c,
		hc: httpx.NewResilientClientLatencyToleranceHigh(nil),
	}
}



func (f *Fetcher) Watch() error {
	var rules []Rule
	for _, source := range f.c.RuleRepositoryURLs() {
		interim, err := f.fetch(source)
		if err != nil {
			f.r.Logger().WithError(err).Errorf("Skipping access rules from repository because fetching failed: %s", source)
			continue
		}

		for _, rule := range interim {
			if err := f.r.RuleValidator().Validate(&rule); err != nil {
				f.r.Logger().WithError(err).Errorf("Skipping access rule because validation failed: %s", rule.ID)
				continue
			}
			rules = append(rules, rule)
		}
	}

	return f.r.RuleRepository().Set(context.Background(), rules)
}

func (f *Fetcher) fetch(source url.URL) ([]Rule, error) {
	switch source.Scheme {
	case "http":
		fallthrough
	case "https":
		return f.fetchRemote(source.String())
	case "file":
		p := strings.Replace(source.String(), "file://", "", 1)
		if path.Ext(p) == ".json" {
			return f.fetchFile(p)
		}
		return f.fetchDir(p)
	case "inline":
		src, err := base64.StdEncoding.DecodeString(strings.Replace(source.String(), "inline://", "", 1))
		if err != nil {
			return nil, errors.Wrapf(err, "rule: ")
		}
		return f.decode(bytes.NewBuffer(src))
	}
	return nil, errors.Errorf("rule: source url uses an unknown scheme: %s", source)
}

func (f *Fetcher) fetchRemote(source string) ([]Rule, error) {
	res, err := f.hc.Get(source)
	if err != nil {
		return nil, errors.Wrapf(err, "rule: ")
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, errors.Errorf("rule: expected http response status code 200 but got %d when fetching: %s", source)
	}

	return f.decode(res.Body)
}

func (f *Fetcher) fetchDir(source string) ([]Rule, error) {
	var rules []Rule
	if err := filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return errors.Wrapf(err, "rule: ")
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

func (f *Fetcher) fetchFile(source string) ([]Rule, error) {
	fp, err := os.Open(source)
	if err != nil {
		return nil, errors.Wrapf(err, "rule: ")
	}
	defer fp.Close()

	return f.decode(fp)
}

func (f *Fetcher) decode(r io.Reader) ([]Rule, error) {
	var ks []Rule
	if err := json.NewDecoder(r).Decode(&ks); err != nil {
		return nil, errors.Wrapf(err, "rule: ")
	}
	return ks, nil
}
