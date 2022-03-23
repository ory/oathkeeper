package x

import (
	"context"
	"net/http"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

func TraceRequest(ctx context.Context, req *http.Request) func() {
	tracer := opentracing.GlobalTracer()
	if tracer == nil {
		return func() {}
	}

	parentSpan := opentracing.SpanFromContext(ctx)
	opts := make([]opentracing.StartSpanOption, 0, 2)
	opts = append(opts, ext.SpanKindRPCClient)
	if parentSpan != nil {
		opts = append(opts, opentracing.ChildOf(parentSpan.Context()))
	}

	urlStr := req.URL.String()
	clientSpan := tracer.StartSpan("HTTP "+req.Method, opts...)

	ext.SpanKindRPCClient.Set(clientSpan)
	ext.HTTPUrl.Set(clientSpan, urlStr)
	ext.HTTPMethod.Set(clientSpan, req.Method)

	_ = tracer.Inject(clientSpan.Context(), opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(req.Header))
	return clientSpan.Finish
}
