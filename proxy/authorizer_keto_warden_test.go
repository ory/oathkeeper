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
	"encoding/json"
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/ory/keto/sdk/go/keto/swagger"
	"github.com/ory/oathkeeper/rule"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/url"
	"testing"
)

func mustParseURL(t *testing.T, u string) *url.URL {
	p, err := url.Parse(u)
	require.NoError(t, err)
	return p
}

func TestAuthorizerKetoWarden(t *testing.T) {
	assert.NotEmpty(t, NewAuthorizerKetoWarden(nil).GetID())

	for k, tc := range []struct {
		setup     func(*testing.T, *MockWardenSDK)
		r         *http.Request
		session   *AuthenticationSession
		config    json.RawMessage
		rule      *rule.Rule
		expectErr bool
	}{
		{
			expectErr: true,
		},
		{
			config: []byte(`{ "required_action": "action", "required_resource": "resource" }`),
			rule: &rule.Rule{
				Match: rule.RuleMatch{
					Methods: []string{"POST"},
					URL:     "https://localhost/",
				},
			},
			r: &http.Request{URL: &url.URL{}},
			setup: func(t *testing.T, m *MockWardenSDK) {
				m.EXPECT().IsSubjectAuthorized(gomock.Any()).Return(nil, nil, errors.New("foo"))
			},
			session:   new(AuthenticationSession),
			expectErr: true,
		},
		{
			config: []byte(`{ "required_action": "action", "required_resource": "resource" }`),
			rule: &rule.Rule{
				Match: rule.RuleMatch{
					Methods: []string{"POST"},
					URL:     "https://localhost/",
				},
			},
			r: &http.Request{URL: &url.URL{}},
			setup: func(t *testing.T, m *MockWardenSDK) {
				m.EXPECT().IsSubjectAuthorized(gomock.Any()).Return(nil, &swagger.APIResponse{Response: &http.Response{StatusCode: http.StatusInternalServerError}}, nil)
			},
			session:   new(AuthenticationSession),
			expectErr: true,
		},
		{
			config: []byte(`{ "required_action": "action", "required_resource": "resource" }`),
			rule: &rule.Rule{
				Match: rule.RuleMatch{
					Methods: []string{"POST"},
					URL:     "https://localhost/",
				},
			},
			r: &http.Request{URL: &url.URL{}},
			setup: func(t *testing.T, m *MockWardenSDK) {
				m.EXPECT().IsSubjectAuthorized(gomock.Any()).Return(
					&swagger.WardenSubjectAuthorizationResponse{Allowed: false},
					&swagger.APIResponse{Response: &http.Response{StatusCode: http.StatusOK}},
					nil,
				)
			},
			session:   new(AuthenticationSession),
			expectErr: true,
		},
		{
			config: []byte(`{ "required_action": "action:$1:$2", "required_resource": "resource:$1:$2" }`),
			rule: &rule.Rule{
				Match: rule.RuleMatch{
					Methods: []string{"POST"},
					URL:     "https://localhost/api/users/<[0-9]+>/<[a-z]+>",
				},
			},
			r: &http.Request{URL: mustParseURL(t, "https://localhost/api/users/1234/abcde")},
			setup: func(t *testing.T, m *MockWardenSDK) {
				m.EXPECT().IsSubjectAuthorized(gomock.Eq(swagger.WardenSubjectAuthorizationRequest{
					Action:   "action:1234:abcde",
					Resource: "resource:1234:abcde",
					Context:  map[string]interface{}{},
					Subject:  "peter",
				})).Return(
					&swagger.WardenSubjectAuthorizationResponse{Allowed: true},
					&swagger.APIResponse{Response: &http.Response{StatusCode: http.StatusOK}},
					nil,
				)
			},
			session:   &AuthenticationSession{Subject: "peter"},
			expectErr: false,
		},
	} {
		t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			sdk := NewMockWardenSDK(c)
			if tc.setup != nil {
				tc.setup(t, sdk)
			}
			a := NewAuthorizerKetoWarden(sdk)
			a.contextCreator = func(r *http.Request) map[string]interface{} {
				return map[string]interface{}{}
			}

			err := a.Authorize(tc.r, tc.session, tc.config, tc.rule)
			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
