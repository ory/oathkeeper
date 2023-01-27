// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package rule

import (
	"net/url"

	"github.com/pkg/errors"

	"github.com/ory/herodot"

	"github.com/ory/oathkeeper/pipeline/authn"
	"github.com/ory/oathkeeper/pipeline/authz"
	pe "github.com/ory/oathkeeper/pipeline/errors"
	"github.com/ory/oathkeeper/pipeline/mutate"
)

type validatorRegistry interface {
	authn.Registry
	authz.Registry
	mutate.Registry
	pe.Registry
}

type Validator interface {
	Validate(r *Rule) error
}

var _ Validator = new(ValidatorDefault)

type ValidatorDefault struct {
	r validatorRegistry
}

func NewValidatorDefault(r validatorRegistry) *ValidatorDefault {
	return &ValidatorDefault{r: r}
}

func (v *ValidatorDefault) validateAuthenticators(r *Rule) error {
	if len(r.Authenticators) == 0 {
		return errors.WithStack(herodot.ErrInternalServerError.WithReason(`Value of "authenticators" must be set and can not be an empty array.`))
	}

	for k, a := range r.Authenticators {
		auth, err := v.r.PipelineAuthenticator(a.Handler)
		if err != nil {
			return herodot.ErrInternalServerError.WithReasonf(`Value "%s" of "authenticators[%d]" is not in list of supported authenticators: %v`, a.Handler, k, v.r.AvailablePipelineAuthenticators()).WithTrace(err).WithDebug(err.Error())
		}

		if err := auth.Validate(a.Config); err != nil {
			return err
		}
	}

	return nil
}

func (v *ValidatorDefault) validateAuthorizer(r *Rule) error {
	if r.Authorizer.Handler == "" {
		return errors.WithStack(herodot.ErrInternalServerError.WithReason(`Value of "authorizer.handler" can not be empty.`))
	}

	auth, err := v.r.PipelineAuthorizer(r.Authorizer.Handler)
	if err != nil {
		return errors.WithStack(herodot.ErrInternalServerError.WithReasonf(`Value "%s" of "authorizer.handler" is not in list of supported authorizers: %v`, r.Authorizer.Handler, v.r.AvailablePipelineAuthorizers()).WithTrace(err).WithDebug(err.Error()))
	}

	return auth.Validate(r.Authorizer.Config)
}

func (v *ValidatorDefault) validateMutators(r *Rule) error {
	if len(r.Mutators) == 0 {
		return errors.WithStack(herodot.ErrInternalServerError.WithReason(`Value of "mutators" must be set and can not be an empty array.`))
	}

	for k, m := range r.Mutators {
		mutator, err := v.r.PipelineMutator(m.Handler)
		if err != nil {
			return herodot.ErrInternalServerError.WithReasonf(`Value "%s" of "mutators[%d]" is not in list of supported mutators: %v`, m.Handler, k,
				v.r.AvailablePipelineMutators()).WithTrace(err).WithDebug(err.Error())
		}

		if err := mutator.Validate(m.Config); err != nil {
			return err
		}
	}

	return nil
}

func (v *ValidatorDefault) validateErrorHandlers(r *Rule) error {
	for k, m := range r.Errors {
		mutator, err := v.r.PipelineErrorHandler(m.Handler)
		if err != nil {
			return herodot.ErrInternalServerError.WithReasonf(`Value "%s" of "errors[%d]" is not in list of supported errors: %v`, m.Handler, k,
				v.r.AvailablePipelineErrorHandlers()).WithTrace(err).WithDebug(err.Error())
		}

		if err := mutator.Validate(m.Config); err != nil {
			return err
		}
	}

	return nil
}

func (v *ValidatorDefault) Validate(r *Rule) error {
	if r.Match == nil {
		return errors.WithStack(herodot.ErrInternalServerError.WithReasonf(`Value "match" is empty but must be set.`))
	}

	if r.Match.GetURL() == "" {
		return errors.WithStack(herodot.ErrInternalServerError.WithReasonf(`Value "%s" of "match.url" field must not be empty.`, r.Match.GetURL()))
	}

	if r.Upstream.URL == "" {
		// Having no upstream URL is fine here because the judge does not need an upstream!
	} else if _, err := url.ParseRequestURI(r.Upstream.URL); err != nil {
		return errors.WithStack(herodot.ErrInternalServerError.WithReasonf(`Value "%s" of "upstream.url" is not a valid url: %s`, r.Upstream.URL, err))
	}

	if err := v.validateAuthenticators(r); err != nil {
		return err
	}

	if err := v.validateAuthorizer(r); err != nil {
		return err
	}

	if err := v.validateMutators(r); err != nil {
		return err
	}

	if err := v.validateErrorHandlers(r); err != nil {
		return err
	}

	return nil
}
