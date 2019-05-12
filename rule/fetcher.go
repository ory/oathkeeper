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

	"github.com/ory/oathkeeper/driver/configuration"

	"github.com/pkg/errors"
	"gopkg.in/square/go-jose.v2"
)

type Fetcher struct {
	c  configuration.Provider
	hc *http.Client
}

func (f *Fetcher) Watch() error {
	set := make(chan []Rule)
	errs := make(chan error)
	sources := f.c.RuleRepositoryURLs()
	for _, source := range sources {
		go f.discover(source, set, errs)
	}

	var rules []Rule
	for i := 0; i < len(sources); i++ {
		select {
		case s := <-set:
			for _, r := range s {
				rules = append(rules, r)
			}
		case err := <-errs:
			return err
		}
	}

	return nil
}

func (f *Fetcher) discover(
	source url.URL,
	set chan []Rule,
	ec chan error,
) {
	if err := f.fetch(&source, set); err != nil {
		ec <- err
	}
}

func (f *Fetcher) fetch(source *url.URL, set chan []Rule) error {
	switch source.Scheme {
	case "http":
		fallthrough
	case "https":
		if err := f.fetchRemote(source.String(), set); err != nil {
			return err
		}
	case "file":
		p := strings.Replace(source.String(), "file://", "", 1)
		if path.Ext(p) == ".json" {
			return f.fetchFile(p, set)
		}
		return f.fetchDir(p)
	case "inline":
		src, err := base64.StdEncoding.DecodeString(strings.Replace(source.String(), "inline://", "", 1))
		if err != nil {
			return errors.Wrapf(err, "rule: ")
		}
		return f.decode(bytes.NewBuffer(src), set)
		// do nothing...
	default:
		return errors.Errorf("rule: source url uses an unknown scheme: %s", source)
	}
	return nil
}

func (f *Fetcher) fetchRemote(source string, set chan []Rule) error {
	res, err := f.hc.Get(source)
	if err != nil {
		return errors.Wrapf(err, "rule: ")
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return errors.Errorf("rule: expected http response status code 200 but got %d when fetching: %s", source)
	}

	return f.decode(res.Body, set)
}

func (f *Fetcher) fetchDir(source string) error {
	return filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return errors.Wrapf(err, "rule: ")
		}
		if info.IsDir() {
			return nil
		}

		return f.fetchFile(path)
	})
}

func (f *Fetcher) fetchFile(source string, set chan []Rule) error {
	fp, err := os.Open(source)
	if err != nil {
		return errors.Wrapf(err, "rule: ")
	}
	defer fp.Close()

	return f.decode(fp, set)
}

func (f *Fetcher) decode(r io.Reader, set chan[]Rule) error {
	var ks []Rule
	if err := json.NewDecoder(r).Decode(&ks); err != nil {
		return errors.Wrapf(err, "rule: ")
	}
	set <- &ks
	return nil
}
