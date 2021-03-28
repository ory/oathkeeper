package authn

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/form3tech-oss/jwt-go"
	"github.com/ory/go-convenience/jwtx"
	"github.com/ory/herodot"
	"github.com/pkg/errors"

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

func (a *AuthenticatorJWT) Authenticate(r *http.Request, session *AuthenticationSession, config json.RawMessage, _ pipeline.Rule) error {
	cf, err := a.Config(config)
	if err != nil {
		return err
	}

	token := helper.BearerTokenFromRequest(r, cf.BearerTokenLocation)
	if token == "" {
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
		return a.tryEnrichResultErr(token, helper.ErrUnauthorized.WithReason(err.Error()).WithTrace(err))
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
	t, _ := jwt.ParseWithClaims(token, jwt.MapClaims{}, nil)
	if t == nil {
		return err
	}
	claims, ok := t.Claims.(jwt.MapClaims)
	if !ok {
		return err
	}
	var claimKeyPairs []string
	for k, v := range claims {
		// print just root content of the JWT, skip nested objects
		if _, ok := v.(map[string]interface{}); ok {
			continue
		}
		if floatVal, ok := v.(float64); ok && v == float64(int64(floatVal)) {
			// JWT JSON decode deserializes numbers as float64
			claimKeyPairs = append(claimKeyPairs, fmt.Sprintf("%s=%v", k, int64(floatVal)))
		} else {
			claimKeyPairs = append(claimKeyPairs, fmt.Sprintf("%s=%v", k, v))
		}
	}
	return err.WithDetail("jwt_claims", fmt.Sprintf("%+v", strings.Join(claimKeyPairs, ", ")))
}
