package metrics

import (
	"net/http"
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
	// Promtheus repository
	Prometheus *PrometheusRepository

	clock timer

	// Silence metrics for specific URL paths
	// it is protected by the mutex
	mutex        sync.RWMutex
	silencePaths map[string]bool
}

// NewMiddleware returns a new *Middleware, yay!
func NewMiddleware(prom *PrometheusRepository, name string) *Middleware {
	return &Middleware{
		Name:         name,
		Prometheus:   prom,
		clock:        &realClock{},
		silencePaths: map[string]bool{},
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

func (m *Middleware) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	start := m.clock.Now()
	next(rw, r)
	latency := m.clock.Since(start)
	res := rw.(negroni.ResponseWriter)

	if _, silent := m.silencePaths[r.URL.Path]; !silent {
		m.Prometheus.RequestDurationObserve(m.Name, r.RequestURI, r.Method, res.Status())(float64(latency.Seconds()))
		m.Prometheus.UpdateRequest(m.Name, r.RequestURI, r.Method, res.Status())
	}
}
