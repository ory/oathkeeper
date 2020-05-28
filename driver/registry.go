package driver

import (
	"github.com/ory/x/logrusx"

	"github.com/ory/oathkeeper/pipeline/errors"
	"github.com/ory/oathkeeper/proxy"

	"github.com/ory/x/healthx"
	"github.com/ory/x/tracing"

	"github.com/ory/oathkeeper/api"
	"github.com/ory/oathkeeper/credentials"
	"github.com/ory/oathkeeper/driver/configuration"
	"github.com/ory/oathkeeper/pipeline/authn"
	"github.com/ory/oathkeeper/pipeline/authz"
	"github.com/ory/oathkeeper/pipeline/mutate"
	"github.com/ory/oathkeeper/rule"
	"github.com/ory/oathkeeper/x"
)

type Registry interface {
	Init()

	WithConfig(c configuration.Provider) Registry
	WithLogger(l *logrusx.Logger) Registry
	WithBuildInfo(version, hash, date string) Registry
	BuildVersion() string
	BuildDate() string
	BuildHash() string

	ProxyRequestHandler() *proxy.RequestHandler
	HealthHandler() *healthx.Handler
	RuleHandler() *api.RuleHandler
	DecisionHandler() *api.DecisionHandler
	CredentialHandler() *api.CredentialsHandler

	Proxy() *proxy.Proxy
	Tracer() *tracing.Tracer

	authn.Registry
	authz.Registry
	mutate.Registry
	errors.Registry

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
