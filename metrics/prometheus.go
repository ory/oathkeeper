// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

// Package contains the collection of prometheus meters/counters
// and related update methods
package metrics

import (
	"strconv"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/ory/x/logrusx"

	"github.com/ory/oathkeeper/driver"
)

var (
	// RequestTotal provides the total number of requests
	RequestTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ory_oathkeeper_requests_total",
			Help: "Total number of requests",
		},
		[]string{"service", "method", "request", "status_code"},
	)
	// HistogramRequestDuration provides the duration of requests
	HistogramRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "ory_oathkeeper_requests_duration_seconds",
			Help:    "Time spent serving requests.",
			Buckets: prometheus.DefBuckets,
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
	logger   *logrusx.Logger
	Registry *prometheus.Registry
	metrics  []prometheus.Collector
}

// NewConfigurablePrometheusRepository creates a new prometheus repository with the given settings
func NewConfigurablePrometheusRepository(d driver.Driver, logger *logrusx.Logger) *PrometheusRepository {
	RequestTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: d.Configuration().PrometheusMetricsNamePrefix() + "requests_total",
			Help: "Total number of requests",
		},
		[]string{"service", "method", "request", "status_code"},
	)
	HistogramRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    d.Configuration().PrometheusMetricsNamePrefix() + "requests_duration_seconds",
			Help:    "Time spent serving requests.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"service", "method", "request", "status_code"},
	)
	return NewPrometheusRepository(logger)
}

// NewPrometheusRepository creates a new prometheus repository
func NewPrometheusRepository(logger *logrusx.Logger) *PrometheusRepository {
	m := []prometheus.Collector{
		prometheus.NewGoCollector(),
		prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}),
		RequestTotal,
		HistogramRequestDuration,
	}

	r := prometheus.NewRegistry()

	for _, metric := range m {
		if err := r.Register(metric); err != nil {
			logger.WithError(err).Error("Unable to register prometheus metric.")
		}
	}

	mr := &PrometheusRepository{
		logger:   logger,
		Registry: r,
		metrics:  m,
	}

	return mr
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
