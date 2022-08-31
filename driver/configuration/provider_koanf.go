package configuration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/knadh/koanf"
	"github.com/ory/fosite"
	"github.com/ory/go-convenience/stringsx"
	"github.com/ory/gojsonschema"
	schema "github.com/ory/oathkeeper/.schema"
	"github.com/ory/oathkeeper/x"
	"github.com/ory/x/configx"
	"github.com/ory/x/logrusx"
	"github.com/ory/x/osx"
	"github.com/ory/x/tracing"
	"github.com/ory/x/urlx"
	"github.com/ory/x/watcherx"
	"github.com/pkg/errors"
	"github.com/rs/cors"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/require"
)

type (
	KoanfProvider struct {
		source *configx.Provider
		l      *logrusx.Logger
		ctx    context.Context

		enabledMutex sync.RWMutex
		enabledCache map[uint64]bool

		configMutex sync.RWMutex
		configCache map[uint64]json.RawMessage

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
	kp = &KoanfProvider{
		ctx: ctx,
		l:   l,

		enabledCache:  make(map[uint64]bool),
		configCache:   make(map[uint64]json.RawMessage),
		subscriptions: subscriptions{data: make(map[SubscriptionID]callback)},
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

	return kp, nil
}

// Internal watcher to distribute configuration changes.
func (v *KoanfProvider) configChangeHandler(event watcherx.Event, err error) {
	v.subscriptions.RLock()
	defer v.subscriptions.RUnlock()
	for _, cb := range v.subscriptions.data {
		cb := cb
		go cb(event, err) // TODO(hperl): Should we block here?
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

// RemoveWatcher removes the watcher with the given subscription ID.
func (v *KoanfProvider) RemoveWatcher(id SubscriptionID) {
	v.subscriptions.Lock()
	delete(v.subscriptions.data, id)
	v.subscriptions.Unlock()
}

func (v *KoanfProvider) Get(k Key) interface{} {
	return v.source.Get(string(k))
}
func (v *KoanfProvider) String(k Key) string {
	return v.source.String(string(k))
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
	switch val := v.source.Get(ViperKeyAccessRuleRepositories).(type) {
	case string:
		sources = []string{val}
	case []string:
		sources = val
	default:
		sources = v.source.Strings(ViperKeyAccessRuleRepositories)
	}

	repositories := make([]url.URL, len(sources))
	for k, source := range sources {
		repositories[k] = *x.ParseURLOrFatal(v.l, source)
	}

	return repositories
}

// AccessRuleMatchingStrategy returns current MatchingStrategy.
func (v *KoanfProvider) AccessRuleMatchingStrategy() MatchingStrategy {
	return MatchingStrategy(v.source.String(ViperKeyAccessRuleMatchingStrategy))
}

func (v *KoanfProvider) CORSEnabled(iface string) bool {
	_, enabled := v.source.CORS("serve."+iface, cors.Options{})
	return enabled
}

func (v *KoanfProvider) CORSOptions(iface string) cors.Options {
	opts, _ := v.source.CORS("serve."+iface, cors.Options{})
	return opts
}

func (v *KoanfProvider) CORS(iface string) (cors.Options, bool) {
	return v.source.CORS("serve."+iface, cors.Options{})
}

func (v *KoanfProvider) ProxyReadTimeout() time.Duration {
	return v.source.DurationF(ViperKeyProxyReadTimeout, 5*time.Second)
}

func (v *KoanfProvider) ProxyWriteTimeout() time.Duration {
	return v.source.DurationF(ViperKeyProxyWriteTimeout, 10*time.Second)
}

func (v *KoanfProvider) ProxyIdleTimeout() time.Duration {
	return v.source.DurationF(ViperKeyProxyIdleTimeout, 120*time.Second)
}

func (v *KoanfProvider) ProxyServeAddress() string {
	return fmt.Sprintf("%s:%d",
		v.source.String(ViperKeyProxyServeAddressHost),
		v.source.IntF(ViperKeyProxyServeAddressPort, 4455),
	)
}

func (v *KoanfProvider) APIReadTimeout() time.Duration {
	return v.source.DurationF(ViperKeyAPIReadTimeout, 5*time.Second)
}

func (v *KoanfProvider) APIWriteTimeout() time.Duration {
	return v.source.DurationF(ViperKeyAPIWriteTimeout, 10*time.Second)
}

func (v *KoanfProvider) APIIdleTimeout() time.Duration {
	return v.source.DurationF(ViperKeyAPIIdleTimeout, 120*time.Second)
}

func (v *KoanfProvider) APIServeAddress() string {
	return fmt.Sprintf(
		"%s:%d",
		v.source.String(ViperKeyAPIServeAddressHost),
		v.source.IntF(ViperKeyAPIServeAddressPort, 4456),
	)
}

func (v *KoanfProvider) PrometheusServeAddress() string {
	return fmt.Sprintf(
		"%s:%d",
		v.source.String(ViperKeyPrometheusServeAddressHost),
		v.source.IntF(ViperKeyPrometheusServeAddressPort, 9000),
	)
}

func (v *KoanfProvider) PrometheusMetricsPath() string {
	return v.source.StringF(ViperKeyPrometheusServeMetricsPath, "/metrics")
}

func (v *KoanfProvider) PrometheusMetricsNamePrefix() string {
	return v.source.StringF(ViperKeyPrometheusServeMetricsNamePrefix, "ory_oathkeeper_")
}

func (v *KoanfProvider) PrometheusCollapseRequestPaths() bool {
	return v.source.BoolF(ViperKeyPrometheusServeCollapseRequestPaths, true)
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

func (v *KoanfProvider) pipelineIsEnabled(prefix, id string) bool {
	return v.source.Bool(fmt.Sprintf("%s.%s.enabled", prefix, id))
}

func (v *KoanfProvider) PipelineConfig(prefix, id string, override json.RawMessage, dest interface{}) error {
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

	if dest != nil {
		if err := json.NewDecoder(bytes.NewBuffer(marshalled)).Decode(dest); err != nil {
			return errors.WithStack(err)
		}
	}

	rawComponentSchema, err := schema.FS.ReadFile(fmt.Sprintf("pipeline/%s.%s.schema.json", strings.Split(prefix, ".")[0], id))
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

func (v *KoanfProvider) ErrorHandlerConfig(id string, override json.RawMessage, dest interface{}) error {
	return v.PipelineConfig(ViperKeyErrors, id, override, dest)
}

func (v *KoanfProvider) ErrorHandlerFallbackSpecificity() []string {
	return v.source.StringsF(ViperKeyErrorsFallback, []string{"json"})
}

func (v *KoanfProvider) ErrorHandlerIsEnabled(id string) bool {
	return v.pipelineIsEnabled(ViperKeyErrors, id)
}

func (v *KoanfProvider) AuthenticatorIsEnabled(id string) bool {
	return v.pipelineIsEnabled("authenticators", id)
}

func (v *KoanfProvider) AuthenticatorConfig(id string, override json.RawMessage, dest interface{}) error {
	return v.PipelineConfig("authenticators", id, override, dest)
}

func (v *KoanfProvider) AuthenticatorJwtJwkMaxWait() time.Duration {
	return v.source.DurationF(ViperKeyAuthenticatorJwtJwkMaxWait, 1*time.Second)
}

func (v *KoanfProvider) AuthenticatorJwtJwkTtl() time.Duration {
	return v.source.DurationF(ViperKeyAuthenticatorJwtJwkTtl, 30*time.Second)
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
	switch val := v.source.Get(ViperKeyMutatorIDTokenJWKSURL).(type) {
	case string:
		return []string{val}
	default:
		return v.source.Strings(ViperKeyMutatorIDTokenJWKSURL)
	}
}

func (v *KoanfProvider) TracingServiceName() string {
	return v.source.StringF("tracing.service_name", "ORY Oathkeeper")
}

func (v *KoanfProvider) TracingProvider() string {
	return stringsx.Coalesce(
		v.source.String("tracing.provider"),
		os.Getenv("TRACING_PROVIDER"),
	)
}

func (v *KoanfProvider) PrometheusHideRequestPaths() bool {
	return v.source.BoolF(ViperKeyPrometheusServeHideRequestPaths, false)
}

func (v *KoanfProvider) TracingJaegerConfig() *tracing.JaegerConfig {
	var samplingValue float64
	var err error
	if val := v.source.Get("tracing.providers.jaeger.sampling.value"); val != nil {
		samplingValue = v.source.Float64("tracing.providers.jaeger.sampling.value")
	} else {
		def := osx.GetenvDefault("TRACING_PROVIDER_JAEGER_SAMPLING_VALUE", "1")
		samplingValue, err = strconv.ParseFloat(def, 64)
		if err != nil {
			samplingValue = float64(1)
		}
	}

	return &tracing.JaegerConfig{
		LocalAgentAddress: v.source.StringF(
			"tracing.providers.jaeger.local_agent_address",
			os.Getenv("TRACING_PROVIDER_JAEGER_LOCAL_AGENT_ADDRESS")),

		Sampling: &tracing.JaegerSampling{
			Type: stringsx.Coalesce(
				v.source.String("tracing.providers.jaeger.sampling.type"),
				os.Getenv("TRACING_PROVIDER_JAEGER_SAMPLING_TYPE"),
				"const",
			),

			Value: samplingValue,

			ServerURL: stringsx.Coalesce(
				v.source.String("tracing.providers.jaeger.sampling.server_url"),
				os.Getenv("TRACING_PROVIDER_JAEGER_SAMPLING_SERVER_URL"),
			),
		},
		Propagation: stringsx.Coalesce(
			v.source.String("JAEGER_PROPAGATION"), // Standard Jaeger client config
			v.source.String("tracing.providers.jaeger.propagation"),
			os.Getenv("TRACING_PROVIDER_JAEGER_PROPAGATION"),
		),
	}
}
func (v *KoanfProvider) TracingZipkinConfig() *tracing.ZipkinConfig {
	return &tracing.ZipkinConfig{
		ServerURL: stringsx.Coalesce(
			v.source.String("tracing.providers.zipkin.server_url"),
			os.Getenv("TRACING_PROVIDER_ZIPKIN_SERVER_URL"),
		),
	}
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
