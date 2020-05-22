package server

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/julienschmidt/httprouter"
	"github.com/ory/x/reqlog"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"go.opentelemetry.io/otel/api/trace"
	"gotest.tools/assert"

	"github.com/urfave/negroni"
)

func TracingApp(logger *reqlog.Middleware) http.Handler {
	n := negroni.Classic()
	n.Use(logger)

	r := httprouter.New()

	r.GET("/", func(res http.ResponseWriter, req *http.Request, p httprouter.Params) {
		fmt.Fprint(res, "Test OpenTracing logs")
	})
	n.UseHandler(r)
	return n
}

var tracingParams = []struct {
	name           string
	traceID        string
	spanID         string
	tracingIsValid bool
}{
	{"with valid tracing", "82c5500f40667e5500e9ae8e9711553c", "992631f881f78c3b", true},
	{"with invalid tracing", "invalid", "invalid", false},
}

func TestLogTracingDecorator(t *testing.T) {
	for _, tt := range tracingParams {
		t.Run(tt.name, func(t *testing.T) {
			logger, hook := test.NewNullLogger()
			tracingLogger := reqlog.NewMiddlewareFromLogger(logger, "test-opentracing")
			// add before function to simulate pre-existing behaviour
			tracingLogger.Before = func(entry *logrus.Entry, r *http.Request, remoteAddr string) *logrus.Entry {
				return entry.WithField("before_tracing", "value")
			}
			// then add pre-existing fields
			tracingLogger = addOpenTracingFields(tracingLogger)
			ts := httptest.NewServer(TracingApp(tracingLogger))
			defer ts.Close()
			req, err := http.NewRequest("GET", ts.URL, nil)
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("traceparent", fmt.Sprintf("00-%s-%s-01", tt.traceID, tt.spanID))
			client := &http.Client{}
			res, err := client.Do(req)
			if err != nil {
				t.Fatal(err)
			}
			body, err := ioutil.ReadAll(res.Body)
			res.Body.Close()
			if err != nil {
				t.Fatal(err)
			}

			exp := "Test OpenTracing logs"
			if exp != string(body) {
				t.Fatalf("Expected body content: %s got: %s", exp, body)
			}
			for _, entry := range hook.Entries {
				// the first before should always be called
				if _, ok := entry.Data["before_tracing"]; !ok {
					t.Fatalf("Cannot extract before_tracing field from the log entry. Previous BeforeFunc was not called correctly")
				}

				if traceID, ok := entry.Data["trace_id"]; !ok {
					assert.Equal(t, false, tt.tracingIsValid)
				} else {
					traceID, ok := traceID.(trace.ID)
					if !ok {
						t.Fatalf("Cannot extract trace id from log entry. Expected %s", traceID)
					}
					spanID, ok := entry.Data["span_id"].(trace.SpanID)
					if !ok && tt.tracingIsValid {
						t.Fatalf("Cannot extract span id from log entry. Expected %s", spanID)
					}
					assert.Equal(t, tt.traceID, fmt.Sprintf("%s", traceID))
					assert.Equal(t, tt.spanID, fmt.Sprintf("%s", spanID))
				}
			}
		})
	}
}
