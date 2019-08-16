package configuration

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/rs/cors"
	"github.com/sirupsen/logrus"

	"github.com/ory/viper"

	"github.com/ory/fosite"
	"github.com/ory/x/corsx"
	"github.com/ory/x/urlx"
	"github.com/ory/x/viperx"
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

func BindEnvs() {
	if err := viper.BindEnv(
		ViperKeyProxyReadTimeout,
		ViperKeyProxyWriteTimeout,
		ViperKeyProxyIdleTimeout,
		ViperKeyProxyServeAddressHost,
		ViperKeyProxyServeAddressPort,
		ViperKeyAPIServeAddressHost,
		ViperKeyAPIServeAddressPort,
		ViperKeyAccessRuleRepositories,
		ViperKeyAuthenticatorAnonymousIsEnabled,
		ViperKeyAuthenticatorAnonymousIdentifier,
		ViperKeyAuthenticatorNoopIsEnabled,
		ViperKeyAuthenticatorCookieSessionIsEnabled,
		ViperKeyAuthenticatorCookieSessionCheckSessionURL,
		ViperKeyAuthenticatorCookieSessionOnly,
		ViperKeyAuthenticatorJWTIsEnabled,
		ViperKeyAuthenticatorJWTJWKSURIs,
		ViperKeyAuthenticatorJWTScopeStrategy,
		ViperKeyAuthenticatorOAuth2ClientCredentialsIsEnabled,
		ViperKeyAuthenticatorClientCredentialsTokenURL,
		ViperKeyAuthenticatorOAuth2TokenIntrospectionIsEnabled,
		ViperKeyAuthenticatorOAuth2TokenIntrospectionScopeStrategy,
		ViperKeyAuthenticatorOAuth2TokenIntrospectionIntrospectionURL,
		ViperKeyAuthenticatorOAuth2TokenIntrospectionPreAuthorizationEnabled,
		ViperKeyAuthenticatorOAuth2TokenIntrospectionPreAuthorizationClientID,
		ViperKeyAuthenticatorOAuth2TokenIntrospectionPreAuthorizationClientSecret,
		ViperKeyAuthenticatorOAuth2TokenIntrospectionPreAuthorizationScope,
		ViperKeyAuthenticatorOAuth2TokenIntrospectionPreAuthorizationTokenURL,
		ViperKeyAuthenticatorUnauthorizedIsEnabled,
		ViperKeyAuthorizerAllowIsEnabled,
		ViperKeyAuthorizerDenyIsEnabled,
		ViperKeyAuthorizerKetoEngineACPORYIsEnabled,
		ViperKeyAuthorizerKetoEngineACPORYBaseURL,
		ViperKeyMutatorCookieIsEnabled,
		ViperKeyMutatorHeaderIsEnabled,
		ViperKeyMutatorNoopIsEnabled,
		ViperKeyMutatorHydratorIsEnabled,
		ViperKeyMutatorIDTokenIsEnabled,
		ViperKeyMutatorIDTokenIssuerURL,
		ViperKeyMutatorIDTokenJWKSURL,
		ViperKeyMutatorIDTokenTTL,
	); err != nil {
		panic(err.Error())
	}
}

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
