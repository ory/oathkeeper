package telemetry

import (
	"net/http"
	"sync"
)

type Middleware struct {
	Requests int64
	sync.Mutex
}

func (m *Middleware) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	next(rw, r)
	m.Lock()
	m.Requests++
	m.Unlock()
}
