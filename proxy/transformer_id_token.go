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

package proxy

import (
	"bytes"
	"encoding/json"
	"github.com/ory/oathkeeper/credentials"
	"github.com/ory/oathkeeper/driver/configuration"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/ory/oathkeeper/rule"
	"github.com/pborman/uuid"
	"github.com/pkg/errors"
)

type TransformerIDTokenRegistry interface {
	JWTSigner() *credentials.DefaultSigner
}

type CredentialsIDTokenConfig struct {
	Audience []string `json:"aud"`
}

type TransformerIDToken struct {
	c configuration.Provider
	r TransformerIDTokenRegistry
}

func NewCredentialsIssuerIDToken(
	c configuration.Provider,
	r TransformerIDTokenRegistry,
) *TransformerIDToken {
	return &TransformerIDToken{
		r: r,
		c: c,
	}
}

func (a *TransformerIDToken) GetID() string {
	return "id_token"
}

func (a *TransformerIDToken) Transform(r *http.Request, session *AuthenticationSession, config json.RawMessage, rl *rule.Rule) (http.Header, error) {
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

	claims["exp"] = now.Add(a.c.TransformerIDTokenTTL()).Unix()
	claims["jti"] = uuid.New()
	claims["iat"] = now.Unix()
	claims["iss"] = a.c.TransformerIDTokenIssuerURL().String()
	claims["nbf"] = now.Unix()
	claims["sub"] = session.Subject

	var token *jwt.Token
	switch a.km.Algorithm() {
	case "RS256":
		token = jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	case "HS256":
		token = jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	default:
		return nil, errors.Errorf("Encountered unknown signing algorithm %s while signing ID Token", a.km.Algorithm())
	}

	token.Header["kid"] = a.km.PublicKeyID()

	signed, err := token.SignedString(privateKey)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	headers := http.Header{}
	headers.Set("Authorization", "Bearer "+signed)
	return headers, nil
}

func (a *TransformerIDToken) Validate() error {
	if !a.c.TransformerIDTokenIsEnabled() {
		return errors.WithStack(ErrAuthenticatorNotEnabled.WithReasonf("Transformer % is disabled per configuration.", a.GetID()))
	}

	return nil
}
