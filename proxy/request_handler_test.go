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
	"testing"

	"github.com/ory/oathkeeper/pipeline/authn"
	"github.com/ory/oathkeeper/pipeline/authz"
	"github.com/ory/oathkeeper/pipeline/mutate"

	"github.com/stretchr/testify/require"

	"github.com/ory/oathkeeper/rule"
)

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
			j:         NewRequestHandler(nil, []authn.Authenticator{}, []authz.Authorizer{}, []mutate.Mutator{}),
			rule: rule.Rule{
				Authenticators: []rule.RuleHandler{},
				Authorizer:     rule.RuleHandler{},
				Mutator:        rule.RuleHandler{},
			},
		},
		{
			expectErr: true,
			r:         newTestRequest(t, "http://localhost"),
			j:         NewRequestHandler(nil, []authn.Authenticator{authn.NewAuthenticatorNoOp()}, []authz.Authorizer{authz.NewAuthorizerAllow()}, []mutate.Mutator{mutate.NewMutatorNoop()}),
			rule: rule.Rule{
				Authenticators: []rule.RuleHandler{},
				Authorizer:     rule.RuleHandler{},
				Mutator:        rule.RuleHandler{},
			},
		},
		{
			expectErr: false,
			r:         newTestRequest(t, "http://localhost"),
			j:         NewRequestHandler(nil, []authn.Authenticator{authn.NewAuthenticatorNoOp()}, []authz.Authorizer{authz.NewAuthorizerAllow()}, []mutate.Mutator{mutate.NewMutatorNoop()}),
			rule: rule.Rule{
				Authenticators: []rule.RuleHandler{{Handler: authn.NewAuthenticatorNoOp().GetID()}},
				Authorizer:     rule.RuleHandler{Handler: authz.NewAuthorizerAllow().GetID()},
				Mutator:        rule.RuleHandler{Handler: mutate.NewMutatorNoop().GetID()},
			},
		},
		{
			expectErr: true,
			r:         newTestRequest(t, "http://localhost"),
			j:         NewRequestHandler(nil, []authn.Authenticator{authn.NewAuthenticatorAnonymous("anonymous")}, []authz.Authorizer{}, []mutate.Mutator{}),
			rule: rule.Rule{
				Authenticators: []rule.RuleHandler{{Handler: authn.NewAuthenticatorAnonymous("").GetID()}},
				Authorizer:     rule.RuleHandler{},
				Mutator:        rule.RuleHandler{},
			},
		},
		{
			expectErr: true,
			r:         newTestRequest(t, "http://localhost"),
			j:         NewRequestHandler(nil, []authn.Authenticator{authn.NewAuthenticatorAnonymous("anonymous")}, []authz.Authorizer{authz.NewAuthorizerAllow()}, []mutate.Mutator{}),
			rule: rule.Rule{
				Authenticators: []rule.RuleHandler{{Handler: authn.NewAuthenticatorAnonymous("").GetID()}},
				Authorizer:     rule.RuleHandler{Handler: "allow"},
				Mutator:        rule.RuleHandler{},
			},
		},
		{
			expectErr: false,
			r:         newTestRequest(t, "http://localhost"),
			j:         NewRequestHandler(nil, []authn.Authenticator{authn.NewAuthenticatorAnonymous("anonymous")}, []authz.Authorizer{authz.NewAuthorizerAllow()}, []mutate.Mutator{mutate.NewMutatorNoop()}),
			rule: rule.Rule{
				Authenticators: []rule.RuleHandler{{Handler: authn.NewAuthenticatorAnonymous("").GetID()}},
				Authorizer:     rule.RuleHandler{Handler: "allow"},
				Mutator:        rule.RuleHandler{Handler: "noop"},
			},
		},
	} {
		t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
			_, err := tc.j.HandleRequest(tc.r, &tc.rule)
			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
