package configuration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/imdario/mergo"
	"github.com/pkg/errors"
	"github.com/rs/cors"
	"github.com/sirupsen/logrus"

	"github.com/ory/gojsonschema"
	"github.com/ory/x/jsonx"

	"github.com/ory/viper"

	"github.com/ory/fosite"
	"github.com/ory/x/corsx"
	"github.com/ory/x/urlx"
	"github.com/ory/x/viperx"

	"github.com/ory/oathkeeper/x"
)

var _ Provider = new(ViperProvider)

const (
	ViperKeyProxyReadTimeout       = "serve.proxy.timeout.read"
	ViperKeyProxyWriteTimeout      = "serve.proxy.timeout.write"
	ViperKeyProxyIdleTimeout       = "serve.proxy.timeout.idle"
	ViperKeyProxyServeAddressHost  = "serve.proxy.host"
	ViperKeyProxyServeAddressPort  = "serve.proxy.port"
	ViperKeyAPIServeAddressHost    = "serve.api.host"
	ViperKeyAPIServeAddressPort    = "serve.api.port"
	ViperKeyAccessRuleRepositories = "access_rules.repositories"
)

// Authorizers
const (
	ViperKeyAuthorizerAllowIsEnabled = "authorizers.allow.enabled"

	ViperKeyAuthorizerDenyIsEnabled = "authorizers.deny.enabled"

	ViperKeyAuthorizerKetoEngineACPORYIsEnabled = "authorizers.keto_engine_acp_ory.enabled"
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
	ViperKeyAuthenticatorJWTIsEnabled = "authenticators.jwt.enabled"

	// oauth2_client_credentials
	ViperKeyAuthenticatorOAuth2ClientCredentialsIsEnabled = "authenticators.oauth2_client_credentials.enabled"

	// oauth2_token_introspection
	ViperKeyAuthenticatorOAuth2TokenIntrospectionIsEnabled = "authenticators.oauth2_introspection.enabled"

	// unauthorized
	ViperKeyAuthenticatorUnauthorizedIsEnabled = "authenticators.unauthorized.enabled"
)

type ViperProvider struct {
	l logrus.FieldLogger
}

func NewViperProvider(l logrus.FieldLogger) *ViperProvider {
	return &ViperProvider{l: l}
}

func (v *ViperProvider) AccessRuleRepositories() []url.URL {
	sources := viperx.GetStringSlice(v.l, ViperKeyAccessRuleRepositories, []string{})
	repositories := make([]url.URL, len(sources))
	for k, source := range sources {
		repositories[k] = *urlx.ParseOrFatal(v.l, source)
	}

	return repositories
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

func (v *ViperProvider) APIServeAddress() string {
	return fmt.Sprintf(
		"%s:%d",
		viperx.GetString(v.l, ViperKeyAPIServeAddressHost, ""),
		viperx.GetInt(v.l, ViperKeyAPIServeAddressPort, 4456),
	)
}

func (v *ViperProvider) ParseURLs(sources []string) ([]url.URL, error) {
	r := make([]url.URL, len(sources))
	for k, u := range sources {
		p, err := url.Parse(u)
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
	return viperx.GetBool(v.l, fmt.Sprintf("%s.%s.enabled", prefix, id), false)
}

func (v *ViperProvider) PipelineConfig(prefix, id string, override json.RawMessage, dest interface{}) error {
	// we need to create a copy for config otherwise we will accidentally override values
	config, err := x.Deepcopy(viper.GetStringMap(fmt.Sprintf("%s.%s.config", prefix, id)))
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
		if err := jsonx.NewStrictDecoder(bytes.NewBuffer(marshalled)).Decode(dest); err != nil {
			return errors.WithStack(err)
		}
	}

	rawComponentSchema, err := schemas.Find(fmt.Sprintf("pipeline/%s.%s.schema.json", prefix, id))
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

func (v *ViperProvider) AuthenticatorIsEnabled(id string) bool {
	return v.pipelineIsEnabled("authenticators", id)
}

func (v *ViperProvider) AuthenticatorConfig(id string, override json.RawMessage, dest interface{}) error {
	return v.PipelineConfig("authenticators", id, override, dest)
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
