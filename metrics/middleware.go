package metrics

import (
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/urfave/negroni"
)

type timer interface {
	Now() time.Time
	Since(time.Time) time.Duration
}

type realClock struct{}

func (rc *realClock) Now() time.Time {
	return time.Now()
}

func (rc *realClock) Since(t time.Time) time.Duration {
	return time.Since(t)
}

// Middleware is a middleware handler that logs the request as it goes in and the response as it goes out.
type Middleware struct {
	// Name is the name of the application as recorded in latency metrics
	Name string
	// Prometheus repository
	Prometheus *PrometheusRepository

	clock timer

	// Silence metrics for specific URL paths
	// it is protected by the mutex
	mutex         sync.RWMutex
	silencePaths  map[string]bool
	collapsePaths bool
}

// NewMiddleware returns a new *Middleware, yay!
func NewMiddleware(prom *PrometheusRepository, name string) *Middleware {
	return &Middleware{
		Name:          name,
		Prometheus:    prom,
		clock:         &realClock{},
		silencePaths:  map[string]bool{},
		collapsePaths: true,
	}
}

// ExcludePaths adds new URL paths to be ignored during logging. The URL u is parsed, hence the returned error
func (m *Middleware) ExcludePaths(paths ...string) *Middleware {
	for _, path := range paths {
		m.mutex.Lock()
		m.silencePaths[path] = true
		m.mutex.Unlock()
	}
	return m
}

// CollapsePaths if set to true, forces the value of the "request" label
// of the prometheus request metrics to be collapsed to the first context path segment only.
// eg. (when set to true):
//    - /decisions/service/my-service -> /decisions
//    - /decisions -> /decisions
func (m *Middleware) CollapsePaths(flag bool) *Middleware {
	m.mutex.Lock()
	m.collapsePaths = flag
	m.mutex.Unlock()
	return m
}

func (m *Middleware) getFirstPathSegment(requestURI string) string {
	// Will split /my/example/uri in []string{"", "my", "example/uri"}
	uriSegments := strings.SplitN(requestURI, "/", 3)
	if len(uriSegments) > 1 {
		return "/" + uriSegments[1]
	}
	return "/"

}

func (m *Middleware) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	start := m.clock.Now()
	next(rw, r)
	latency := m.clock.Since(start)
	res := rw.(negroni.ResponseWriter)

	if _, silent := m.silencePaths[r.URL.Path]; !silent {
		requestURI := r.RequestURI
		if m.collapsePaths {
			requestURI = m.getFirstPathSegment(requestURI)
		}
		m.Prometheus.RequestDurationObserve(m.Name, requestURI, r.Method, res.Status())(float64(latency.Seconds()))
		m.Prometheus.UpdateRequest(m.Name, requestURI, r.Method, res.Status())
	}
}
