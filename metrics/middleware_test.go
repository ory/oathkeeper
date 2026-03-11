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
			_, _ = fmt.Fprint(res, "OK")
		})
	}
	n.UseHandler(r)
	return n
}

var prometheusParams = []struct {
	name               string
	collapsePaths      bool
	collapsePathsDepth int
	hidePaths          bool
	expectedMetrics    string
}{
	{"Not collapsed paths", false, 1, false, metricsNotCollapsed},
	{"Collapsed paths", true, 1, false, metricsCollapsed},
	{"Hidden not collapsed paths", false, 1, true, metricsHidden},
	{"Hidden collapsed paths", true, 1, true, metricsHidden},
}

func TestPrometheusRequestTotalMetrics(t *testing.T) {
	for _, tt := range prometheusParams {
		t.Run(tt.name, func(t *testing.T) {
			// re-initialize to prevent double counts
			RequestTotal.Reset()

			promRepo := NewTestPrometheusRepository(RequestTotal)
			promMiddleware := NewMiddleware(promRepo, "test")
			promMiddleware.CollapsePaths(tt.collapsePaths, tt.collapsePathsDepth)
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
	name               string
	collapsePaths      bool
	collapsePathsDepth int
	expectedMetrics    string
}{
	{"Not collapsed paths", false, 1, configurableMetricsNotCollapsed},
	{"Collapsed paths", true, 1, configurableMetricsCollapsed},
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
			promMiddleware.CollapsePaths(tt.collapsePaths, tt.collapsePathsDepth)

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
	name          string
	originalPath  string
	firstSegment  string
	collapseDepth int
}{
	// depth = 1 (default behaviour)
	{"root path depth 1", "/", "/", 1},
	{"single segment depth 1", "/test", "/test", 1},
	{"single segment depth 1 with query", "/test?foo=bar", "/test", 1},
	{"two segments depth 1", "/test/path", "/test", 1},
	{"multiple segments depth 1", "/test/path/segments", "/test", 1},
	{"with query depth 1", "/test/path?foo=bar", "/test", 1},
	{"three segments depth 1", "/test/path/segments", "/test", 1},
	{"four segments depth 1", "/test/path/segments/foo", "/test", 1},

	// depth = 2
	{"root path depth 2", "/", "/", 2},
	{"single segment depth 2", "/test", "/test", 2},
	{"single segment depth 2 with query", "/test?foo=bar", "/test", 2},
	{"two segments depth 2", "/test/path", "/test/path", 2},
	{"two segments depth 2 with query", "/test/path?foo=bar", "/test/path", 2},
	{"multiple segments depth 2", "/test/path/segments", "/test/path", 2},
	{"with query depth 2", "/test/path/segments?foo=bar", "/test/path", 2},

	// depth = 3
	{"root path depth 3", "/", "/", 3},
	{"single segment depth 3", "/test", "/test", 3},
	{"two segments depth 3", "/test/path", "/test/path", 3},
	{"two segments depth 3 with query", "/test/path?foo=bar", "/test/path", 3},
	{"multiple segments depth 3", "/test/path/segments", "/test/path/segments", 3},
	{"with query depth 3", "/test/path/segments?foo=bar", "/test/path/segments", 3},
	{"four segments depth 3", "/test/path/segments/foo", "/test/path/segments", 3},
	{"five segments depth 3", "/test/path/segments/foo/bar", "/test/path/segments", 3},
}

func TestMiddlewareGetFirstNPathSegments(t *testing.T) {
	promMiddleware := NewMiddleware(nil, "test")

	for _, tt := range requestURIParams {
		t.Run(tt.name, func(t *testing.T) {
			promMiddleware.CollapsePaths(true, tt.collapseDepth)
			collapsed := promMiddleware.getFirstNPathSegments(tt.originalPath, tt.collapseDepth)
			if collapsed != tt.firstSegment {
				t.Fatalf("Expected first segment: %s to be equal to: %s", collapsed, tt.firstSegment)
			}
		})
	}
}
