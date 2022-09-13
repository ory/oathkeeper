package middleware

import (
	"context"

	"github.com/pkg/errors"

	_ "github.com/ory/jsonschema/v3/fileloader"
	_ "github.com/ory/jsonschema/v3/httploader"
	"github.com/ory/x/configx"
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
	}

	middleware struct{ dependencies }

	options struct {
		logger             *logrusx.Logger
		configFile         string
		registryAddr       *driver.Registry
		configProviderAddr *configuration.Provider
	}

	Option func(*options)
)

// WithConfigFile sets the path to the config file to use for the middleware.
func WithConfigFile(configFile string) Option {
	return func(o *options) { o.configFile = configFile }
}

// New creates an Oathkeeper middleware from the options. By default, it tries
// to read the configuration from the file "oathkeeper.yaml".
func New(ctx context.Context, opts ...Option) (*middleware, error) {
	o := options{
		logger:     logrusx.New("Ory Oathkeeper Middleware", x.Version),
		configFile: "oathkeeper.yaml",
	}
	for _, opt := range opts {
		opt(&o)
	}

	c, err := configuration.NewKoanfProvider(
		ctx, nil, o.logger,
		configx.WithContext(ctx),
		configx.WithLogger(o.logger),
		configx.WithConfigFiles(o.configFile),
		configx.DisableEnvLoading(),
	)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	r := driver.NewRegistry(c).WithLogger(o.logger).WithBuildInfo(x.Version, x.Commit, x.Date)
	if o.registryAddr != nil {
		*o.registryAddr = r
	}
	if o.configProviderAddr != nil {
		*o.configProviderAddr = c
	}

	m := &middleware{r}

	return m, nil
}
