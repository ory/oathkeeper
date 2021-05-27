package configuration

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"hash/crc64"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/imdario/mergo"
	"github.com/pkg/errors"
	"github.com/rs/cors"

	"github.com/ory/viper"
	"github.com/ory/x/logrusx"
	"github.com/ory/x/urlx"

	"github.com/ory/go-convenience/stringsx"

	"github.com/ory/fosite"
	"github.com/ory/gojsonschema"
	"github.com/ory/x/corsx"
	"github.com/ory/x/tracing"
	"github.com/ory/x/viperx"

	"github.com/ory/oathkeeper/x"
)

var _ Provider = new(ViperProvider)

func init() {
	// The JSON error handler is the default error handler and must be enabled by default.
	viper.SetDefault(ViperKeyErrorsJSONIsEnabled, true)
}

const (
	ViperKeyProxyReadTimeout                    = "serve.proxy.timeout.read"
	ViperKeyProxyWriteTimeout                   = "serve.proxy.timeout.write"
	ViperKeyProxyIdleTimeout                    = "serve.proxy.timeout.idle"
	ViperKeyProxyServeAddressHost               = "serve.proxy.host"
	ViperKeyProxyServeAddressPort               = "serve.proxy.port"
	ViperKeyAPIServeAddressHost                 = "serve.api.host"
	ViperKeyAPIServeAddressPort                 = "serve.api.port"
	ViperKeyAPIReadTimeout                      = "serve.api.timeout.read"
	ViperKeyAPIWriteTimeout                     = "serve.api.timeout.write"
	ViperKeyAPIIdleTimeout                      = "serve.api.timeout.idle"
	ViperKeyPrometheusServeAddressHost          = "serve.prometheus.host"
	ViperKeyPrometheusServeAddressPort          = "serve.prometheus.port"
	ViperKeyPrometheusServeMetricsPath          = "serve.prometheus.metrics_path"
	ViperKeyPrometheusServeCollapseRequestPaths = "serve.prometheus.collapse_request_paths"
	ViperKeyAccessRuleRepositories              = "access_rules.repositories"
	ViperKeyAccessRuleMatchingStrategy          = "access_rules.matching_strategy"
)

// Authorizers
const (
	ViperKeyAuthorizerAllowIsEnabled = "authorizers.allow.enabled"

	ViperKeyAuthorizerDenyIsEnabled = "authorizers.deny.enabled"

	ViperKeyAuthorizerKetoEngineACPORYIsEnabled = "authorizers.keto_engine_acp_ory.enabled"

	ViperKeyAuthorizerRemoteIsEnabled = "authorizers.remote.enabled"

	ViperKeyAuthorizerRemoteJSONIsEnabled = "authorizers.remote_json.enabled"
)

// Mutators
const (
	ViperKeyMutatorCookieIsEnabled = "mutators.cookie.enabled"

	ViperKeyMutatorHeaderIsEnabled = "mutators.header.enabled"

	ViperKeyMutatorNoopIsEnabled = "mutators.noop.enabled"

	ViperKeyMutatorHydratorIsEnabled = "mutators.hydrator.enabled"

	ViperKeyMutatorIDTokenIsEnabled = "mutators.id_token.enabled"
	ViperKeyMutatorIDTokenJWKSURL   = "mutators.id_token.config.jwks_url"
)

// Authenticators
const (
	// anonymous
	ViperKeyAuthenticatorAnonymousIsEnabled = "authenticators.anonymous.enabled"

	// noop
	ViperKeyAuthenticatorNoopIsEnabled = "authenticators.noop.enabled"

	// cookie session
	ViperKeyAuthenticatorCookieSessionIsEnabled = "authenticators.cookie_session.enabled"

	// jwt
	ViperKeyAuthenticatorJwtIsEnabled  = "authenticators.jwt.enabled"
	ViperKeyAuthenticatorJwtJwkMaxWait = "authenticators.jwt.config.jwks_max_wait"
	ViperKeyAuthenticatorJwtJwkTtl     = "authenticators.jwt.config.jwks_ttl"

	// oauth2_client_credentials
	ViperKeyAuthenticatorOAuth2ClientCredentialsIsEnabled = "authenticators.oauth2_client_credentials.enabled"

	// oauth2_token_introspection
	ViperKeyAuthenticatorOAuth2TokenIntrospectionIsEnabled = "authenticators.oauth2_introspection.enabled"

	// unauthorized
	ViperKeyAuthenticatorUnauthorizedIsEnabled = "authenticators.unauthorized.enabled"
)

// Errors
const (
	ViperKeyErrors                         = "errors.handlers"
	ViperKeyErrorsFallback                 = "errors.fallback"
	ViperKeyErrorsJSONIsEnabled            = ViperKeyErrors + ".json.enabled"
	ViperKeyErrorsRedirectIsEnabled        = ViperKeyErrors + ".redirect.enabled"
	ViperKeyErrorsWWWAuthenticateIsEnabled = ViperKeyErrors + ".www_authenticate.enabled"
)

type ViperProvider struct {
	l *logrusx.Logger

	enabledMutex sync.RWMutex
	enabledCache map[uint64]bool

	configMutex sync.RWMutex
	configCache map[uint64]json.RawMessage
}

func NewViperProvider(l *logrusx.Logger) *ViperProvider {
	return &ViperProvider{
		l:            l,
		enabledCache: make(map[uint64]bool),
		configCache:  make(map[uint64]json.RawMessage),
	}
}

func (v *ViperProvider) AccessRuleRepositories() []url.URL {
	sources := viperx.GetStringSlice(v.l, ViperKeyAccessRuleRepositories, []string{})
	repositories := make([]url.URL, len(sources))
	for k, source := range sources {
		repositories[k] = *x.ParseURLOrFatal(v.l, source)
	}

	return repositories
}

// AccessRuleMatchingStrategy returns current MatchingStrategy.
func (v *ViperProvider) AccessRuleMatchingStrategy() MatchingStrategy {
	return MatchingStrategy(viperx.GetString(v.l, ViperKeyAccessRuleMatchingStrategy, ""))
}

func (v *ViperProvider) CORSEnabled(iface string) bool {
	return corsx.IsEnabled(v.l, "serve."+iface)
}

func (v *ViperProvider) CORSOptions(iface string) cors.Options {
	return corsx.ParseOptions(v.l, "serve."+iface)
}

func (v *ViperProvider) ProxyReadTimeout() time.Duration {
	return viperx.GetDuration(v.l, ViperKeyProxyReadTimeout, time.Second*5, "PROXY_SERVER_READ_TIMEOUT")
}

func (v *ViperProvider) ProxyWriteTimeout() time.Duration {
	return viperx.GetDuration(v.l, ViperKeyProxyWriteTimeout, time.Second*10, "PROXY_SERVER_WRITE_TIMEOUT")
}

func (v *ViperProvider) ProxyIdleTimeout() time.Duration {
	return viperx.GetDuration(v.l, ViperKeyProxyIdleTimeout, time.Second*120, "PROXY_SERVER_IDLE_TIMEOUT")
}

func (v *ViperProvider) ProxyServeAddress() string {
	return fmt.Sprintf(
		"%s:%d",
		viperx.GetString(v.l, ViperKeyProxyServeAddressHost, ""),
		viperx.GetInt(v.l, ViperKeyProxyServeAddressPort, 4455),
	)
}

func (v *ViperProvider) APIReadTimeout() time.Duration {
	return viperx.GetDuration(v.l, ViperKeyAPIReadTimeout, time.Second*5)
}

func (v *ViperProvider) APIWriteTimeout() time.Duration {
	return viperx.GetDuration(v.l, ViperKeyAPIWriteTimeout, time.Second*10)
}

func (v *ViperProvider) APIIdleTimeout() time.Duration {
	return viperx.GetDuration(v.l, ViperKeyAPIIdleTimeout, time.Second*120)
}

func (v *ViperProvider) APIServeAddress() string {
	return fmt.Sprintf(
		"%s:%d",
		viperx.GetString(v.l, ViperKeyAPIServeAddressHost, ""),
		viperx.GetInt(v.l, ViperKeyAPIServeAddressPort, 4456),
	)
}

func (v *ViperProvider) PrometheusServeAddress() string {
	return fmt.Sprintf(
		"%s:%d",
		viperx.GetString(v.l, ViperKeyPrometheusServeAddressHost, ""),
		viperx.GetInt(v.l, ViperKeyPrometheusServeAddressPort, 9000),
	)
}

func (v *ViperProvider) PrometheusMetricsPath() string {
	return viperx.GetString(v.l, ViperKeyPrometheusServeMetricsPath, "/metrics")
}

func (v *ViperProvider) PrometheusCollapseRequestPaths() bool {
	return viperx.GetBool(v.l, ViperKeyPrometheusServeCollapseRequestPaths, true)
}

func (v *ViperProvider) ParseURLs(sources []string) ([]url.URL, error) {
	r := make([]url.URL, len(sources))
	for k, u := range sources {
		p, err := urlx.Parse(u)
		if err != nil {
			return nil, err
		}
		r[k] = *p
	}

	return r, nil
}

func (v *ViperProvider) getURL(value string, key string) *url.URL {
	u, err := url.ParseRequestURI(value)
	if err != nil {
		v.l.WithError(err).Errorf(`Configuration key "%s" is missing or malformed.`, key)
		return nil
	}

	return u
}

func (v *ViperProvider) ToScopeStrategy(value string, key string) fosite.ScopeStrategy {
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

func (v *ViperProvider) pipelineIsEnabled(prefix, id string) bool {
	hash, err := v.hashPipelineConfig(prefix, id, nil)
	if err != nil {
		return false
	}

	v.enabledMutex.RLock()
	e, ok := v.enabledCache[hash]
	v.enabledMutex.RUnlock()

	if ok {
		return e
	}

	v.enabledMutex.Lock()
	v.enabledCache[hash] = viperx.GetBool(v.l, fmt.Sprintf("%s.%s.enabled", prefix, id), false)
	v.enabledMutex.Unlock()

	return v.enabledCache[hash]
}

func (v *ViperProvider) hashPipelineConfig(prefix, id string, override json.RawMessage) (uint64, error) {
	ts := viper.ConfigChangeAt().UnixNano()
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(ts))

	slices := [][]byte{
		[]byte(prefix),
		[]byte(id),
		[]byte(override),
		[]byte(b),
	}

	var hashSlices []byte
	for _, s := range slices {
		hashSlices = append(hashSlices, s...)
	}

	return crc64.Checksum(hashSlices, crc64.MakeTable(crc64.ECMA)), nil
}

func (v *ViperProvider) PipelineConfig(prefix, id string, override json.RawMessage, dest interface{}) error {
	hash, err := v.hashPipelineConfig(prefix, id, override)
	if err != nil {
		return errors.WithStack(err)
	}

	v.configMutex.RLock()
	c, ok := v.configCache[hash]
	v.configMutex.RUnlock()

	if ok {
		if dest != nil {
			if err := json.NewDecoder(bytes.NewBuffer(c)).Decode(dest); err != nil {
				return errors.WithStack(err)
			}
		}

		return nil
	}

	// we need to create a copy for config otherwise we will accidentally override values
	config, err := x.Deepcopy(viperx.GetStringMapConfig(stringsx.Splitx(fmt.Sprintf("%s.%s.config", prefix, id), ".")...))
	if err != nil {
		return errors.WithStack(err)
	}

	if len(override) != 0 {
		var overrideMap map[string]interface{}
		if err := json.Unmarshal(override, &overrideMap); err != nil {
			return errors.WithStack(err)
		}

		if err := mergo.Merge(&config, &overrideMap, mergo.WithOverride); err != nil {
			return errors.WithStack(err)
		}
	}

	marshalled, err := json.Marshal(config)
	if err != nil {
		return errors.WithStack(err)
	}

	if dest != nil {
		if err := json.NewDecoder(bytes.NewBuffer(marshalled)).Decode(dest); err != nil {
			return errors.WithStack(err)
		}
	}

	rawComponentSchema, err := schemas.Find(fmt.Sprintf("pipeline/%s.%s.schema.json", strings.Split(prefix, ".")[0], id))
	if err != nil {
		return errors.WithStack(err)
	}

	rawRootSchema, err := schemas.Find("config.schema.json")
	if err != nil {
		return errors.WithStack(err)
	}

	sbl := gojsonschema.NewSchemaLoader()
	if err := sbl.AddSchemas(gojsonschema.NewBytesLoader(rawRootSchema)); err != nil {
		return errors.WithStack(err)
	}

	schema, err := sbl.Compile(gojsonschema.NewBytesLoader(rawComponentSchema))
	if err != nil {
		return errors.WithStack(err)
	}

	if result, err := schema.Validate(gojsonschema.NewBytesLoader(marshalled)); err != nil {
		return errors.WithStack(err)
	} else if !result.Valid() {
		return errors.WithStack(result.Errors())
	}

	v.configMutex.Lock()
	v.configCache[hash] = marshalled
	v.configMutex.Unlock()

	return nil
}

func (v *ViperProvider) ErrorHandlerConfig(id string, override json.RawMessage, dest interface{}) error {
	return v.PipelineConfig(ViperKeyErrors, id, override, dest)
}

func (v *ViperProvider) ErrorHandlerFallbackSpecificity() []string {
	return viperx.GetStringSlice(v.l, ViperKeyErrorsFallback, []string{"json"})
}

func (v *ViperProvider) ErrorHandlerIsEnabled(id string) bool {
	return v.pipelineIsEnabled(ViperKeyErrors, id)
}

func (v *ViperProvider) AuthenticatorIsEnabled(id string) bool {
	return v.pipelineIsEnabled("authenticators", id)
}

func (v *ViperProvider) AuthenticatorConfig(id string, override json.RawMessage, dest interface{}) error {
	return v.PipelineConfig("authenticators", id, override, dest)
}

func (v *ViperProvider) AuthenticatorJwtJwkMaxWait() time.Duration {
	return viperx.GetDuration(v.l, ViperKeyAuthenticatorJwtJwkMaxWait, time.Second)
}

func (v *ViperProvider) AuthenticatorJwtJwkTtl() time.Duration {
	return viperx.GetDuration(v.l, ViperKeyAuthenticatorJwtJwkTtl, time.Second*30)
}

func (v *ViperProvider) AuthorizerIsEnabled(id string) bool {
	return v.pipelineIsEnabled("authorizers", id)
}

func (v *ViperProvider) AuthorizerConfig(id string, override json.RawMessage, dest interface{}) error {
	return v.PipelineConfig("authorizers", id, override, dest)
}

func (v *ViperProvider) MutatorIsEnabled(id string) bool {
	return v.pipelineIsEnabled("mutators", id)
}

func (v *ViperProvider) MutatorConfig(id string, override json.RawMessage, dest interface{}) error {
	return v.PipelineConfig("mutators", id, override, dest)
}

func (v *ViperProvider) JSONWebKeyURLs() []string {
	return viperx.GetStringSlice(v.l, ViperKeyMutatorIDTokenJWKSURL, []string{})
}

func (v *ViperProvider) TracingServiceName() string {
	return viperx.GetString(v.l, "tracing.service_name", "ORY Oathkeeper")
}

func (v *ViperProvider) TracingProvider() string {
	return viperx.GetString(v.l, "tracing.provider", "", "TRACING_PROVIDER")
}

func (v *ViperProvider) TracingJaegerConfig() *tracing.JaegerConfig {
	return &tracing.JaegerConfig{
		LocalAgentHostPort: viperx.GetString(v.l, "tracing.providers.jaeger.local_agent_address", "", "TRACING_PROVIDER_JAEGER_LOCAL_AGENT_ADDRESS"),
		SamplerType:        viperx.GetString(v.l, "tracing.providers.jaeger.sampling.type", "const", "TRACING_PROVIDER_JAEGER_SAMPLING_TYPE"),
		SamplerValue:       viperx.GetFloat64(v.l, "tracing.providers.jaeger.sampling.value", float64(1), "TRACING_PROVIDER_JAEGER_SAMPLING_VALUE"),
		SamplerServerURL:   viperx.GetString(v.l, "tracing.providers.jaeger.sampling.server_url", "", "TRACING_PROVIDER_JAEGER_SAMPLING_SERVER_URL"),
		Propagation: stringsx.Coalesce(
			viper.GetString("JAEGER_PROPAGATION"), // Standard Jaeger client config
			viperx.GetString(v.l, "tracing.providers.jaeger.propagation", "", "TRACING_PROVIDER_JAEGER_PROPAGATION"),
		),
	}
}
