// Copyright 2021 Ory GmbH
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package configuration

import (
	"encoding/json"
	"net/url"
	"time"

	"github.com/gobuffalo/packr/v2"

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
