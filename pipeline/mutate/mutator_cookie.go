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

const cookieHeader = "Cookie"

type CredentialsCookiesConfig struct {
	Cookies map[string]string `json:"cookies"`
}

type MutatorCookie struct {
	t *template.Template
	c configuration.Provider
}

func NewMutatorCookie(c configuration.Provider) *MutatorCookie {
	return &MutatorCookie{c: c, t: x.NewTemplate("cookie")}
}

func (a *MutatorCookie) GetID() string {
	return "cookie"
}

func (a *MutatorCookie) WithCache(t *template.Template) {
	a.t = t
}

func (a *MutatorCookie) Mutate(r *http.Request, session *authn.AuthenticationSession, config json.RawMessage, rl pipeline.Rule) error {
	// Cache request cookies
	requestCookies := r.Cookies()

	req := http.Request{Header: map[string][]string{}}

	// Keep track of rule cookies in a map
	cookies := map[string]bool{}

	cfg, err := a.config(config)
	if err != nil {
		return err
	}

	for cookie, templateString := range cfg.Cookies {
		var tmpl *template.Template
		var err error

		templateId := fmt.Sprintf("%s:%s", rl.GetID(), cookie)
		tmpl = a.t.Lookup(templateId)
		if tmpl == nil {
			tmpl, err = a.t.New(templateId).Parse(templateString)
			if err != nil {
				return errors.Wrapf(err, `error parsing cookie template "%s" in rule "%s"`, templateString, rl.GetID())
			}
		}

		cookieValue := bytes.Buffer{}
		err = tmpl.Execute(&cookieValue, session)
		if err != nil {
			return errors.Wrapf(err, `error executing cookie template "%s" in rule "%s"`, templateString, rl.GetID())
		}

		req.AddCookie(&http.Cookie{
			Name:  cookie,
			Value: cookieValue.String(),
		})

		cookies[cookie] = true
	}

	// Re-add previously set cookies that do not coincide with rule cookies
	for _, cookie := range requestCookies {
		// Test if cookie is handled by rule
		if _, ok := cookies[cookie.Name]; !ok {
			// Re-add cookie if not handled by rule
			req.AddCookie(cookie)
		}
	}

	session.SetHeader(cookieHeader, req.Header.Get(cookieHeader))

	return nil
}

func (a *MutatorCookie) Validate(config json.RawMessage) error {
	if !a.c.MutatorIsEnabled(a.GetID()) {
		return NewErrMutatorNotEnabled(a)
	}

	_, err := a.config(config)
	return err
}

func (a *MutatorCookie) config(config json.RawMessage) (*CredentialsCookiesConfig, error) {
	var c CredentialsCookiesConfig
	if err := a.c.MutatorConfig(a.GetID(), config, &c); err != nil {
		return nil, NewErrMutatorMisconfigured(a, err)
	}

	return &c, nil
}
