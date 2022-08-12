package metrics

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/ory/oathkeeper/driver"
	"github.com/ory/viper"
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
	`
	metricsCollapsed string = metricMetadata + rootMetric + `
	ory_oathkeeper_requests_total{method="GET",request="/hello",service="test",status_code="200"} 1
	`
	serverContextPaths []string = []string{"/", "/hello/world"}

	configurableMetricMetadata string = `
	# HELP http_requests_total Total number of requests
	# TYPE http_requests_total counter
	`
	configurableRootMetric string = `
	http_requests_total{method="GET",request="/",service="test",status_code="200"} 1
	`
	configurableMetricsNotCollapsed string = configurableMetricMetadata + configurableRootMetric + `
	http_requests_total{method="GET",request="/hello/world",service="test",status_code="200"} 1
	`
	configurableMetricsCollapsed string = configurableMetricMetadata + configurableRootMetric + `
	http_requests_total{method="GET",request="/hello",service="test",status_code="200"} 1
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

	for _, path := range serverContextPaths {
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
	expectedMetrics string
}{
	{"Not collapsed paths", false, metricsNotCollapsed},
	{"Collapsed paths", true, metricsCollapsed},
}

func TestPrometheusRequestTotalMetrics(t *testing.T) {
	for _, tt := range prometheusParams {
		t.Run(tt.name, func(t *testing.T) {
			// re-initialize to prevent double counts
			RequestTotal.Reset()

			promRepo := NewTestPrometheusRepository(RequestTotal)
			promMiddleware := NewMiddleware(promRepo, "test")
			promMiddleware.CollapsePaths(tt.collapsePaths)

			ts := httptest.NewServer(PrometheusTestApp(promMiddleware))
			defer ts.Close()

			for _, path := range serverContextPaths {
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

			viper.SetConfigType("yaml")

			// We set the fallback to first run www_authenticate. But because the error is not_found, as
			// is defined in the when clause, we should see a json error instead!
			require.NoError(t, viper.ReadConfig(bytes.NewBufferString(`
serve:
  prometheus:
    metric_name_prefix: http_
`)))
			logger := logrusx.New("ORY Oathkeeper", "1")
			d := driver.NewDefaultDriver(logger, "1", "test", time.Now().String())
			promRepo := NewConfigurablePrometheusRepository(d, logger)
			promMiddleware := NewMiddleware(promRepo, "test")
			promMiddleware.CollapsePaths(tt.collapsePaths)

			ts := httptest.NewServer(PrometheusTestApp(promMiddleware))
			defer ts.Close()

			for _, path := range serverContextPaths {
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
