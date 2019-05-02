package driver

import (
	"github.com/ory/oathkeeper/driver/configuration"
	"github.com/ory/oathkeeper/proxy"
	"github.com/sirupsen/logrus"
)

type RegistryBase struct {
	buildVersion string
	buildHash    string
	buildDate    string
	l            logrus.FieldLogger
	c            configuration.Provider

	authn []proxy.Authenticator
}

func (m *RegistryBase) BuildVersion() string {
	return m.buildVersion
}

func (m *RegistryBase) BuildDate() string {
	return m.buildDate
}

func (m *RegistryBase) BuildHash() string {
	return m.buildHash
}

func (m *RegistryBase) WithConfig(c configuration.Provider) Registry {
	m.c = c
	return m
}

func (m *RegistryBase) WithBuildInfo(version, hash, date string) Registry {
	m.buildVersion = version
	m.buildHash = hash
	m.buildDate = date
	return m
}

func (m *RegistryBase) Authenticators() []proxy.Authenticator {
	if m.authn != nil {
		return m.authn
	}

	var authn []proxy.Authenticator

	if m.c.AuthenticatorAnonymousEnabled() {
		authn = append(authn, proxy.NewAuthenticatorAnonymous(m.c))
	}

	if m.c.AuthenticatorNoopEnabled() {
		authn = append(authn, proxy.NewAuthenticatorNoOp())
	}

	if m.c.AuthenticatorJWTEnabled() {
		authn = append(authn, proxy.NewAuthenticatorJWT(m.c))
	}

}
