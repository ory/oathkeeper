package rule

import (
	"bytes"
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
}

type FetcherDefault struct {
	c  configuration.Provider
	r  fetcherRegistry
	hc *http.Client
}

func NewFetcherDefault(
	c configuration.Provider,
	r fetcherRegistry,
) *FetcherDefault {
	return &FetcherDefault{
		r: r, c: c,
		hc: httpx.NewResilientClientLatencyToleranceHigh(nil),
	}
}

func (f *FetcherDefault) Fetch() ([]Rule, error) {
	var rules []Rule
	sources := f.c.AccessRuleRepositories()
	for _, source := range sources {
		interim, err := f.fetch(source)
		if err != nil {
			return nil, err
		}
		rules = append(rules, interim...)
	}

	return rules, nil
}

func (f *FetcherDefault) fetch(source url.URL) ([]Rule, error) {
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
	var ks []Rule
	d := json.NewDecoder(r)
	d.DisallowUnknownFields()
	if err := d.Decode(&ks); err != nil {
		return nil, errors.WithStack(err)
	}
	return ks, nil
}
