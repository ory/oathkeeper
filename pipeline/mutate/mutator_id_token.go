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
	"text/template"
	"time"

	"github.com/ory/x/jsonx"

	"github.com/dgrijalva/jwt-go"

	"github.com/ory/oathkeeper/credentials"
	"github.com/ory/oathkeeper/driver/configuration"
	"github.com/ory/oathkeeper/pipeline"
	"github.com/ory/oathkeeper/pipeline/authn"
	"github.com/pborman/uuid"
	"github.com/pkg/errors"
)

type MutatorIDTokenRegistry interface {
	credentials.SignerRegistry
}

type MutatorIDToken struct {
	c configuration.Provider
	r MutatorIDTokenRegistry
	t *template.Template
}

type CredentialsIDTokenConfig struct {
	Claims jwt.MapClaims `json:"claims"`
}

func NewMutatorIDToken(c configuration.Provider, r MutatorIDTokenRegistry) *MutatorIDToken {
	return &MutatorIDToken{r: r, c: c, t: newTemplate("id_token")}
}

func (a *MutatorIDToken) GetID() string {
	return "id_token"
}

func (a *MutatorIDToken) WithCache(t *template.Template) {
	a.t = t
}

func (a *MutatorIDToken) Mutate(r *http.Request, session *authn.AuthenticationSession, config json.RawMessage, rl pipeline.Rule) error {
	var claims = jwt.MapClaims{}
	if len(config) != 0 {

		var err error

		templateID := rl.GetID()
		tmpl := a.t.Lookup(templateID)
		if tmpl == nil {
			tmpl, err = a.t.New(templateID).Parse(string(config))
			if err != nil {
				return errors.Wrapf(err, `error parsing claims template in rule "%s"`, rl.GetID())
			}
		}

		b := bytes.Buffer{}
		if err := tmpl.Execute(&b, session); err != nil {
			return errors.Wrapf(err, `error executing claims template in rule "%s"`, rl.GetID())
		}

		var cc CredentialsIDTokenConfig
		if err := jsonx.NewStrictDecoder(bytes.NewBuffer(b.Bytes())).Decode(&cc); err != nil {
			return errors.WithStack(err)
		}

		claims = cc.Claims
	}

	now := time.Now().UTC()
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
		return err
	}

	session.SetHeader("Authorization", "Bearer "+signed)
	return nil
}

func (a *MutatorIDToken) Validate() error {
	if !a.c.MutatorIDTokenIsEnabled() {
		return errors.WithStack(ErrMutatorNotEnabled.WithReasonf(`Mutator "%s" is disabled per configuration.`, a.GetID()))
	}

	if a.c.MutatorIDTokenIssuerURL() == nil {
		return errors.WithStack(ErrMutatorNotEnabled.WithReasonf(`Configuration for mutator "%s" did not specify any values for configuration key "%s" and is thus disabled.`, a.GetID(), configuration.ViperKeyMutatorIDTokenIssuerURL))
	}

	if a.c.MutatorIDTokenJWKSURL() == nil {
		return errors.WithStack(ErrMutatorNotEnabled.WithReasonf(`Configuration for mutator "%s" did not specify any values for configuration key "%s" and is thus disabled.`, a.GetID(), configuration.ViperKeyMutatorIDTokenJWKSURL))
	}

	return nil
}
