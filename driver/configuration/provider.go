// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package configuration

import (
	"encoding/json"
	"net/url"
	"testing"
	"time"

	"github.com/rs/cors"

	"github.com/ory/fosite"
	"github.com/ory/x/configx"
	"github.com/ory/x/otelx"
)

// MatchingStrategy defines matching strategy such as Regexp or Glob.
// Empty string defaults to "regexp".
type MatchingStrategy string

// Possible matching strategies.
const (
	Regexp                  MatchingStrategy = "regexp"
	Glob                    MatchingStrategy = "glob"
	DefaultMatchingStrategy                  = Regexp
)

type Configuration interface {
	Get(k Key) interface{}
	String(k Key) string
	AllSettings() map[string]interface{}
	Source() *configx.Provider

	AddWatcher(cb callback) SubscriptionID

	CORSEnabled(iface string) bool
	CORSOptions(iface string) cors.Options
	CORS(iface string) (cors.Options, bool)

	ProxyTrustForwardedHeaders() bool

	AuthenticatorConfig(id string, overrides json.RawMessage, destination interface{}) error
	AuthenticatorIsEnabled(id string) bool
	AuthenticatorJwtJwkMaxWait() time.Duration
	AuthenticatorJwtJwkTtl() time.Duration

	ErrorHandlerConfig(id string, override json.RawMessage, dest interface{}) error
	ErrorHandlerIsEnabled(id string) bool
	ErrorHandlerFallbackSpecificity() []string

	AuthorizerConfig(id string, overrides json.RawMessage, destination interface{}) error
	AuthorizerIsEnabled(id string) bool

	MutatorConfig(id string, overrides json.RawMessage, destination interface{}) error
	MutatorIsEnabled(id string) bool

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
	PrometheusMetricsNamePrefix() string
	PrometheusHideRequestPaths() bool
	PrometheusCollapseRequestPaths() bool

	ToScopeStrategy(value string, key string) fosite.ScopeStrategy
	ParseURLs(sources []string) ([]url.URL, error)
	JSONWebKeyURLs() []string

	TracingServiceName() string
	TracingConfig() *otelx.Config

	TLSConfig(daemon string) *TLSConfig

	SetForTest(t testing.TB, key string, value interface{})
}

type Provider interface {
	Config() Configuration
}
