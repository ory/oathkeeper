package driver

import (
	"context"
	"sync"
	"time"

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

	authenticators map[string]authn.Authenticator
	authorizers    map[string]authz.Authorizer
	mutators       map[string]mutate.Mutator
	errors         map[string]ep.Handler

	ruleRepositoryLock sync.Mutex
}

func (r *RegistryMemory) Init() {
	go func() {
		if err := r.RuleFetcher().Watch(context.Background()); err != nil {
			r.Logger().WithError(err).Fatal("Access rule watcher terminated with an error.")
		}
	}()
	_ = r.RuleRepository()
}

func (r *RegistryMemory) RuleFetcher() rule.Fetcher {
	if r.ruleFetcher == nil {
		r.ruleFetcher = rule.NewFetcherDefault(r.c, r)
	}
	return r.ruleFetcher
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

func (r *RegistryMemory) HealthHandler() *healthx.Handler {
	if r.healthxHandler == nil {
		r.healthxHandler = healthx.NewHandler(r.Writer(), r.BuildVersion(), healthx.ReadyCheckers{})
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
		r.ruleRepository = rule.NewRepositoryMemory(r)
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
		r.credentialsFetcher = credentials.NewFetcherDefault(r.Logger(), time.Second, time.Second*30)
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
			authn.NewAuthenticatorJWT(r.c, r),
			authn.NewAuthenticatorNoOp(r.c),
			authn.NewAuthenticatorOAuth2ClientCredentials(r.c),
			authn.NewAuthenticatorOAuth2Introspection(r.c),
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
