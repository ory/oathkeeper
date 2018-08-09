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

	"github.com/ory/oathkeeper/helper"
	"github.com/ory/oathkeeper/rule"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type RequestHandler struct {
	Logger                 logrus.FieldLogger
	AuthorizationHandlers  map[string]Authorizer
	AuthenticationHandlers map[string]Authenticator
	CredentialIssuers      map[string]CredentialsIssuer
	Issuer                 string
}

func NewRequestHandler(
	l logrus.FieldLogger,
	authenticationHandlers []Authenticator,
	authorizationHandlers []Authorizer,
	credentialIssuers []CredentialsIssuer,
) *RequestHandler {
	if l == nil {
		l = logrus.New()
	}

	j := &RequestHandler{
		Logger:                 l,
		AuthorizationHandlers:  map[string]Authorizer{},
		AuthenticationHandlers: map[string]Authenticator{},
		CredentialIssuers:      map[string]CredentialsIssuer{},
	}

	for _, h := range authorizationHandlers {
		j.AuthorizationHandlers[h.GetID()] = h
	}

	for _, h := range authenticationHandlers {
		j.AuthenticationHandlers[h.GetID()] = h
	}

	for _, h := range credentialIssuers {
		j.CredentialIssuers[h.GetID()] = h
	}

	return j
}

func (d *RequestHandler) HandleRequest(r *http.Request, rl *rule.Rule) error {
	var err error
	var session *AuthenticationSession
	var found bool

	if len(rl.Authenticators) == 0 {
		err = errors.New("No authentication handler was set in the rule")
		d.Logger.WithError(err).
			WithField("granted", false).
			WithField("access_url", r.URL.String()).
			WithField("reason_id", "authentication_handler_missing").
			Warn("No authentication handler was set in the rule")
		return err
	}

	for _, a := range rl.Authenticators {
		anh, ok := d.AuthenticationHandlers[a.Handler]
		if !ok {
			d.Logger.
				WithField("granted", false).
				WithField("access_url", r.URL.String()).
				WithField("authentication_handler", a.Handler).
				WithField("reason_id", "unknown_authentication_handler").
				Warn("Unknown authentication handler requested")
			return errors.New("Unknown authentication handler requested")
		}

		session, err = anh.Authenticate(r, a.Config, rl)
		if err != nil {
			switch errors.Cause(err).Error() {
			case ErrAuthenticatorNotResponsible.Error():
				// The authentication handler is not responsible for handling this request, skip to the next handler
				break
			//case ErrAuthenticatorBypassed.Error():
			// The authentication handler says that no further authentication/authorization is required, and the request should
			// be forwarded to its final destination.
			//return nil
			default:
				d.Logger.WithError(err).
					WithField("granted", false).
					WithField("access_url", r.URL.String()).
					WithField("authentication_handler", a.Handler).
					WithField("reason_id", "authentication_handler_error").
					Warn("The authentication handler encountered an error")
				return err
			}
		} else {
			// The first authenticator that matches must return the session
			found = true
			break
		}
	}

	if !found {
		err := errors.WithStack(helper.ErrUnauthorized)
		d.Logger.WithError(err).
			WithField("granted", false).
			WithField("access_url", r.URL.String()).
			WithField("reason_id", "authentication_handler_no_match").
			Warn("No authentication handler was responsible for handling the authentication request")
		return err
	}

	azh, ok := d.AuthorizationHandlers[rl.Authorizer.Handler]
	if !ok {
		d.Logger.
			WithField("granted", false).
			WithField("access_url", r.URL.String()).
			WithField("authorization_handler", rl.Authorizer.Handler).
			WithField("reason_id", "unknown_authorization_handler").
			Warn("Unknown authentication handler requested")
		return errors.New("Unknown authorization handler requested")
	}

	if err := azh.Authorize(r, session, rl.Authorizer.Config, rl); err != nil {
		d.Logger.
			WithError(err).
			WithField("granted", false).
			WithField("access_url", r.URL.String()).
			WithField("authorization_handler", rl.Authorizer.Handler).
			WithField("reason_id", "authorization_handler_error").
			Warn("The authorization handler encountered an error")
		return err
	}

	sh, ok := d.CredentialIssuers[rl.CredentialsIssuer.Handler]
	if !ok {
		d.Logger.
			WithField("granted", false).
			WithField("access_url", r.URL.String()).
			WithField("session_handler", rl.CredentialsIssuer.Handler).
			WithField("reason_id", "unknown_credential_issuer").
			Warn("Unknown credential issuer requested")
		return errors.New("Unknown credential issuer requested")
	}

	if err := sh.Issue(r, session, rl.CredentialsIssuer.Config, rl); err != nil {
		return err
	}

	return nil
}
