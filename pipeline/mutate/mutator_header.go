// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package mutate

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"text/template"

	"github.com/ory/oathkeeper/driver/configuration"
	"github.com/ory/oathkeeper/pipeline"
	"github.com/ory/oathkeeper/pipeline/authn"
	"github.com/ory/oathkeeper/x"

	"github.com/pkg/errors"
)

type MutatorHeaderConfig struct {
	Headers map[string]string `json:"headers"`
}

type MutatorHeader struct {
	c configuration.Provider
	t *template.Template
}

func NewMutatorHeader(c configuration.Provider) *MutatorHeader {
	return &MutatorHeader{c: c, t: x.NewTemplate("header")}
}

func (a *MutatorHeader) GetID() string {
	return "header"
}

func (a *MutatorHeader) WithCache(t *template.Template) {
	a.t = t
}

func (a *MutatorHeader) Mutate(_ *http.Request, session *authn.AuthenticationSession, config json.RawMessage, rl pipeline.Rule) error {
	cfg, err := a.config(config)
	if err != nil {
		return err
	}

	for hdr, templateString := range cfg.Headers {
		var tmpl *template.Template
		var err error

		templateId := fmt.Sprintf("%s:%s", rl.GetID(), hdr)
		tmpl = a.t.Lookup(templateId)
		if tmpl == nil {
			tmpl, err = a.t.New(templateId).Parse(templateString)
			if err != nil {
				return errors.Wrapf(err, `error parsing headers template "%s" in rule "%s"`, templateString, rl.GetID())
			}
		}

		headerValue := bytes.Buffer{}
		err = tmpl.Execute(&headerValue, session)
		if err != nil {
			return errors.Wrapf(err, `error executing headers template "%s" in rule "%s"`, templateString, rl.GetID())
		}
		session.SetHeader(hdr, headerValue.String())
	}

	return nil
}

func (a *MutatorHeader) Validate(config json.RawMessage) error {
	if !a.c.MutatorIsEnabled(a.GetID()) {
		return NewErrMutatorNotEnabled(a)
	}

	_, err := a.config(config)
	return err
}

func (a *MutatorHeader) config(config json.RawMessage) (*MutatorHeaderConfig, error) {
	var c MutatorHeaderConfig
	if err := a.c.MutatorConfig(a.GetID(), config, &c); err != nil {
		return nil, NewErrMutatorMisconfigured(a, err)
	}

	return &c, nil
}
