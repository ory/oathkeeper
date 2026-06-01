// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"github.com/ory/x/healthx"
	"github.com/ory/x/httpx"
	"github.com/ory/x/logrusx"
	"github.com/ory/x/otelx"

	"github.com/ory/oathkeeper/api"
	"github.com/ory/oathkeeper/credentials"
	"github.com/ory/oathkeeper/driver/configuration"
	"github.com/ory/oathkeeper/pipeline/authn"
	"github.com/ory/oathkeeper/pipeline/authz"
	"github.com/ory/oathkeeper/pipeline/errors"
	"github.com/ory/oathkeeper/pipeline/mutate"
	"github.com/ory/oathkeeper/proxy"
	"github.com/ory/oathkeeper/rule"
)

type Registry interface {
	Init()

	Config() configuration.Configuration

	ProxyRequestHandler() proxy.RequestHandler
	HealthxReadyCheckers() healthx.ReadyCheckers
	HealthHandler() *healthx.Handler
	RuleHandler() *api.RuleHandler
	DecisionHandler() *api.DecisionHandler
	CredentialHandler() *api.CredentialsHandler

	Proxy() *proxy.Proxy

	otelx.Provider

	authn.Registry
	authz.Registry
	mutate.Registry
	errors.Registry

	rule.Registry
	credentials.FetcherRegistry
	credentials.SignerRegistry
	credentials.VerifierRegistry

	httpx.WriterProvider
	logrusx.Provider
}

func NewRegistry(c configuration.Configuration, l *logrusx.Logger) Registry {
	return NewRegistryMemory(c, l)
}
