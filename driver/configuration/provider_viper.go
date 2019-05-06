package configuration

import (
	"github.com/ory/fosite"
	"github.com/sirupsen/logrus"
	"net/url"
	"strings"
)

type ViperProvider struct {
	l logrus.FieldLogger
}

func (v *ViperProvider) getURL(value string, key string) *url.URL {
	u, err := url.ParseRequestURI(value)
	if err != nil {
		v.l.WithError(err).Errorf(`Configuration key "%s" is missing or malformed.`, key)
		return nil
	}

	return u
}

func (v *ViperProvider) toScopeStrategy(value string, key string) fosite.ScopeStrategy {
	switch strings.ToLower(value) {
	case "hierarchic":
		return fosite.HierarchicScopeStrategy
	case "exact":
		return fosite.ExactScopeStrategy
	case "wildcard":
		return fosite.WildcardScopeStrategy
	case "none":
		return nil
	default:
		v.l.Errorf(`Configuration key "%s" declares unknown scope strategy "%s", only "hierarchic", "exact", "wildcard", "none" are supported. Falling back to strategy "none".`, key, value)
		return nil
	}
}
