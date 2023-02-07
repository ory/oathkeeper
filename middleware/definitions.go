// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"

	"github.com/pkg/errors"
	"google.golang.org/grpc"

	_ "github.com/ory/jsonschema/v3/fileloader"
	_ "github.com/ory/jsonschema/v3/httploader"
	"github.com/ory/x/configx"
	"github.com/ory/x/healthx"
	"github.com/ory/x/logrusx"

	"github.com/ory/oathkeeper/driver"
	"github.com/ory/oathkeeper/driver/configuration"
	"github.com/ory/oathkeeper/proxy"
	"github.com/ory/oathkeeper/rule"
	"github.com/ory/oathkeeper/x"
)

type (
	dependencies interface {
		Logger() *logrusx.Logger
		RuleMatcher() rule.Matcher
		ProxyRequestHandler() proxy.RequestHandler
		HealthxReadyCheckers() healthx.ReadyCheckers
	}

	middleware struct{ dependencies }

	Middleware interface {
		UnaryInterceptor() grpc.UnaryServerInterceptor
		StreamInterceptor() grpc.StreamServerInterceptor
		HealthxReadyCheckers() healthx.ReadyCheckers
	}

	options struct {
		logger             *logrusx.Logger
		configFile         string
		registryAddr       *driver.Registry
		configProviderAddr *configuration.Provider
		configProviderOpts []configx.OptionModifier
	}

	Option func(*options)
)

// WithConfigFile sets the path to the config file to use for the middleware.
func WithConfigFile(configFile string) Option {
	return func(o *options) { o.configFile = configFile }
}

// WithLogger sets the logger for the middleware.
func WithLogger(logger *logrusx.Logger) Option {
	return func(o *options) { o.logger = logger }
}

// WithConfigOption sets a config option for the middleware. The following
// options will be set regardless:
// - configx.WithContext
// - configx.WithLogger
// - configx.WithConfigFiles
// - configx.DisableEnvLoading
func WithConfigOption(option configx.OptionModifier) Option {
	return func(o *options) {
		o.configProviderOpts = append(o.configProviderOpts, option)
	}
}

// New creates an Oathkeeper middleware from the options. By default, it tries
// to read the configuration from the file "oathkeeper.yaml".
func New(ctx context.Context, opts ...Option) (Middleware, error) {
	o := options{
		logger:     logrusx.New("Ory Oathkeeper Middleware", x.Version),
		configFile: "oathkeeper.yaml",
	}
	for _, opt := range opts {
		opt(&o)
	}

	c, err := configuration.NewKoanfProvider(
		ctx, nil, o.logger,
		append(o.configProviderOpts,
			configx.WithContext(ctx),
			configx.WithLogger(o.logger),
			configx.WithConfigFiles(o.configFile),
			configx.DisableEnvLoading(),
		)...,
	)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	r := driver.NewRegistry(c).WithLogger(o.logger).WithBuildInfo(x.Version, x.Commit, x.Date)
	r.Init()
	if o.registryAddr != nil {
		*o.registryAddr = r
	}
	if o.configProviderAddr != nil {
		*o.configProviderAddr = c
	}

	m := &middleware{r}

	return m, nil
}

func (m *middleware) HealthxReadyCheckers() healthx.ReadyCheckers {
	return m.dependencies.HealthxReadyCheckers()
}
