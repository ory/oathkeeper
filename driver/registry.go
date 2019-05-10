package driver

import (
	"github.com/sirupsen/logrus"

	"github.com/ory/oathkeeper/proxy"

	"github.com/ory/oathkeeper/api"
	"github.com/ory/oathkeeper/credentials"
	"github.com/ory/oathkeeper/driver/configuration"
	"github.com/ory/oathkeeper/pipeline/authn"
	"github.com/ory/oathkeeper/pipeline/authz"
	"github.com/ory/oathkeeper/pipeline/mutate"
	"github.com/ory/oathkeeper/rule"
	"github.com/ory/oathkeeper/x"
	"github.com/ory/x/healthx"
)

type Registry interface {
	WithConfig(c configuration.Provider) Registry
	WithLogger(l logrus.FieldLogger) Registry
	WithBuildInfo(version, hash, date string) Registry
	BuildVersion() string
	BuildDate() string
	BuildHash() string

	ProxyRequestHandler() *proxy.RequestHandler
	HealthHandler() *healthx.Handler
	RuleHandler() *api.RuleHandler
	JudgeHandler() *api.JudgeHandler
	CredentialHandler() *api.CredentialsHandler

	Proxy() *proxy.Proxy

	authn.Registry
	authz.Registry
	mutate.Registry

	rule.Registry
	credentials.FetcherRegistry
	credentials.SignerRegistry
	credentials.VerifierRegistry

	x.RegistryWriter
	x.RegistryLogger
}

func NewRegistry(c configuration.Provider) Registry {
	return NewRegistryMemory().WithConfig(c)
}
