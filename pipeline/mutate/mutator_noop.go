// Copyright © 2022 Ory Corp

package mutate

import (
	"encoding/json"
	"net/http"

	"github.com/ory/oathkeeper/driver/configuration"
	"github.com/ory/oathkeeper/pipeline"
	"github.com/ory/oathkeeper/pipeline/authn"
)

type MutatorNoop struct{ c configuration.Provider }

func NewMutatorNoop(c configuration.Provider) *MutatorNoop {
	return &MutatorNoop{c: c}
}

func (a *MutatorNoop) GetID() string {
	return "noop"
}

func (a *MutatorNoop) Mutate(r *http.Request, session *authn.AuthenticationSession, config json.RawMessage, _ pipeline.Rule) error {
	session.Header = r.Header
	return nil
}

func (a *MutatorNoop) Validate(config json.RawMessage) error {
	if !a.c.MutatorIsEnabled(a.GetID()) {
		return NewErrMutatorNotEnabled(a)
	}

	if err := a.c.MutatorConfig(a.GetID(), config, nil); err != nil {
		return NewErrMutatorMisconfigured(a, err)
	}
	return nil
}
