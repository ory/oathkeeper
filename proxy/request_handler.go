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

	"github.com/ory/herodot"

	"github.com/ory/oathkeeper/x"

	"github.com/ory/oathkeeper/pipeline/authn"
	"github.com/ory/oathkeeper/pipeline/authz"
	pe "github.com/ory/oathkeeper/pipeline/errors"
	"github.com/ory/oathkeeper/pipeline/mutate"

	"github.com/pkg/errors"

	"github.com/ory/oathkeeper/helper"
	"github.com/ory/oathkeeper/rule"
)

type requestHandlerRegistry interface {
	x.RegistryWriter
	x.RegistryLogger

	authn.Registry
	authz.Registry
	mutate.Registry
	pe.Registry
}

type RequestHandler struct {
	r requestHandlerRegistry
}

func NewRequestHandler(r requestHandlerRegistry) *RequestHandler {
	return &RequestHandler{r: r}
}

func (d *RequestHandler) HandleError(w http.ResponseWriter, r *http.Request, rl *rule.Rule, err error) {
	for name := range rl.Errors {
		handler, err := d.r.PipelineErrorHandler(name)
		if err != nil {
			d.r.Writer().WriteError(w, r, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("Unable to find error handler named: %s. This is a configuration issue and should be reported to the administrator.", name)))
			return
		}

		re := rl.Errors[name]
		if err := handler.Validate(re.Config); err != nil {
			d.r.Writer().WriteError(w, r, err)
			return
		}

		if e := handler.Handle(w, r, re.Config, rl, err); e == nil {
			return
		} else if errors.Cause(e) == pe.ErrHandlerNotResponsible {
			// loop
		} else {
			d.r.Writer().WriteError(w, r, errors.WithStack(herodot.ErrInternalServerError.WithReasonf(`Unable to execute error handler. This is either a bug or a configuration issue and should be reported to the administrator. Returned error: "%s". Original error: "%s"`, name, err, e)))
			return
		}
	}

	for name, handler := range d.r.AvailablePipelineErrorHandlers() {
		if e := handler.Handle(w, r, nil, rl, err); e == nil {
			return
		} else if errors.Cause(e) == pe.ErrHandlerNotResponsible {
			// loop
		} else {
			d.r.Writer().WriteError(w, r, errors.WithStack(herodot.ErrInternalServerError.WithReasonf(`Unable to execute error handler. This is either a bug or a configuration issue and should be reported to the administrator. Returned error: "%s". Original error: "%s"`, name, err, e)))
			return
		}
	}

	d.r.Writer().WriteError(w, r, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("Unable to handle HTTP request because no matching error handling strategy was found. This is a bug and should be reported to: http://github.com/ory/oathkeeper")))
}

func (d *RequestHandler) HandleRequest(r *http.Request, rl *rule.Rule) (session *authn.AuthenticationSession, err error) {
	var found bool

	fields := map[string]interface{}{
		"http_method":     r.Method,
		"http_url":        r.URL.String(),
		"http_host":       r.Host,
		"http_user_agent": r.UserAgent(),
		"rule_id":         rl.ID,
	}

	if len(rl.Authenticators) == 0 {
		err = errors.New("No authentication handler was set in the rule")
		d.r.Logger().WithError(err).
			WithFields(fields).
			WithField("granted", false).
			WithField("reason_id", "authentication_handler_missing").
			Warn("No authentication handler was set in the rule")
		return nil, err
	}

	for _, a := range rl.Authenticators {
		anh, err := d.r.PipelineAuthenticator(a.Handler)
		if err != nil {
			d.r.Logger().WithError(err).
				WithFields(fields).
				WithField("granted", false).
				WithField("authentication_handler", a.Handler).
				WithField("reason_id", "unknown_authentication_handler").
				Warn("Unknown authentication handler requested")
			return nil, err
		}

		if err := anh.Validate(a.Config); err != nil {
			d.r.Logger().WithError(err).
				WithFields(fields).
				WithField("granted", false).
				WithField("authentication_handler", a.Handler).
				WithField("reason_id", "invalid_authentication_handler").
				Warn("Unable to validate use of authentication handler")
			return nil, err
		}

		session, err = anh.Authenticate(r, a.Config, rl)
		if err != nil {
			switch errors.Cause(err).Error() {
			case authn.ErrAuthenticatorNotResponsible.Error():
				// The authentication handler is not responsible for handling this request, skip to the next handler
				break
			// case ErrAuthenticatorBypassed.Error():
			// The authentication handler says that no further authentication/authorization is required, and the request should
			// be forwarded to its final destination.
			// return nil
			default:
				d.r.Logger().WithError(err).
					WithFields(fields).
					WithField("granted", false).
					WithField("authentication_handler", a.Handler).
					WithField("reason_id", "authentication_handler_error").
					Warn("The authentication handler encountered an error")
				return nil, err
			}
		} else {
			// The first authenticator that matches must return the session
			found = true
			fields["subject"] = session.Subject
			break
		}
	}

	if !found {
		err := errors.WithStack(helper.ErrUnauthorized)
		d.r.Logger().WithError(err).
			WithFields(fields).
			WithField("granted", false).
			WithField("reason_id", "authentication_handler_no_match").
			Warn("No authentication handler was responsible for handling the authentication request")
		return nil, err
	}

	azh, err := d.r.PipelineAuthorizer(rl.Authorizer.Handler)
	if err != nil {
		d.r.Logger().WithError(err).
			WithFields(fields).
			WithField("granted", false).
			WithField("authorization_handler", rl.Authorizer.Handler).
			WithField("reason_id", "unknown_authorization_handler").
			Warn("Unknown authentication handler requested")
		return nil, err
	}

	if err := azh.Validate(rl.Authorizer.Config); err != nil {
		d.r.Logger().WithError(err).
			WithFields(fields).
			WithField("granted", false).
			WithField("authorization_handler", rl.Authorizer.Handler).
			WithField("reason_id", "invalid_authorization_handler").
			Warn("Unable to validate use of authorization handler")
		return nil, err
	}

	if err := azh.Authorize(r, session, rl.Authorizer.Config, rl); err != nil {
		d.r.Logger().
			WithError(err).
			WithFields(fields).
			WithField("granted", false).
			WithField("authorization_handler", rl.Authorizer.Handler).
			WithField("reason_id", "authorization_handler_error").
			Warn("The authorization handler encountered an error")
		return nil, err
	}

	if len(rl.Mutators) == 0 {
		err = errors.New("No mutation handler was set in the rule")
		d.r.Logger().WithError(err).
			WithFields(fields).
			WithField("granted", false).
			WithField("reason_id", "mutation_handler_missing").
			Warn("No mutation handler was set in the rule")
		return nil, err
	}

	for _, m := range rl.Mutators {
		sh, err := d.r.PipelineMutator(m.Handler)
		if err != nil {
			d.r.Logger().WithError(err).
				WithFields(fields).
				WithField("granted", false).
				WithField("access_url", r.URL.String()).
				WithField("mutation_handler", m.Handler).
				WithField("reason_id", "unknown_mutation_handler").
				Warn("Unknown mutator requested")
			return nil, err
		}

		if err := sh.Validate(m.Config); err != nil {
			d.r.Logger().WithError(err).
				WithFields(fields).
				WithField("granted", false).
				WithField("mutation_handler", m.Handler).
				WithField("reason_id", "invalid_mutation_handler").
				Warn("Invalid mutator requested")
			return nil, err
		}

		if err := sh.Mutate(r, session, m.Config, rl); err != nil {
			d.r.Logger().WithError(err).
				WithFields(fields).
				WithField("granted", false).
				WithField("mutation_handler", m.Handler).
				WithField("reason_id", "mutation_handler_error").
				Warn("The mutation handler encountered an error")
			return nil, err
		}
	}

	return session, nil
}
