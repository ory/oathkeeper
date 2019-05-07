package driver

import (
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/ory/herodot"
	"github.com/ory/oathkeeper/api"
	"github.com/ory/oathkeeper/credential"
	"github.com/ory/oathkeeper/driver/configuration"
	"github.com/ory/oathkeeper/pipeline/authn"
	"github.com/ory/oathkeeper/pipeline/authz"
	"github.com/ory/oathkeeper/pipeline/mutate"
	"github.com/ory/oathkeeper/rule"
	"github.com/ory/x/healthx"
)

var _ Registry = new(RegistryMemory)

type RegistryMemory struct {
	sync.RWMutex

	buildVersion string
	buildHash    string
	buildDate    string
	l            logrus.FieldLogger
	c            configuration.Provider

	ch *api.CredentialHandler

	cf credential.Fetcher

	authenticators map[string]authn.Authenticator
	authorizers    map[string]authz.Authorizer
	mutators       map[string]mutate.Mutator
}

func (m *RegistryMemory) Init() error {
	return nil
}

func NewRegistryMemory() *RegistryMemory {
	return &RegistryMemory{}
}

func (m *RegistryMemory) BuildVersion() string {
	return m.buildVersion
}

func (m *RegistryMemory) BuildDate() string {
	return m.buildDate
}

func (m *RegistryMemory) BuildHash() string {
	return m.buildHash
}

func (m *RegistryMemory) WithConfig(c configuration.Provider) Registry {
	m.c = c
	return m
}

func (m *RegistryMemory) WithBuildInfo(version, hash, date string) Registry {
	m.buildVersion = version
	m.buildHash = hash
	m.buildDate = date
	return m
}

func (m *RegistryMemory) WithLogger(l logrus.FieldLogger) Registry {
	m.l = l
	return m
}

func (m *RegistryMemory) CredentialHandler() *api.CredentialHandler {
	if m.ch == nil {
		m.ch = api.NewCredentialHandler(m.c, m)
	}

	return m.ch
}

func (m *RegistryMemory) HealthHandler() *healthx.Handler {
	panic("implement me")
}

func (m *RegistryMemory) RuleValidator() rule.ValidatorDefault {
	panic("implement me")
}

func (m *RegistryMemory) RuleManager() rule.Repository {
	panic("implement me")
}

func (m *RegistryMemory) Writer() herodot.Writer {
	panic("implement me")
}

func (m *RegistryMemory) Logger() logrus.FieldLogger {
	panic("implement me")
}

func (m *RegistryMemory) RuleHandler() *api.RuleHandler {
	panic("implement me")
}

func (m *RegistryMemory) JudgeHandler() *api.JudgeHandler {
	panic("implement me")
}

func (m *RegistryMemory) CredentialsFetcher() credential.Fetcher {
	if m.cf == nil {
		m.cf = credential.NewFetcherDefault(m.Logger(), time.Second, time.Second*30)
	}

	return m.cf
}

func (m *RegistryMemory) CredentialsSigner() credential.Signer {
	panic("implement me")
}

func (m *RegistryMemory) CredentialsVerifier() credential.Verifier {
	panic("implement me")
}

func (m *RegistryMemory) AvailablePipelineAuthenticators() (available []string) {
	m.prepareAuthn()
	m.RLock()
	defer m.RUnlock()

	available = make([]string, 0, len(m.authenticators))
	for k := range m.authenticators {
		available = append(available, k)
	}
	return
}

func (m *RegistryMemory) PipelineAuthenticator(id string) (authn.Authenticator, error) {
	m.prepareAuthn()
	m.RLock()
	defer m.RUnlock()

	a, ok := m.authenticators[id]
	if !ok {
		return nil, errors.WithStack(ErrPipelineHandlerNotFound)
	}
	return a, nil
}

func (m *RegistryMemory) AvailablePipelineAuthorizers() (available []string) {
	m.prepareAuthz()
	m.RLock()
	defer m.RUnlock()

	available = make([]string, 0, len(m.authorizers))
	for k := range m.authorizers {
		available = append(available, k)
	}
	return
}

func (m *RegistryMemory) PipelineAuthorizer(id string) (authz.Authorizer, error) {
	m.prepareAuthz()
	m.RLock()
	defer m.RUnlock()

	a, ok := m.authorizers[id]
	if !ok {
		return nil, errors.WithStack(ErrPipelineHandlerNotFound)
	}
	return a, nil
}

func (m *RegistryMemory) AvailablePipelineMutators() (available []string) {
	m.prepareMutators()
	m.RLock()
	defer m.RUnlock()

	available = make([]string, 0, len(m.mutators))
	for k := range m.mutators {
		available = append(available, k)
	}

	return
}

func (m *RegistryMemory) PipelineMutator(id string) (mutate.Mutator, error) {
	m.prepareMutators()
	m.RLock()
	defer m.RUnlock()

	a, ok := m.mutators[id]
	if !ok {
		return nil, errors.WithStack(ErrPipelineHandlerNotFound)
	}
	return a, nil
}

func (m *RegistryMemory) prepareAuthn() {
	m.Lock()
	defer m.Unlock()
	if m.authenticators == nil {
		interim := []authn.Authenticator{authn.NewAuthenticatorNoOp(m.c)}

		m.authenticators = map[string]authn.Authenticator{}
		for _, a := range interim {
			m.authenticators[a.GetID()] = a
		}
	}
}

func (m *RegistryMemory) prepareAuthz() {
	m.Lock()
	defer m.Unlock()
	if m.authorizers == nil {
		interim := []authz.Authorizer{authz.NewAuthorizerAllow(m.c)}

		m.authorizers = map[string]authz.Authorizer{}
		for _, a := range interim {
			m.authorizers[a.GetID()] = a
		}
	}
}

func (m *RegistryMemory) prepareMutators() {
	m.Lock()
	defer m.Unlock()
	if m.mutators == nil {
		interim := []mutate.Mutator{mutate.NewMutatorNoop(m.c)}

		m.mutators = map[string]mutate.Mutator{}
		for _, a := range interim {
			m.mutators[a.GetID()] = a
		}
	}
}
