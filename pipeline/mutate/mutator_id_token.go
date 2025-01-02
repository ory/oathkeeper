// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

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

	"github.com/golang-jwt/jwt/v5"

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

	tokenCache *ristretto.Cache[string, *idTokenCacheContainer]
}

type CredentialsIDTokenConfig struct {
	Claims    string             `json:"claims"`
	IssuerURL string             `json:"issuer_url"`
	JWKSURL   string             `json:"jwks_url"`
	TTL       string             `json:"ttl"`
	Cache     IdTokenCacheConfig `json:"cache"`
}

type IdTokenCacheConfig struct {
	Enabled bool `json:"enabled"`
	MaxCost int  `json:"max_cost"`
}

func (c *CredentialsIDTokenConfig) ClaimsTemplateID() string {
	return fmt.Sprintf("%x", md5.Sum([]byte(c.Claims)))
}

func NewMutatorIDToken(c configuration.Provider, r MutatorIDTokenRegistry) *MutatorIDToken {
	return &MutatorIDToken{r: r, c: c, templates: x.NewTemplate("id_token")}
}

func (a *MutatorIDToken) GetID() string {
	return "id_token"
}

func (a *MutatorIDToken) WithCache(t *template.Template) {
	a.templates = t
}

type idTokenCacheContainer struct {
	ExpiresAt time.Time
	Token     string
}

func (a *MutatorIDToken) cacheKey(config *CredentialsIDTokenConfig, ttl time.Duration, claims []byte, session *authn.AuthenticationSession) string {
	return fmt.Sprintf("%x",
		md5.Sum([]byte(fmt.Sprintf("%s|%s|%s|%s|%s", config.IssuerURL, ttl, config.JWKSURL, claims, session.Subject))),
	)
}

func (a *MutatorIDToken) tokenFromCache(config *CredentialsIDTokenConfig, session *authn.AuthenticationSession, claims []byte, ttl time.Duration) (string, bool) {
	if !config.Cache.Enabled {
		return "", false
	}

	key := a.cacheKey(config, ttl, claims, session)

	item, found := a.tokenCache.Get(key)
	if !found {
		return "", false
	}

	if item.ExpiresAt.Before(time.Now().Add(ttl * 1 / 10)) {
		a.tokenCache.Del(key)
		return "", false
	}

	return item.Token, true
}

func (a *MutatorIDToken) tokenToCache(config *CredentialsIDTokenConfig, session *authn.AuthenticationSession, claims []byte, ttl time.Duration, expiresAt time.Time, token string) {
	if !config.Cache.Enabled {
		return
	}

	key := a.cacheKey(config, ttl, claims, session)
	a.tokenCache.SetWithTTL(
		key,
		&idTokenCacheContainer{
			ExpiresAt: expiresAt,
			Token:     token,
		},
		0,
		ttl,
	)
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
		if err := json.Unmarshal(templateClaims, &claims); err != nil {
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
	c := CredentialsIDTokenConfig{
		Cache: IdTokenCacheConfig{
			Enabled: true, // default to true
		},
	}
	if err := a.c.MutatorConfig(a.GetID(), config, &c); err != nil {
		return nil, NewErrMutatorMisconfigured(a, err)
	}

	if c.TTL == "" {
		c.TTL = "15m"
	}

	cost := int64(c.Cache.MaxCost)
	if cost == 0 {
		cost = 1 << 25
	}

	if a.tokenCache == nil || a.tokenCache.MaxCost() != cost {
		cache, err := ristretto.NewCache(&ristretto.Config[string, *idTokenCacheContainer]{
			// Guessed approximation of max number of items.
			NumCounters: cost * 4,
			// Allocate a max
			MaxCost: cost,
			// This is a best-practice value.
			BufferItems: 64,
			Cost: func(container *idTokenCacheContainer) int64 {
				return int64(len(container.Token))
			},
		})

		if err != nil {
			return nil, err
		}
		a.tokenCache = cache
	}

	return &c, nil
}
