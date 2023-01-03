// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package mutate

import (
	"encoding/json"
	"net/http"

	"github.com/pkg/errors"

	"github.com/ory/oathkeeper/pipeline"
	"github.com/ory/oathkeeper/pipeline/authn"
)

type MutatorBroken struct {
	enabled bool
}

func NewMutatorBroken(enabled bool) *MutatorBroken {
	return &MutatorBroken{
		enabled: enabled,
	}
}

func (a *MutatorBroken) GetID() string {
	return "broken"
}

func (a *MutatorBroken) Mutate(r *http.Request, session *authn.AuthenticationSession, config json.RawMessage, _ pipeline.Rule) error {
	return errors.New("forced denial of credentials")
}

func (a *MutatorBroken) Validate(_ json.RawMessage) error {
	if !a.enabled {
		return errors.WithStack(ErrMutatorNotEnabled.WithReasonf(`Mutator "%s" is disabled per configuration.`, a.GetID()))
	}

	return nil
}
