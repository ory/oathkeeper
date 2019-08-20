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
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"text/template"
	"time"

	"github.com/ory/x/jsonx"

	"github.com/dgrijalva/jwt-go"

	"github.com/pborman/uuid"
	"github.com/pkg/errors"

	"github.com/ory/oathkeeper/credentials"
	"github.com/ory/oathkeeper/driver/configuration"
	"github.com/ory/oathkeeper/pipeline"
	"github.com/ory/oathkeeper/pipeline/authn"
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
	Claims string `json:"claims"`
}

func (c *CredentialsIDTokenConfig) ClaimsTemplateID() string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(c.Claims)))
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
	if len(config) == 0 {
		config = json.RawMessage("{}")
	}

	var c CredentialsIDTokenConfig
	if err := jsonx.NewStrictDecoder(bytes.NewBuffer(config)).Decode(&c); err != nil {
		return errors.WithStack(err)
	}

	if len(c.Claims) > 0 {
		t := a.t.Lookup(c.ClaimsTemplateID())
		if t == nil {
			var err error
			t, err = a.t.New(c.ClaimsTemplateID()).Parse(c.Claims)
			if err != nil {
				return errors.Wrapf(err, `error parsing claims template in rule "%s"`, rl.GetID())
			}
		}

		var b bytes.Buffer
		if err := t.Execute(&b, session); err != nil {
			return errors.Wrapf(err, `error executing claims template in rule "%s"`, rl.GetID())
		}

		if err := json.NewDecoder(&b).Decode(&claims); err != nil {
			return errors.WithStack(err)
		}
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
