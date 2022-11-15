// Copyright © 2022 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package authz

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"text/template"

	"github.com/pkg/errors"

	"github.com/ory/oathkeeper/driver/configuration"
	"github.com/ory/oathkeeper/helper"
	"github.com/ory/oathkeeper/pipeline"
	"github.com/ory/oathkeeper/pipeline/authn"
	"github.com/ory/oathkeeper/x"
)

type AuthorizerAllowIf struct {
	c configuration.Provider

	t *template.Template
}

type AuthorizerAllowIfConfiguration struct {
	Key   string `json:"key"`
	Op    string `json:"op"`
	Value string `json:"value"`
}

// PayloadTemplateID returns a string with which to associate the payload template.
func (c *AuthorizerAllowIfConfiguration) PayloadTemplateID() string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(c.Key)))
}

func NewAuthorizerAllowIf(c configuration.Provider) *AuthorizerAllowIf {
	return &AuthorizerAllowIf{
		c: c,
		t: x.NewTemplate("allow_if"),
	}
}

func (a *AuthorizerAllowIf) GetID() string {
	return "allow_if"
}

func (a *AuthorizerAllowIf) Authorize(r *http.Request, session *authn.AuthenticationSession, config json.RawMessage, _ pipeline.Rule) error {
	c, err := a.Config(config)
	if err != nil {
		return err
	}
	templateID := c.PayloadTemplateID()
	t := a.t.Lookup(templateID)
	if t == nil {
		var err error
		t, err = a.t.New(templateID).Parse(c.Key)
		if err != nil {
			return errors.WithStack(err)
		}
	}
	var actualValue bytes.Buffer
	if err := t.Execute(&actualValue, session); err != nil {
		return errors.WithStack(err)
	}
	ok, err := a.DoOperation(actualValue.String(), c.Value, c.Op)
	if err != nil {
		return err
	}
	if !ok {
		return errors.WithStack(helper.ErrForbidden)
	}

	return nil
}

func (a *AuthorizerAllowIf) Validate(config json.RawMessage) error {
	if !a.c.AuthorizerIsEnabled(a.GetID()) {
		return NewErrAuthorizerNotEnabled(a)
	}

	_, err := a.Config(config)
	return err
}

func (a *AuthorizerAllowIf) Config(config json.RawMessage) (*AuthorizerAllowIfConfiguration, error) {
	var c AuthorizerAllowIfConfiguration
	if err := a.c.AuthorizerConfig(a.GetID(), config, &c); err != nil {
		return nil, NewErrAuthorizerMisconfigured(a, err)
	}

	return &c, nil
}

func (a *AuthorizerAllowIf) DoOperation(key, value, op string) (bool, error) {
	switch strings.ToLower(op) {
	case "equals":
		return key == value, nil
	default:
		return false, fmt.Errorf("operator %s is not supported", op)
	}
}
