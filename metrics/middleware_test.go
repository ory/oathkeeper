// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package metrics

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/ory/oathkeeper/driver"
	"github.com/ory/oathkeeper/x"
	"github.com/ory/x/configx"
	"github.com/ory/x/logrusx"

	"github.com/julienschmidt/httprouter"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/urfave/negroni"
)

var (
	metricMetadata string = `
	# HELP ory_oathkeeper_requests_total Total number of requests
	# TYPE ory_oathkeeper_requests_total counter
	`
	rootMetric string = `
	ory_oathkeeper_requests_total{method="GET",request="/",service="test",status_code="200"} 1
	`
	metricsNotCollapsed string = metricMetadata + rootMetric + `
	ory_oathkeeper_requests_total{method="GET",request="/hello/world",service="test",status_code="200"} 1
	ory_oathkeeper_requests_total{method="GET",request="/hello/world?foo=bar",service="test",status_code="200"} 1
	ory_oathkeeper_requests_total{method="GET",request="/hello?foo=bar",service="test",status_code="200"} 1
	`
	metricsCollapsed string = metricMetadata + rootMetric + `
	ory_oathkeeper_requests_total{method="GET",request="/hello",service="test",status_code="200"} 3
	`
	metricsHidden string = metricMetadata + `
	ory_oathkeeper_requests_total{method="GET",request="",service="test",status_code="200"} 4
	`

	serverConfigPaths  []string = []string{"/", "/hello", "/hello/world"}
	serverRequestPaths []string = []string{"/", "/hello?foo=bar", "/hello/world", "/hello/world?foo=bar"}

	configurableMetricMetadata string = `
	# HELP http_requests_total Total number of requests
	# TYPE http_requests_total counter
	`
	configurableRootMetric string = `
	http_requests_total{method="GET",request="/",service="test",status_code="200"} 1
	`
	configurableMetricsNotCollapsed string = configurableMetricMetadata + configurableRootMetric + `
	http_requests_total{method="GET",request="/hello?foo=bar",service="test",status_code="200"} 1
	http_requests_total{method="GET",request="/hello/world",service="test",status_code="200"} 1
	http_requests_total{method="GET",request="/hello/world?foo=bar",service="test",status_code="200"} 1
	`
	configurableMetricsCollapsed string = configurableMetricMetadata + configurableRootMetric + `
	http_requests_total{method="GET",request="/hello",service="test",status_code="200"} 3
	`
)

func NewTestPrometheusRepository(collector prometheus.Collector) *PrometheusRepository {
	r := prometheus.NewRegistry()

	pr := &PrometheusRepository{
		Registry: r,
		metrics:  []prometheus.Collector{collector},
	}

	return pr
}

func PrometheusTestApp(middleware *Middleware) http.Handler {
	n := negroni.Classic()
	n.Use(middleware)

	r := httprouter.New()

	for _, path := range serverConfigPaths {
		r.GET(path, func(res http.ResponseWriter, req *http.Request, p httprouter.Params) {
			fmt.Fprint(res, "OK")
		})
	}
	n.UseHandler(r)
	return n
}

var prometheusParams = []struct {
	name            string
	collapsePaths   bool
	hidePaths       bool
	expectedMetrics string
}{
	{"Not collapsed paths", false, false, metricsNotCollapsed},
	{"Collapsed paths", true, false, metricsCollapsed},
	{"Hidden not collapsed paths", false, true, metricsHidden},
	{"Hidden collapsed paths", true, true, metricsHidden},
}

func TestPrometheusRequestTotalMetrics(t *testing.T) {
	for _, tt := range prometheusParams {
		t.Run(tt.name, func(t *testing.T) {
			// re-initialize to prevent double counts
			RequestTotal.Reset()

			promRepo := NewTestPrometheusRepository(RequestTotal)
			promMiddleware := NewMiddleware(promRepo, "test")
			promMiddleware.CollapsePaths(tt.collapsePaths)
			promMiddleware.HidePaths(tt.hidePaths)

			ts := httptest.NewServer(PrometheusTestApp(promMiddleware))
			defer ts.Close()

			for _, path := range serverRequestPaths {
				req, err := http.NewRequest("GET", ts.URL+path, nil)
				if err != nil {
					t.Fatal(err)
				}
				client := &http.Client{}
				_, err = client.Do(req)
				if err != nil {
					t.Fatal(err)
				}
			}
			if err := testutil.CollectAndCompare(RequestTotal, strings.NewReader(tt.expectedMetrics), "ory_oathkeeper_requests_total"); err != nil {
				t.Fatal(err)
			}
		})
	}
}

var configurablePrometheusParams = []struct {
	name            string
	collapsePaths   bool
	expectedMetrics string
}{
	{"Not collapsed paths", false, configurableMetricsNotCollapsed},
	{"Collapsed paths", true, configurableMetricsCollapsed},
}

func TestConfigurablePrometheusRequestTotalMetrics(t *testing.T) {
	for _, tt := range configurablePrometheusParams {
		t.Run(tt.name, func(t *testing.T) {
			// re-initialize to prevent double counts
			RequestTotal.Reset()

			logger := logrusx.New("ORY Oathkeeper", "1")
			d := driver.NewDefaultDriver(logger, "1", "test", time.Now().String(), nil,
				configx.WithConfigFiles(x.WriteFile(t, `
serve:
  prometheus:
    metric_name_prefix: http_
`)),
			)
			promRepo := NewConfigurablePrometheusRepository(d, logger)
			promMiddleware := NewMiddleware(promRepo, "test")
			promMiddleware.CollapsePaths(tt.collapsePaths)

			ts := httptest.NewServer(PrometheusTestApp(promMiddleware))
			defer ts.Close()

			for _, path := range serverRequestPaths {
				req, err := http.NewRequest("GET", ts.URL+path, nil)
				if err != nil {
					t.Fatal(err)
				}
				client := &http.Client{}
				_, err = client.Do(req)
				if err != nil {
					t.Fatal(err)
				}
			}
			if err := testutil.CollectAndCompare(RequestTotal, strings.NewReader(tt.expectedMetrics), "http_requests_total"); err != nil {
				t.Fatal(err)
			}
		})
	}
}

var requestURIParams = []struct {
	name         string
	originalPath string
	firstSegment string
}{
	{"root path", "/", "/"},
	{"single segment", "/test", "/test"},
	{"two segments", "/test/path", "/test"},
	{"multiple segments", "/test/path/segments", "/test"},
}

func TestMiddlewareGetFirstPathSegment(t *testing.T) {
	promMiddleware := NewMiddleware(nil, "test")

	for _, tt := range requestURIParams {
		t.Run(tt.name, func(t *testing.T) {
			promMiddleware.CollapsePaths(true)
			collapsed := promMiddleware.getFirstPathSegment(tt.originalPath)
			if collapsed != tt.firstSegment {
				t.Fatalf("Expected first segment: %s to be equal to: %s", collapsed, tt.firstSegment)
			}
		})
	}
}
