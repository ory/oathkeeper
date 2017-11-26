package evaluator

import (
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/ory/hydra/sdk/go/hydra"
	"github.com/ory/hydra/sdk/go/hydra/swagger"
	"github.com/ory/ladon/compiler"
	"github.com/ory/oathkeeper/rule"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mustCompileRegex(t *testing.T, pattern string) *regexp.Regexp {
	exp, err := compiler.CompileRegex(pattern, '<', '>')
	require.NoError(t, err)
	return exp
}

func mustGenerateURL(t *testing.T, u string) *url.URL {
	up, err := url.Parse(u)
	require.NoError(t, err)
	return up
}

func TestEvaluator(t *testing.T) {
	we := NewWardenEvaluator(nil, nil, nil, "")
	publicRule := rule.Rule{MatchesMethods: []string{"GET"}, MatchesURLCompiled: mustCompileRegex(t, "http://localhost/users/<[0-9]+>"), Mode: rule.AnonymousMode}
	bypassACPRule := rule.Rule{MatchesMethods: []string{"GET"}, MatchesURLCompiled: mustCompileRegex(t, "http://localhost/users/<[0-9]+>"), Mode: rule.AuthenticatedMode}
	privateRuleWithSubstitution := rule.Rule{
		MatchesMethods:     []string{"POST"},
		MatchesURLCompiled: mustCompileRegex(t, "http://localhost/users/<[0-9]+>"),
		RequiredResource:   "users:$1",
		RequiredAction:     "get:$1",
		RequiredScopes:     []string{"users.create"},
		Mode:               rule.PolicyMode,
	}
	privateRuleWithoutSubstitution := rule.Rule{
		MatchesMethods:     []string{"POST"},
		MatchesURLCompiled: mustCompileRegex(t, "http://localhost/users<$|/([0-9]+)>"),
		RequiredResource:   "users",
		RequiredAction:     "get",
		RequiredScopes:     []string{"users.create"},
		Mode:               rule.PolicyMode,
	}
	privateRuleWithPartialSubstitution := rule.Rule{
		MatchesMethods:     []string{"POST"},
		MatchesURLCompiled: mustCompileRegex(t, "http://localhost/users<$|/([0-9]+)>"),
		RequiredResource:   "users:$2",
		RequiredAction:     "get",
		RequiredScopes:     []string{"users.create"},
		Mode:               rule.PolicyMode,
	}

	for k, tc := range []struct {
		d     string
		rules []rule.Rule
		r     *http.Request
		e     func(*testing.T, *Session, error)
		mock  func(*gomock.Controller) hydra.SDK
	}{
		{
			d:     "request is allowed anonymously because it matches a public rule and does not contain a bearer token",
			rules: []rule.Rule{publicRule},
			r:     &http.Request{Method: "GET", Host: "localhost", URL: mustGenerateURL(t, "http://localhost/users/1234")},
			e: func(t *testing.T, s *Session, err error) {
				require.NoError(t, err)
				assert.Empty(t, s.ClientID)
				assert.Empty(t, s.User)
				assert.True(t, s.Anonymous)
			},
			mock: func(c *gomock.Controller) hydra.SDK {
				return NewMockSDK(c)
			},
		},
		{
			d:     "request is denied because no rule exists",
			rules: []rule.Rule{},
			r:     &http.Request{Method: "GET", Host: "localhost", URL: mustGenerateURL(t, "http://localhost/users/1234")},
			e: func(t *testing.T, s *Session, err error) {
				require.Error(t, err)
			},
			mock: func(c *gomock.Controller) hydra.SDK {
				return NewMockSDK(c)
			},
		},
		{
			d:     "request is denied because no rule matches",
			rules: []rule.Rule{},
			r:     &http.Request{Method: "POST", Host: "localhost", URL: mustGenerateURL(t, "http://localhost/users/1234")},
			e: func(t *testing.T, s *Session, err error) {
				require.Error(t, err)
			},
			mock: func(c *gomock.Controller) hydra.SDK {
				return NewMockSDK(c)
			},
		},
		{
			d:     "request is denied because multiple rules match",
			rules: []rule.Rule{publicRule, publicRule},
			r:     &http.Request{Method: "GET", Host: "localhost", URL: mustGenerateURL(t, "http://localhost/users/1234")},
			e: func(t *testing.T, s *Session, err error) {
				require.Error(t, err)
			},
			mock: func(c *gomock.Controller) hydra.SDK {
				return NewMockSDK(c)
			},
		},
		{
			d:     "request is allowed anonymously because it matches a public rule, but token introspection fails with a network connection issue",
			rules: []rule.Rule{publicRule},
			r:     &http.Request{Method: "GET", Host: "localhost", Header: http.Header{"Authorization": []string{"bEaReR token"}}, URL: mustGenerateURL(t, "http://localhost/users/1234")},
			e: func(t *testing.T, s *Session, err error) {
				require.NoError(t, err)
				assert.Empty(t, s.ClientID)
				assert.Empty(t, s.User)
				assert.True(t, s.Anonymous)
			},
			mock: func(c *gomock.Controller) hydra.SDK {
				s := NewMockSDK(c)
				s.EXPECT().IntrospectOAuth2Token(gomock.Eq("token"), gomock.Eq("")).Return(nil, nil, errors.New("error"))
				return s
			},
		},
		{
			d:     "request is allowed anonymously because it matches a public rule, but token introspection fails with status code 400",
			rules: []rule.Rule{publicRule},
			r:     &http.Request{Method: "GET", Host: "localhost", Header: http.Header{"Authorization": []string{"bEaReR token"}}, URL: mustGenerateURL(t, "http://localhost/users/1234")},
			e: func(t *testing.T, s *Session, err error) {
				require.NoError(t, err)
				assert.Empty(t, s.ClientID)
				assert.Empty(t, s.User)
				assert.True(t, s.Anonymous)
			},
			mock: func(c *gomock.Controller) hydra.SDK {
				s := NewMockSDK(c)
				s.EXPECT().IntrospectOAuth2Token(gomock.Eq("token"), gomock.Eq("")).Return(nil, &swagger.APIResponse{Response: &http.Response{StatusCode: 400}}, nil)
				return s
			},
		},
		{
			d:     "request is allowed anonymously because it matches a public rule, but token introspection says the token is no longer active",
			rules: []rule.Rule{publicRule},
			r:     &http.Request{Method: "GET", Host: "localhost", Header: http.Header{"Authorization": []string{"bEaReR token"}}, URL: mustGenerateURL(t, "http://localhost/users/1234")},
			e: func(t *testing.T, s *Session, err error) {
				require.NoError(t, err)
				assert.Empty(t, s.ClientID)
				assert.Empty(t, s.User)
				assert.True(t, s.Anonymous)
			},
			mock: func(c *gomock.Controller) hydra.SDK {
				s := NewMockSDK(c)
				s.EXPECT().IntrospectOAuth2Token(gomock.Eq("token"), gomock.Eq("")).Return(&swagger.OAuth2TokenIntrospection{Active: false, Sub: "user", ClientId: "client"}, &swagger.APIResponse{Response: &http.Response{StatusCode: http.StatusOK}}, nil)
				return s
			},
		},
		{
			d:     "request is allowed because it matches a public rule, and token introspection succeeds",
			rules: []rule.Rule{publicRule},
			r:     &http.Request{Method: "GET", Host: "localhost", Header: http.Header{"Authorization": []string{"bEaReR token"}}, URL: mustGenerateURL(t, "http://localhost/users/1234")},
			e: func(t *testing.T, s *Session, err error) {
				require.NoError(t, err)
				assert.Equal(t, "client", s.ClientID)
				assert.Equal(t, "user", s.User)
				assert.False(t, s.Anonymous)
			},
			mock: func(c *gomock.Controller) hydra.SDK {
				s := NewMockSDK(c)
				s.EXPECT().IntrospectOAuth2Token(gomock.Eq("token"), gomock.Eq("")).Return(&swagger.OAuth2TokenIntrospection{Active: true, Sub: "user", ClientId: "client"}, &swagger.APIResponse{Response: &http.Response{StatusCode: http.StatusOK}}, nil)
				return s
			},
		},
		{
			d:     "request is not allowed because it matches a rule without access control policies, but token introspection fails",
			rules: []rule.Rule{bypassACPRule},
			r:     &http.Request{Method: "GET", Host: "localhost", Header: http.Header{"Authorization": []string{"bEaReR token"}}, URL: mustGenerateURL(t, "http://localhost/users/1234")},
			e: func(t *testing.T, s *Session, err error) {
				require.Error(t, err)
			},
			mock: func(c *gomock.Controller) hydra.SDK {
				s := NewMockSDK(c)
				s.EXPECT().IntrospectOAuth2Token(gomock.Eq("token"), gomock.Eq("")).Return(&swagger.OAuth2TokenIntrospection{Active: false}, &swagger.APIResponse{Response: &http.Response{StatusCode: http.StatusOK}}, nil)
				return s
			},
		},
		{
			d:     "request is not allowed because it matches a rule without access control policies, but token introspection fails with network error",
			rules: []rule.Rule{bypassACPRule},
			r:     &http.Request{Method: "GET", Host: "localhost", Header: http.Header{"Authorization": []string{"bEaReR token"}}, URL: mustGenerateURL(t, "http://localhost/users/1234")},
			e: func(t *testing.T, s *Session, err error) {
				require.Error(t, err)
			},
			mock: func(c *gomock.Controller) hydra.SDK {
				s := NewMockSDK(c)
				s.EXPECT().IntrospectOAuth2Token(gomock.Eq("token"), gomock.Eq("")).Return(nil, nil, errors.New("Some error"))
				return s
			},
		},
		{
			d:     "request is not allowed because it matches a rule without access control policies, but token introspection fails with wrong status code",
			rules: []rule.Rule{bypassACPRule},
			r:     &http.Request{Method: "GET", Host: "localhost", Header: http.Header{"Authorization": []string{"bEaReR token"}}, URL: mustGenerateURL(t, "http://localhost/users/1234")},
			e: func(t *testing.T, s *Session, err error) {
				require.Error(t, err)
			},
			mock: func(c *gomock.Controller) hydra.SDK {
				s := NewMockSDK(c)
				s.EXPECT().IntrospectOAuth2Token(gomock.Eq("token"), gomock.Eq("")).Return(nil, &swagger.APIResponse{Response: &http.Response{StatusCode: http.StatusUnauthorized}}, nil)
				return s
			},
		},
		{
			d:     "request is allowed because it matches a rule without access control policies, and token introspection succeeds",
			rules: []rule.Rule{bypassACPRule},
			r:     &http.Request{Method: "GET", Host: "localhost", Header: http.Header{"Authorization": []string{"bEaReR token"}}, URL: mustGenerateURL(t, "http://localhost/users/1234")},
			e: func(t *testing.T, s *Session, err error) {
				require.NoError(t, err)
				assert.Equal(t, "client", s.ClientID)
				assert.Equal(t, "user", s.User)
				assert.False(t, s.Anonymous)
			},
			mock: func(c *gomock.Controller) hydra.SDK {
				s := NewMockSDK(c)
				s.EXPECT().IntrospectOAuth2Token(gomock.Eq("token"), gomock.Eq("")).Return(&swagger.OAuth2TokenIntrospection{Active: true, Sub: "user", ClientId: "client"}, &swagger.APIResponse{Response: &http.Response{StatusCode: http.StatusOK}}, nil)
				return s
			},
		},
		{
			d:     "request is denied because token is missing and endpoint is not public",
			rules: []rule.Rule{privateRuleWithSubstitution},
			r:     &http.Request{Method: "POST", Host: "localhost", Header: http.Header{"Authorization": []string{"bEaReR"}}, URL: mustGenerateURL(t, "http://localhost/users/1234")},
			e: func(t *testing.T, s *Session, err error) {
				require.Error(t, err)
			},
			mock: func(c *gomock.Controller) hydra.SDK {
				return nil
			},
		},
		{
			d:     "request is denied because warden request fails with a network error and endpoint is not public",
			rules: []rule.Rule{privateRuleWithSubstitution},
			r:     &http.Request{Method: "POST", Host: "localhost", Header: http.Header{"Authorization": []string{"bEaReR token"}}, URL: mustGenerateURL(t, "http://localhost/users/1234")},
			e: func(t *testing.T, s *Session, err error) {
				require.Error(t, err)
			},
			mock: func(c *gomock.Controller) hydra.SDK {
				s := NewMockSDK(c)
				s.EXPECT().DoesWardenAllowTokenAccessRequest(gomock.Any()).Return(nil, nil, errors.New("error)"))
				return s
			},
		},
		{
			d:     "request is denied because warden request fails with a 400 status code and endpoint is not public",
			rules: []rule.Rule{privateRuleWithSubstitution},
			r:     &http.Request{Method: "POST", Host: "localhost", Header: http.Header{"Authorization": []string{"bEaReR token"}}, URL: mustGenerateURL(t, "http://localhost/users/1234")},
			e: func(t *testing.T, s *Session, err error) {
				require.Error(t, err)
			},
			mock: func(c *gomock.Controller) hydra.SDK {
				s := NewMockSDK(c)
				s.EXPECT().DoesWardenAllowTokenAccessRequest(gomock.Any()).Return(nil, &swagger.APIResponse{Response: &http.Response{StatusCode: http.StatusBadRequest}}, nil)
				return s
			},
		},
		{
			d:     "request is denied because warden request fails with allowed=false",
			rules: []rule.Rule{privateRuleWithSubstitution},
			r:     &http.Request{Method: "POST", Host: "localhost", Header: http.Header{"Authorization": []string{"bEaReR token"}}, URL: mustGenerateURL(t, "http://localhost/users/1234")},
			e: func(t *testing.T, s *Session, err error) {
				require.Error(t, err)
			},
			mock: func(c *gomock.Controller) hydra.SDK {
				s := NewMockSDK(c)
				s.EXPECT().DoesWardenAllowTokenAccessRequest(gomock.Any()).Return(&swagger.WardenTokenAccessRequestResponse{Allowed: false}, &swagger.APIResponse{Response: &http.Response{StatusCode: http.StatusOK}}, nil)
				return s
			},
		},
		{
			d:     "request is allowed because token is valid and allowed (rule with substitution)",
			rules: []rule.Rule{privateRuleWithSubstitution},
			r:     &http.Request{RemoteAddr: "127.0.0.1:1234", Method: "POST", Host: "localhost", Header: http.Header{"Authorization": []string{"bEaReR token"}}, URL: mustGenerateURL(t, "http://localhost/users/1234")},
			e: func(t *testing.T, s *Session, err error) {
				require.NoError(t, err)
			},
			mock: func(c *gomock.Controller) hydra.SDK {
				s := NewMockSDK(c)
				s.EXPECT().DoesWardenAllowTokenAccessRequest(gomock.Eq(swagger.WardenTokenAccessRequest{
					Token:    "token",
					Resource: "users:1234",
					Action:   "get:1234",
					Scopes:   []string{"users.create"},
					Context:  map[string]interface{}{"remoteIpAddress": "127.0.0.1"},
				})).Return(&swagger.WardenTokenAccessRequestResponse{Allowed: true}, &swagger.APIResponse{Response: &http.Response{StatusCode: http.StatusOK}}, nil)
				return s
			},
		},
		{
			d:     "request is allowed because token is valid and allowed (rule with partial substitution)",
			rules: []rule.Rule{privateRuleWithPartialSubstitution},
			r:     &http.Request{RemoteAddr: "127.0.0.1:1234", Method: "POST", Host: "localhost", Header: http.Header{"Authorization": []string{"bEaReR token"}}, URL: mustGenerateURL(t, "http://localhost/users/1234")},
			e: func(t *testing.T, s *Session, err error) {
				require.NoError(t, err)
			},
			mock: func(c *gomock.Controller) hydra.SDK {
				s := NewMockSDK(c)
				s.EXPECT().DoesWardenAllowTokenAccessRequest(gomock.Eq(swagger.WardenTokenAccessRequest{
					Token:    "token",
					Resource: "users:1234",
					Action:   "get",
					Scopes:   []string{"users.create"},
					Context:  map[string]interface{}{"remoteIpAddress": "127.0.0.1"},
				})).Return(&swagger.WardenTokenAccessRequestResponse{Allowed: true}, &swagger.APIResponse{Response: &http.Response{StatusCode: http.StatusOK}}, nil)
				return s
			},
		},
		{
			d:     "request is allowed because token is valid and allowed (rule with partial substitution and path parameter)",
			rules: []rule.Rule{privateRuleWithoutSubstitution},
			r:     &http.Request{RemoteAddr: "127.0.0.1:1234", Method: "POST", Host: "localhost", Header: http.Header{"Authorization": []string{"bEaReR token"}}, URL: mustGenerateURL(t, "http://localhost/users/1234")},
			e: func(t *testing.T, s *Session, err error) {
				require.NoError(t, err)
			},
			mock: func(c *gomock.Controller) hydra.SDK {
				s := NewMockSDK(c)
				s.EXPECT().DoesWardenAllowTokenAccessRequest(gomock.Eq(swagger.WardenTokenAccessRequest{
					Token:    "token",
					Resource: "users",
					Action:   "get",
					Scopes:   []string{"users.create"},
					Context:  map[string]interface{}{"remoteIpAddress": "127.0.0.1"},
				})).Return(&swagger.WardenTokenAccessRequestResponse{Allowed: true}, &swagger.APIResponse{Response: &http.Response{StatusCode: http.StatusOK}}, nil)
				return s
			},
		},
		{
			d:     "request is allowed because token is valid and allowed (rule without substitution and path parameter)",
			rules: []rule.Rule{privateRuleWithoutSubstitution},
			r:     &http.Request{RemoteAddr: "127.0.0.1:1234", Method: "POST", Host: "localhost", Header: http.Header{"Authorization": []string{"bEaReR token"}}, URL: mustGenerateURL(t, "http://localhost/users")},
			e: func(t *testing.T, s *Session, err error) {
				require.NoError(t, err)
			},
			mock: func(c *gomock.Controller) hydra.SDK {
				s := NewMockSDK(c)
				s.EXPECT().DoesWardenAllowTokenAccessRequest(gomock.Eq(swagger.WardenTokenAccessRequest{
					Token:    "token",
					Resource: "users",
					Action:   "get",
					Scopes:   []string{"users.create"},
					Context:  map[string]interface{}{"remoteIpAddress": "127.0.0.1"},
				})).Return(&swagger.WardenTokenAccessRequestResponse{Allowed: true}, &swagger.APIResponse{Response: &http.Response{StatusCode: http.StatusOK}}, nil)
				return s
			},
		},
	} {
		t.Run(fmt.Sprintf("case=%d/description=%s", k, tc.d), func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			we.Matcher = &rule.CachedMatcher{Rules: tc.rules}
			we.Hydra = tc.mock(ctrl)

			s, err := we.EvaluateAccessRequest(tc.r)
			tc.e(t, s, err)
		})
	}
}
