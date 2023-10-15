// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/ory/herodot"
	"github.com/ory/x/otelx"

	"github.com/ory/oathkeeper/rule"
)

var ErrDenied = herodot.ErrUnauthorized

// httpRequest builds an HTTP request equivalent that is used for rule matching.
func (m *middleware) httpRequest(ctx context.Context, fullMethod string) (*http.Request, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, fmt.Errorf("no metadata found in context")
	}
	m.Logger().WithField("middleware", "oathkeeper").WithField("metadata", m.Logger().HTTPHeadersRedacted(http.Header(md))).Debug("using request metadata to build http header")

	authorities := md.Get(":authority")
	if len(authorities) != 1 {
		return nil, fmt.Errorf("no authority found in metadata")
	}

	header := make(http.Header)
	for key, vals := range md {
		k := strings.ToLower(key)
		if k == "authorization" || strings.HasPrefix(k, "x-") {
			for _, val := range vals {
				header.Add(key, val)
			}
		}
	}

	u := &url.URL{
		Host:   authorities[0],
		Path:   fullMethod,
		Scheme: "grpc",
	}

	return (&http.Request{
		Method:     "POST",
		Proto:      "HTTP/2",
		ProtoMajor: 2,
		URL:        u,
		Host:       u.Host,
		Header:     header,
	}).WithContext(ctx), nil
}

var (
	_ grpc.UnaryServerInterceptor  = new(middleware).unaryInterceptor
	_ grpc.StreamServerInterceptor = new(middleware).streamInterceptor
)

// UnaryInterceptor returns the gRPC unary interceptor of the middleware.
func (m *middleware) UnaryInterceptor() grpc.UnaryServerInterceptor {
	return m.unaryInterceptor
}

func (m *middleware) unaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	traceCtx, span := trace.SpanFromContext(ctx).TracerProvider().Tracer("oathkeeper/middleware").Start(ctx, "middleware.UnaryInterceptor")
	defer otelx.End(span, &err)

	log := m.Logger().WithField("middleware", "oathkeeper")

	httpReq, err := m.httpRequest(traceCtx, info.FullMethod)
	if err != nil {
		log.WithError(err).Warn("could not build HTTP request")
		span.SetAttributes(attribute.String("oathkeeper.verdict", "denied"))
		span.SetStatus(codes.Error, err.Error())
		return nil, ErrDenied
	}
	log = log.WithRequest(httpReq)

	log.Debug("matching HTTP request build from gRPC")

	r, err := m.RuleMatcher().Match(traceCtx, httpReq.Method, httpReq.URL, rule.ProtocolGRPC)
	if err != nil {
		log.WithError(err).Warn("could not find a matching rule")
		span.SetAttributes(attribute.String("oathkeeper.verdict", "denied"))
		span.SetStatus(codes.Error, err.Error())
		return nil, ErrDenied
	}

	_, err = m.ProxyRequestHandler().HandleRequest(httpReq, r)
	if err != nil {
		log.WithError(err).Warn("failed to handle request")
		span.SetAttributes(attribute.String("oathkeeper.verdict", "denied"))
		span.SetStatus(codes.Error, err.Error())
		return nil, ErrDenied
	}

	log.Info("access request granted")
	span.SetAttributes(attribute.String("oathkeeper.verdict", "allowed"))
	span.End()
	return handler(ctx, req)
}

// StreamInterceptor returns the gRPC stream interceptor of the middleware.
func (m *middleware) StreamInterceptor() grpc.StreamServerInterceptor {
	return m.streamInterceptor
}

func (m *middleware) streamInterceptor(
	srv interface{},
	stream grpc.ServerStream,
	info *grpc.StreamServerInfo,
	handler grpc.StreamHandler) (err error) {
	ctx := stream.Context()
	ctx, span := trace.SpanFromContext(ctx).TracerProvider().Tracer("oathkeeper/middleware").Start(ctx, "middleware.StreamInterceptor")
	otelx.End(span, &err)

	log := m.Logger().WithField("middleware", "oathkeeper")

	httpReq, err := m.httpRequest(ctx, info.FullMethod)
	if err != nil {
		log.WithError(err).Warn("could not build HTTP request")
		span.SetAttributes(attribute.String("oathkeeper.verdict", "denied"))
		span.SetStatus(codes.Error, err.Error())
		return ErrDenied
	}
	log = log.WithRequest(httpReq)

	log.Debug("matching HTTP request build from gRPC")

	r, err := m.RuleMatcher().Match(ctx, httpReq.Method, httpReq.URL, rule.ProtocolGRPC)
	if err != nil {
		log.WithError(err).Warn("could not find a matching rule")
		span.SetAttributes(attribute.String("oathkeeper.verdict", "denied"))
		span.SetStatus(codes.Error, err.Error())
		return ErrDenied
	}

	_, err = m.ProxyRequestHandler().HandleRequest(httpReq, r)
	if err != nil {
		log.WithError(err).Warn("failed to handle request")
		span.SetAttributes(attribute.String("oathkeeper.verdict", "denied"))
		span.SetStatus(codes.Error, err.Error())
		return ErrDenied
	}

	log.Info("access request granted")
	span.SetAttributes(attribute.String("oathkeeper.verdict", "allowed"))
	span.End()
	return handler(srv, stream)
}
