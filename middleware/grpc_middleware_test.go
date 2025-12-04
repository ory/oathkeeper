// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package middleware_test

//go:generate mockgen -destination=grpc_mock_server_test.go -package=middleware_test google.golang.org/grpc/interop/grpc_testing TestServiceServer

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/interop/grpc_testing"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"

	"github.com/ory/oathkeeper/driver"
	"github.com/ory/oathkeeper/driver/configuration"
	"github.com/ory/oathkeeper/middleware"
	"github.com/ory/oathkeeper/rule"
)

func testClient(t *testing.T, l *bufconn.Listener, dialOpts ...grpc.DialOption) grpc_testing.TestServiceClient {
	conn, err := grpc.Dial("bufnet", //nolint:staticcheck // grpc.Dial is adequate for tests in grpc v1.x
		append(dialOpts,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithAuthority("myproject.apis.ory.sh"),
			grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) { return l.Dial() }),
		)...,
	)
	require.NoError(t, err)
	t.Cleanup(func() { conn.Close() }) //nolint:errcheck,gosec // cleanup best effort

	return grpc_testing.NewTestServiceClient(conn)
}

func testTokenCheckServer(t *testing.T) *httptest.Server {
	s := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("authorization") != "Bearer correct token" {
				t.Logf("denied request %+v", r)
				w.WriteHeader(http.StatusForbidden)
				return
			}
			t.Logf("allowed request %+v", r)
			io.WriteString(w, "{}") //nolint:errcheck,gosec // best-effort response write
		}))
	t.Cleanup(s.Close)
	return s
}

func writeTestConfig(t *testing.T, pattern string, content string) string {
	f, err := os.CreateTemp(t.TempDir(), pattern)
	if err != nil {
		t.Error(err)
		return ""
	}
	defer f.Close()            //nolint:errcheck
	io.WriteString(f, content) //nolint:errcheck,gosec // helper ignores write errors

	return f.Name()
}

type testToken string

func (t testToken) GetRequestMetadata(context.Context, ...string) (map[string]string, error) {
	return map[string]string{"authorization": "Bearer " + string(t)}, nil
}
func (t testToken) RequireTransportSecurity() bool { return false }

type upstream struct {
	*MockTestServiceServer
	grpc_testing.UnsafeTestServiceServer
}

func TestMiddleware(t *testing.T) {
	ctx := context.Background()

	tokenCheckServer := testTokenCheckServer(t)

	ctrl := gomock.NewController(t)
	upstream := upstream{MockTestServiceServer: NewMockTestServiceServer(ctrl)}

	config := writeTestConfig(t, "config-*.yaml", fmt.Sprintf(`
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
`, tokenCheckServer.URL))

	regPtr := new(driver.Registry)
	mw, err := middleware.New(ctx,
		middleware.WithConfigFile(config),
		middleware.WithRegistry(regPtr),
	)
	require.NoError(t, err)
	reg := *regPtr

	l := bufconn.Listen(1024 * 1024)
	s := grpc.NewServer(
		grpc.UnaryInterceptor(mw.UnaryInterceptor()),
		grpc.StreamInterceptor(mw.StreamInterceptor()),
	)
	grpc_testing.RegisterTestServiceServer(s, upstream)
	go func() {
		if err := s.Serve(l); err != nil {
			t.Logf("Server exited with error: %v", err)
		}
	}()
	t.Cleanup(s.Stop)

	upstream.EXPECT().
		EmptyCall(gomock.Any(), gomock.Any()).
		AnyTimes().
		Return(&grpc_testing.Empty{}, nil)

	cases := []struct {
		name     string
		rules    map[configuration.MatchingStrategy][]rule.Rule
		dialOpts []grpc.DialOption
		assert   assert.ErrorAssertionFunc
	}{
		{
			name: "should fail on empty ruleset",
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
					Match: &rule.MatchGRPC{
						Authority:  "<([0-9a-zA-Z\\-]+)>.apis.ory.sh",
						FullMethod: "grpc.testing.TestService/EmptyCall",
					},
					Authenticators: []rule.Handler{{Handler: "anonymous"}},
					Authorizer:     rule.Handler{Handler: "allow"},
					Mutators:       []rule.Handler{{Handler: "noop"}},
				}},
				configuration.Glob: {{
					Match: &rule.MatchGRPC{
						Authority:  "<**>.apis.ory.sh",
						FullMethod: "grpc.testing.TestService/EmptyCall",
					},
					Authenticators: []rule.Handler{{Handler: "anonymous"}},
					Authorizer:     rule.Handler{Handler: "allow"},
					Mutators:       []rule.Handler{{Handler: "noop"}},
				}},
			},
			assert: assert.NoError,
		},
		{
			name: "should succeed on matching rule with wildcard",
			rules: map[configuration.MatchingStrategy][]rule.Rule{
				configuration.Regexp: {{
					Match: &rule.MatchGRPC{
						Authority:  "<.*>",
						FullMethod: "<.*>Empty<.*>",
					},
					Authenticators: []rule.Handler{{Handler: "anonymous"}},
					Authorizer:     rule.Handler{Handler: "allow"},
					Mutators:       []rule.Handler{{Handler: "noop"}},
				}},
				configuration.Glob: {{
					Match: &rule.MatchGRPC{
						Authority:  "<**>.apis.ory.sh",
						FullMethod: "grpc.testing.TestService/Empty<**>",
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
					Match: &rule.MatchGRPC{
						Authority:  "<[0-9a-zA-Z]+>.apis.ory.sh",
						FullMethod: "grpc.testing.TestService/Empty<[a-zA-Z]+>-mismatch",
					},
					Authenticators: []rule.Handler{{Handler: "anonymous"}},
					Authorizer:     rule.Handler{Handler: "allow"},
					Mutators:       []rule.Handler{{Handler: "noop"}},
				}},
				configuration.Glob: {{
					Match: &rule.MatchGRPC{
						Authority:  "<**>.apis.ory.sh",
						FullMethod: "grpc.testing.TestService/Empty<**>-mismatch",
					},
					Authenticators: []rule.Handler{{Handler: "anonymous"}},
					Authorizer:     rule.Handler{Handler: "allow"},
					Mutators:       []rule.Handler{{Handler: "noop"}},
				}},
			},
			assert: assertErrDenied,
		},
		{
			name: "should fail because not a gRPC rule",
			rules: map[configuration.MatchingStrategy][]rule.Rule{
				configuration.Regexp: {{
					Match: &rule.Match{
						Methods: []string{"POST"},
						URL:     "grpc://<[0-9a-zA-Z]+>.apis.ory.sh/grpc.testing.TestService/EmptyCall",
					},
					Authenticators: []rule.Handler{{Handler: "anonymous"}},
					Authorizer:     rule.Handler{Handler: "allow"},
					Mutators:       []rule.Handler{{Handler: "noop"}},
				}},
				configuration.Glob: {{
					Match: &rule.Match{
						Methods: []string{"POST"},
						URL:     "grpc://<**>.apis.ory.sh/grpc.testing.TestService/EmptyCall",
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
					Match: &rule.MatchGRPC{
						Authority:  "<([0-9a-zA-Z\\-]+)>.apis.ory.sh",
						FullMethod: "grpc.testing.TestService/EmptyCall",
					},
					Authenticators: []rule.Handler{{Handler: "bearer_token"}},
					Authorizer:     rule.Handler{Handler: "allow"},
					Mutators:       []rule.Handler{{Handler: "noop"}},
				}},
				configuration.Glob: {{
					Match: &rule.MatchGRPC{
						Authority:  "<**>.apis.ory.sh",
						FullMethod: "grpc.testing.TestService/EmptyCall",
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
					Match: &rule.MatchGRPC{
						Authority:  "<([0-9a-zA-Z\\-]+)>.apis.ory.sh",
						FullMethod: "grpc.testing.TestService/EmptyCall",
					},
					Authenticators: []rule.Handler{{Handler: "bearer_token"}},
					Authorizer:     rule.Handler{Handler: "allow"},
					Mutators:       []rule.Handler{{Handler: "noop"}},
				}},
				configuration.Glob: {{
					Match: &rule.MatchGRPC{
						Authority:  "<**>.apis.ory.sh",
						FullMethod: "grpc.testing.TestService/EmptyCall",
					},
					Authenticators: []rule.Handler{{Handler: "bearer_token"}},
					Authorizer:     rule.Handler{Handler: "allow"},
					Mutators:       []rule.Handler{{Handler: "noop"}},
				}},
			},
			dialOpts: []grpc.DialOption{grpc.WithPerRPCCredentials(
				testToken("correct token"))},
			assert: assert.NoError,
		},
		{
			name: "should fail when incorrect bearer token is supplied",
			rules: map[configuration.MatchingStrategy][]rule.Rule{
				configuration.Regexp: {{
					Match: &rule.MatchGRPC{
						Authority:  "<([0-9a-zA-Z\\-]+)>.apis.ory.sh",
						FullMethod: "grpc.testing.TestService/EmptyCall",
					},
					Authenticators: []rule.Handler{{Handler: "bearer_token"}},
					Authorizer:     rule.Handler{Handler: "allow"},
					Mutators:       []rule.Handler{{Handler: "noop"}},
				}},
				configuration.Glob: {{
					Match: &rule.MatchGRPC{
						Authority:  "<**>.apis.ory.sh",
						FullMethod: "grpc.testing.TestService/EmptyCall",
					},
					Authenticators: []rule.Handler{{Handler: "bearer_token"}},
					Authorizer:     rule.Handler{Handler: "allow"},
					Mutators:       []rule.Handler{{Handler: "noop"}},
				}},
			},
			dialOpts: []grpc.DialOption{grpc.WithPerRPCCredentials(
				testToken("incorrect token"))},
			assert: assertErrDenied,
		},
	}

	strategies := []configuration.MatchingStrategy{configuration.Regexp, configuration.Glob}

	for _, tc := range cases {
		t.Run("case="+tc.name, func(t *testing.T) {
			client := testClient(t, l, tc.dialOpts...)
			for _, s := range strategies {
				t.Run("strategy="+string(s), func(t *testing.T) {
					require.NoError(t, reg.RuleRepository().SetMatchingStrategy(ctx, s))
					require.NoError(t, reg.RuleRepository().Set(ctx, tc.rules[s]))

					_, err := client.EmptyCall(ctx, &grpc_testing.Empty{})
					tc.assert(t, err)

					_, err = client.UnaryCall(ctx, &grpc_testing.SimpleRequest{})
					assertErrDenied(t, err)

					stream, _ := client.StreamingOutputCall(ctx, &grpc_testing.StreamingOutputCallRequest{})
					_, err = stream.Recv()
					assertErrDenied(t, err)
				})
			}
		})
	}
}

func assertErrDenied(t assert.TestingT, err error, _ ...interface{}) bool {
	s, ok := status.FromError(err)
	assert.Truef(t, ok, "error %v is not a status.Error (type: %T)", err, err)
	assert.Equal(t, codes.Unauthenticated, s.Code())
	return true
}

// Test that the middleware config does not read values from the environment.
func TestMiddleware_EnvironmentIsolation(t *testing.T) {
	ctx := context.Background()
	envVals := []string{"true", "false"}
	for _, envVal := range envVals {
		t.Run("AUTHENTICATORS_NOOP_ENABLED="+envVal, func(t *testing.T) {
			t.Setenv("AUTHENTICATORS_NOOP_ENABLED", envVal)

			configFile := writeTestConfig(t, "config-*.yaml", "")
			configPtr := new(configuration.Provider)
			_, err := middleware.New(ctx,
				middleware.WithConfigFile(configFile),
				middleware.WithConfigProvider(configPtr),
			)
			require.NoError(t, err)
			config := *configPtr

			assert.Falsef(t, config.Get("authenticators.noop.enabled").(bool), "was: %v", config.Get("authenticators.noop.enabled"))
		})
	}
}

func TestMiddleware_LoadRulesFromJSON(t *testing.T) {
	ctx := context.Background()

	jsonRule := `
	{
  "authenticators": [
   {
    "handler": "noop"
   }
  ],
  "authorizer": {
   "handler": "allow"
  },
  "id": "some-rule-id",
  "match": {
   "methods": [
    "POST"
   ],
   "url": "<(https|http)>://example.com:8080/service/webhooks<(|/.*)>"
  },
  "mutators": [
   {
    "handler": "noop"
   }
  ],
  "upstream": {
   "preserve_host": true,
   "strip_path": "/service",
   "url": "http://example.svc.cluster.local"
  }
}`
	var expected rule.Rule
	require.NoError(t, json.Unmarshal([]byte(jsonRule), &expected))
	rulesFile := writeTestConfig(t, "access-rules-*.json", "["+jsonRule+"]")

	configFile := writeTestConfig(t, "config-*.yaml", fmt.Sprintf(`
access_rules:
  matching_strategy: regexp
  repositories:
  - file://%s
authenticators:
  noop:
    enabled: true
authorizers:
  allow:
    enabled: true
mutators:
  noop:
    enabled: true
`, rulesFile))

	regPtr := new(driver.Registry)
	_, err := middleware.New(ctx,
		middleware.WithConfigFile(configFile),
		middleware.WithRegistry(regPtr),
	)
	require.NoError(t, err)
	reg := *regPtr

	time.Sleep(100 * time.Millisecond)

	actual, err := reg.RuleRepository().Get(ctx, "some-rule-id")
	require.NoError(t, err)
	assert.Equal(t, &expected, actual)
}
