package envoycheck_test

import (
	"context"
	"fmt"
	"net"
	"testing"

	authv3 "github.com/envoyproxy/go-control-plane/envoy/service/auth/v3"
	"github.com/ory/oathkeeper/driver"
	"github.com/ory/oathkeeper/driver/configuration"
	"github.com/ory/oathkeeper/internal"
	"github.com/ory/oathkeeper/rule"
	"github.com/ory/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/test/bufconn"
)

type testClient struct {
	Registry *driver.RegistryMemory
	authv3.AuthorizationClient
}

func newClient(t *testing.T) *testClient {
	ctx := context.Background()
	conf := internal.NewConfigurationWithDefaults()
	viper.Set(configuration.ViperKeyAuthenticatorNoopIsEnabled, true)
	viper.Set(configuration.ViperKeyAuthenticatorUnauthorizedIsEnabled, true)
	viper.Set(configuration.ViperKeyAuthenticatorAnonymousIsEnabled, true)
	viper.Set(configuration.ViperKeyAuthorizerAllowIsEnabled, true)
	viper.Set(configuration.ViperKeyAuthorizerDenyIsEnabled, true)
	viper.Set(configuration.ViperKeyMutatorNoopIsEnabled, true)
	viper.Set(configuration.ViperKeyErrorsWWWAuthenticateIsEnabled, true)
	reg := internal.NewRegistry(conf)

	l := bufconn.Listen(1024 * 1024)
	s := grpc.NewServer()
	authv3.RegisterAuthorizationServer(s, reg.EnvoyCheckServer())
	go func() {
		if err := s.Serve(l); err != nil {
			t.Logf("Server exited with error: %v", err)
		}
	}()
	t.Cleanup(s.Stop)

	conn, err := grpc.DialContext(ctx, "bufnet",
		grpc.WithInsecure(),
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) { return l.Dial() }))
	require.NoError(t, err)
	t.Cleanup(func() { conn.Close() })

	return &testClient{
		Registry:            reg,
		AuthorizationClient: authv3.NewAuthorizationClient(conn),
	}
}

func checkGetRequest(path string) *authv3.CheckRequest {
	return &authv3.CheckRequest{
		Attributes: &authv3.AttributeContext{
			Request: &authv3.AttributeContext_Request{
				Http: &authv3.AttributeContext_HttpRequest{
					Path:   path,
					Method: "GET",
				},
			},
		},
	}
}

func denied(t *testing.T, res *authv3.CheckResponse) {
	assert.Equal(t, int32(codes.PermissionDenied), res.Status.Code)
}
func allowed(t *testing.T, res *authv3.CheckResponse) {
	assert.Equal(t, int32(codes.OK), res.Status.Code)
}
func hasHeader(key, value string) func(*testing.T, *authv3.CheckResponse) {
	return func(t *testing.T, res *authv3.CheckResponse) {
		for _, header := range res.GetOkResponse().GetHeaders() {
			if header.Header.Key == key {
				assert.Equal(t, value, header.Header.Value)
				return
			}
		}
		t.Errorf("could not find header %q in %+v", key, res.GetOkResponse().GetHeaders())
	}
}
func noHeader(key string) func(*testing.T, *authv3.CheckResponse) {
	return func(t *testing.T, res *authv3.CheckResponse) {
		for _, header := range res.GetOkResponse().GetHeaders() {
			if header.Header.Key == key {
				t.Errorf("found header %q in %+v, should have been empty", key, res.GetOkResponse().GetHeaders())
				return
			}
		}
	}
}
func multiple(asserts ...func(*testing.T, *authv3.CheckResponse)) func(*testing.T, *authv3.CheckResponse) {
	return func(t *testing.T, cr *authv3.CheckResponse) {
		for _, assert := range asserts {
			assert(t, cr)
		}
	}
}

func TestCheckServer(t *testing.T) {
	ctx := context.Background()
	client := newClient(t)

	ruleNoOpAuthenticator := rule.Rule{
		Match:          &rule.Match{Methods: []string{"GET"}, URL: "http://example.com/authn-noop/<[0-9]+>"},
		Authenticators: []rule.Handler{{Handler: "noop"}},
		Authorizer:     rule.Handler{Handler: "allow"},
		Mutators:       []rule.Handler{{Handler: "noop"}},
		Upstream:       rule.Upstream{URL: ""},
	}
	ruleNoOpAuthenticatorGLOB := rule.Rule{
		Match:          &rule.Match{Methods: []string{"GET"}, URL: "http://example.com/authn-noop/<[0-9]*>"},
		Authenticators: []rule.Handler{{Handler: "noop"}},
		Authorizer:     rule.Handler{Handler: "allow"},
		Mutators:       []rule.Handler{{Handler: "noop"}},
		Upstream:       rule.Upstream{URL: ""},
	}
	ruleNoOpAuthenticatorModifyUpstream := rule.Rule{
		Match:          &rule.Match{Methods: []string{"GET"}, URL: "http://example.com/strip-path/authn-noop/<[0-9]+>"},
		Authenticators: []rule.Handler{{Handler: "noop"}},
		Authorizer:     rule.Handler{Handler: "allow"},
		Mutators:       []rule.Handler{{Handler: "noop"}},
		Upstream:       rule.Upstream{URL: "", StripPath: "/strip-path/", PreserveHost: true},
	}

	ruleNoOpAuthenticatorModifyUpstreamGLOB := rule.Rule{
		Match:          &rule.Match{Methods: []string{"GET"}, URL: "http://example.com/strip-path/authn-noop/<[0-9]*>"},
		Authenticators: []rule.Handler{{Handler: "noop"}},
		Authorizer:     rule.Handler{Handler: "allow"},
		Mutators:       []rule.Handler{{Handler: "noop"}},
		Upstream:       rule.Upstream{URL: "", StripPath: "/strip-path/", PreserveHost: true},
	}

	for k, tc := range []struct {
		d           string
		rulesRegexp []rule.Rule
		rulesGlob   []rule.Rule
		request     *authv3.CheckRequest
		assert      func(*testing.T, *authv3.CheckResponse)
	}{
		{
			d:           "should fail on empty check request",
			rulesRegexp: []rule.Rule{},
			rulesGlob:   []rule.Rule{},
			request:     &authv3.CheckRequest{},
			assert:      denied,
		},
		{
			d:           "should fail because url does exist but is matched by two rulesRegexp",
			rulesRegexp: []rule.Rule{ruleNoOpAuthenticator, ruleNoOpAuthenticator},
			rulesGlob:   []rule.Rule{ruleNoOpAuthenticatorGLOB, ruleNoOpAuthenticatorGLOB},
			request:     checkGetRequest("http://example.com/authn-noop/1234"),
			assert:      denied,
		},
		{
			d:           "should pass",
			rulesRegexp: []rule.Rule{ruleNoOpAuthenticator},
			rulesGlob:   []rule.Rule{ruleNoOpAuthenticatorGLOB},
			request: &authv3.CheckRequest{
				Attributes: &authv3.AttributeContext{
					Request: &authv3.AttributeContext_Request{
						Http: &authv3.AttributeContext_HttpRequest{
							Headers: map[string]string{
								"Authorization": "bearer token",
							},
							Path:   "http://example.com/authn-noop/1234",
							Method: "GET",
						},
					},
				},
			},
			assert: multiple(
				allowed,
				hasHeader("Authorization", "bearer token"),
			),
		},
		{
			d:           "should pass and strip path",
			rulesRegexp: []rule.Rule{ruleNoOpAuthenticatorModifyUpstream},
			rulesGlob:   []rule.Rule{ruleNoOpAuthenticatorModifyUpstreamGLOB},
			request: &authv3.CheckRequest{
				Attributes: &authv3.AttributeContext{
					Request: &authv3.AttributeContext_Request{
						Http: &authv3.AttributeContext_HttpRequest{
							Headers: map[string]string{
								"Authorization": "bearer token",
							},
							Path:   "http://example.com/strip-path/authn-noop/1234",
							Method: "GET",
						},
					},
				},
			},
			assert: multiple(
				allowed,
				hasHeader("Authorization", "bearer token"),
			),
		},
		{
			d: "should pass with anonymous and everything else set to noop",
			rulesRegexp: []rule.Rule{{
				Match:          &rule.Match{Methods: []string{"GET"}, URL: "http://example.com/authn-anon/authz-allow/cred-noop/<[0-9]+>"},
				Authenticators: []rule.Handler{{Handler: "anonymous"}},
				Authorizer:     rule.Handler{Handler: "allow"},
				Mutators:       []rule.Handler{{Handler: "noop"}},
				Upstream:       rule.Upstream{URL: ""},
			}},
			rulesGlob: []rule.Rule{{
				Match:          &rule.Match{Methods: []string{"GET"}, URL: "http://example.com/authn-anon/authz-allow/cred-noop/<[0-9]*>"},
				Authenticators: []rule.Handler{{Handler: "anonymous"}},
				Authorizer:     rule.Handler{Handler: "allow"},
				Mutators:       []rule.Handler{{Handler: "noop"}},
				Upstream:       rule.Upstream{URL: ""},
			}},
			request: checkGetRequest("http://example.com/authn-anon/authz-allow/cred-noop/1234"),
			assert: multiple(
				allowed,
				noHeader("Content-Length"),
			),
		},
		{
			d: "should fail when one of the mutators fails",
			rulesRegexp: []rule.Rule{{
				Match:          &rule.Match{Methods: []string{"GET"}, URL: "http://example.com/authn-anonymous/authz-allow/cred-broken/<[0-9]+>"},
				Authenticators: []rule.Handler{{Handler: "anonymous"}},
				Authorizer:     rule.Handler{Handler: "allow"},
				Mutators:       []rule.Handler{{Handler: "noop"}, {Handler: "broken"}},
				Upstream:       rule.Upstream{URL: ""},
			}},
			rulesGlob: []rule.Rule{{
				Match:          &rule.Match{Methods: []string{"GET"}, URL: "http://example.com/authn-anonymous/authz-allow/cred-broken/<[0-9]*>"},
				Authenticators: []rule.Handler{{Handler: "anonymous"}},
				Authorizer:     rule.Handler{Handler: "allow"},
				Mutators:       []rule.Handler{{Handler: "noop"}, {Handler: "broken"}},
				Upstream:       rule.Upstream{URL: ""},
			}},
			request: checkGetRequest("http://example.com/authn-anonymous/authz-allow/cred-broken/1234"),
			assert:  denied,
		},
		{
			d: "should not pass Content-Length from client",
			rulesRegexp: []rule.Rule{{
				Match:          &rule.Match{Methods: []string{"GET"}, URL: "http://example.com/authn-anon/authz-allow/cred-noop/<[0-9]+>"},
				Authenticators: []rule.Handler{{Handler: "anonymous"}},
				Authorizer:     rule.Handler{Handler: "allow"},
				Mutators:       []rule.Handler{{Handler: "noop"}},
				Upstream:       rule.Upstream{URL: ""},
			}},
			rulesGlob: []rule.Rule{{
				Match:          &rule.Match{Methods: []string{"GET"}, URL: "http://example.com/authn-anon/authz-allow/cred-noop/<[0-9]*>"},
				Authenticators: []rule.Handler{{Handler: "anonymous"}},
				Authorizer:     rule.Handler{Handler: "allow"},
				Mutators:       []rule.Handler{{Handler: "noop"}},
				Upstream:       rule.Upstream{URL: ""},
			}},
			request: &authv3.CheckRequest{
				Attributes: &authv3.AttributeContext{
					Request: &authv3.AttributeContext_Request{
						Http: &authv3.AttributeContext_HttpRequest{
							Headers: map[string]string{
								"Content-Length": "1337",
							},
							Path:   "http://example.com/authn-anon/authz-allow/cred-noop/1234",
							Method: "GET",
						},
					},
				},
			},
			assert: allowed,
		},
	} {
		t.Run(fmt.Sprintf("case=%d/description=%s", k, tc.d), func(t *testing.T) {
			testFunc := func(t *testing.T, strategy configuration.MatchingStrategy) {
				require.NoError(t, client.Registry.RuleRepository().SetMatchingStrategy(ctx, strategy))
				res, err := client.Check(ctx, tc.request)
				require.NoError(t, err)

				tc.assert(t, res)
			}
			t.Run("strategy=regexp", func(t *testing.T) {
				client.Registry.RuleRepository().(*rule.RepositoryMemory).WithRules(tc.rulesRegexp)
				testFunc(t, configuration.Regexp)
			})
			t.Run("strategy=glob", func(t *testing.T) {
				client.Registry.RuleRepository().(*rule.RepositoryMemory).WithRules(tc.rulesGlob)
				testFunc(t, configuration.Glob)
			})
		})
	}
}
