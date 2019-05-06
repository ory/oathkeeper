package driver

import (
	"github.com/ory/herodot"
	"github.com/ory/oathkeeper/api"
	"github.com/ory/oathkeeper/credential"
	"github.com/ory/oathkeeper/driver/configuration"
	authn2 "github.com/ory/oathkeeper/pipeline/authn"
	"github.com/ory/oathkeeper/rule"
	"github.com/ory/x/healthx"
	"github.com/sirupsen/logrus"
	"time"
)

var _ Registry = new(RegistryMemory)

type RegistryMemory struct {
	buildVersion string
	buildHash    string
	buildDate    string
	l            logrus.FieldLogger
	c            configuration.Provider

	ch *api.CredentialHandler

	cf credential.Fetcher

	authn []authn2.Authenticator
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

func (m *RegistryMemory) Authenticators() []authn2.Authenticator {
	if m.authn != nil {
		return m.authn
	}

	//var authn []authn2.Authenticator
	return nil
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

func (m *RegistryMemory) RuleValidator() rule.Validator {
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
		m.cf = credential.NewFetcherDefault(m.Logger(), time.Second, time.Second * 30)
	}

	return m.cf
}

func (m *RegistryMemory) CredentialsSigner() credential.Signer {
	panic("implement me")
}

func (m *RegistryMemory) CredentialsVerifier() credential.Verifier {
	panic("implement me")
}