package configuration

import (
	"encoding/json"
	"net/url"
	"time"

	"github.com/gobuffalo/packr/v2"
	"github.com/sirupsen/logrus"

	"github.com/ory/fosite"

	"github.com/rs/cors"
)

var schemas = packr.New("schemas", "../../.schemas")

type Provider interface {
	CORSEnabled(iface string) bool
	CORSOptions(iface string) cors.Options

	ProviderAuthenticators
	ProviderAuthorizers
	ProviderMutators

	ProxyReadTimeout() time.Duration
	ProxyWriteTimeout() time.Duration
	ProxyIdleTimeout() time.Duration

	AccessRuleRepositories() []url.URL

	ProxyServeAddress() string
	APIServeAddress() string

	ToScopeStrategy(value string, key string) fosite.ScopeStrategy
	ParseURLs(sources []string) ([]url.URL, error)
	JSONWebKeyURLs() []string
}

type ProviderAuthenticators interface {
	AuthenticatorConfig(id string, overrides json.RawMessage, destination interface{}) error
	AuthenticatorIsEnabled(id string) bool
}

type ProviderAuthorizers interface {
	AuthorizerConfig(id string, overrides json.RawMessage, destination interface{}) error
	AuthorizerIsEnabled(id string) bool
}

type ProviderMutators interface {
	MutatorConfig(id string, overrides json.RawMessage, destination interface{}) error
	MutatorIsEnabled(id string) bool
}

func MustValidate(l logrus.FieldLogger, p Provider) {
}
