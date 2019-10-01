package authn

import (
	"encoding/json"
	"net/http"

	"github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"

	"github.com/ory/go-convenience/jwtx"
	"github.com/ory/herodot"

	"github.com/ory/oathkeeper/credentials"
	"github.com/ory/oathkeeper/driver/configuration"
	"github.com/ory/oathkeeper/helper"
	"github.com/ory/oathkeeper/pipeline"
)

type AuthenticatorJWTRegistry interface {
	credentials.VerifierRegistry
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

func (a *AuthenticatorJWT) Authenticate(r *http.Request, config json.RawMessage, _ pipeline.Rule) (*AuthenticationSession, error) {
	cf, err := a.Config(config)
	if err != nil {
		return nil, err
	}

	token := helper.BearerTokenFromRequest(r, cf.BearerTokenLocation)
	if token == "" {
		return nil, errors.WithStack(ErrAuthenticatorNotResponsible)
	}

	if len(cf.AllowedAlgorithms) == 0 {
		cf.AllowedAlgorithms = []string{"RS256"}
	}

	jwksu, err := a.c.ParseURLs(cf.JWKSURLs)
	if err != nil {
		return nil, err
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
		return nil, helper.ErrUnauthorized.WithReason(err.Error()).WithTrace(err)
	}

	claims, ok := pt.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("Expected JSON Web Token claims to be of type jwt.MapClaims but got: %T", pt.Claims))
	}

	parsedClaims := jwtx.ParseMapStringInterfaceClaims(claims)
	return &AuthenticationSession{
		Subject: parsedClaims.Subject,
		Extra:   claims,
	}, nil
}
