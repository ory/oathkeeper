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
	"crypto/md5"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"text/template"
	"time"

	"github.com/dgraph-io/ristretto"

	"github.com/form3tech-oss/jwt-go"

	"github.com/pborman/uuid"
	"github.com/pkg/errors"

	"github.com/ory/oathkeeper/credentials"
	"github.com/ory/oathkeeper/driver/configuration"
	"github.com/ory/oathkeeper/pipeline"
	"github.com/ory/oathkeeper/pipeline/authn"
	"github.com/ory/oathkeeper/x"
	"github.com/ory/x/urlx"
)

type MutatorIDTokenRegistry interface {
	credentials.SignerRegistry
}

type MutatorIDToken struct {
	c             configuration.Provider
	r             MutatorIDTokenRegistry
	templates     *template.Template
	templatesLock sync.Mutex

	tokenCache        *ristretto.Cache
	tokenCacheEnabled bool
}

type CredentialsIDTokenConfig struct {
	Claims    string `json:"claims"`
	IssuerURL string `json:"issuer_url"`
	JWKSURL   string `json:"jwks_url"`
	TTL       string `json:"ttl"`
}

func (c *CredentialsIDTokenConfig) ClaimsTemplateID() string {
	return fmt.Sprintf("%x", md5.Sum([]byte(c.Claims)))
}

func NewMutatorIDToken(c configuration.Provider, r MutatorIDTokenRegistry) *MutatorIDToken {
	cache, _ := ristretto.NewCache(&ristretto.Config{
		NumCounters: 10000,
		MaxCost:     1 << 25,
		BufferItems: 64,
	})
	return &MutatorIDToken{r: r, c: c, templates: x.NewTemplate("id_token"), tokenCache: cache, tokenCacheEnabled: true}
}

func (a *MutatorIDToken) GetID() string {
	return "id_token"
}

func (a *MutatorIDToken) WithCache(t *template.Template) {
	a.templates = t
}

func (a *MutatorIDToken) SetCaching(token bool) {
	a.tokenCacheEnabled = token
}

type idTokenCacheContainer struct {
	ExpiresAt time.Time
	TTL       time.Duration
	Token     string
}

func (a *MutatorIDToken) cacheKey(config *CredentialsIDTokenConfig, ttl time.Duration, claims []byte, session *authn.AuthenticationSession) string {
	return fmt.Sprintf("%x",
		md5.Sum([]byte(fmt.Sprintf("%s|%s|%s|%s|%s", config.IssuerURL, ttl, config.JWKSURL, claims, session.Subject))),
	)
}

func (a *MutatorIDToken) tokenFromCache(config *CredentialsIDTokenConfig, session *authn.AuthenticationSession, claims []byte, ttl time.Duration) (string, bool) {
	if !a.tokenCacheEnabled {
		return "", false
	}

	key := a.cacheKey(config, ttl, claims, session)

	item, found := a.tokenCache.Get(key)
	if !found {
		return "", false
	}

	container := item.(*idTokenCacheContainer)
	if container.ExpiresAt.Before(time.Now().Add(ttl * 1 / 10)) {
		a.tokenCache.Del(key)
		return "", false
	}

	return container.Token, true
}

func (a *MutatorIDToken) tokenToCache(config *CredentialsIDTokenConfig, session *authn.AuthenticationSession, claims []byte, ttl time.Duration, expiresAt time.Time, token string) {
	if !a.tokenCacheEnabled {
		return
	}

	key := a.cacheKey(config, ttl, claims, session)
	a.tokenCache.Set(key, &idTokenCacheContainer{
		TTL:       ttl,
		ExpiresAt: expiresAt,
		Token:     token,
	}, 0)
}

func (a *MutatorIDToken) Mutate(r *http.Request, session *authn.AuthenticationSession, config json.RawMessage, rl pipeline.Rule) error {
	var claims = jwt.MapClaims{}
	c, err := a.Config(config)
	if err != nil {
		return err
	}

	ttl, err := time.ParseDuration(c.TTL)
	if err != nil {
		return errors.WithStack(err)
	}

	var templateClaims []byte
	if len(c.Claims) > 0 {
		t := a.templates.Lookup(c.ClaimsTemplateID())
		if t == nil {
			var err error
			a.templatesLock.Lock()
			t, err = a.templates.New(c.ClaimsTemplateID()).Parse(c.Claims)
			a.templatesLock.Unlock()
			if err != nil {
				return errors.Wrapf(err, `error parsing claims template in rule "%s"`, rl.GetID())
			}
		}

		var b bytes.Buffer
		if err := t.Execute(&b, session); err != nil {
			return errors.Wrapf(err, `error executing claims template in rule "%s"`, rl.GetID())
		}

		templateClaims = b.Bytes()
		if err := json.NewDecoder(&b).Decode(&claims); err != nil {
			return errors.WithStack(err)
		}
	}

	if token, ok := a.tokenFromCache(c, session, templateClaims, ttl); ok {
		session.SetHeader("Authorization", "Bearer "+token)
		return nil
	}

	now := time.Now().UTC()
	exp := now.Add(ttl)
	claims["exp"] = exp.Unix()
	claims["jti"] = uuid.New()
	claims["iat"] = now.Unix()
	claims["iss"] = c.IssuerURL
	claims["nbf"] = now.Unix()
	claims["sub"] = session.Subject

	jwks, err := urlx.Parse(c.JWKSURL)
	if err != nil {
		return errors.WithStack(err)
	}

	signed, err := a.r.CredentialsSigner().Sign(r.Context(), jwks, claims)
	if err != nil {
		return err
	}

	a.tokenToCache(c, session, templateClaims, ttl, exp, signed)
	session.SetHeader("Authorization", "Bearer "+signed)
	return nil
}

func (a *MutatorIDToken) Validate(config json.RawMessage) error {
	if !a.c.MutatorIsEnabled(a.GetID()) {
		return NewErrMutatorNotEnabled(a)
	}

	_, err := a.Config(config)
	return err
}

func (a *MutatorIDToken) Config(config json.RawMessage) (*CredentialsIDTokenConfig, error) {
	var c CredentialsIDTokenConfig
	if err := a.c.MutatorConfig(a.GetID(), config, &c); err != nil {
		return nil, NewErrMutatorMisconfigured(a, err)
	}

	if c.TTL == "" {
		c.TTL = "15m"
	}

	return &c, nil
}
