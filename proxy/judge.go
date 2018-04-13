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

package decision

import (
	"net/http"

	"strings"

	"github.com/ory/keto/sdk/go/keto"
	"github.com/ory/keto/sdk/go/keto/swagger"
	"github.com/ory/oathkeeper/helper"
	"github.com/ory/oathkeeper/rule"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/tomasen/realip"
)

type Judge struct {
	Logger  logrus.FieldLogger
	Matcher rule.Matcher
	Issuer  string
}

func NewWardenEvaluator(l logrus.FieldLogger, m rule.Matcher, i string) *Judge {
	if l == nil {
		l = logrus.New()
	}

	return &Judge{
		Matcher: m,
		Logger:  l,
		Issuer:  i,
	}
}

var reasons = map[string]string{
	"no_rule_match":                               "Unable to match a rule",
	"passthrough":                                 "Access is granted because rule is set to passthrough",
	"anonymous_without_credentials":               "Access is granted although no bearer token was given, because the rule allows anonymous access",
	"anonymous_introspection_network_error":       "Access is granted although token introspection endpoint returned a network error, because the rule allows anonymous access",
	"anonymous_introspection_http_error":          "Access is granted although token introspection endpoint returned a http error, because the rule allows anonymous access",
	"anonymous_introspection_invalid_credentials": "Access is granted although token introspection endpoint could not validate the bearer token, because the rule allows anonymous access",
	"anonymous_with_valid_credentials":            "Rule allows anonymous access and valid access credentials have been provided",
	"missing_credentials":                         "Rule requires a valid bearer token, but no bearer token was given",
	"introspection_network_error":                 "Rule requires a valid bearer token, but token introspection endpoint returned a network error",
	"introspection_http_error":                    "Rule requires a valid bearer token, but token introspection endpoint returned a http error",
	"introspection_invalid_credentials":           "Rule requires a valid bearer token, but token introspection endpoint could not validate the token",
	"introspection_valid":                         "Rule requires a valid bearer token, which was confirmed by the token introspection endpoint",
	"policy_decision_point_network_error":         "Rule requires policy-based access control decision, which failed due to a network error",
	"policy_decision_point_http_error":            "Rule requires policy-based access control decision, which failed due to a http error",
	"policy_decision_point_access_forbidden":      "Rule requires policy-based access control decision, which was denied",
	"policy_decision_point_access_granted":        "Rule requires policy-based access control decision, which was granted",
	"unknown_mode":                                "Rule defines a unknown mod ",
}

func (d *Judge) EvaluateAccessRequest(r *http.Request) (*Session, error) {
	var u = *r.URL
	u.Host = r.Host
	u.Scheme = "http"
	if r.TLS != nil {
		u.Scheme = "https"
	}

	token := helper.BearerTokenFromRequest(r)
	var tokenID = token
	if len(token) >= 5 {
		tokenID = token[:5]
	}

	rl, err := d.Matcher.MatchRule(r.Method, &u)
	if err != nil {
		d.Logger.WithError(err).
			WithField("granted", false).
			WithField("user", "").
			WithField("access_url", u.String()).
			WithField("token", tokenID).
			WithField("reason", reasons["no_rule_match"]).
			WithField("reason_id", "no_rule_match").
			Warn("Access request denied")
		return nil, err
	}

	switch rl.Mode {
	case rule.BypassMode:
		d.Logger.
			WithField("granted", true).
			WithField("user", "").
			WithField("access_url", u.String()).
			WithField("token", tokenID).
			WithField("rule", rl.ID).
			WithField("mode", rl.Mode).
			WithField("reason", reasons["passthrough"]).
			WithField("reason_id", "passthrough").
			Infoln("Access request granted")
		return &Session{Issuer: d.Issuer, User: "", Anonymous: true, ClientID: "", Disabled: true}, nil
	case rule.AnonymousMode:
		if token == "" {
			d.Logger.
				WithField("granted", true).
				WithField("user", "").
				WithField("access_url", u.String()).
				WithField("token", tokenID).
				WithField("rule", rl.ID).
				WithField("reason", reasons["anonymous_without_credentials"]).
				WithField("reason_id", "anonymous_without_credentials").
				Infoln("Access request granted")
			return &Session{Issuer: d.Issuer, User: "", Anonymous: true, ClientID: ""}, nil
		}

		introspection, response, err := d.Hydra.IntrospectOAuth2Token(token, "")
		if err != nil {
			d.Logger.WithError(err).
				WithField("granted", true).
				WithField("user", "").
				WithField("access_url", u.String()).
				WithField("token", tokenID).
				WithField("mode", rl.Mode).
				WithField("reason", reasons["anonymous_without_credentials_failed_introspection"]).
				WithField("reason_id", "anonymous_without_credentials_failed_introspection").
				Infoln("Access request granted")
			return &Session{Issuer: d.Issuer, User: "", Anonymous: true, ClientID: ""}, nil
		} else if response.StatusCode != http.StatusOK {
			d.Logger.
				WithField("granted", true).
				WithField("user", "").
				WithField("status_code", response.StatusCode).
				WithField("token", tokenID).
				WithField("access_url", u.String()).
				WithField("mode", rl.Mode).
				WithField("reason", reasons["anonymous_introspection_http_error"]).
				WithField("reason_id", "anonymous_introspection_http_error").
				Infoln("Access request granted")
			return &Session{Issuer: d.Issuer, User: "", Anonymous: true, ClientID: ""}, nil
		} else if !introspection.Active {
			d.Logger.
				WithField("granted", true).
				WithField("user", "").
				WithField("access_url", u.String()).
				WithField("token", tokenID).
				WithField("rule", rl.ID).
				WithField("mode", rl.Mode).
				WithField("reason", reasons["anonymous_introspection_invalid_credentials"]).
				WithField("reason_id", "anonymous_introspection_invalid_credentials").
				Infoln("Access request granted")
			return &Session{Issuer: d.Issuer, User: "", Anonymous: true, ClientID: ""}, nil
		}

		d.Logger.
			WithField("granted", true).
			WithField("user", introspection.Sub).
			WithField("access_url", u.String()).
			WithField("token", tokenID).
			WithField("rule", rl.ID).
			WithField("mode", rl.Mode).
			WithField("reason", reasons["anonymous_with_valid_credentials"]).
			WithField("reason_id", "anonymous_with_valid_credentials").
			Infoln("Access request granted")
		return &Session{
			Issuer:    d.Issuer,
			User:      introspection.Sub,
			ClientID:  introspection.ClientId,
			Anonymous: false,
			Extra:     introspection.Ext,
		}, nil
	case rule.AuthenticatedMode:
		if token == "" {
			d.Logger.WithError(err).
				WithField("granted", false).
				WithField("user", "").
				WithField("access_url", u.String()).
				WithField("token", tokenID).
				WithField("mode", rl.Mode).
				WithField("reason", reasons["missing_credentials"]).
				WithField("reason_id", "missing_credentials").
				Warn("Access request denied")
			return nil, errors.WithStack(helper.ErrMissingBearerToken)
		}

		introspection, response, err := d.Hydra.IntrospectOAuth2Token(token, strings.Join(rl.RequiredScopes, " "))
		if err != nil {
			d.Logger.WithError(err).
				WithField("granted", false).
				WithField("user", "").
				WithField("access_url", u.String()).
				WithField("token", tokenID).
				WithField("mode", rl.Mode).
				WithField("reason", reasons["introspection_network_error"]).
				WithField("reason_id", "introspection_network_error").
				Warn("Access request denied")
			return nil, errors.WithStack(err)
		} else if response.StatusCode != http.StatusOK {
			d.Logger.WithError(err).
				WithField("granted", false).
				WithField("user", "").
				WithField("access_url", u.String()).
				WithField("status_code", response.StatusCode).
				WithField("token", tokenID).
				WithField("mode", rl.Mode).
				WithField("reason", reasons["introspection_http_error"]).
				WithField("reason_id", "introspection_http_error").
				Warn("Access request denied")
			return nil, errors.Errorf("Token introspection expects status code %d but got %d", http.StatusOK, response.StatusCode)
		} else if !introspection.Active {
			d.Logger.WithError(err).
				WithField("granted", false).
				WithField("user", "").
				WithField("access_url", u.String()).
				WithField("status_code", response.StatusCode).
				WithField("token", tokenID).
				WithField("mode", rl.Mode).
				WithField("reason", reasons["introspection_invalid_credentials"]).
				WithField("reason_id", "introspection_invalid_credentials").
				Warn("Access request denied")
			return nil, errors.WithStack(helper.ErrUnauthorized)
		}

		d.Logger.
			WithField("granted", true).
			WithField("user", introspection.Sub).
			WithField("access_url", u.String()).
			WithField("token", tokenID).
			WithField("rule", rl.ID).
			WithField("mode", rl.Mode).
			WithField("reason", reasons["introspection_valid"]).
			WithField("reason_id", "introspection_valid").
			Infoln("Access request granted")
		return &Session{
			Issuer:    d.Issuer,
			User:      introspection.Sub,
			ClientID:  introspection.ClientId,
			Anonymous: false,
			Extra:     introspection.Ext,
		}, nil
	case rule.PolicyMode:
		if token == "" {
			d.Logger.WithError(err).
				WithField("granted", false).
				WithField("user", "").
				WithField("access_url", u.String()).
				WithField("token", tokenID).
				WithField("mode", rl.Mode).
				WithField("reason", reasons["missing_credentials"]).
				WithField("reason_id", "missing_credentials").
				Warn("Access request denied")
			return nil, errors.WithStack(helper.ErrMissingBearerToken)
		}

		introspection, response, err := d.Hydra.DoesWardenAllowTokenAccessRequest(d.prepareAccessRequests(r, u.String(), token, rl))
		if err != nil {
			d.Logger.WithError(err).
				WithField("granted", false).
				WithField("user", "").
				WithField("access_url", u.String()).
				WithField("token", tokenID).
				WithField("mode", rl.Mode).
				WithField("reason", reasons["policy_decision_point_network_error"]).
				WithField("reason_id", "policy_decision_point_network_error").
				Warn("Access request denied")
			return nil, errors.WithStack(err)
		} else if response.StatusCode != http.StatusOK {
			d.Logger.WithError(err).
				WithField("granted", false).
				WithField("user", "").
				WithField("access_url", u.String()).
				WithField("status_code", response.StatusCode).
				WithField("token", tokenID).
				WithField("mode", rl.Mode).
				WithField("reason", reasons["policy_decision_point_http_error"]).
				WithField("reason_id", "policy_decision_point_http_error").
				Warn("Access request denied")
			return nil, errors.Errorf("Token introspection expects status code %d but got %d", http.StatusOK, response.StatusCode)
		} else if !introspection.Allowed {
			d.Logger.WithError(err).
				WithField("granted", false).
				WithField("user", "").
				WithField("access_url", u.String()).
				WithField("status_code", response.StatusCode).
				WithField("token", tokenID).
				WithField("mode", rl.Mode).
				WithField("reason", reasons["policy_decision_point_access_forbidden"]).
				WithField("reason_id", "policy_decision_point_access_forbidden").
				Warn("Access request denied")
			return nil, errors.WithStack(helper.ErrForbidden)
		}

		d.Logger.
			WithField("granted", true).
			WithField("user", introspection.Subject).
			WithField("access_url", u.String()).
			WithField("token", tokenID).
			WithField("rule", rl.ID).
			WithField("mode", rl.Mode).
			WithField("reason", reasons["policy_decision_point_access_granted"]).
			WithField("reason_id", "policy_decision_point_access_granted").
			Infoln("Access request granted")
		return &Session{
			Issuer:    d.Issuer,
			User:      introspection.Subject,
			ClientID:  introspection.ClientId,
			Anonymous: false,
			Extra:     introspection.AccessTokenExtra,
		}, nil
	default:
		d.Logger.WithError(err).
			WithField("granted", false).
			WithField("user", "").
			WithField("access_url", u.String()).
			WithField("token", tokenID).
			WithField("mode", rl.Mode).
			WithField("reason", reasons["unknown_mode"]).
			WithField("reason_id", "unknown_mode").
			Warn("Access request denied")
		return nil, errors.Errorf("Unknown rule mode \"%s\"", rl.Mode)
	}
}

func (d *Judge) prepareAccessRequests(r *http.Request, u string, token string, rl *rule.Rule) swagger.WardenTokenAccessRequest {
	return swagger.WardenTokenAccessRequest{
		Scopes:   rl.RequiredScopes,
		Action:   rl.MatchesURLCompiled.ReplaceAllString(u, rl.RequiredAction),
		Resource: rl.MatchesURLCompiled.ReplaceAllString(u, rl.RequiredResource),
		Token:    token,
		Context: map[string]interface{}{
			"remoteIpAddress": realip.RealIP(r),
		},
	}
}
