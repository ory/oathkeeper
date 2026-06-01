// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package mutate

import (
	"encoding/json"
	"net/http"

	"github.com/ory/oathkeeper/pipeline"
	"github.com/ory/oathkeeper/pipeline/authn"
)

type MutatorNoop struct{ d dependencies }

func NewMutatorNoop(d dependencies) *MutatorNoop {
	return &MutatorNoop{d: d}
}

func (a *MutatorNoop) GetID() string { return "noop" }

func (a *MutatorNoop) Mutate(r *http.Request, session *authn.AuthenticationSession, _ json.RawMessage, _ pipeline.Rule) error {
	currentSessionHeaders := session.Header.Clone()
	session.Header = r.Header
	if session.Header == nil {
		session.Header = make(map[string][]string)
	}

	for k, v := range currentSessionHeaders {
		var val string
		if len(v) == 0 {
			val = ""
		} else {
			val = v[0]
		}
		session.SetHeader(k, val)
	}

	return nil
}

func (a *MutatorNoop) Validate(config json.RawMessage) error {
	if !a.d.Config().MutatorIsEnabled(a.GetID()) {
		return NewErrMutatorNotEnabled(a)
	}

	if err := a.d.Config().MutatorConfig(a.GetID(), config, nil); err != nil {
		return NewErrMutatorMisconfigured(a, err)
	}
	return nil
}
