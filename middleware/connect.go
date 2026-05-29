// Copyright © 2026 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"net/http"
	"net/url"

	"connectrpc.com/connect"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/ory/x/otelx"

	"github.com/ory/oathkeeper/rule"
)

func (m *middleware) ConnectInterceptor() connect.Interceptor { return m }

// httpRequestFromConnect builds an HTTP request equivalent that is used for rule matching.
func httpRequestFromConnect(ctx context.Context, header http.Header, spec connect.Spec, method string) *http.Request {
	u := &url.URL{
		Scheme: "rpc",
		Host:   header.Get("Host"),
		Path:   spec.Procedure,
	}

	return (&http.Request{
		Method:     method,
		Proto:      "HTTP/2",
		ProtoMajor: 2,
		URL:        u,
		Host:       u.Host,
		Header:     header,
	}).WithContext(ctx)
}

var _ connect.Interceptor = (*middleware)(nil)

func (m *middleware) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return func(ctx context.Context, req connect.AnyRequest) (_ connect.AnyResponse, err error) {
		traceCtx, span := trace.SpanFromContext(ctx).TracerProvider().Tracer("oathkeeper/middleware").Start(ctx, "middleware.UnaryInterceptorConnect")
		defer otelx.End(span, &err)

		log := m.reg.Logger().WithField("middleware", "oathkeeper")

		httpReq := httpRequestFromConnect(traceCtx, req.Header(), req.Spec(), req.HTTPMethod())
		log = log.WithRequest(httpReq)

		log.Debug("matching HTTP request build from connect")

		r, err := m.reg.RuleMatcher().Match(traceCtx, httpReq.Method, httpReq.URL, rule.ProtocolRPC)
		if err != nil {
			log.WithError(err).Warn("could not find a matching rule")
			span.SetAttributes(attribute.String("oathkeeper.verdict", "denied"))
			span.SetStatus(codes.Error, err.Error())
			return nil, connect.NewError(connect.CodeUnauthenticated, ErrDenied())
		}

		_, err = m.reg.ProxyRequestHandler().HandleRequest(httpReq, r)
		if err != nil {
			log.WithError(err).Warn("failed to handle request")
			span.SetAttributes(attribute.String("oathkeeper.verdict", "denied"))
			span.SetStatus(codes.Error, err.Error())
			return nil, connect.NewError(connect.CodeUnauthenticated, ErrDenied())
		}

		log.Info("access request granted")
		span.SetAttributes(attribute.String("oathkeeper.verdict", "allowed"))
		span.End()
		return next(ctx, req)
	}
}

func (m *middleware) WrapStreamingClient(clientFunc connect.StreamingClientFunc) connect.StreamingClientFunc {
	// Oathkeeper middleware on the client doesn't make any sense.
	return clientFunc
}

func (m *middleware) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return func(ctx context.Context, conn connect.StreamingHandlerConn) (err error) {
		ctx, span := trace.SpanFromContext(ctx).TracerProvider().Tracer("oathkeeper/middleware").Start(ctx, "middleware.StreamInterceptorConnect")
		defer otelx.End(span, &err)

		log := m.reg.Logger().WithField("middleware", "oathkeeper")

		httpReq := httpRequestFromConnect(ctx, conn.RequestHeader(), conn.Spec(), "POST")
		log = log.WithRequest(httpReq)

		log.Debug("matching HTTP request build from connect")

		r, err := m.reg.RuleMatcher().Match(ctx, httpReq.Method, httpReq.URL, rule.ProtocolRPC)
		if err != nil {
			log.WithError(err).Warn("could not find a matching rule")
			span.SetAttributes(attribute.String("oathkeeper.verdict", "denied"))
			span.SetStatus(codes.Error, err.Error())
			return connect.NewError(connect.CodeUnauthenticated, ErrDenied())
		}

		_, err = m.reg.ProxyRequestHandler().HandleRequest(httpReq, r)
		if err != nil {
			log.WithError(err).Warn("failed to handle request")
			span.SetAttributes(attribute.String("oathkeeper.verdict", "denied"))
			span.SetStatus(codes.Error, err.Error())
			return connect.NewError(connect.CodeUnauthenticated, ErrDenied())
		}

		log.Info("access request granted")
		span.SetAttributes(attribute.String("oathkeeper.verdict", "allowed"))
		span.End()
		return next(ctx, conn)
	}
}
