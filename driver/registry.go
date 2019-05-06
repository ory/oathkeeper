package driver

import (
	"github.com/ory/oathkeeper/api"
	"github.com/ory/oathkeeper/credential"
	"github.com/ory/oathkeeper/driver/configuration"
	"github.com/ory/oathkeeper/rule"
	"github.com/ory/oathkeeper/x"
	"github.com/ory/x/healthx"
	"github.com/sirupsen/logrus"
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
	CredentialHandler() *api.CredentialHandler

	rule.Registry
	credential.FetcherRegistry
	credential.SignerRegistry
	credential.VerifierRegistry

	x.RegistryWriter
	x.RegistryLogger
}
