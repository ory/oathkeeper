// Package contains the collection of prometheus meters/counters
// and related update methods
package metrics

import (
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

var (
	// RequestTotal provides the total number of requests
	RequestTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ory_oathkeeper_request_total",
			Help: "Total number of requests",
		},
		[]string{"service", "method", "request", "status_code"},
	)
	// HistogramRequestDuration provides the duration of requests
	HistogramRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "ory_oathkeeper_request_duration_in_seconds",
			Help:    "Time spent serving requests.",
			Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
		},
		[]string{"service", "method", "request", "status_code"},
	)
)

// RequestDurationObserve tracks request durations
type RequestDurationObserve func(histogram *prometheus.HistogramVec, service, request, method string, statusCode int) func(float64)

// UpdateRequest tracks total requests done
type UpdateRequest func(counter *prometheus.CounterVec, service, request, method string, statusCode int)

// PrometheusRepository provides methods to manage prometheus metrics
type PrometheusRepository struct {
	logger                 log.FieldLogger
	requestDurationObserve RequestDurationObserve
	updateRequest          UpdateRequest
	Registry               *prometheus.Registry
	metrics                []prometheus.Collector
}

// NewPrometheusRepository creates a new prometheus repository with the given settings
func NewPrometheusRepository(logger log.FieldLogger) (*PrometheusRepository, error) {
	m := []prometheus.Collector{
		prometheus.NewGoCollector(),
		prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}),
		RequestTotal,
		HistogramRequestDuration,
	}

	r := prometheus.NewRegistry()

	for _, metric := range m {
		if err := r.Register(metric); err != nil {
			logger.Warnf("register prometheus metrics: %+v", err)
		}
	}

	mr := &PrometheusRepository{
		logger:   logger,
		Registry: r,
		metrics:  m,
	}

	return mr, nil
}

// RequestDurationObserve tracks request durations
func (r *PrometheusRepository) RequestDurationObserve(service, request, method string, statusCode int) func(float64) {
	return func(v float64) {
		HistogramRequestDuration.With(prometheus.Labels{
			"service":     service,
			"method":      method,
			"request":     request,
			"status_code": strconv.Itoa(statusCode),
		}).Observe(v)
	}
}

// UpdateRequest tracks total requests done
func (r *PrometheusRepository) UpdateRequest(service, request, method string, statusCode int) {
	RequestTotal.With(prometheus.Labels{
		"service":     service,
		"method":      method,
		"request":     request,
		"status_code": strconv.Itoa(statusCode),
	}).Inc()
}
