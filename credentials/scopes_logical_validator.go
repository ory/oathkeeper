package credentials

import (
	"github.com/ory/herodot"
	"github.com/pkg/errors"
)

type ScopesValidator func(scopeResult map[string]bool) error

func DefaultValidation(scopeResult map[string]bool) error {
	for sc, result := range scopeResult {
		if !result {
			return errors.WithStack(herodot.ErrInternalServerError.WithReasonf(`JSON Web Token is missing required scope "%s".`, sc))
		}
	}

	return nil
}

func AnyValidation(scopeResult map[string]bool) error {
	for _, result := range scopeResult {
		if result {
			return nil
		}
	}

	return errors.WithStack(herodot.ErrInternalServerError.WithReasonf(`JSON Web Token is missing required scope`))
}
