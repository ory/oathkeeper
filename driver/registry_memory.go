package driver

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ory/oathkeeper/driver/health"
	"github.com/ory/oathkeeper/pipeline"
	pe "github.com/ory/oathkeeper/pipeline/errors"
	"github.com/ory/oathkeeper/proxy"
	"github.com/ory/oathkeeper/x"

	"github.com/ory/x/logrusx"
	"github.com/ory/x/tracing"

	"github.com/pkg/errors"

	"github.com/ory/herodot"
	"github.com/ory/x/healthx"

	"github.com/ory/oathkeeper/api"
	"github.com/ory/oathkeeper/credentials"
	"github.com/ory/oathkeeper/driver/configuration"
	"github.com/ory/oathkeeper/pipeline/authn"
	"github.com/ory/oathkeeper/pipeline/authz"
	ep "github.com/ory/oathkeeper/pipeline/errors"
	"github.com/ory/oathkeeper/pipeline/mutate"
	"github.com/ory/oathkeeper/rule"
	rulereadiness "github.com/ory/oathkeeper/rule/readiness"
)

var _ Registry = new(RegistryMemory)

type RegistryMemory struct {
	sync.RWMutex

	buildVersion string
	buildHash    string
	buildDate    string
	logger       *logrusx.Logger
	writer       herodot.Writer
	c            configuration.Provider
	trc          *tracing.Tracer

	ch *api.CredentialsHandler

	credentialsFetcher  credentials.Fetcher
	credentialsVerifier credentials.Verifier
	credentialsSigner   credentials.Signer
	ruleValidator       rule.Validator
	ruleRepository      *rule.RepositoryMemory
	apiRuleHandler      *api.RuleHandler
	apiJudgeHandler     *api.DecisionHandler
	healthxHandler      *healthx.Handler

	proxyRequestHandler *proxy.RequestHandler
	proxyProxy          *proxy.Proxy
	ruleFetcher         rule.Fetcher

	cachedCerts          sync.Map
	upstreamRequestCount int64
	upstreamTransport    *http.Transport

	authenticators map[string]authn.Authenticator
	authorizers    map[string]authz.Authorizer
	mutators       map[string]mutate.Mutator
	errors         map[string]ep.Handler

	healthEventManager *health.DefaultHealthEventManager

	ruleRepositoryLock sync.Mutex
}

type certCache struct {
	path       string
	stat       *os.FileInfo
	lastUpdate time.Time
}

func (r *RegistryMemory) Init() {
	go func() {
		if err := r.RuleFetcher().Watch(context.Background()); err != nil {
			r.Logger().WithError(err).Fatal("Access rule watcher terminated with an error.")
		}
	}()
	r.HealthEventManager().Watch(context.Background())
	_ = r.RuleRepository()
}

func (r *RegistryMemory) RuleFetcher() rule.Fetcher {
	if r.ruleFetcher == nil {
		r.ruleFetcher = rule.NewFetcherDefault(r.c, r)
	}
	return r.ruleFetcher
}

func (r *RegistryMemory) WithRuleFetcher(fetcher rule.Fetcher) Registry {
	r.ruleFetcher = fetcher
	return r
}

func (r *RegistryMemory) ProxyRequestHandler() *proxy.RequestHandler {
	if r.proxyRequestHandler == nil {
		r.proxyRequestHandler = proxy.NewRequestHandler(r, r.c)
	}
	return r.proxyRequestHandler
}

func (r *RegistryMemory) RuleMatcher() rule.Matcher {
	_ = r.RuleRepository() // make sure `r.ruleRepository` is set
	return r.ruleRepository
}

func (r *RegistryMemory) isTransportCacheValid(certFile string) (bool, error) {
	// No cached transport yet, but configuration requires it, so handle as if it was dirty so we get one.
	if r.upstreamTransport == nil {
		return true, nil
	}

	// Check whether this is the first time we see this file.
	cacheItem, ok := r.cachedCerts.Load(certFile)
	if !ok {
		return true, nil
	}

	cache := cacheItem.(*certCache)

	// If the last time the transport was accessed is beyond the cache TTL then
	// a refresh of the cache is forced whether the file was updated or not.
	// This ensures that the cache is always refreshed at fixed intervals
	// regardless of the environment.
	if time.Since(cache.lastUpdate) > r.c.ProxyServeTransportCacheTimeToLive() {
		return true, nil
	}

	// Transport is dirty if the certificate bytes changed or the modification time,
	// but only check this based on the cache.refresh_frequency option to reduce IO impact.
	caRefreshFrequency := r.c.ProxyServeTransportCacheRefreshFrequency()
	if caRefreshFrequency <= 0 {
		return false, nil // Periodic refresh of the Ca is diabled.
	}

	if r.upstreamRequestCount < int64(caRefreshFrequency) {
		return false, nil
	}
	r.upstreamRequestCount = 0

	stat, err := os.Stat(certFile)
	if err != nil {
		return false, err
	}

	if stat.Size() != (*cache.stat).Size() || stat.ModTime() != (*cache.stat).ModTime() {
		return true, nil
	}

	return false, nil
}

func createTransportWithCerts(certs []string) (*http.Transport, error) {
	transport := &(*http.DefaultTransport.(*http.Transport)) // shallow copy

	// Get the SystemCertPool or continue with an empty pool on error
	rootCAs, err := x509.SystemCertPool()
	if err != nil {
		return nil, err
	}

	for _, cert := range certs {
		data, err := ioutil.ReadFile(cert)
		if err != nil {
			return nil, err
		}

		// Append our cert to the system pool
		if ok := rootCAs.AppendCertsFromPEM(data); !ok {
			return nil, errors.New("No certs appended, only system certs present, did you specify the correct cert file?")
		}
	}

	transport.TLSClientConfig = &tls.Config{
		InsecureSkipVerify: false,
		RootCAs:            rootCAs,
	}

	return transport, nil
}

// UpstreamTransport decides the transport to use for the upstream for the request.
func (r *RegistryMemory) UpstreamTransport(req *http.Request) (http.RoundTripper, error) {
	atomic.AddInt64(&r.upstreamRequestCount, 1)

	// Use req to decide the transport per request iff need be.

	// Not using custom certificates on upstreams, so go with default transport no need to use cache
	certs := r.c.ProxyServeTransportCerts()
	if certs == nil {
		return http.DefaultTransport, nil
	}

	// Decide wether to create a transport or use the cached one. Update the cached transport if it is dirty.
	refreshTransport := false
	for _, cert := range certs {
		dirty, err := r.isTransportCacheValid(cert)
		if err != nil {
			return nil, err
		}

		// TODO: Should we check cached certs and purge them if they are no longer used?

		// Cache the transport and cert metadata
		if dirty {
			stat, err := os.Stat(cert)
			if err != nil {
				return nil, err
			}

			r.cachedCerts.Store(cert, &certCache{
				path:       cert,
				stat:       &stat,
				lastUpdate: time.Now(),
			})

			refreshTransport = true
		}
	}

	if refreshTransport {
		transport, err := createTransportWithCerts(certs)
		if err != nil {
			return nil, err
		}

		r.upstreamTransport = transport
	}

	return r.upstreamTransport, nil
}

func NewRegistryMemory() *RegistryMemory {
	return &RegistryMemory{}
}

func (r *RegistryMemory) BuildVersion() string {
	return r.buildVersion
}

func (r *RegistryMemory) BuildDate() string {
	return r.buildDate
}

func (r *RegistryMemory) BuildHash() string {
	return r.buildHash
}

func (r *RegistryMemory) WithConfig(c configuration.Provider) Registry {
	r.c = c
	return r
}

func (r *RegistryMemory) WithBuildInfo(version, hash, date string) Registry {
	r.buildVersion = version
	r.buildHash = hash
	r.buildDate = date
	return r
}

func (r *RegistryMemory) WithLogger(l *logrusx.Logger) Registry {
	r.logger = l
	return r
}

func (r *RegistryMemory) CredentialHandler() *api.CredentialsHandler {
	if r.ch == nil {
		r.ch = api.NewCredentialHandler(r.c, r)
	}

	return r.ch
}

func (r *RegistryMemory) HealthEventManager() health.EventManager {
	if r.healthEventManager == nil {
		var err error
		rulesReadinessChecker := rulereadiness.NewReadinessHealthChecker()
		if r.healthEventManager, err = health.NewDefaultHealthEventManager(rulesReadinessChecker); err != nil {
			r.logger.WithError(err).Fatal("unable to instantiate new health event manager")
		}
	}
	return r.healthEventManager
}

func (r *RegistryMemory) HealthHandler() *healthx.Handler {
	r.RLock()
	defer r.RUnlock()

	if r.healthxHandler == nil {
		r.healthxHandler = healthx.NewHandler(r.Writer(), r.BuildVersion(), r.HealthEventManager().HealthxReadyCheckers())
	}
	return r.healthxHandler
}

func (r *RegistryMemory) RuleValidator() rule.Validator {
	if r.ruleValidator == nil {
		r.ruleValidator = rule.NewValidatorDefault(r)
	}
	return r.ruleValidator
}

func (r *RegistryMemory) RuleRepository() rule.Repository {
	if r.ruleRepository == nil {
		r.ruleRepository = rule.NewRepositoryMemory(r, r.HealthEventManager())
	}
	return r.ruleRepository
}

func (r *RegistryMemory) Writer() herodot.Writer {
	if r.writer == nil {
		r.writer = herodot.NewJSONWriter(r.Logger().Logger)
	}
	return r.writer
}

func (r *RegistryMemory) Logger() *logrusx.Logger {
	if r.logger == nil {
		r.logger = logrusx.New("ORY Oathkeeper", x.Version)
	}
	return r.logger
}

func (r *RegistryMemory) RuleHandler() *api.RuleHandler {
	if r.apiRuleHandler == nil {
		r.apiRuleHandler = api.NewRuleHandler(r)
	}
	return r.apiRuleHandler
}

func (r *RegistryMemory) DecisionHandler() *api.DecisionHandler {
	if r.apiJudgeHandler == nil {
		r.apiJudgeHandler = api.NewJudgeHandler(r)
	}
	return r.apiJudgeHandler
}

func (r *RegistryMemory) CredentialsFetcher() credentials.Fetcher {
	if r.credentialsFetcher == nil {
		r.credentialsFetcher = credentials.NewFetcherDefault(r.Logger(), r.c.AuthenticatorJwtJwkMaxWait(), r.c.AuthenticatorJwtJwkTtl())
	}

	return r.credentialsFetcher
}

func (r *RegistryMemory) CredentialsSigner() credentials.Signer {
	if r.credentialsSigner == nil {
		r.credentialsSigner = credentials.NewSignerDefault(r)
	}

	return r.credentialsSigner
}

func (r *RegistryMemory) CredentialsVerifier() credentials.Verifier {
	if r.credentialsVerifier == nil {
		r.credentialsVerifier = credentials.NewVerifierDefault(r)
	}

	return r.credentialsVerifier
}

func (r *RegistryMemory) AvailablePipelineErrorHandlers() pe.Handlers {
	r.prepareErrors()
	r.RLock()
	defer r.RUnlock()

	hs := make(pe.Handlers, 0, len(r.errors))
	for _, e := range r.errors {
		hs = append(hs, e)
	}

	return hs
}

func (r *RegistryMemory) PipelineErrorHandler(id string) (pe.Handler, error) {
	r.prepareErrors()
	r.RLock()
	defer r.RUnlock()

	a, ok := r.errors[id]
	if !ok {
		return nil, errors.WithStack(pipeline.ErrPipelineHandlerNotFound)
	}
	return a, nil
}

func (r *RegistryMemory) prepareErrors() {
	r.Lock()
	defer r.Unlock()

	if r.errors == nil {
		interim := []ep.Handler{
			ep.NewErrorJSON(r.c, r),
			ep.NewErrorRedirect(r.c, r),
			ep.NewErrorWWWAuthenticate(r.c, r),
		}

		r.errors = map[string]ep.Handler{}
		for _, a := range interim {
			r.errors[a.GetID()] = a
		}
	}
}

func (r *RegistryMemory) AvailablePipelineAuthenticators() (available []string) {
	r.prepareAuthn()
	r.RLock()
	defer r.RUnlock()

	available = make([]string, 0, len(r.authenticators))
	for k := range r.authenticators {
		available = append(available, k)
	}
	return
}

func (r *RegistryMemory) PipelineAuthenticator(id string) (authn.Authenticator, error) {
	r.prepareAuthn()
	r.RLock()
	defer r.RUnlock()

	a, ok := r.authenticators[id]
	if !ok {
		return nil, errors.WithStack(pipeline.ErrPipelineHandlerNotFound)
	}
	return a, nil
}

func (r *RegistryMemory) AvailablePipelineAuthorizers() (available []string) {
	r.prepareAuthz()
	r.RLock()
	defer r.RUnlock()

	available = make([]string, 0, len(r.authorizers))
	for k := range r.authorizers {
		available = append(available, k)
	}
	return
}

func (r *RegistryMemory) PipelineAuthorizer(id string) (authz.Authorizer, error) {
	r.prepareAuthz()
	r.RLock()
	defer r.RUnlock()

	a, ok := r.authorizers[id]
	if !ok {
		return nil, errors.WithStack(pipeline.ErrPipelineHandlerNotFound)
	}
	return a, nil
}

func (r *RegistryMemory) AvailablePipelineMutators() (available []string) {
	r.prepareMutators()
	r.RLock()
	defer r.RUnlock()

	available = make([]string, 0, len(r.mutators))
	for k := range r.mutators {
		available = append(available, k)
	}

	return
}

func (r *RegistryMemory) Proxy() *proxy.Proxy {
	if r.proxyProxy == nil {
		r.proxyProxy = proxy.NewProxy(r)
	}

	return r.proxyProxy
}

func (r *RegistryMemory) PipelineMutator(id string) (mutate.Mutator, error) {
	r.prepareMutators()
	r.RLock()
	defer r.RUnlock()

	a, ok := r.mutators[id]
	if !ok {
		return nil, errors.WithStack(pipeline.ErrPipelineHandlerNotFound)
	}
	return a, nil
}

func (r *RegistryMemory) WithBrokenPipelineMutator() *RegistryMemory {
	r.prepareMutators()
	r.mutators["broken"] = mutate.NewMutatorBroken(true)
	return r
}

func (r *RegistryMemory) prepareAuthn() {
	r.Lock()
	defer r.Unlock()
	if r.authenticators == nil {
		interim := []authn.Authenticator{
			authn.NewAuthenticatorAnonymous(r.c),
			authn.NewAuthenticatorCookieSession(r.c),
			authn.NewAuthenticatorBearerToken(r.c),
			authn.NewAuthenticatorJWT(r.c, r),
			authn.NewAuthenticatorNoOp(r.c),
			authn.NewAuthenticatorOAuth2ClientCredentials(r.c),
			authn.NewAuthenticatorOAuth2Introspection(r.c, r.Logger()),
			authn.NewAuthenticatorUnauthorized(r.c),
		}

		r.authenticators = map[string]authn.Authenticator{}
		for _, a := range interim {
			r.authenticators[a.GetID()] = a
		}
	}
}

func (r *RegistryMemory) prepareAuthz() {
	r.Lock()
	defer r.Unlock()
	if r.authorizers == nil {
		interim := []authz.Authorizer{
			authz.NewAuthorizerAllow(r.c),
			authz.NewAuthorizerDeny(r.c),
			authz.NewAuthorizerKetoEngineACPORY(r.c),
			authz.NewAuthorizerRemote(r.c),
			authz.NewAuthorizerRemoteJSON(r.c),
		}

		r.authorizers = map[string]authz.Authorizer{}
		for _, a := range interim {
			r.authorizers[a.GetID()] = a
		}
	}
}

func (r *RegistryMemory) prepareMutators() {
	r.Lock()
	defer r.Unlock()
	if r.mutators == nil {
		interim := []mutate.Mutator{
			mutate.NewMutatorCookie(r.c),
			mutate.NewMutatorHeader(r.c),
			mutate.NewMutatorIDToken(r.c, r),
			mutate.NewMutatorNoop(r.c),
			mutate.NewMutatorHydrator(r.c, r),
		}

		r.mutators = map[string]mutate.Mutator{}
		for _, a := range interim {
			r.mutators[a.GetID()] = a
		}
	}
}

func (r *RegistryMemory) Tracer() *tracing.Tracer {
	if r.trc == nil {
		r.trc = &tracing.Tracer{
			ServiceName:  r.c.TracingServiceName(),
			JaegerConfig: r.c.TracingJaegerConfig(),
			Provider:     r.c.TracingProvider(),
			Logger:       r.Logger(),
		}

		if err := r.trc.Setup(); err != nil {
			r.Logger().WithError(err).Fatalf("Unable to initialize Tracer.")
		}
	}

	return r.trc
}
