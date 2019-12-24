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

package proxy_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/ory/viper"

	"github.com/ory/oathkeeper/driver/configuration"
	"github.com/ory/oathkeeper/internal"
	"github.com/ory/x/urlx"

	"github.com/stretchr/testify/require"

	"github.com/ory/oathkeeper/rule"
)

func newTestRequest(u string) *http.Request {
	return &http.Request{URL: urlx.ParseOrPanic(u)}
}

func TestRequestHandler(t *testing.T) {
	for k, tc := range []struct {
		d         string
		setup     func()
		rule      rule.Rule
		r         *http.Request
		expectErr bool
	}{
		{
			d:         "should fail because the rule is missing authn, authz, and mutator",
			expectErr: true,
			r:         newTestRequest("http://localhost"),
			rule: rule.Rule{
				Authenticators: []rule.Handler{},
				Authorizer:     rule.Handler{},
				Mutators:       []rule.Handler{},
			},
		},
		{
			d: "should fail because the rule is missing authn, authz, and mutator even when some pipelines are enabled",
			setup: func() {
				viper.Set(configuration.ViperKeyAuthenticatorNoopIsEnabled, true)
				viper.Set(configuration.ViperKeyAuthorizerAllowIsEnabled, true)
				viper.Set(configuration.ViperKeyMutatorNoopIsEnabled, true)
			},
			expectErr: true,
			r:         newTestRequest("http://localhost"),
			rule: rule.Rule{
				Authenticators: []rule.Handler{},
				Authorizer:     rule.Handler{},
				Mutators:       []rule.Handler{},
			},
		},
		{
			d: "should pass",
			setup: func() {
				viper.Set(configuration.ViperKeyAuthenticatorNoopIsEnabled, true)
				viper.Set(configuration.ViperKeyAuthorizerAllowIsEnabled, true)
				viper.Set(configuration.ViperKeyMutatorNoopIsEnabled, true)
			},
			expectErr: false,
			r:         newTestRequest("http://localhost"),
			rule: rule.Rule{
				Authenticators: []rule.Handler{{Handler: "noop"}},
				Authorizer:     rule.Handler{Handler: "allow"},
				Mutators:       []rule.Handler{{Handler: "noop"}},
			},
		},
		{
			d: "should fail when authn is set but not authz nor mutator",
			setup: func() {
				viper.Set(configuration.ViperKeyAuthenticatorAnonymousIsEnabled, true)
				viper.Set(configuration.ViperKeyAuthorizerAllowIsEnabled, true)
				viper.Set(configuration.ViperKeyMutatorNoopIsEnabled, true)
			},
			expectErr: true,
			r:         newTestRequest("http://localhost"),
			rule: rule.Rule{
				Authenticators: []rule.Handler{{Handler: "anonymous"}},
				Authorizer:     rule.Handler{},
				Mutators:       []rule.Handler{},
			},
		},
		{
			d: "should fail when authn, authz is set but not mutator",
			setup: func() {
				viper.Set(configuration.ViperKeyAuthenticatorAnonymousIsEnabled, true)
				viper.Set(configuration.ViperKeyAuthorizerAllowIsEnabled, true)
				viper.Set(configuration.ViperKeyMutatorNoopIsEnabled, true)
			},
			expectErr: true,
			r:         newTestRequest("http://localhost"),
			rule: rule.Rule{
				Authenticators: []rule.Handler{{Handler: "anonymous"}},
				Authorizer:     rule.Handler{Handler: "allow"},
				Mutators:       []rule.Handler{},
			},
		},
		{
			d: "should fail when authn is invalid because not enabled",
			setup: func() {
				viper.Set(configuration.ViperKeyAuthenticatorAnonymousIsEnabled, false)
				viper.Set(configuration.ViperKeyAuthorizerAllowIsEnabled, true)
				viper.Set(configuration.ViperKeyMutatorNoopIsEnabled, true)
			},
			expectErr: true,
			r:         newTestRequest("http://localhost"),
			rule: rule.Rule{
				Authenticators: []rule.Handler{{Handler: "anonymous"}},
				Authorizer:     rule.Handler{Handler: "allow"},
				Mutators:       []rule.Handler{{Handler: "noop"}},
			},
		},
		{
			d: "should fail when authz is invalid because not enabled",
			setup: func() {
				viper.Set(configuration.ViperKeyAuthenticatorAnonymousIsEnabled, true)
				viper.Set(configuration.ViperKeyAuthorizerAllowIsEnabled, false)
				viper.Set(configuration.ViperKeyMutatorNoopIsEnabled, true)
			},
			expectErr: true,
			r:         newTestRequest("http://localhost"),
			rule: rule.Rule{
				Authenticators: []rule.Handler{{Handler: "anonymous"}},
				Authorizer:     rule.Handler{Handler: "allow"},
				Mutators:       []rule.Handler{{Handler: "noop"}},
			},
		},
		{
			d: "should fail when mutator is invalid because not enabled",
			setup: func() {
				viper.Set(configuration.ViperKeyAuthenticatorAnonymousIsEnabled, true)
				viper.Set(configuration.ViperKeyAuthorizerAllowIsEnabled, true)
				viper.Set(configuration.ViperKeyMutatorNoopIsEnabled, false)
			},
			expectErr: true,
			r:         newTestRequest("http://localhost"),
			rule: rule.Rule{
				Authenticators: []rule.Handler{{Handler: "anonymous"}},
				Authorizer:     rule.Handler{Handler: "allow"},
				Mutators:       []rule.Handler{{Handler: "noop"}},
			},
		},
		{
			d: "should fail when authn does not exist",
			setup: func() {
				viper.Set(configuration.ViperKeyAuthenticatorAnonymousIsEnabled, true)
				viper.Set(configuration.ViperKeyAuthorizerAllowIsEnabled, true)
				viper.Set(configuration.ViperKeyMutatorNoopIsEnabled, true)
			},
			expectErr: true,
			r:         newTestRequest("http://localhost"),
			rule: rule.Rule{
				Authenticators: []rule.Handler{{Handler: "invalid-id"}},
				Authorizer:     rule.Handler{Handler: "allow"},
				Mutators:       []rule.Handler{{Handler: "noop"}},
			},
		},
		{
			d: "should fail when authz does not exist",
			setup: func() {
				viper.Set(configuration.ViperKeyAuthenticatorAnonymousIsEnabled, true)
				viper.Set(configuration.ViperKeyAuthorizerAllowIsEnabled, true)
				viper.Set(configuration.ViperKeyMutatorNoopIsEnabled, true)
			},
			expectErr: true,
			r:         newTestRequest("http://localhost"),
			rule: rule.Rule{
				Authenticators: []rule.Handler{{Handler: "anonymous"}},
				Authorizer:     rule.Handler{Handler: "invalid-id"},
				Mutators:       []rule.Handler{{Handler: "noop"}},
			},
		},
		{
			d: "should fail when mutator does not exist",
			setup: func() {
				viper.Set(configuration.ViperKeyAuthenticatorAnonymousIsEnabled, true)
				viper.Set(configuration.ViperKeyAuthorizerAllowIsEnabled, true)
				viper.Set(configuration.ViperKeyMutatorNoopIsEnabled, true)
			},
			expectErr: true,
			r:         newTestRequest("http://localhost"),
			rule: rule.Rule{
				Authenticators: []rule.Handler{{Handler: "anonymous"}},
				Authorizer:     rule.Handler{Handler: "allow"},
				Mutators:       []rule.Handler{{Handler: "invalid-id"}},
			},
		},
	} {
		t.Run(fmt.Sprintf("case=%d/description=%s", k, tc.d), func(t *testing.T) {

			conf := internal.NewConfigurationWithDefaults()
			reg := internal.NewRegistry(conf)

			if tc.setup != nil {
				tc.setup()
			}

			_, err := reg.ProxyRequestHandler().HandleRequest(tc.r, &tc.rule)
			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
