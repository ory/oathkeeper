// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package configuration

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/dgraph-io/ristretto"

	"github.com/google/uuid"
	"github.com/knadh/koanf/v2"
	"github.com/pkg/errors"
	"github.com/rs/cors"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/require"

	"github.com/ory/fosite"
	"github.com/ory/gojsonschema"
	"github.com/ory/x/configx"
	"github.com/ory/x/logrusx"
	"github.com/ory/x/otelx"
	"github.com/ory/x/stringsx"
	"github.com/ory/x/urlx"
	"github.com/ory/x/watcherx"

	schema "github.com/ory/oathkeeper/spec"
	"github.com/ory/oathkeeper/x"
)

type (
	KoanfProvider struct {
		source *configx.Provider
		l      *logrusx.Logger
		ctx    context.Context

		configValidationCache *ristretto.Cache[string, bool]

		subscriptions subscriptions
	}

	callback       = func(event watcherx.Event, err error)
	SubscriptionID uuid.UUID
	subscriptions  struct {
		data map[SubscriptionID]callback
		sync.RWMutex
	}
)

var _ Provider = new(KoanfProvider)

func NewKoanfProvider(ctx context.Context, flags *pflag.FlagSet, l *logrusx.Logger, opts ...configx.OptionModifier) (kp *KoanfProvider, err error) {
	maxItems := int64(5000)
	cache, _ := ristretto.NewCache(&ristretto.Config[string, bool]{
		NumCounters:        maxItems * 10,
		MaxCost:            maxItems,
		BufferItems:        64,
		Metrics:            false,
		IgnoreInternalCost: true,
		Cost: func(value bool) int64 {
			return 1
		},
	})

	kp = &KoanfProvider{
		ctx: ctx,
		l:   l,

		configValidationCache: cache,
		subscriptions:         subscriptions{data: make(map[SubscriptionID]callback)},
	}
	kp.source, err = configx.New(
		ctx,
		schema.Config,
		append(opts,
			configx.WithFlags(flags),
			configx.WithStderrValidationReporter(),
			configx.WithLogrusWatcher(l),
			configx.WithContext(ctx),
			configx.AttachWatcher(kp.configChangeHandler),
		)...,
	)
	if err != nil {
		return nil, err
	}

	l.UseConfig(kp.source)

	for k, v := range kp.source.All() {
		l.Infof("Loaded config: %v = %v", k, v)
	}

	return kp, nil
}

// Internal watcher to distribute configuration changes.
func (v *KoanfProvider) configChangeHandler(event watcherx.Event, err error) {
	v.subscriptions.RLock()
	defer v.subscriptions.RUnlock()
	for _, cb := range v.subscriptions.data {
		cb := cb
		go cb(event, err)
	}
}

// AddWatcher ensures that the callback is called when the configuration
// changes. The returned subscription can be used to remove the watcher.
func (v *KoanfProvider) AddWatcher(cb callback) SubscriptionID {
	sID := SubscriptionID(uuid.New())

	v.subscriptions.Lock()
	v.subscriptions.data[sID] = cb
	v.subscriptions.Unlock()

	return sID
}

func (v *KoanfProvider) Get(k Key) interface{} {
	return v.source.Get(k)
}
func (v *KoanfProvider) String(k Key) string {
	return v.source.String(k)
}
func (v *KoanfProvider) AllSettings() map[string]interface{} {
	return v.source.All()
}
func (v *KoanfProvider) Source() *configx.Provider {
	return v.source
}

func (v *KoanfProvider) SetForTest(t testing.TB, key string, value interface{}) {
	if original := v.source.Get(key); original != nil {
		t.Cleanup(func() { require.NoError(t, v.source.Set(key, original)) })
	} else {
		t.Cleanup(func() { v.source.Delete(key) })
	}
	require.NoError(t, v.source.Set(key, value))
}

func (v *KoanfProvider) AccessRuleRepositories() []url.URL {
	var sources []string

	// The config supports both a single string and a list of strings.
	switch val := v.source.Get(AccessRuleRepositories).(type) {
	case string:
		sources = []string{val}
	case []string:
		sources = val
	default:
		sources = v.source.Strings(AccessRuleRepositories)
	}

	repositories := make([]url.URL, len(sources))
	for k, source := range sources {
		repositories[k] = *x.ParseURLOrFatal(v.l, source)
	}

	return repositories
}

// AccessRuleMatchingStrategy returns current MatchingStrategy.
func (v *KoanfProvider) AccessRuleMatchingStrategy() MatchingStrategy {
	return MatchingStrategy(v.source.String(AccessRuleMatchingStrategy))
}

func (v *KoanfProvider) CORSEnabled(iface string) bool {
	_, enabled := v.CORS(iface)
	return enabled
}

func (v *KoanfProvider) CORSOptions(iface string) cors.Options {
	opts, _ := v.CORS(iface)
	return opts
}

func (v *KoanfProvider) CORS(iface string) (cors.Options, bool) {
	return v.source.CORS("serve."+iface, cors.Options{})
}

func (v *KoanfProvider) ProxyReadTimeout() time.Duration {
	return v.source.DurationF(ProxyReadTimeout, 5*time.Second)
}

func (v *KoanfProvider) ProxyWriteTimeout() time.Duration {
	return v.source.DurationF(ProxyWriteTimeout, 10*time.Second)
}

func (v *KoanfProvider) ProxyIdleTimeout() time.Duration {
	return v.source.DurationF(ProxyIdleTimeout, 120*time.Second)
}

func (v *KoanfProvider) ProxyServeAddress() string {
	return fmt.Sprintf("%s:%d",
		v.source.String(ProxyServeAddressHost),
		v.source.IntF(ProxyServeAddressPort, 4455),
	)
}

func (v *KoanfProvider) APIReadTimeout() time.Duration {
	return v.source.DurationF(APIReadTimeout, 5*time.Second)
}

func (v *KoanfProvider) APIWriteTimeout() time.Duration {
	return v.source.DurationF(APIWriteTimeout, 10*time.Second)
}

func (v *KoanfProvider) APIIdleTimeout() time.Duration {
	return v.source.DurationF(APIIdleTimeout, 120*time.Second)
}

func (v *KoanfProvider) APIServeAddress() string {
	return fmt.Sprintf(
		"%s:%d",
		v.source.String(APIServeAddressHost),
		v.source.IntF(APIServeAddressPort, 4456),
	)
}

func (v *KoanfProvider) PrometheusServeAddress() string {
	return fmt.Sprintf(
		"%s:%d",
		v.source.String(PrometheusServeAddressHost),
		v.source.IntF(PrometheusServeAddressPort, 9000),
	)
}

func (v *KoanfProvider) PrometheusMetricsPath() string {
	return v.source.StringF(PrometheusServeMetricsPath, "/metrics")
}

func (v *KoanfProvider) PrometheusMetricsNamePrefix() string {
	return v.source.StringF(PrometheusServeMetricsNamePrefix, "ory_oathkeeper_")
}

func (v *KoanfProvider) PrometheusCollapseRequestPaths() bool {
	return v.source.BoolF(PrometheusServeCollapseRequestPaths, true)
}

func (v *KoanfProvider) ParseURLs(sources []string) ([]url.URL, error) {
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

func (v *KoanfProvider) getURL(value string, key string) *url.URL {
	u, err := url.ParseRequestURI(value)
	if err != nil {
		v.l.WithError(err).Errorf(`Configuration key "%s" is missing or malformed.`, key)
		return nil
	}

	return u
}

func (v *KoanfProvider) ToScopeStrategy(value string, key string) fosite.ScopeStrategy {
	switch s := stringsx.SwitchExact(strings.ToLower(value)); {
	case s.AddCase("hierarchic"):
		return fosite.HierarchicScopeStrategy
	case s.AddCase("exact"):
		return fosite.ExactScopeStrategy
	case s.AddCase("wildcard"):
		return fosite.WildcardScopeStrategy
	case s.AddCase("none"):
		return nil
	default:
		v.l.Errorf(`Configuration key "%s" declares unknown scope strategy: "%s". Falling back to strategy "none".`, key, s.ToUnknownCaseErr())
		return nil
	}
}

func (v *KoanfProvider) pipelineIsEnabled(prefix, id string) bool {
	return v.source.Bool(fmt.Sprintf("%s.%s.enabled", prefix, id))
}

func (v *KoanfProvider) hashPipelineConfig(prefix, id string, marshalled []byte) string {
	slices := [][]byte{
		[]byte(prefix),
		[]byte(id),
		marshalled,
	}

	var hashSlices []byte
	for _, s := range slices {
		hashSlices = append(hashSlices, s...)
	}
	return fmt.Sprintf("%x", sha256.Sum256(hashSlices))
}

func (v *KoanfProvider) PipelineConfig(prefix, id string, override json.RawMessage, dest interface{}) error {
	if dest == nil {
		return nil
	}
	pipelineCfg := v.source.Cut(fmt.Sprintf("%s.%s.config", prefix, id))

	if len(override) != 0 {
		overrideCfg := koanf.New(".")
		if err := overrideCfg.Load(configx.NewKoanfMemory(v.ctx, override), nil); err != nil {
			return errors.WithStack(err)
		}
		pipelineCfg.Merge(overrideCfg)
	}

	marshalled, err := json.Marshal(pipelineCfg.Raw())
	if err != nil {
		return errors.WithStack(err)
	}

	hash := v.hashPipelineConfig(prefix, id, marshalled)
	item, found := v.configValidationCache.Get(hash)
	if !found || !item {
		if err = v.validatePipelineConfig(prefix, id, marshalled); err != nil {
			return errors.WithStack(err)
		}
		v.configValidationCache.Set(hash, true, 0)
	}

	if err := json.NewDecoder(bytes.NewBuffer(marshalled)).Decode(dest); err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (v *KoanfProvider) validatePipelineConfig(prefix, id string, marshalled []byte) error {
	rawComponentSchema, err := schema.FS.ReadFile(fmt.Sprintf(
		"pipeline/%s.%s.schema.json", strings.Split(prefix, ".")[0], id))
	if err != nil {
		return errors.WithStack(err)
	}

	rawRootSchema, err := schema.FS.ReadFile("config.schema.json")
	if err != nil {
		return errors.WithStack(err)
	}

	sbl := gojsonschema.NewSchemaLoader()
	if err := sbl.AddSchemas(gojsonschema.NewBytesLoader(rawRootSchema)); err != nil {
		return errors.WithStack(err)
	}

	scheme, err := sbl.Compile(gojsonschema.NewBytesLoader(rawComponentSchema))
	if err != nil {
		return errors.WithStack(err)
	}

	if result, err := scheme.Validate(gojsonschema.NewBytesLoader(marshalled)); err != nil {
		return errors.WithStack(err)
	} else if !result.Valid() {
		return errors.WithStack(result.Errors())
	}

	return nil
}

func (v *KoanfProvider) ErrorHandlerConfig(id string, override json.RawMessage, dest interface{}) error {
	return v.PipelineConfig(ErrorsHandlers, id, override, dest)
}

func (v *KoanfProvider) ErrorHandlerFallbackSpecificity() []string {
	return v.source.StringsF(ErrorsFallback, []string{"json"})
}

func (v *KoanfProvider) ErrorHandlerIsEnabled(id string) bool {
	return v.pipelineIsEnabled(ErrorsHandlers, id)
}

func (v *KoanfProvider) AuthenticatorIsEnabled(id string) bool {
	return v.pipelineIsEnabled("authenticators", id)
}

func (v *KoanfProvider) ProxyTrustForwardedHeaders() bool {
	return v.source.Bool(ProxyTrustForwardedHeaders)
}

func (v *KoanfProvider) AuthenticatorConfig(id string, override json.RawMessage, dest interface{}) error {
	return v.PipelineConfig("authenticators", id, override, dest)
}

func (v *KoanfProvider) AuthenticatorJwtJwkMaxWait() time.Duration {
	return v.source.DurationF(AuthenticatorJwtJwkMaxWait, 1*time.Second)
}

func (v *KoanfProvider) AuthenticatorJwtJwkTtl() time.Duration {
	return v.source.DurationF(AuthenticatorJwtJwkTtl, 30*time.Second)
}

func (v *KoanfProvider) AuthorizerIsEnabled(id string) bool {
	return v.pipelineIsEnabled("authorizers", id)
}

func (v *KoanfProvider) AuthorizerConfig(id string, override json.RawMessage, dest interface{}) error {
	return v.PipelineConfig("authorizers", id, override, dest)
}

func (v *KoanfProvider) MutatorIsEnabled(id string) bool {
	return v.pipelineIsEnabled("mutators", id)
}

func (v *KoanfProvider) MutatorConfig(id string, override json.RawMessage, dest interface{}) error {
	return v.PipelineConfig("mutators", id, override, dest)
}

func (v *KoanfProvider) JSONWebKeyURLs() []string {
	switch val := v.source.Get(MutatorIDTokenJWKSURL).(type) {
	case string:
		return []string{val}
	default:
		return v.source.Strings(MutatorIDTokenJWKSURL)
	}
}

func (v *KoanfProvider) TracingServiceName() string {
	return v.source.StringF("tracing.service_name", "ORY Oathkeeper")
}

func (v *KoanfProvider) TracingConfig() *otelx.Config {
	return v.source.TracingConfig(v.TracingServiceName())
}

func (v *KoanfProvider) PrometheusHideRequestPaths() bool {
	return v.source.BoolF(PrometheusServeHideRequestPaths, false)
}

type TLSConfig struct {
	Key  TLSData `mapstructure:"key"`
	Cert TLSData `mapstructure:"cert"`
}
type TLSData struct {
	Path   string `mapstructure:"path"`
	Base64 string `mapstructure:"base64"`
}

func (v *KoanfProvider) TLSConfig(daemon string) *TLSConfig {
	c := new(TLSConfig)
	if err := v.source.Unmarshal("serve."+daemon+".tls", c); err != nil {
		v.l.Logger.Warnf("Failed to unmarshal TLS config for %s: %v", daemon, err)
	}
	return c
}
