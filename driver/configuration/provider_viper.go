package configuration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/ory/x/viperx"
	"github.com/rs/cors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/ory/fosite"
	"github.com/ory/x/corsx"
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

type ViperProvider struct {
	l logrus.FieldLogger
}

func NewViperProvider(l logrus.FieldLogger) *ViperProvider {
	return &ViperProvider{l: l}
}

func (v *ViperProvider) AccessRuleRepositories() []AccessRuleRepository {
	var w = func() []AccessRuleRepository {
		v.l.Warnf(`All request will be disallowed because no access rule repositories have been defined or the configuration is invalid. To define one or more access rule repositories please configure key "%s"`, ViperKeyAccessRuleRepositories)
		return []AccessRuleRepository{}
	}

	// This makes me really angry. Really, really angry.
	var simple []map[string]interface{}
	if err := viper.UnmarshalKey(ViperKeyAccessRuleRepositories, &simple); err != nil {
		v.l.WithError(err).Errorf(`Unable to decode configuration key "%s" to internal representation. Make sure that the configuration values are valid.`, ViperKeyAccessRuleRepositories)
		return w()
	}

	var b bytes.Buffer
	if err := json.NewEncoder(&b).Encode(simple); err != nil {
		v.l.WithError(err).Errorf(`Unable to encode configuration key "%s" to internal representation. Make sure that the configuration values are valid.`, ViperKeyAccessRuleRepositories)
		return w()
	}

	d := json.NewDecoder(&b)
	d.DisallowUnknownFields()

	var repos []AccessRuleRepository
	if err := d.Decode(&repos); err != nil {
		v.l.WithError(err).Errorf(`Unable to decode configuration key "%s" to internal representation. Make sure that the configuration values are valid.`, ViperKeyAccessRuleRepositories)
		return w()
	}

	return repos
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
