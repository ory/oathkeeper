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
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/ory/hydra/sdk/go/hydra/swagger"
	"github.com/ory/oathkeeper/rule"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOAuth2Juror(t *testing.T) {
	type testCase struct {
		prepare       func(sdk *MockSDK)
		token         string
		expectErr     bool
		expectSession *Session
	}

	var runTest = func(tc testCase, j *JurorOAuth2Introspection, rl *rule.Rule) func(t *testing.T) {
		return func(t *testing.T) {
			c := gomock.NewController(t)
			sdk := NewMockSDK(c)
			j.H = sdk
			if tc.prepare != nil {
				tc.prepare(sdk)
			}

			r := &http.Request{Header: http.Header{"Authorization": {"Bearer " + tc.token}}}
			s, err := j.Try(r, rl, new(url.URL))

			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)

				if tc.expectSession != nil {
					assert.EqualValues(t, tc.expectSession, s)
				}
			}
		}
	}

	t.Run("suite=regular", func(t *testing.T) {
		j := &JurorOAuth2Introspection{L: logrus.New()}
		assert.Equal(t, "oauth2_introspection", j.GetID())

		rl := &rule.Rule{ID: "1234", Mode: "foo"}
		for k, tc := range []testCase{
			{
				token:     "",
				expectErr: true,
			},
			{
				token: "foo-token",
				prepare: func(sdk *MockSDK) {
					sdk.EXPECT().IntrospectOAuth2Token(gomock.Eq("foo-token"), gomock.Eq("")).Return(nil, nil, errors.New("network error"))
				},
				expectErr: true,
			},
			{
				token: "foo-token",
				prepare: func(sdk *MockSDK) {
					sdk.EXPECT().IntrospectOAuth2Token(gomock.Eq("foo-token"), gomock.Eq("")).Return(nil, &swagger.APIResponse{
						Response: &http.Response{StatusCode: http.StatusInternalServerError},
					}, nil)
				},
				expectErr: true,
			},
			{
				token: "foo-token",
				prepare: func(sdk *MockSDK) {
					sdk.EXPECT().IntrospectOAuth2Token(gomock.Eq("foo-token"), gomock.Eq("")).Return(&swagger.OAuth2TokenIntrospection{
						Active: false,
					}, &swagger.APIResponse{
						Response: &http.Response{StatusCode: http.StatusOK},
					}, nil)
				},
				expectErr: true,
			},
			{
				token: "foo-token",
				prepare: func(sdk *MockSDK) {
					sdk.EXPECT().IntrospectOAuth2Token(gomock.Eq("foo-token"), gomock.Eq("")).Return(&swagger.OAuth2TokenIntrospection{
						Active: true,
					}, &swagger.APIResponse{
						Response: &http.Response{StatusCode: http.StatusOK},
					}, nil)
				},
				expectErr: false,
			},
		} {
			t.Run(fmt.Sprintf("case=%d", k), runTest(tc, j, rl))
		}
	})

	t.Run("suite=anonymous", func(t *testing.T) {
		j := &JurorOAuth2Introspection{
			AllowAnonymous: true,
			L:              logrus.New(),
		}
		assert.Equal(t, "oauth2_introspection_anonymous", j.GetID())

		rl := &rule.Rule{ID: "1234", Mode: "foo"}
		for k, tc := range []testCase{
			{
				token:     "",
				expectErr: false,
			},
			{
				token: "foo-token",
				prepare: func(sdk *MockSDK) {
					sdk.EXPECT().IntrospectOAuth2Token(gomock.Eq("foo-token"), gomock.Eq("")).Return(nil, nil, errors.New("network error"))
				},
				expectErr: false,
			},
			{
				token: "foo-token",
				prepare: func(sdk *MockSDK) {
					sdk.EXPECT().IntrospectOAuth2Token(gomock.Eq("foo-token"), gomock.Eq("")).Return(nil, &swagger.APIResponse{
						Response: &http.Response{StatusCode: http.StatusInternalServerError},
					}, nil)
				},
				expectErr: false,
			},
			{
				token: "foo-token",
				prepare: func(sdk *MockSDK) {
					sdk.EXPECT().IntrospectOAuth2Token(gomock.Eq("foo-token"), gomock.Eq("")).Return(&swagger.OAuth2TokenIntrospection{
						Active: false,
					}, &swagger.APIResponse{
						Response: &http.Response{StatusCode: http.StatusOK},
					}, nil)
				},
				expectErr: false,
			},
			{
				token: "foo-token",
				prepare: func(sdk *MockSDK) {
					sdk.EXPECT().IntrospectOAuth2Token(gomock.Eq("foo-token"), gomock.Eq("")).Return(&swagger.OAuth2TokenIntrospection{
						Active: true,
					}, &swagger.APIResponse{
						Response: &http.Response{StatusCode: http.StatusOK},
					}, nil)
				},
				expectErr: false,
			},
		} {
			t.Run(fmt.Sprintf("case=%d", k), runTest(tc, j, rl))
		}
	})
}
