// Copyright © 2026 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/ory/rpctest/gen/go/ory/rpctest"
	"github.com/ory/rpctest/gen/go/ory/rpctest/rpctestconnect"

	"github.com/ory/oathkeeper/driver/configuration"
	"github.com/ory/oathkeeper/middleware"
	"github.com/ory/oathkeeper/rule"
)

//go:generate go tool mockgen -destination=connect_mock_server_test.go -package=middleware_test github.com/ory/rpctest/gen/go/ory/rpctest/rpctestconnect TestServiceHandler

func TestConnectMiddleware(t *testing.T) {
	tokenCheckServer := testTokenCheckServer(t)
	config := writeTestConfigf(t, `
authenticators:
  noop:
    enabled: true
  anonymous:
    enabled: true
  bearer_token:
    enabled: true
    config:
      check_session_url: %s
authorizers:
  allow:
    enabled: true
mutators:
  noop:
    enabled: true
`, tokenCheckServer.URL)

	mw, err := middleware.New(t.Context(), middleware.WithConfigFile(config))
	require.NoError(t, err)
	reg := mw.Registry()

	ctrl := gomock.NewController(t)
	upstream := NewMockTestServiceHandler(ctrl)
	upstream.EXPECT().
		Unary(gomock.Any(), gomock.Any()).
		AnyTimes().
		Return(connect.NewResponse(&rpctest.UnaryResponse{}), nil)
	mux := http.NewServeMux()
	mux.Handle(rpctestconnect.NewTestServiceHandler(upstream, connect.WithInterceptors(mw.ConnectInterceptor())))

	ts := httptest.NewUnstartedServer(mux)
	ts.EnableHTTP2 = true
	ts.StartTLS()
	t.Cleanup(ts.Close)

	connectClient := func(bearer *string) rpctestconnect.TestServiceClient {
		return rpctestconnect.NewTestServiceClient(ts.Client(), ts.URL, connect.WithInterceptors(connect.UnaryInterceptorFunc(func(f connect.UnaryFunc) connect.UnaryFunc {
			return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
				if bearer != nil {
					req.Header().Set("Authorization", "Bearer "+*bearer)
				}
				req.Header().Set("Host", "myproject.example.com")
				return f(ctx, req)
			}
		})))
	}

	cases := []struct {
		name   string
		rules  map[configuration.MatchingStrategy][]rule.Rule
		bearer *string
		assert assert.ErrorAssertionFunc
	}{
		{
			name: "should deny on empty ruleset",
			rules: map[configuration.MatchingStrategy][]rule.Rule{
				configuration.Regexp: {},
				configuration.Glob:   {},
			},
			assert: assertErrDenied,
		},
		{
			name: "should succeed on matching rule",
			rules: map[configuration.MatchingStrategy][]rule.Rule{
				configuration.Regexp: {{
					Match: &rule.MatchRPC{
						Authority:  "<([0-9a-zA-Z\\-]+)>.example.com",
						FullMethod: "ory.rpctest.TestService/Unary",
					},
					Authenticators: []rule.Handler{{Handler: "anonymous"}},
					Authorizer:     rule.Handler{Handler: "allow"},
					Mutators:       []rule.Handler{{Handler: "noop"}},
				}},
				configuration.Glob: {{
					Match: &rule.MatchRPC{
						Authority:  "<**>.example.com",
						FullMethod: "ory.rpctest.TestService/Unary",
					},
					Authenticators: []rule.Handler{{Handler: "anonymous"}},
					Authorizer:     rule.Handler{Handler: "allow"},
					Mutators:       []rule.Handler{{Handler: "noop"}},
				}},
			},
			assert: assert.NoError,
		},
		{
			name: "should block on matching rule when authorizer denies",
			rules: map[configuration.MatchingStrategy][]rule.Rule{
				configuration.Regexp: {{
					Match: &rule.MatchRPC{
						Authority:  "<([0-9a-zA-Z\\-]+)>.example.com",
						FullMethod: "ory.rpctest.TestService/Unary",
					},
					Authenticators: []rule.Handler{{Handler: "anonymous"}},
					Authorizer:     rule.Handler{Handler: "deny"},
					Mutators:       []rule.Handler{{Handler: "noop"}},
				}},
				configuration.Glob: {{
					Match: &rule.MatchRPC{
						Authority:  "<**>.example.com",
						FullMethod: "ory.rpctest.TestService/Unary",
					},
					Authenticators: []rule.Handler{{Handler: "anonymous"}},
					Authorizer:     rule.Handler{Handler: "deny"},
					Mutators:       []rule.Handler{{Handler: "noop"}},
				}},
			},
			assert: assertErrDenied,
		},
		{
			name: "should succeed on matching rule with wildcard",
			rules: map[configuration.MatchingStrategy][]rule.Rule{
				configuration.Regexp: {{
					Match: &rule.MatchRPC{
						Authority:  "<.*>",
						FullMethod: "<.*>U<.*>",
					},
					Authenticators: []rule.Handler{{Handler: "anonymous"}},
					Authorizer:     rule.Handler{Handler: "allow"},
					Mutators:       []rule.Handler{{Handler: "noop"}},
				}},
				configuration.Glob: {{
					Match: &rule.MatchRPC{
						Authority:  "<**>.example.com",
						FullMethod: "ory.rpctest.TestService/U<**>",
					},
					Authenticators: []rule.Handler{{Handler: "anonymous"}},
					Authorizer:     rule.Handler{Handler: "allow"},
					Mutators:       []rule.Handler{{Handler: "noop"}},
				}},
			},
			assert: assert.NoError,
		},
		{
			name: "should fail on mis-matching rule with wildcard",
			rules: map[configuration.MatchingStrategy][]rule.Rule{
				configuration.Regexp: {{
					Match: &rule.MatchRPC{
						Authority:  "<[0-9a-zA-Z]+>.example.com",
						FullMethod: "ory.rpctest.TestService/U<[a-zA-Z]+>-mismatch",
					},
					Authenticators: []rule.Handler{{Handler: "anonymous"}},
					Authorizer:     rule.Handler{Handler: "allow"},
					Mutators:       []rule.Handler{{Handler: "noop"}},
				}},
				configuration.Glob: {{
					Match: &rule.MatchRPC{
						Authority:  "<**>.example.com",
						FullMethod: "ory.rpctest.TestService/U<**>-mismatch",
					},
					Authenticators: []rule.Handler{{Handler: "anonymous"}},
					Authorizer:     rule.Handler{Handler: "allow"},
					Mutators:       []rule.Handler{{Handler: "noop"}},
				}},
			},
			assert: assertErrDenied,
		},
		{
			name: "should fail because not an rpc rule",
			rules: map[configuration.MatchingStrategy][]rule.Rule{
				configuration.Regexp: {{
					Match: &rule.Match{
						Methods: []string{"POST"},
						URL:     "rpc://<[0-9a-zA-Z]+>.example.com/ory.rpctest.TestService/Unary",
					},
					Authenticators: []rule.Handler{{Handler: "anonymous"}},
					Authorizer:     rule.Handler{Handler: "allow"},
					Mutators:       []rule.Handler{{Handler: "noop"}},
				}},
				configuration.Glob: {{
					Match: &rule.Match{
						Methods: []string{"POST"},
						URL:     "rpc://<**>.example.com/ory.rpctest.TestService/Unary",
					},
					Authenticators: []rule.Handler{{Handler: "anonymous"}},
					Authorizer:     rule.Handler{Handler: "allow"},
					Mutators:       []rule.Handler{{Handler: "noop"}},
				}},
			},
			assert: assertErrDenied,
		},
		{
			name: "should fail when no bearer token is supplied",
			rules: map[configuration.MatchingStrategy][]rule.Rule{
				configuration.Regexp: {{
					Match: &rule.MatchRPC{
						Authority:  "<([0-9a-zA-Z\\-]+)>.example.com",
						FullMethod: "ory.rpctest.TestService/Unary",
					},
					Authenticators: []rule.Handler{{Handler: "bearer_token"}},
					Authorizer:     rule.Handler{Handler: "allow"},
					Mutators:       []rule.Handler{{Handler: "noop"}},
				}},
				configuration.Glob: {{
					Match: &rule.MatchRPC{
						Authority:  "<**>.example.com",
						FullMethod: "ory.rpctest.TestService/Unary",
					},
					Authenticators: []rule.Handler{{Handler: "bearer_token"}},
					Authorizer:     rule.Handler{Handler: "allow"},
					Mutators:       []rule.Handler{{Handler: "noop"}},
				}},
			},
			assert: assertErrDenied,
		},
		{
			name: "should succeed when correct bearer token is supplied",
			rules: map[configuration.MatchingStrategy][]rule.Rule{
				configuration.Regexp: {{
					Match: &rule.MatchRPC{
						Authority:  "<([0-9a-zA-Z\\-]+)>.example.com",
						FullMethod: "ory.rpctest.TestService/Unary",
					},
					Authenticators: []rule.Handler{{Handler: "bearer_token"}},
					Authorizer:     rule.Handler{Handler: "allow"},
					Mutators:       []rule.Handler{{Handler: "noop"}},
				}},
				configuration.Glob: {{
					Match: &rule.MatchRPC{
						Authority:  "<**>.example.com",
						FullMethod: "ory.rpctest.TestService/Unary",
					},
					Authenticators: []rule.Handler{{Handler: "bearer_token"}},
					Authorizer:     rule.Handler{Handler: "allow"},
					Mutators:       []rule.Handler{{Handler: "noop"}},
				}},
			},
			bearer: new("correct token"),
			assert: assert.NoError,
		},
		{
			name: "should fail when incorrect bearer token is supplied",
			rules: map[configuration.MatchingStrategy][]rule.Rule{
				configuration.Regexp: {{
					Match: &rule.MatchRPC{
						Authority:  "<([0-9a-zA-Z\\-]+)>.example.com",
						FullMethod: "ory.rpctest.TestService/Unary",
					},
					Authenticators: []rule.Handler{{Handler: "bearer_token"}},
					Authorizer:     rule.Handler{Handler: "allow"},
					Mutators:       []rule.Handler{{Handler: "noop"}},
				}},
				configuration.Glob: {{
					Match: &rule.MatchRPC{
						Authority:  "<**>.example.com",
						FullMethod: "ory.rpctest.TestService/Unary",
					},
					Authenticators: []rule.Handler{{Handler: "bearer_token"}},
					Authorizer:     rule.Handler{Handler: "allow"},
					Mutators:       []rule.Handler{{Handler: "noop"}},
				}},
			},
			bearer: new("incorrect token"),
			assert: assertErrDenied,
		},
	}

	strategies := []configuration.MatchingStrategy{configuration.Regexp, configuration.Glob}

	for _, tc := range cases {
		t.Run("case="+tc.name, func(t *testing.T) {
			grpcClient := testClient(t, ts, tc.bearer)
			connectClient := connectClient(tc.bearer)
			for _, s := range strategies {
				t.Run("strategy="+string(s), func(t *testing.T) {
					require.NoError(t, reg.RuleRepository().SetMatchingStrategy(t.Context(), s))
					require.NoError(t, reg.RuleRepository().Set(t.Context(), tc.rules[s]))

					_, err := grpcClient.Unary(t.Context(), &rpctest.UnaryRequest{})
					tc.assert(t, err)
					_, err = connectClient.Unary(t.Context(), connect.NewRequest(&rpctest.UnaryRequest{}))
					tc.assert(t, err)

					srvStream, err := grpcClient.ServerStream(t.Context(), &rpctest.ServerStreamRequest{})
					require.NoError(t, err)
					_, err = srvStream.Recv()
					assertErrDenied(t, err)
					cSrvStream, err := connectClient.ServerStream(t.Context(), connect.NewRequest(&rpctest.ServerStreamRequest{}))
					require.NoError(t, err)
					cSrvStream.Receive()
					assertErrDenied(t, cSrvStream.Err())

					clStream, err := grpcClient.ClientStream(t.Context())
					require.NoError(t, err)
					_, err = clStream.CloseAndRecv()
					assertErrDenied(t, err)
					cClStream := connectClient.ClientStream(t.Context())
					require.NoError(t, cClStream.Send(&rpctest.ClientStreamRequest{}))
					_, err = cClStream.CloseAndReceive()
					assertErrDenied(t, err)

					bidiStream, err := grpcClient.BidiStream(t.Context())
					require.NoError(t, err)
					_, err = bidiStream.Recv()
					assertErrDenied(t, err)
					cBidiStream := connectClient.BidiStream(t.Context())
					require.NoError(t, cBidiStream.CloseRequest())
					_, err = cBidiStream.Receive()
					assertErrDenied(t, err)
				})
			}
		})
	}
}
