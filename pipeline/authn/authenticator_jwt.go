// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package authn

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/trace"

	"github.com/ory/herodot"
	"github.com/ory/oathkeeper/credentials"
	"github.com/ory/oathkeeper/driver/configuration"
	"github.com/ory/oathkeeper/helper"
	"github.com/ory/oathkeeper/pipeline"
	"github.com/ory/x/jwtx"
	"github.com/ory/x/otelx"
)

type AuthenticatorJWTRegistry interface {
	credentials.VerifierRegistry
	Tracer() trace.Tracer
}

type AuthenticatorOAuth2JWTConfiguration struct {
	Scope               []string                    `json:"required_scope"`
	Audience            []string                    `json:"target_audience"`
	Issuers             []string                    `json:"trusted_issuers"`
	AllowedAlgorithms   []string                    `json:"allowed_algorithms"`
	JWKSURLs            []string                    `json:"jwks_urls"`
	ScopeStrategy       string                      `json:"scope_strategy"`
	BearerTokenLocation *helper.BearerTokenLocation `json:"token_from"`
}

type AuthenticatorJWT struct {
	c configuration.Provider
	r AuthenticatorJWTRegistry
}

func NewAuthenticatorJWT(
	c configuration.Provider,
	r AuthenticatorJWTRegistry,
) *AuthenticatorJWT {
	return &AuthenticatorJWT{
		c: c,
		r: r,
	}
}

func (a *AuthenticatorJWT) GetID() string {
	return "jwt"
}

func (a *AuthenticatorJWT) Validate(config json.RawMessage) error {
	if !a.c.AuthenticatorIsEnabled(a.GetID()) {
		return NewErrAuthenticatorNotEnabled(a)
	}

	_, err := a.Config(config)
	return err
}

func (a *AuthenticatorJWT) Config(config json.RawMessage) (*AuthenticatorOAuth2JWTConfiguration, error) {
	var c AuthenticatorOAuth2JWTConfiguration
	if err := a.c.AuthenticatorConfig(a.GetID(), config, &c); err != nil {
		return nil, NewErrAuthenticatorMisconfigured(a, err)
	}

	return &c, nil
}

func (a *AuthenticatorJWT) Authenticate(r *http.Request, session *AuthenticationSession, config json.RawMessage, _ pipeline.Rule) (err error) {
	ctx, span := a.r.Tracer().Start(r.Context(), "pipeline.authn.AuthenticatorJWT.Authenticate")
	defer otelx.End(span, &err)
	r = r.WithContext(ctx)

	cf, err := a.Config(config)
	if err != nil {
		return err
	}

	token := helper.BearerTokenFromRequest(r, cf.BearerTokenLocation)
	if token == "" {
		return errors.WithStack(ErrAuthenticatorNotResponsible)
	}

	// If the token is not a JWT, declare ourselves not responsible. This enables using fallback authenticators (i. e.
	// bearer_token or oauth2_introspection) for different token types at the same location.
	if len(strings.Split(token, ".")) != 3 {
		return errors.WithStack(ErrAuthenticatorNotResponsible)
	}

	if len(cf.AllowedAlgorithms) == 0 {
		cf.AllowedAlgorithms = []string{"RS256"}
	}

	jwksu, err := a.c.ParseURLs(cf.JWKSURLs)
	if err != nil {
		return err
	}

	pt, err := a.r.CredentialsVerifier().Verify(r.Context(), token, &credentials.ValidationContext{
		Algorithms:    cf.AllowedAlgorithms,
		KeyURLs:       jwksu,
		Scope:         cf.Scope,
		Issuers:       cf.Issuers,
		Audiences:     cf.Audience,
		ScopeStrategy: a.c.ToScopeStrategy(cf.ScopeStrategy, "authenticators.jwt.Config.scope_strategy"),
	})
	if err != nil {
		de := herodot.ToDefaultError(err, "")
		r := fmt.Sprintf("%+v", de)
		return a.tryEnrichResultErr(token, helper.ErrUnauthorized.WithReason(r).WithTrace(err))
	}

	claims, ok := pt.Claims.(jwt.MapClaims)
	if !ok {
		return errors.WithStack(herodot.ErrInternalServerError.WithReasonf("Expected JSON Web Token claims to be of type jwt.MapClaims but got: %T", pt.Claims))
	}

	session.Subject = jwtx.ParseMapStringInterfaceClaims(claims).Subject
	session.Extra = claims

	return nil
}

func (a *AuthenticatorJWT) tryEnrichResultErr(token string, err *herodot.DefaultError) *herodot.DefaultError {
	t, _ := jwt.ParseWithClaims(token, jwt.MapClaims{}, nil, jwt.WithIssuedAt())
	if t == nil {
		return err
	}
	claims, ok := t.Claims.(jwt.MapClaims)
	if !ok {
		return err
	}
	jsonVal, err2 := json.Marshal(claims)
	if err2 != nil {
		return err
	}
	return err.WithDetail("jwt_claims", string(jsonVal))
}
