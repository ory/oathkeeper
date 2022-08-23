package middleware_test

//go:generate mockgen -destination=grpc_mock_server_test.go -package=middleware_test google.golang.org/grpc/test/grpc_testing TestServiceServer

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	gomock "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
	grpcTesting "google.golang.org/grpc/test/grpc_testing"

	"github.com/ory/oathkeeper/driver"
	"github.com/ory/oathkeeper/driver/configuration"
	"github.com/ory/oathkeeper/middleware"
	"github.com/ory/oathkeeper/rule"
)

func testClient(t *testing.T, l *bufconn.Listener, dialOpts ...grpc.DialOption) grpcTesting.TestServiceClient {
	conn, err := grpc.Dial("bufnet",
		append(dialOpts,
			grpc.WithInsecure(),
			grpc.WithAuthority("myproject.apis.ory.sh"),
			grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) { return l.Dial() }),
		)...,
	)
	require.NoError(t, err)
	t.Cleanup(func() { conn.Close() })

	return grpcTesting.NewTestServiceClient(conn)
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
			io.WriteString(w, "{}")
		}))
	t.Cleanup(s.Close)
	return s
}

func writeTestConfig(t *testing.T, content string) string {
	f, err := os.CreateTemp(t.TempDir(), "config-*.yaml")
	if err != nil {
		t.Error(err)
		return ""
	}
	defer f.Close()
	io.WriteString(f, content)

	return f.Name()
}

type testToken string

func (t testToken) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	return map[string]string{"authorization": "Bearer " + string(t)}, nil
}
func (t testToken) RequireTransportSecurity() bool { return false }

func TestMiddleware(t *testing.T) {
	ctx := context.Background()

	tokenCheckServer := testTokenCheckServer(t)

	ctrl := gomock.NewController(t)
	upstream := NewMockTestServiceServer(ctrl)

	config := writeTestConfig(t, fmt.Sprintf(`
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
	mw := middleware.New(middleware.WithConfig(config), middleware.WithRegistry(regPtr))
	reg := *regPtr

	l := bufconn.Listen(1024 * 1024)
	s := grpc.NewServer(
		grpc.UnaryInterceptor(mw.UnaryInterceptor()),
		grpc.StreamInterceptor(mw.StreamInterceptor()),
	)
	grpcTesting.RegisterTestServiceServer(s, upstream)
	go func() {
		if err := s.Serve(l); err != nil {
			t.Logf("Server exited with error: %v", err)
		}
	}()
	t.Cleanup(s.Stop)

	upstream.EXPECT().
		EmptyCall(gomock.Any(), gomock.Any()).
		AnyTimes().
		Return(&grpcTesting.Empty{}, nil)

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
					Match:          &rule.Match{Methods: []string{"POST"}, URL: "grpc://<([0-9a-zA-Z\\-]+)>.apis.ory.sh/grpc.testing.TestService/EmptyCall"},
					Authenticators: []rule.Handler{{Handler: "anonymous"}},
					Authorizer:     rule.Handler{Handler: "allow"},
					Mutators:       []rule.Handler{{Handler: "noop"}},
				}},
				configuration.Glob: {{
					Match:          &rule.Match{Methods: []string{"POST"}, URL: "grpc://<**>.apis.ory.sh/grpc.testing.TestService/EmptyCall"},
					Authenticators: []rule.Handler{{Handler: "anonymous"}},
					Authorizer:     rule.Handler{Handler: "allow"},
					Mutators:       []rule.Handler{{Handler: "noop"}},
				}},
			},
			assert: assert.NoError,
		},
		{
			name: "should fail when no bearer token is supplied",
			rules: map[configuration.MatchingStrategy][]rule.Rule{
				configuration.Regexp: {{
					Match:          &rule.Match{Methods: []string{"POST"}, URL: "grpc://<([0-9a-zA-Z\\-]+)>.apis.ory.sh/grpc.testing.TestService/EmptyCall"},
					Authenticators: []rule.Handler{{Handler: "bearer_token"}},
					Authorizer:     rule.Handler{Handler: "allow"},
					Mutators:       []rule.Handler{{Handler: "noop"}},
				}},
				configuration.Glob: {{
					Match:          &rule.Match{Methods: []string{"POST"}, URL: "grpc://<**>.apis.ory.sh/grpc.testing.TestService/EmptyCall"},
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
					Match:          &rule.Match{Methods: []string{"POST"}, URL: "grpc://<([0-9a-zA-Z\\-]+)>.apis.ory.sh/grpc.testing.TestService/EmptyCall"},
					Authenticators: []rule.Handler{{Handler: "bearer_token"}},
					Authorizer:     rule.Handler{Handler: "allow"},
					Mutators:       []rule.Handler{{Handler: "noop"}},
				}},
				configuration.Glob: {{
					Match:          &rule.Match{Methods: []string{"POST"}, URL: "grpc://<**>.apis.ory.sh/grpc.testing.TestService/EmptyCall"},
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
					Match:          &rule.Match{Methods: []string{"POST"}, URL: "grpc://<([0-9a-zA-Z\\-]+)>.apis.ory.sh/grpc.testing.TestService/EmptyCall"},
					Authenticators: []rule.Handler{{Handler: "bearer_token"}},
					Authorizer:     rule.Handler{Handler: "allow"},
					Mutators:       []rule.Handler{{Handler: "noop"}},
				}},
				configuration.Glob: {{
					Match:          &rule.Match{Methods: []string{"POST"}, URL: "grpc://<**>.apis.ory.sh/grpc.testing.TestService/EmptyCall"},
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

					require.NoError(t, reg.RuleRepository().
						SetMatchingStrategy(ctx, s))
					reg.RuleRepository().(*rule.RepositoryMemory).
						WithRules(tc.rules[s])

					_, err := client.EmptyCall(ctx, &grpcTesting.Empty{})
					tc.assert(t, err)

					_, err = client.UnaryCall(ctx, &grpcTesting.SimpleRequest{})
					assertErrDenied(t, err)

					stream, _ := client.StreamingOutputCall(ctx, &grpcTesting.StreamingOutputCallRequest{})
					_, err = stream.Recv()
					assertErrDenied(t, err)
				})
			}
		})
	}
}

func assertErrDenied(t assert.TestingT, err error, _ ...interface{}) bool {
	return assert.ErrorIs(t, err, middleware.ErrDenied)
}
