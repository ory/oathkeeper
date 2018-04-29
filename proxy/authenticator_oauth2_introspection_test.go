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
	"errors"
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/ory/fosite"
	"github.com/ory/keto/authentication"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
)

func TestAuthenticatorOAuth2Introspection(t *testing.T) {
	a := NewAuthenticatorOAuth2Introspection("", "", "", "", []string{}, fosite.ExactScopeStrategy)
	assert.NotEmpty(t, a.GetID())

	for k, tc := range []struct {
		setup      func(*testing.T, *MockauthenticatorOAuth2IntrospectionHelper)
		r          *http.Request
		config     json.RawMessage
		expectErr  bool
		expectSess *AuthenticationSession
	}{
		{
			r:         &http.Request{Header: http.Header{}},
			expectErr: true,
		},
		{
			r:      &http.Request{Header: http.Header{"Authorization": {"bearer token"}}},
			config: []byte(`{ "required_scope": ["scope-a"] }`),
			setup: func(t *testing.T, m *MockauthenticatorOAuth2IntrospectionHelper) {
				m.EXPECT().Introspect(gomock.Eq("token"), gomock.Eq([]string{"scope-a"}), gomock.Eq(a.scopeStrategy)).Return(nil, errors.New("some error"))
			},
			expectErr: true,
		},
		{
			r:      &http.Request{Header: http.Header{"Authorization": {"bearer token"}}},
			config: []byte(`{ "required_scope": ["scope-a"], "trusted_issuers": ["foo", "bar"]}`),
			setup: func(t *testing.T, m *MockauthenticatorOAuth2IntrospectionHelper) {
				m.EXPECT().Introspect(gomock.Eq("token"), gomock.Eq([]string{"scope-a"}), gomock.Eq(a.scopeStrategy)).Return(&authentication.IntrospectionResponse{
					Subject:  "subject",
					Audience: []string{"audience"},
					Issuer:   "issuer",
					Username: "username",
					Extra:    map[string]interface{}{"extra": "foo"},
				}, nil)
			},
			expectErr: true,
		},
		{
			r:      &http.Request{Header: http.Header{"Authorization": {"bearer token"}}},
			config: []byte(`{ "required_scope": ["scope-a"], "target_audience": ["foo", "bar"]}`),
			setup: func(t *testing.T, m *MockauthenticatorOAuth2IntrospectionHelper) {
				m.EXPECT().Introspect(gomock.Eq("token"), gomock.Eq([]string{"scope-a"}), gomock.Eq(a.scopeStrategy)).Return(&authentication.IntrospectionResponse{
					Subject:  "subject",
					Audience: []string{"audience"},
					Issuer:   "issuer",
					Username: "username",
					Extra:    map[string]interface{}{"extra": "foo"},
				}, nil)
			},
			expectErr: true,
		},
		{
			r:      &http.Request{Header: http.Header{"Authorization": {"bearer token"}}},
			config: []byte(`{ "required_scope": ["scope-a"] }`),
			setup: func(t *testing.T, m *MockauthenticatorOAuth2IntrospectionHelper) {
				m.EXPECT().Introspect(gomock.Eq("token"), gomock.Eq([]string{"scope-a"}), gomock.Eq(a.scopeStrategy)).Return(&authentication.IntrospectionResponse{
					Subject:  "subject",
					Audience: []string{"audience"},
					Issuer:   "issuer",
					Username: "username",
					Extra:    map[string]interface{}{"extra": "foo"},
				}, nil)
			},
			expectErr: false,
		},
		{
			r:      &http.Request{Header: http.Header{"Authorization": {"bearer token"}}},
			config: []byte(`{ "required_scope": ["scope-a"], "trusted_issuers": ["issuer", "issuer-bar"], "target_audience": ["audience"] }`),
			setup: func(t *testing.T, m *MockauthenticatorOAuth2IntrospectionHelper) {
				m.EXPECT().Introspect(gomock.Eq("token"), gomock.Eq([]string{"scope-a"}), gomock.Eq(a.scopeStrategy)).Return(&authentication.IntrospectionResponse{
					Subject:  "subject",
					Audience: []string{"audience"},
					Issuer:   "issuer",
					Username: "username",
					Extra:    map[string]interface{}{"extra": "foo"},
				}, nil)
			},
			expectErr: false,
		},
	} {
		t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			m := NewMockauthenticatorOAuth2IntrospectionHelper(ctrl)
			if tc.setup != nil {
				tc.setup(t, m)
			}

			a.helper = m

			sess, err := a.Authenticate(tc.r, tc.config, nil)
			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			if tc.expectSess != nil {
				assert.Equal(t, tc.expectSess, sess)
			}
		})
	}
}
