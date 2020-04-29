package configuration

import (
	"encoding/json"
	"net/url"
	"time"

	"github.com/gobuffalo/packr/v2"
	"github.com/sirupsen/logrus"

	"github.com/ory/fosite"
	"github.com/ory/x/tracing"

	"github.com/rs/cors"
)

var schemas = packr.New("schemas", "../../.schema")

const (
	ForbiddenStrategyErrorType = "forbidden"
)

// MatchingStrategy defines matching strategy such as Regexp or Glob.
// Empty string defaults to "regexp".
type MatchingStrategy string

// Possible matching strategies.
const (
	Regexp MatchingStrategy = "regexp"
	Glob   MatchingStrategy = "glob"
)

type Provider interface {
	CORSEnabled(iface string) bool
	CORSOptions(iface string) cors.Options

	ProviderAuthenticators
	ProviderErrorHandlers
	ProviderAuthorizers
	ProviderMutators

	ProxyReadTimeout() time.Duration
	ProxyWriteTimeout() time.Duration
	ProxyIdleTimeout() time.Duration

	AccessRuleRepositories() []url.URL
	AccessRuleMatchingStrategy() MatchingStrategy

	ProxyServeAddress() string
	APIServeAddress() string
	PrometheusServeAddress() string

	PrometheusMetricsPath() string

	ToScopeStrategy(value string, key string) fosite.ScopeStrategy
	ParseURLs(sources []string) ([]url.URL, error)
	JSONWebKeyURLs() []string

	TracingServiceName() string
	TracingProvider() string
	TracingJaegerConfig() *tracing.JaegerConfig
}

type ProviderErrorHandlers interface {
	ErrorHandlerConfig(id string, override json.RawMessage, dest interface{}) error
	ErrorHandlerIsEnabled(id string) bool
	ErrorHandlerFallbackSpecificity() []string
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
