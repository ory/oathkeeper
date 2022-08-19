package envoycheck

import (
	"context"
	"errors"
	"net/http"
	"net/url"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	authv3 "github.com/envoyproxy/go-control-plane/envoy/service/auth/v3"
	"github.com/ory/oathkeeper/proxy"
	"github.com/ory/oathkeeper/rule"
	"github.com/ory/x/logrusx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type (
	dependencies interface {
		Logger() *logrusx.Logger
		RuleMatcher() rule.Matcher
		ProxyRequestHandler() proxy.RequestHandler
	}

	// Server answers requests from Envoy's ext_authz HTTP filter by consulting
	// the configured rule set.
	Server struct{ dependencies }
)

var (
	denied = &authv3.CheckResponse{
		Status: status.New(codes.PermissionDenied, "denied").Proto(),
	}

	ErrInvalidCheckRequest = errors.New("invalid check request")

	contentLengthHeader = http.CanonicalHeaderKey("content-length")
)

func NewServer(d dependencies) *Server {
	return &Server{d}
}

func toHttpRequest(req *authv3.CheckRequest) (*http.Request, error) {
	if req == nil || req.Attributes == nil || req.Attributes.Request == nil || req.Attributes.Request.Http == nil {
		return nil, ErrInvalidCheckRequest
	}
	url, err := url.Parse(req.Attributes.Request.Http.Path)
	if err != nil {
		return nil, err
	}

	header := make(http.Header)
	for k, v := range req.Attributes.Request.Http.Headers {
		header.Add(k, v)
	}

	return &http.Request{
		Method: req.Attributes.Request.Http.Method,
		URL:    url,
		Header: header,
		Host:   req.Attributes.Request.Http.Host,
		Proto:  req.Attributes.Request.Http.Protocol,
	}, nil
}

func toProtoHeader(header http.Header) (result []*corev3.HeaderValueOption) {
	result = make([]*corev3.HeaderValueOption, 0, len(header))
	for k, vv := range header {
		// Avoid copying the original Content-Length header from the client
		if k == contentLengthHeader {
			continue
		}

		for i, v := range vv {
			result = append(result, &corev3.HeaderValueOption{
				Header: &corev3.HeaderValue{Key: k, Value: v},
				Append: wrapperspb.Bool(i > 0), // replace first, append rest
			})
		}
	}
	return
}

func (s *Server) Check(ctx context.Context, req *authv3.CheckRequest) (*authv3.CheckResponse, error) {
	log := s.Logger().WithField("method", "envoycheck.Check")

	if err := req.Validate(); err != nil {
		return nil, err
	}

	httpReq, err := toHttpRequest(req)
	if err != nil {
		log.WithError(err).
			Warn("could not extract HTTP information from request")
		return denied, nil
	}
	log = log.WithRequest(httpReq)

	rule, err := s.RuleMatcher().Match(ctx, httpReq.Method, httpReq.URL)
	if err != nil {
		log.WithError(err).Warn("could not find a matching rule")
		return denied, nil
	}

	session, err := s.ProxyRequestHandler().HandleRequest(httpReq, rule)
	if err != nil {
		log.WithError(err).Warn("failed to handle request")
		return denied, nil
	}

	log.Info("access request granted")
	return &authv3.CheckResponse{
		Status: status.New(codes.OK, "allowed").Proto(),
		HttpResponse: &authv3.CheckResponse_OkResponse{
			OkResponse: &authv3.OkHttpResponse{
				Headers: toProtoHeader(session.Header),
			},
		},
	}, nil
}
