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
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/ory/oathkeeper/rsakey"
	"github.com/ory/oathkeeper/rule"
	"github.com/pborman/uuid"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type CredentialsIDTokenConfig struct {
	Audience []string `json:"aud"`
}

type CredentialsIDToken struct {
	km       rsakey.Manager
	l        logrus.FieldLogger
	lifetime time.Duration
	issuer   string
}

func NewCredentialsIssuerIDToken(
	k rsakey.Manager,
	l logrus.FieldLogger,
	lifetime time.Duration,
	issuer string,
) *CredentialsIDToken {
	return &CredentialsIDToken{
		km:       k,
		l:        l,
		lifetime: lifetime,
		issuer:   issuer,
	}
}

func (a *CredentialsIDToken) GetID() string {
	return "id_token"
}

type Claims struct {
	Audience  []string `json:"aud,omitempty"`
	ExpiresAt int64    `json:"exp,omitempty"`
	Id        string   `json:"jti,omitempty"`
	IssuedAt  int64    `json:"iat,omitempty"`
	Issuer    string   `json:"iss,omitempty"`
	NotBefore int64    `json:"nbf,omitempty"`
	Subject   string   `json:"sub,omitempty"`
}

func (c *Claims) Valid() error {
	return nil
}

func (a *CredentialsIDToken) Issue(r *http.Request, session *AuthenticationSession, config json.RawMessage, rl *rule.Rule) error {
	privateKey, err := a.km.PrivateKey()
	if err != nil {
		return errors.WithStack(err)
	}
	if len(config) == 0 {
		config = []byte("{}")
	}

	var cc CredentialsIDTokenConfig
	d := json.NewDecoder(bytes.NewBuffer(config))
	d.DisallowUnknownFields()
	if err := d.Decode(&cc); err != nil {
		return errors.WithStack(err)
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

	claims["exp"] = now.Add(a.lifetime).Unix()
	claims["jti"] = uuid.New()
	claims["iat"] = now.Unix()
	claims["iss"] = a.issuer
	claims["nbf"] = now.Unix()
	claims["sub"] = session.Subject

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = a.km.PublicKeyID()

	signed, err := token.SignedString(privateKey)
	if err != nil {
		return errors.WithStack(err)
	}

	r.Header.Set("Authorization", "Bearer "+signed)
	return nil
}
