package configuration

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/ory/x/watcherx"

	"github.com/imdario/mergo"
	"github.com/pkg/errors"
	"github.com/rs/cors"

	"github.com/ory/oathkeeper/embedx"

	"github.com/ory/fosite"
	"github.com/ory/gojsonschema"
	"github.com/ory/oathkeeper/x"
	"github.com/ory/x/configx"
	"github.com/ory/x/logrusx"
	"github.com/ory/x/tracing"
	"github.com/ory/x/urlx"
)

var _ Provider = new(ViperProvider)

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
	ViperKeyAuthorizerAllowIsEnabled            = "authorizers.allow.enabled"
	ViperKeyAuthorizerDenyIsEnabled             = "authorizers.deny.enabled"
	ViperKeyAuthorizerKetoEngineACPORYIsEnabled = "authorizers.keto_engine_acp_ory.enabled"
	ViperKeyAuthorizerRemoteIsEnabled           = "authorizers.remote.enabled"
	ViperKeyAuthorizerRemoteJSONIsEnabled       = "authorizers.remote_json.enabled"
)

// Mutators
const (
	ViperKeyMutatorCookieIsEnabled   = "mutators.cookie.enabled"
	ViperKeyMutatorHeaderIsEnabled   = "mutators.header.enabled"
	ViperKeyMutatorNoopIsEnabled     = "mutators.noop.enabled"
	ViperKeyMutatorHydratorIsEnabled = "mutators.hydrator.enabled"
	ViperKeyMutatorIDTokenIsEnabled  = "mutators.id_token.enabled"
	ViperKeyMutatorIDTokenJWKSURL    = "mutators.id_token.config.jwks_url"
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

	watchersLock sync.RWMutex
	watchers     []func(watcherx.Event)

	p *configx.Provider
}

func NewViperProvider(ctx context.Context, l *logrusx.Logger, opts ...configx.OptionModifier) (*ViperProvider, error) {
	prov := &ViperProvider{
		l:            l,
		enabledCache: make(map[uint64]bool),
		configCache:  make(map[uint64]json.RawMessage),
	}

	// TODO: check settings
	opts = append([]configx.OptionModifier{
		configx.WithStderrValidationReporter(),
		configx.OmitKeysFromTracing("dsn", "courier.smtp.connection_uri", "secrets.default", "secrets.cookie", "secrets.cipher", "client_secret"),
		configx.WithImmutables("serve", "profiling", "log"),
		configx.WithLogrusWatcher(l),
		configx.WithLogger(l),
		configx.WithContext(ctx),
		configx.AttachWatcher(func(e watcherx.Event, err error) {
			if err != nil {
				l.WithError(err).Errorf("failed to process config event: %s", e)
				return
			}
			prov.onChange(e)
		}),
		// TODO: add watch for the rules
	}, opts...)

	p, err := configx.New(ctx, []byte(embedx.ConfigSchema), opts...)
	if err != nil {
		return nil, err
	}

	prov.p = p

	l.UseConfig(p)

	if !p.SkipValidation() {
		// TODO: validate schemas
	}

	return prov, nil
}

func (v *ViperProvider) onChange(e watcherx.Event) {
	v.watchersLock.RLock()
	defer v.watchersLock.RUnlock()

	for _, fn := range v.watchers {
		fn(e)
	}
}

func (v *ViperProvider) AddWatcher(fn func(watcherx.Event)) {
	v.watchersLock.Lock()
	defer v.watchersLock.Unlock()

	v.watchers = append(v.watchers, fn)
}

func (v *ViperProvider) AccessRuleRepositories() []url.URL {
	sources := v.p.Strings(ViperKeyAccessRuleRepositories)
	repositories := make([]url.URL, len(sources))
	for k, source := range sources {
		repositories[k] = *x.ParseURLOrFatal(v.l, source)
	}

	return repositories
}

// AccessRuleMatchingStrategy returns current MatchingStrategy.
func (v *ViperProvider) AccessRuleMatchingStrategy() MatchingStrategy {
	return MatchingStrategy(v.p.String(ViperKeyAccessRuleMatchingStrategy))
}

func (v *ViperProvider) CORS(iface string) (cors.Options, bool) {
	return v.p.CORS("serve."+iface, cors.Options{
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE"},
		AllowedHeaders:   []string{"Authorization", "Content-Type", "Cookie"},
		ExposedHeaders:   []string{"Content-Type", "Set-Cookie"},
		AllowCredentials: true,
	})
}

func (v *ViperProvider) ProxyReadTimeout() time.Duration {
	return v.p.DurationF(ViperKeyProxyReadTimeout, time.Second*5)
}

func (v *ViperProvider) ProxyWriteTimeout() time.Duration {
	return v.p.DurationF(ViperKeyProxyWriteTimeout, time.Second*10)
}

func (v *ViperProvider) ProxyIdleTimeout() time.Duration {
	return v.p.DurationF(ViperKeyProxyIdleTimeout, time.Second*120)
}

func (v *ViperProvider) ProxyServeAddress() string {
	return fmt.Sprintf(
		"%s:%d",
		v.p.String(ViperKeyProxyServeAddressHost),
		v.p.IntF(ViperKeyProxyServeAddressPort, 4455),
	)
}

func (v *ViperProvider) APIReadTimeout() time.Duration {
	return v.p.DurationF(ViperKeyAPIReadTimeout, time.Second*5)
}

func (v *ViperProvider) APIWriteTimeout() time.Duration {
	return v.p.DurationF(ViperKeyAPIWriteTimeout, time.Second*10)
}

func (v *ViperProvider) APIIdleTimeout() time.Duration {
	return v.p.DurationF(ViperKeyAPIIdleTimeout, time.Second*120)
}

func (v *ViperProvider) APIServeAddress() string {
	return fmt.Sprintf(
		"%s:%d",
		v.p.String(ViperKeyAPIServeAddressHost),
		v.p.IntF(ViperKeyAPIServeAddressPort, 4456),
	)
}

func (v *ViperProvider) PrometheusServeAddress() string {
	return fmt.Sprintf(
		"%s:%d",
		v.p.String(ViperKeyPrometheusServeAddressHost),
		v.p.IntF(ViperKeyPrometheusServeAddressPort, 9000),
	)
}

func (v *ViperProvider) PrometheusMetricsPath() string {
	return v.p.StringF(ViperKeyPrometheusServeMetricsPath, "/metrics")
}

func (v *ViperProvider) PrometheusCollapseRequestPaths() bool {
	return v.p.BoolF(ViperKeyPrometheusServeCollapseRequestPaths, true)
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
	return v.p.Bool(fmt.Sprintf("%s.%s.enabled", prefix, id))
}

func (v *ViperProvider) PipelineConfig(prefix, id string, override json.RawMessage, dest interface{}) error {
	if err := v.p.Unmarshal(fmt.Sprintf("%s.%s.config", prefix, id), dest); err != nil {
		return errors.WithStack(err)
	}

	if len(override) != 0 {
		var overrideMap map[string]interface{}
		if err := json.Unmarshal(override, &overrideMap); err != nil {
			return errors.WithStack(err)
		}

		if err := mergo.Map(dest, &overrideMap, mergo.WithOverride); err != nil {
			return errors.WithStack(err)
		}
	}

	// TODO: do we really need the following checks?
	marshalled, err := json.Marshal(dest)
	if err != nil {
		return errors.WithStack(err)
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

	return nil
}

func (v *ViperProvider) ErrorHandlerConfig(id string, override json.RawMessage, dest interface{}) error {
	return v.PipelineConfig(ViperKeyErrors, id, override, dest)
}

func (v *ViperProvider) ErrorHandlerFallbackSpecificity() []string {
	return v.p.StringsF(ViperKeyErrorsFallback, []string{"json"})
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
	return v.p.DurationF(ViperKeyAuthenticatorJwtJwkMaxWait, time.Second)
}

func (v *ViperProvider) AuthenticatorJwtJwkTtl() time.Duration {
	return v.p.DurationF(ViperKeyAuthenticatorJwtJwkTtl, time.Second*30)
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
	return v.p.Strings(ViperKeyMutatorIDTokenJWKSURL)
}

func (v *ViperProvider) Tracing() *tracing.Config {
	return v.p.TracingConfig("ORY Oathkeeper")
}

func (v *ViperProvider) Source() *configx.Provider {
	return v.p
}
