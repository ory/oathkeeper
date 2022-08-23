package middleware

import (
	_ "github.com/ory/jsonschema/v3/fileloader"
	_ "github.com/ory/jsonschema/v3/httploader"
	schema "github.com/ory/oathkeeper/.schema"
	"github.com/ory/oathkeeper/driver"
	"github.com/ory/oathkeeper/driver/configuration"
	"github.com/ory/oathkeeper/proxy"
	"github.com/ory/oathkeeper/rule"
	"github.com/ory/oathkeeper/x"
	"github.com/ory/x/logrusx"
	"github.com/ory/x/viperx"
)

type (
	dependencies interface {
		Logger() *logrusx.Logger
		RuleMatcher() rule.Matcher
		ProxyRequestHandler() proxy.RequestHandler
	}

	middleware struct{ dependencies }

	options struct {
		logger       *logrusx.Logger
		configFile   string
		registryAddr *driver.Registry
	}

	Option func(*options)
)

// WithLogger uses the specified logger in the middleware.
func WithLogger(l *logrusx.Logger) Option {
	return func(o *options) { o.logger = l }
}

// WithConfig sets the path to the config file to use for the middleware.
func WithConfig(configFile string) Option {
	return func(o *options) { o.configFile = configFile }
}

// New creates an Oathkeeper middleware from the options. By default, it tries
// to read the configuration from the file "oathkeeper.yaml".
func New(opts ...Option) *middleware {
	o := options{
		logger:     logrusx.New("Ory Oathkeeper Middleware", x.Version),
		configFile: "oathkeeper.yaml",
	}
	for _, opt := range opts {
		opt(&o)
	}

	initializeConfig(o.logger, o.configFile)
	c := configuration.NewViperProvider(o.logger)

	r := driver.NewRegistry(c).WithLogger(o.logger).WithBuildInfo(x.Version, x.Commit, x.Date)
	if o.registryAddr != nil {
		*o.registryAddr = r
	}

	m := &middleware{r}
	m.watchAndValidateViper()

	return m
}

func (m *middleware) watchAndValidateViper() {
	schema, err := schema.FS.ReadFile("config.schema.json")
	if err != nil {
		m.Logger().WithError(err).Fatal("Unable to open configuration JSON Schema.")
	}
	viperx.WatchAndValidateViper(m.Logger(), schema, "ORY Oathkeeper", []string{"serve", "profiling", "log"}, "")
}
