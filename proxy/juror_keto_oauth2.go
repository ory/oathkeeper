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
	"net/http"
	"net/url"
	"time"

	"github.com/ory/keto/sdk/go/keto"
	"github.com/ory/keto/sdk/go/keto/swagger"
	"github.com/ory/oathkeeper/helper"
	"github.com/ory/oathkeeper/rule"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/tomasen/realip"
)

type JurorWardenOAuth2 struct {
	L              logrus.FieldLogger
	K              keto.WardenSDK
	AllowAnonymous bool
	AnonymousName  string
}

func NewJurorWardenOAuth2(
	l logrus.FieldLogger,
	k keto.WardenSDK,
	allowAnonymous bool,
	anonymousName string,
) *JurorWardenOAuth2 {
	return &JurorWardenOAuth2{
		L:              l,
		K:              k,
		AllowAnonymous: allowAnonymous,
		AnonymousName:  anonymousName,
	}
}

func (j *JurorWardenOAuth2) GetID() string {
	// const PolicyMode = "policy"
	if j.AllowAnonymous {
		return "keto_warden_oauth2_anonymous"
	}

	return "keto_warden_oauth2"
}

func contextFromRequest(r *http.Request) map[string]interface{} {
	return map[string]interface{}{
		"remoteIpAddress": realip.RealIP(r),
		"requestedAt":     time.Now().UTC(),
	}
}

func (j *JurorWardenOAuth2) Try(r *http.Request, rl *rule.Rule, u *url.URL) (*Session, error) {
	var oauthSession *swagger.WardenOAuth2AuthorizationResponse
	var defaultSession *swagger.WardenSubjectAuthorizationResponse
	var isAuthorized bool
	var response *swagger.APIResponse
	var err error
	var subject string

	token, _ := getBearerToken(r)
	if token == "" {
		if j.AllowAnonymous {
			defaultSession, response, err = j.K.IsSubjectAuthorized(swagger.WardenSubjectAuthorizationRequest{
				Action:   rl.MatchesURLCompiled.ReplaceAllString(u.String(), rl.RequiredAction),
				Resource: rl.MatchesURLCompiled.ReplaceAllString(u.String(), rl.RequiredResource),
				Context:  contextFromRequest(r),
				Subject:  j.AnonymousName,
			})
			if defaultSession != nil {
				isAuthorized = defaultSession.Allowed
			}
			subject = j.AnonymousName
		} else {
			j.L.
				WithFields(toLogFields(r, u, false, rl, "")).
				WithField("reason", "Rule requires a valid bearer token, but no bearer token was given").
				WithField("reason_id", "missing_credentials").
				Warn("Access request denied")
			return nil, errors.WithStack(helper.ErrMissingBearerToken)
		}
	} else {
		oauthSession, response, err = j.K.IsOAuth2AccessTokenAuthorized(swagger.WardenOAuth2AuthorizationRequest{
			Scopes:   rl.RequiredScopes,
			Action:   rl.MatchesURLCompiled.ReplaceAllString(u.String(), rl.RequiredAction),
			Resource: rl.MatchesURLCompiled.ReplaceAllString(u.String(), rl.RequiredResource),
			Token:    token,
			Context:  contextFromRequest(r),
		})
		if oauthSession != nil {
			isAuthorized = oauthSession.Allowed
			subject = oauthSession.Subject
		}
	}

	if err != nil {
		j.L.WithError(err).
			WithFields(toLogFields(r, u, false, rl, subject)).
			WithField("reason", "Rule requires policy-based access control decision, which failed due to a network error").
			WithField("reason_id", "policy_decision_point_network_error").
			Warn("Access request denied")
		return nil, errors.WithStack(err)
	} else if response.StatusCode != http.StatusOK {
		j.L.WithError(err).
			WithFields(toLogFields(r, u, false, rl, subject)).
			WithField("status_code", response.StatusCode).
			WithField("reason", "Rule requires policy-based access control decision, which failed due to a http error").
			WithField("reason_id", "policy_decision_point_http_error").
			Warn("Access request denied")
		return nil, errors.Errorf("Token introspection expects status code %d but got %d", http.StatusOK, response.StatusCode)
	} else if !isAuthorized {
		j.L.WithError(err).
			WithFields(toLogFields(r, u, false, rl, subject)).
			WithField("reason", "Rule requires policy-based access control decision, which was denied").
			WithField("reason_id", "policy_decision_point_access_forbidden").
			Warn("Access request denied")
		return nil, errors.WithStack(helper.ErrForbidden)
	}

	j.L.
		WithFields(toLogFields(r, u, true, rl, subject)).
		WithField("reason", "Rule requires policy-based access control decision, which was granted").
		WithField("reason_id", "policy_decision_point_access_granted").
		Infoln("Access request granted")

	if defaultSession != nil {
		return &Session{
			Subject:   subject,
			Anonymous: true,
		}, nil
	}

	return &Session{
		Subject:   subject,
		ClientID:  oauthSession.ClientId,
		Anonymous: false,
		Extra:     oauthSession.AccessTokenExtra,
	}, nil
}
