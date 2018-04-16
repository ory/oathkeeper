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
	"github.com/sirupsen/logrus"
	"net/http"
	"github.com/ory/oathkeeper/rule"
	"net/url"
	"github.com/pkg/errors"
	"github.com/ory/oathkeeper/helper"
	"github.com/ory/keto/sdk/go/keto"
	"github.com/ory/keto/sdk/go/keto/swagger"
	"github.com/tomasen/realip"
)

type JurorWardenOAuth2 struct {
	L logrus.FieldLogger
	K keto.WardenSDK
}

func (j *JurorWardenOAuth2) GetID() string {
	// const PolicyMode = "policy"
	return "warden"
}

func (j *JurorWardenOAuth2) Try(r *http.Request, rl *rule.Rule, u *url.URL) (*Session, error) {
	token, _ := getBearerToken(r)
	if token == "" {
		j.L.
			WithFields(toLogFields(r, u, false, rl, "")).
			WithField("reason", "Rule requires a valid bearer token, but no bearer token was given").
			WithField("reason_id", "missing_credentials").
			Warn("Access request denied")
		return nil, errors.WithStack(helper.ErrMissingBearerToken)
	}

	introspection, response, err := j.K.IsOAuth2AccessTokenAuthorized(swagger.WardenOAuth2AccessRequest{
		Scopes:   rl.RequiredScopes,
		Action:   rl.MatchesURLCompiled.ReplaceAllString(u.String(), rl.RequiredAction),
		Resource: rl.MatchesURLCompiled.ReplaceAllString(u.String(), rl.RequiredResource),
		Token:    token,
		Context: map[string]interface{}{
			"remoteIpAddress": realip.RealIP(r),
		},
	})
	if err != nil {
		j.L.WithError(err).
			WithFields(toLogFields(r, u, false, rl, "")).
			WithField("reason", "Rule requires policy-based access control decision, which failed due to a network error").
			WithField("reason_id", "policy_decision_point_network_error").
			Warn("Access request denied")
		return nil, errors.WithStack(err)
	} else if response.StatusCode != http.StatusOK {
		j.L.WithError(err).
			WithFields(toLogFields(r, u, false, rl, "")).
			WithField("status_code", response.StatusCode).
			WithField("reason", "Rule requires policy-based access control decision, which failed due to a http error").
			WithField("reason_id", "policy_decision_point_http_error").
			Warn("Access request denied")
		return nil, errors.Errorf("Token introspection expects status code %d but got %d", http.StatusOK, response.StatusCode)
	} else if !introspection.Allowed {
		j.L.WithError(err).
			WithFields(toLogFields(r, u, false, rl, "")).
			WithField("reason", "Rule requires policy-based access control decision, which was denied").
			WithField("reason_id", "policy_decision_point_access_forbidden").
			Warn("Access request denied")
		return nil, errors.WithStack(helper.ErrForbidden)
	}

	j.L.
		WithFields(toLogFields(r, u, true, rl, introspection.Subject)).
		WithField("reason", "Rule requires policy-based access control decision, which was granted").
		WithField("reason_id", "policy_decision_point_access_granted").
		Infoln("Access request granted")
	return &Session{
		User:      introspection.Subject,
		ClientID:  introspection.ClientId,
		Anonymous: false,
		Extra:     introspection.AccessTokenExtra,
	}, nil
}
