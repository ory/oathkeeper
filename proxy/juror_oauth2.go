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
	"strings"

	"github.com/ory/hydra/sdk/go/hydra"
	"github.com/ory/oathkeeper/helper"
	"github.com/ory/oathkeeper/rule"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type JurorOAuth2Introspection struct {
	L              logrus.FieldLogger
	H              hydra.OAuth2API
	AllowAnonymous bool
}

func (j *JurorOAuth2Introspection) GetID() string {
	if j.AllowAnonymous {
		return "oauth2_introspection_anonymous"
	}

	return "oauth2_introspection"
}

func (j *JurorOAuth2Introspection) Try(r *http.Request, rl *rule.Rule, u *url.URL) (*Session, error) {
	token, _ := getBearerToken(r)

	if token == "" && !j.AllowAnonymous {
		err := errors.WithStack(helper.ErrMissingBearerToken)
		j.L.WithError(err).
			WithFields(toLogFields(r, u, false, rl, "")).
			WithField("reason", "Rule requires a valid bearer token, but no bearer token was given").
			WithField("reason_id", "missing_credentials").
			Warn("Access request denied")

		return nil, err
	} else if token == "" {
		j.L.
			WithFields(toLogFields(r, u, true, rl, "")).
			WithField("reason", "Access is granted although no bearer token was given, because the rule allows anonymous access").
			WithField("reason_id", "anonymous_without_credentials").
			Infoln("Access request granted")

		return &Session{
			Subject:   "",
			Anonymous: true,
			ClientID:  "",
		}, nil
	}

	introspection, response, err := j.H.IntrospectOAuth2Token(token, strings.Join(rl.RequiredScopes, " "))
	if err != nil && !j.AllowAnonymous {
		j.L.WithError(err).
			WithFields(toLogFields(r, u, false, rl, "")).
			WithField("reason", "Rule requires a valid bearer token, but token introspection endpoint returned a network error").
			WithField("reason_id", "introspection_network_error").
			Warn("Access request denied")

		return nil, errors.WithStack(err)
	} else if err != nil {
		j.L.WithError(err).
			WithFields(toLogFields(r, u, true, rl, "")).
			WithField("reason", "Access is granted although an error occurred during token introspection because the rule allows anonymous access").
			WithField("reason_id", "anonymous_without_credentials_failed_introspection").
			Infoln("Access request granted")

		return &Session{
			Subject:   "",
			Anonymous: true,
			ClientID:  "",
		}, nil
	}

	if response.StatusCode != http.StatusOK && !j.AllowAnonymous {
		j.L.WithError(err).
			WithFields(toLogFields(r, u, false, rl, "")).
			WithField("status_code", response.StatusCode).
			WithField("reason", "Rule requires a valid bearer token, but token introspection endpoint returned a http error").
			WithField("reason_id", "introspection_http_error").
			Warn("Access request denied")

		return nil, errors.Errorf("Token introspection expects status code %d but got %d", http.StatusOK, response.StatusCode)
	} else if response.StatusCode != http.StatusOK {
		j.L.
			WithFields(toLogFields(r, u, true, rl, "")).
			WithField("status_code", response.StatusCode).
			WithField("reason", "Access is granted although token introspection endpoint returned a http error, because the rule allows anonymous access").
			WithField("reason_id", "anonymous_introspection_http_error").
			Infoln("Access request granted")

		return &Session{
			Subject:   "",
			Anonymous: true,
			ClientID:  "",
		}, nil
	}

	if !introspection.Active && !j.AllowAnonymous {
		j.L.WithError(err).
			WithFields(toLogFields(r, u, false, rl, "")).
			WithField("status_code", response.StatusCode).
			WithField("reason", "Rule requires a valid bearer token, but token introspection endpoint could not validate the token").
			WithField("reason_id", "introspection_invalid_credentials").
			Warn("Access request denied")

		return nil, errors.WithStack(helper.ErrUnauthorized)
	} else if !introspection.Active {
		j.L.
			WithFields(toLogFields(r, u, true, rl, "")).
			WithField("reason", "Access is granted although token introspection endpoint could not validate the bearer token, because the rule allows anonymous access").
			WithField("reason_id", "anonymous_introspection_invalid_credentials").
			Infoln("Access request granted")

		return &Session{
			Subject:   "",
			Anonymous: true,
			ClientID:  "",
		}, nil
	}

	j.L.
		WithFields(toLogFields(r, u, true, rl, introspection.Sub)).
		WithField("reason", "Rule allows anonymous access and valid access credentials have been provided").
		WithField("reason_id", "anonymous_with_valid_credentials").
		Infoln("Access request granted")

	return &Session{
		Subject:   introspection.Sub,
		ClientID:  introspection.ClientId,
		Anonymous: false,
		Extra:     introspection.Ext,
	}, nil
}
