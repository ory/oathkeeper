/*
 * Copyright © 2017-2018 Aeneas Rekkas <aeneas+oss@aeneas.io>
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * @author       Aeneas Rekkas <aeneas+oss@aeneas.io>
 * @copyright  2017-2018 Aeneas Rekkas <aeneas+oss@aeneas.io>
 * @license  	   Apache-2.0
 */

package proxy

import (
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"testing"

	"github.com/ory/ladon/compiler"
	"github.com/ory/oathkeeper/rule"
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

func newTestRequest(t *testing.T, u string) *http.Request {
	p, err := url.Parse(u)
	require.NoError(t, err)

	return &http.Request{
		URL: p,
	}
}

func TestRequestHandler(t *testing.T) {
	for k, tc := range []struct {
		rule      rule.Rule
		r         *http.Request
		expectErr bool
		j         *RequestHandler
	}{
		{
			expectErr: true,
			r:         newTestRequest(t, "http://localhost"),
			j:         NewRequestHandler(nil, []Authenticator{}, []Authorizer{}, []CredentialsIssuer{}),
			rule: rule.Rule{
				Authenticators:    []rule.RuleHandler{},
				Authorizer:        rule.RuleHandler{},
				CredentialsIssuer: rule.RuleHandler{},
			},
		},
		{
			expectErr: true,
			r:         newTestRequest(t, "http://localhost"),
			j:         NewRequestHandler(nil, []Authenticator{NewAuthenticatorNoOp()}, []Authorizer{NewAuthorizerAllow()}, []CredentialsIssuer{NewCredentialsIssuerNoOp()}),
			rule: rule.Rule{
				Authenticators:    []rule.RuleHandler{},
				Authorizer:        rule.RuleHandler{},
				CredentialsIssuer: rule.RuleHandler{},
			},
		},
		{
			expectErr: false,
			r:         newTestRequest(t, "http://localhost"),
			j:         NewRequestHandler(nil, []Authenticator{NewAuthenticatorNoOp()}, []Authorizer{NewAuthorizerAllow()}, []CredentialsIssuer{NewCredentialsIssuerNoOp()}),
			rule: rule.Rule{
				Authenticators:    []rule.RuleHandler{{Handler: NewAuthenticatorNoOp().GetID()}},
				Authorizer:        rule.RuleHandler{Handler: NewAuthorizerAllow().GetID()},
				CredentialsIssuer: rule.RuleHandler{Handler: NewCredentialsIssuerNoOp().GetID()},
			},
		},
		{
			expectErr: true,
			r:         newTestRequest(t, "http://localhost"),
			j:         NewRequestHandler(nil, []Authenticator{NewAuthenticatorAnonymous("anonymous")}, []Authorizer{}, []CredentialsIssuer{}),
			rule: rule.Rule{
				Authenticators:    []rule.RuleHandler{{Handler: NewAuthenticatorAnonymous("").GetID()}},
				Authorizer:        rule.RuleHandler{},
				CredentialsIssuer: rule.RuleHandler{},
			},
		},
		{
			expectErr: true,
			r:         newTestRequest(t, "http://localhost"),
			j:         NewRequestHandler(nil, []Authenticator{NewAuthenticatorAnonymous("anonymous")}, []Authorizer{NewAuthorizerAllow()}, []CredentialsIssuer{}),
			rule: rule.Rule{
				Authenticators:    []rule.RuleHandler{{Handler: NewAuthenticatorAnonymous("").GetID()}},
				Authorizer:        rule.RuleHandler{Handler: "allow"},
				CredentialsIssuer: rule.RuleHandler{},
			},
		},
		{
			expectErr: false,
			r:         newTestRequest(t, "http://localhost"),
			j:         NewRequestHandler(nil, []Authenticator{NewAuthenticatorAnonymous("anonymous")}, []Authorizer{NewAuthorizerAllow()}, []CredentialsIssuer{NewCredentialsIssuerNoOp()}),
			rule: rule.Rule{
				Authenticators:    []rule.RuleHandler{{Handler: NewAuthenticatorAnonymous("").GetID()}},
				Authorizer:        rule.RuleHandler{Handler: "allow"},
				CredentialsIssuer: rule.RuleHandler{Handler: "noop"},
			},
		},
	} {
		t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
			if tc.expectErr {
				require.Error(t, tc.j.HandleRequest(tc.r, &tc.rule))
			} else {
				require.NoError(t, tc.j.HandleRequest(tc.r, &tc.rule))
			}
		})
	}
}
