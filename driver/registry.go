package driver

import (
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

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

var (
	ErrPipelineHandlerNotFound = errors.New("requested pipeline handler does not exist")
)

type Registry interface {
	Init() error

	WithConfig(c configuration.Provider) Registry
	WithLogger(l logrus.FieldLogger) Registry
	WithBuildInfo(version, hash, date string) Registry
	BuildVersion() string
	BuildDate() string
	BuildHash() string

	HealthHandler() *healthx.Handler
	RuleHandler() *api.RuleHandler
	JudgeHandler() *api.JudgeHandler
	CredentialHandler() *api.CredentialsHandler

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
