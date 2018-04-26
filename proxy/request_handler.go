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

	"github.com/ory/oathkeeper/rule"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/tools/go/gcimporter15/testdata"
)

type RequestHandler struct {
	Logger                 logrus.FieldLogger
	Matcher                rule.Matcher
	AuthorizationHandlers  map[string]Authorizer
	AuthenticationHandlers map[string]Authenticator
	SessionHandlers        map[string]SessionHandler
	Issuer                 string
}

func NewRequestHandler(l logrus.FieldLogger, m rule.Matcher, i string, jury Jury) *RequestHandler {
	if l == nil {
		l = logrus.New()
	}

	j := &RequestHandler{Matcher: m, Logger: l, Issuer: i, Authorizers: map[string]Juror{}}
	for _, juror := range jury {
		j.Authorizers[juror.GetID()] = juror
	}

	return j
}

func (d *RequestHandler) HandleRequest(r *http.Request) (error) {
	var u = *r.URL
	u.Host = r.Host
	u.Scheme = "http"
	if r.TLS != nil {
		u.Scheme = "https"
	}

	rl, err := d.Matcher.MatchRule(r.Method, &u)
	if err != nil {
		d.Logger.WithError(err).
			WithField("granted", false).
			WithField("user", "").
			WithField("access_url", u.String()).
			WithField("reason", "Unable to match a rule").
			WithField("reason_id", "no_rule_match").
			Warn("Access request denied")
		return err
	}

	var session *AuthenticationSession
	for _, a := range rl.Authenticators {
		anh, ok := d.AuthenticationHandlers[a.Handler]
		if !ok {
			d.Logger.
				WithField("granted", false).
				WithField("access_url", u.String()).
				WithField("authentication_handler", a.Handler).
				WithField("reason_id", "unknown_authentication_handler").
				Warn("Unknown authentication handler requested")
			return errors.New("Unknown authentication handler requested")
		}

		if session, err = anh.Authenticate(r); errors.Cause(ErrAuthenticatorNotResponsible).Error() == ErrAuthenticatorNotResponsible.Error() {
			// The authentication handler is not responsible for handling this request, skip to the next handler
		} else if errors.Cause(ErrAuthenticatorBypassed).Error() == ErrAuthenticatorBypassed.Error() {
			// The authentication handler says that no further authentication/authorization is required, and the request should
			// be forwarded to its final destination.
			return nil
		} else if err != nil {
			d.Logger.WithError(err).
				WithField("granted", false).
				WithField("access_url", u.String()).
				WithField("authentication_handler", a.Handler).
				WithField("reason_id", "authentication_handler_error").
				Warn("The authentication handler encountered an error")
			return err
		} else {
			// The first authenticator that matches must return the session
			break
		}
	}

	azh, ok := d.AuthorizationHandlers[rl.Authorizer.Handler]
	if !ok {
		d.Logger.
			WithField("granted", false).
			WithField("access_url", u.String()).
			WithField("authorization_handler", rl.Authorizer.Handler).
			WithField("reason_id", "unknown_authorization_handler").
			Warn("Unknown authentication handler requested")
		return errors.New("Unknown authorization handler requested")
	}

	if err := azh.Authorize(r, session); err != nil {
		return err

	}

	sh, ok := d.SessionHandlers[rl.Session.Handler]
	if !ok {
		d.Logger.
			WithField("granted", false).
			WithField("access_url", u.String()).
			WithField("session_handler", rl.Session.Handler).
			WithField("reason_id", "unknown_session_handler").
			Warn("Unknown session handler requested")
		return errors.New("Unknown session handler requested")
	}

	if err := sh.CreateSession(r,session); err != nil {
		return err
	}

	return nil
}
