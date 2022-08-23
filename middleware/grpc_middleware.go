package middleware

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/ory/herodot"

	"github.com/ory/oathkeeper/rule"
)

var ErrDenied = herodot.ErrUnauthorized

// httpRequest builds an HTTP request equivalent that is used for rule matching.
func (m *middleware) httpRequest(ctx context.Context, fullMethod string) (*http.Request, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, fmt.Errorf("no metadata found in context")
	}
	m.Logger().WithField("middleware", "oathkeeper").Debugf("request metadata: %+v", md)

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

	return &http.Request{
		Method:     "POST",
		Proto:      "HTTP/2",
		ProtoMajor: 2,
		URL:        u,
		Host:       u.Host,
		Header:     header,
	}, nil
}

// UnaryInterceptor returns the gRPC unary interceptor of the middleware.
func (m *middleware) UnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler) (resp interface{}, err error) {

		log := m.Logger().WithField("middleware", "oathkeeper")

		httpReq, err := m.httpRequest(ctx, info.FullMethod)
		if err != nil {
			log.WithError(err).Warn("could not build HTTP request")
			return nil, ErrDenied
		}
		log.Debugf("http request: %v", httpReq)

		r, err := m.RuleMatcher().Match(ctx, httpReq.Method, httpReq.URL, rule.ProtocolGRPC)
		if err != nil {
			log.WithError(err).Warn("could not find a matching rule")
			return nil, ErrDenied
		}

		_, err = m.ProxyRequestHandler().HandleRequest(httpReq, r)
		if err != nil {
			log.WithError(err).Warn("failed to handle request")
			return nil, ErrDenied
		}

		log.Info("access request granted")
		return handler(ctx, req)
	}
}

// StreamInterceptor returns the gRPC stream interceptor of the middleware.
func (m *middleware) StreamInterceptor() grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		stream grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler) error {

		log := m.Logger().WithField("middleware", "oathkeeper")
		ctx := stream.Context()

		httpReq, err := m.httpRequest(ctx, info.FullMethod)
		if err != nil {
			log.WithError(err).Warn("could not build HTTP request")
			return ErrDenied
		}
		log.Debugf("http request: %v", httpReq)

		r, err := m.RuleMatcher().Match(ctx, httpReq.Method, httpReq.URL, rule.ProtocolHTTP)
		if err != nil {
			log.WithError(err).Warn("could not find a matching rule")
			return ErrDenied
		}

		_, err = m.ProxyRequestHandler().HandleRequest(httpReq, r)
		if err != nil {
			log.WithError(err).Warn("failed to handle request")
			return ErrDenied
		}

		log.Info("access request granted")
		return handler(srv, stream)
	}
}
