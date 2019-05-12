package rule

import (
	"github.com/ory/oathkeeper/driver/configuration"
	"github.com/pkg/errors"
	"net/url"
	"path"
)

type Fetcher struct {
	c configuration.Provider
}

func (f *Fetcher) Watch() error {
	for _, source := range f.c.RuleRepositoryURLs() {
		switch source.Scheme {
		case "http":
			fallthrough
		case "https":
			if err := f.watchRemote(source); err != nil {
				return err
			}
		case "file":
			if path.Ext(source.String()) == ".json" {
				if err := f.watchFile(source); err != nil {
					return err
				}
			} else {
				if err := f.watchDir(source); err != nil {
					return err
				}
			}
		case "inline":
			// do nothing...
		default:
			return errors.Errorf("rule: source url uses an unknown scheme: %s", source)
		}
	}
}

func (f *Fetcher) watchRemote(source url.URL) error {
	return nil
}

func (f *Fetcher) watchDir(source url.URL) error {
	return nil
}

func (f *Fetcher) watchFile(source url.URL) error {
	return nil
}
