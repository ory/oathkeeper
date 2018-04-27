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
	"regexp"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/ory/keto/sdk/go/keto/swagger"
	"github.com/ory/oathkeeper/rule"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKetoJuror(t *testing.T) {
	type testCase struct {
		prepare       func(sdk *MockWardenSDK)
		token         string
		expectErr     bool
		expectSession *Session
	}
	u, _ := url.Parse("http://localhost/")

	var runTest = func(tc testCase, j *JurorWardenOAuth2, rl *rule.Rule) func(t *testing.T) {
		return func(t *testing.T) {
			c := gomock.NewController(t)
			sdk := NewMockWardenSDK(c)
			j.K = sdk
			if tc.prepare != nil {
				tc.prepare(sdk)
			}

			r := &http.Request{Header: http.Header{"Authorization": {"Bearer " + tc.token}}}
			s, err := j.Try(r, rl, u)

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
		j := &JurorWardenOAuth2{L: logrus.New(), AnonymousName: "anonymous"}
		assert.Equal(t, "keto_warden_oauth2", j.GetID())

		matches, _ := regexp.Compile(".*")
		rl := &rule.Rule{ID: "1234", Mode: "foo", MatchesURLCompiled: matches}
		for k, tc := range []testCase{
			{
				token:     "",
				expectErr: true,
			},
			{
				token: "foo-token",
				prepare: func(sdk *MockWardenSDK) {
					sdk.EXPECT().
						IsOAuth2AccessTokenAuthorized(gomock.Any()).
						Return(nil, nil, errors.New("Network error"))
				},
				expectErr: true,
			},
			{
				token: "foo-token",
				prepare: func(sdk *MockWardenSDK) {
					sdk.EXPECT().
						IsOAuth2AccessTokenAuthorized(gomock.Any()).
						Return(&swagger.WardenOAuth2AccessTokenAuthorizationResponse{}, &swagger.APIResponse{Response: &http.Response{StatusCode: http.StatusInternalServerError}}, nil)
				},
				expectErr: true,
			},
			{
				token: "foo-token",
				prepare: func(sdk *MockWardenSDK) {
					sdk.EXPECT().
						IsOAuth2AccessTokenAuthorized(gomock.Any()).
						Return(&swagger.WardenOAuth2AccessTokenAuthorizationResponse{Allowed: false}, &swagger.APIResponse{Response: &http.Response{StatusCode: http.StatusOK}}, nil)
				},
				expectErr: true,
			},
			{
				token: "foo-token",
				prepare: func(sdk *MockWardenSDK) {
					sdk.EXPECT().
						IsOAuth2AccessTokenAuthorized(gomock.Any()).
						Return(&swagger.WardenOAuth2AccessTokenAuthorizationResponse{Allowed: true}, &swagger.APIResponse{Response: &http.Response{StatusCode: http.StatusOK}}, nil)
				},
			},
		} {
			t.Run(fmt.Sprintf("case=%d", k), runTest(tc, j, rl))
		}
	})

	t.Run("suite=regular", func(t *testing.T) {
		j := &JurorWardenOAuth2{L: logrus.New(), AnonymousName: "anonymous", AllowAnonymous: true}
		assert.Equal(t, "keto_warden_oauth2_anonymous", j.GetID())

		matches, _ := regexp.Compile(".*")
		rl := &rule.Rule{ID: "1234", Mode: "foo", MatchesURLCompiled: matches}
		for k, tc := range []testCase{
			{
				token: "foo-token",
				prepare: func(sdk *MockWardenSDK) {
					sdk.EXPECT().
						IsOAuth2AccessTokenAuthorized(gomock.Any()).
						Return(nil, nil, errors.New("Network error"))
				},
				expectErr: true,
			},
			{
				prepare: func(sdk *MockWardenSDK) {
					sdk.EXPECT().
						IsSubjectAuthorized(gomock.Any()).
						Return(nil, nil, errors.New("Network error"))
				},
				expectErr: true,
			},
			{
				token: "foo-token",
				prepare: func(sdk *MockWardenSDK) {
					sdk.EXPECT().
						IsOAuth2AccessTokenAuthorized(gomock.Any()).
						Return(&swagger.WardenOAuth2AccessTokenAuthorizationResponse{}, &swagger.APIResponse{Response: &http.Response{StatusCode: http.StatusInternalServerError}}, nil)
				},
				expectErr: true,
			},
			{
				prepare: func(sdk *MockWardenSDK) {
					sdk.EXPECT().
						IsSubjectAuthorized(gomock.Any()).
						Return(&swagger.WardenSubjectAuthorizationResponse{}, &swagger.APIResponse{Response: &http.Response{StatusCode: http.StatusInternalServerError}}, nil)
				},
				expectErr: true,
			},
			{
				token: "foo-token",
				prepare: func(sdk *MockWardenSDK) {
					sdk.EXPECT().
						IsOAuth2AccessTokenAuthorized(gomock.Any()).
						Return(&swagger.WardenOAuth2AccessTokenAuthorizationResponse{Allowed: false}, &swagger.APIResponse{Response: &http.Response{StatusCode: http.StatusOK}}, nil)
				},
				expectErr: true,
			},
			{
				prepare: func(sdk *MockWardenSDK) {
					sdk.EXPECT().
						IsSubjectAuthorized(gomock.Any()).
						Return(&swagger.WardenSubjectAuthorizationResponse{Allowed: false}, &swagger.APIResponse{Response: &http.Response{StatusCode: http.StatusOK}}, nil)
				},
				expectErr: true,
			},
			{
				token: "foo-token",
				prepare: func(sdk *MockWardenSDK) {
					sdk.EXPECT().
						IsOAuth2AccessTokenAuthorized(gomock.Any()).
						Return(&swagger.WardenOAuth2AccessTokenAuthorizationResponse{Allowed: true}, &swagger.APIResponse{Response: &http.Response{StatusCode: http.StatusOK}}, nil)
				},
			},
			{
				prepare: func(sdk *MockWardenSDK) {
					sdk.EXPECT().
						IsSubjectAuthorized(gomock.Any()).
						Return(&swagger.WardenSubjectAuthorizationResponse{Allowed: true}, &swagger.APIResponse{Response: &http.Response{StatusCode: http.StatusOK}}, nil)
				},
			},
		} {
			t.Run(fmt.Sprintf("case=%d", k), runTest(tc, j, rl))
		}
	})
}
