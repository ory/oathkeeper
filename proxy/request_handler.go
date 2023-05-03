// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package proxy

import (
	"encoding/json"
	"net/http"

	"github.com/ory/herodot"
	"github.com/ory/x/errorsx"

	"github.com/ory/oathkeeper/driver/configuration"
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

type RequestHandler interface {
	HandleError(w http.ResponseWriter, r *http.Request, rl *rule.Rule, handleErr error)
	HandleRequest(r *http.Request, rl *rule.Rule) (session *authn.AuthenticationSession, err error)
	InitializeAuthnSession(r *http.Request, rl *rule.Rule) *authn.AuthenticationSession
}

type requestHandler struct {
	r requestHandlerRegistry
	c configuration.Provider
}

type whenConfig struct {
	When pe.Whens `json:"when"`
}

func NewRequestHandler(r requestHandlerRegistry, c configuration.Provider) RequestHandler {
	return &requestHandler{r: r, c: c}
}

// matchesWhen
func (d *requestHandler) matchesWhen(w http.ResponseWriter, r *http.Request, h pe.Handler, config json.RawMessage, handleErr error) error {
	var when whenConfig
	if err := d.c.ErrorHandlerConfig(h.GetID(), config, &when); err != nil {
		d.r.Writer().WriteError(w, r, pe.NewErrErrorHandlerMisconfigured(h, err))
		return err
	}

	if err := pe.MatchesWhen(when.When, r, handleErr); err != nil {
		if errorsx.Cause(err) == pe.ErrHandlerNotResponsible {
			return err
		}
		d.r.Writer().WriteError(w, r, errors.WithStack(herodot.ErrInternalServerError.WithReasonf(`Unable to execute error handler "%s". This is either a bug or a configuration issue and should be reported to the administrator. Returned error: "%s". Original error: "%s"`, h.GetID(), err, handleErr)))
		return err
	}

	return nil
}

func (d *requestHandler) HandleError(w http.ResponseWriter, r *http.Request, rl *rule.Rule, handleErr error) {
	if rl == nil {
		// Create a new, empty rule.
		rl = new(rule.Rule)
	}

	var h pe.Handler
	var config json.RawMessage
	for _, re := range rl.Errors {
		handler, err := d.r.PipelineErrorHandler(re.Handler)
		if err != nil {
			d.r.Writer().WriteError(w, r, errors.WithStack(herodot.ErrInternalServerError.WithReasonf(
				"Unable to find error handler named: %s. This is a configuration issue and should be reported to the administrator.", re.Handler,
			)))
			return
		}

		if err := d.matchesWhen(w, r, handler, re.Config, handleErr); errorsx.Cause(err) == pe.ErrHandlerNotResponsible {
			continue
		} else if err != nil {
			// error was handled already by d.matchesWhen
			return
		}

		if h != nil {
			d.r.Writer().WriteError(w, r, errors.WithStack(herodot.ErrInternalServerError.WithReasonf(
				`Found more than one error handlers to be responsible for this request. This is a configuration error that needs to be resolved by the system administrator."`,
			)))
			return
		}

		h = handler
		config = re.Config
	}

	if h == nil {
		for _, name := range d.c.ErrorHandlerFallbackSpecificity() {
			if !d.c.ErrorHandlerIsEnabled(name) {
				d.r.Writer().WriteError(w, r, errors.WithStack(herodot.ErrInternalServerError.WithReasonf(
					`Fallback error handler "%s" was requested but is disabled or unknown. This is a configuration issue and should be reported to the administrator.`, name,
				)))
				return
			}

			handler, err := d.r.PipelineErrorHandler(name)
			if err != nil {
				d.r.Writer().WriteError(w, r, errors.WithStack(herodot.ErrInternalServerError.WithReasonf(
					`Unable to find fallback error handler named "%s". This is a configuration issue and should be reported to the administrator.`, name,
				)))
				return
			}

			if err := d.matchesWhen(w, r, handler, nil, handleErr); errorsx.Cause(err) == pe.ErrHandlerNotResponsible {
				continue
			} else if err != nil {
				// error was handled already by d.matchesWhen
				return
			}

			h = handler
			break
		}
	}

	if h == nil {
		d.r.Writer().WriteError(w, r, errors.WithStack(herodot.ErrInternalServerError.WithReasonf(
			"Unable to handle HTTP request because no matching error handling strategy was found. This is a bug and should be reported to: http://github.com/ory/oathkeeper",
		)))
		return
	}

	if err := h.Validate(config); err != nil {
		d.r.Writer().WriteError(w, r, err)
		return
	}

	if err := h.Handle(w, r, config, rl, handleErr); err != nil {
		d.r.Writer().WriteError(w, r, errors.WithStack(herodot.ErrInternalServerError.WithReasonf(
			`Unable to execute error handler "%s". This is either a bug or a configuration issue and should be reported to the administrator. Returned error: "%s". Original error: "%s"`, h.GetID(), err, handleErr,
		)))
		return
	}
}

func (d *requestHandler) HandleRequest(r *http.Request, rl *rule.Rule) (session *authn.AuthenticationSession, err error) {
	var found bool

	fields := map[string]interface{}{
		"http_method":     r.Method,
		"http_url":        r.URL.String(),
		"http_host":       r.Host,
		"http_user_agent": r.UserAgent(),
		"rule_id":         rl.ID,
	}

	// initialize the session used during all the flow
	session = d.InitializeAuthnSession(r, rl)

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

		err = anh.Authenticate(r, session, a.Config, rl)
		if err != nil {
			switch errors.Cause(err).Error() {
			case authn.ErrAuthenticatorNotResponsible.Error():
				// The authentication handler is not responsible for handling this request, skip to the next handler
				break
			// case ErrAuthenticatorBypassed.Error():
			// The authentication handler says that no further authentication/authorization is required, and the request should
			// be forwarded to its final destination.
			// return nil
			case helper.ErrUnauthorized.ErrorField:
				d.r.Logger().Info(err)
				return nil, err
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

// InitializeAuthnSession creates an authentication session and initializes it with a Match context if possible
func (d *requestHandler) InitializeAuthnSession(r *http.Request, rl *rule.Rule) *authn.AuthenticationSession {

	session := &authn.AuthenticationSession{
		Subject: "",
	}

	values, err := rl.ExtractRegexGroups(d.c.AccessRuleMatchingStrategy(), r.URL)
	if err != nil {
		d.r.Logger().WithError(err).
			WithField("rule_id", rl.ID).
			WithField("access_url", r.URL.String()).
			WithField("reason_id", "capture_groups_error").
			Warn("Unable to capture the groups for the MatchContext")
	} else {
		session.MatchContext = authn.MatchContext{
			RegexpCaptureGroups: values,
			URL:                 r.URL,
			Method:              r.Method,
			Header:              r.Header,
		}
	}

	return session
}
