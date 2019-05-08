/*
 * Copyright © 2017-2018 Aeneas Rekkas <aeneas+oss@aeneas.io>
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * @author       Aeneas Rekkas <aeneas+oss@aeneas.io>
 * @copyright  2017-2018 Aeneas Rekkas <aeneas+oss@aeneas.io>
 * @license  	   Apache-2.0
 */

package mutate

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	"github.com/ory/oathkeeper/credentials"
	"github.com/ory/oathkeeper/driver/configuration"
	"github.com/ory/oathkeeper/pipeline"
	"github.com/ory/oathkeeper/pipeline/authn"

	"github.com/dgrijalva/jwt-go"
	"github.com/pborman/uuid"
	"github.com/pkg/errors"
)

type MutatorIDTokenRegistry interface {
	credentials.SignerRegistry
}

type CredentialsIDTokenConfig struct {
	Audience []string `json:"aud"`
}

type MutatorIDToken struct {
	c configuration.Provider
	r MutatorIDTokenRegistry
}

func NewMutatorIDToken(
	c configuration.Provider,
	r MutatorIDTokenRegistry,
) *MutatorIDToken {
	return &MutatorIDToken{
		r: r,
		c: c,
	}
}

func (a *MutatorIDToken) GetID() string {
	return "id_token"
}

func (a *MutatorIDToken) Mutate(r *http.Request, session *authn.AuthenticationSession, config json.RawMessage, _ pipeline.Rule) (http.Header, error) {
	if len(config) == 0 {
		config = []byte("{}")
	}

	var cc CredentialsIDTokenConfig
	d := json.NewDecoder(bytes.NewBuffer(config))
	d.DisallowUnknownFields()
	if err := d.Decode(&cc); err != nil {
		return nil, errors.WithStack(err)
	}

	now := time.Now().UTC()
	claims := jwt.MapClaims{}
	if session.Extra != nil {
		for k, v := range session.Extra {
			claims[k] = v
		}
	}

	if len(cc.Audience) > 0 {
		claims["aud"] = cc.Audience
	}

	claims["exp"] = now.Add(a.c.MutatorIDTokenTTL()).Unix()
	claims["jti"] = uuid.New()
	claims["iat"] = now.Unix()
	claims["iss"] = a.c.MutatorIDTokenIssuerURL().String()
	claims["nbf"] = now.Unix()
	claims["sub"] = session.Subject

	signed, err := a.r.CredentialsSigner().Sign(
		r.Context(),
		a.c.MutatorIDTokenJWKSURL(),
		claims,
	)
	if err != nil {
		return nil, err
	}

	headers := http.Header{}
	headers.Set("Authorization", "Bearer "+signed)
	return headers, nil
}

func (a *MutatorIDToken) Validate() error {
	if !a.c.MutatorIDTokenIsEnabled() {
		return errors.WithStack(authn.ErrAuthenticatorNotEnabled.WithReasonf("Mutator % is disabled per configuration.", a.GetID()))
	}

	if a.c.MutatorIDTokenIssuerURL() == nil {
		return errors.WithStack(authn.ErrAuthenticatorNotEnabled.WithReasonf(`Configuration for transformer % did not specify any values for configuration key "%s" and is thus disabled.`, a.GetID(), configuration.ViperKeyMutatorIDTokenIssuerURL))
	}

	if a.c.MutatorIDTokenJWKSURL() == nil {
		return errors.WithStack(authn.ErrAuthenticatorNotEnabled.WithReasonf(`Configuration for transformer % did not specify any values for configuration key "%s" and is thus disabled.`, a.GetID(), configuration.ViperKeyMutatorIDTokenJWKSURL))
	}

	return nil
}
