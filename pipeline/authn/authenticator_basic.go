// Copyright © 2022 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package authn

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/ory/oathkeeper/driver/configuration"
	"github.com/ory/oathkeeper/helper"
	"github.com/ory/oathkeeper/pipeline"
)

type AuthenticatorBasicConfiguration struct {
	Credentials string `json:"credentials"`
}

type AuthenticatorBasic struct {
	c configuration.Provider
}

func NewAuthenticatorBasic(
	c configuration.Provider,
) *AuthenticatorBasic {
	return &AuthenticatorBasic{
		c: c,
	}
}

func (a *AuthenticatorBasic) GetID() string {
	return "basic"
}

func (a *AuthenticatorBasic) Validate(config json.RawMessage) error {
	if !a.c.AuthenticatorIsEnabled(a.GetID()) {
		return NewErrAuthenticatorNotEnabled(a)
	}

	_, err := a.Config(config)
	return err
}

func (a *AuthenticatorBasic) Config(config json.RawMessage) (*AuthenticatorBasicConfiguration, error) {
	var c AuthenticatorBasicConfiguration
	if err := a.c.AuthenticatorConfig(a.GetID(), config, &c); err != nil {
		return nil, NewErrAuthenticatorMisconfigured(a, err)
	}

	return &c, nil
}

func (a *AuthenticatorBasic) Authenticate(r *http.Request, session *AuthenticationSession, config json.RawMessage, _ pipeline.Rule) error {
	cf, err := a.Config(config)
	if err != nil {
		return err
	}

	authorization := r.Header.Get("Authorization")
	if authorization == "" {
		return helper.ErrUnauthorized
	}

	token, err := BasicTokenFromHeader(authorization)
	if err != nil {
		return helper.ErrUnauthorized.WithReason("Basic token is not correctly base64 encoded")
	}

	h := sha256.New()
	h.Write([]byte(token))
	hash := hex.EncodeToString(h.Sum(nil))

	if hash == cf.Credentials {
		return nil
	}

	return helper.ErrUnauthorized
}

func BasicTokenFromHeader(header string) (string, error) {
	split := strings.SplitN(header, " ", 2)
	if len(split) != 2 || !strings.EqualFold(strings.ToLower(split[0]), "basic") {
		return "", nil
	}
	rawDecodedText, err := base64.StdEncoding.DecodeString(split[1])
	if err != nil {
		return "", err
	}

	return string(rawDecodedText), nil
}
