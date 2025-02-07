// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package credentials

import (
	"github.com/pkg/errors"

	"github.com/ory/herodot"
)

type ScopeValidation func(scopeResult map[string]bool) error

func DefaultValidation(scopeResult map[string]bool) error {
	for sc, result := range scopeResult {
		if !result {
			return errors.WithStack(herodot.ErrInternalServerError.WithReasonf(`JSON Web Token is missing required scope "%s"`, sc))
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
