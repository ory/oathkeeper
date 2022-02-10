package configuration

import (
	"encoding/json"
	"net/url"
	"time"

	"github.com/gobuffalo/packr/v2"
	"github.com/ory/x/configx"
	"github.com/rs/cors"

	"github.com/ory/fosite"
	"github.com/ory/x/tracing"
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
	CORS(iface string) (cors.Options, bool)

	ProviderAuthenticators
	ProviderErrorHandlers
	ProviderAuthorizers
	ProviderMutators

	ProxyReadTimeout() time.Duration
	ProxyWriteTimeout() time.Duration
	ProxyIdleTimeout() time.Duration

	APIReadTimeout() time.Duration
	APIWriteTimeout() time.Duration
	APIIdleTimeout() time.Duration

	AccessRuleRepositories() []url.URL
	AccessRuleMatchingStrategy() MatchingStrategy

	ProxyServeAddress() string
	APIServeAddress() string

	PrometheusServeAddress() string
	PrometheusMetricsPath() string
	PrometheusCollapseRequestPaths() bool

	ToScopeStrategy(value string, key string) fosite.ScopeStrategy
	ParseURLs(sources []string) ([]url.URL, error)
	JSONWebKeyURLs() []string

	Tracing() *tracing.Config

	Source() *configx.Provider
}

type ProviderErrorHandlers interface {
	ErrorHandlerConfig(id string, override json.RawMessage, dest interface{}) error
	ErrorHandlerIsEnabled(id string) bool
	ErrorHandlerFallbackSpecificity() []string
}
type ProviderAuthenticators interface {
	AuthenticatorConfig(id string, overrides json.RawMessage, destination interface{}) error
	AuthenticatorIsEnabled(id string) bool
	AuthenticatorJwtJwkMaxWait() time.Duration
	AuthenticatorJwtJwkTtl() time.Duration
}

type ProviderAuthorizers interface {
	AuthorizerConfig(id string, overrides json.RawMessage, destination interface{}) error
	AuthorizerIsEnabled(id string) bool
}

type ProviderMutators interface {
	MutatorConfig(id string, overrides json.RawMessage, destination interface{}) error
	MutatorIsEnabled(id string) bool
}
