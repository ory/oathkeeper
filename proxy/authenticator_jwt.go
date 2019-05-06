package proxy

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"github.com/ory/oathkeeper/credentials"
	"github.com/ory/oathkeeper/driver/configuration"
	"net/http"
	"strings"

	"github.com/ory/x/stringsx"

	"github.com/dgrijalva/jwt-go"
	"github.com/ory/fosite"
	"github.com/ory/go-convenience/jwtx"
	"github.com/ory/go-convenience/mapx"
	"github.com/ory/go-convenience/stringslice"
	"github.com/ory/oathkeeper/helper"
	"github.com/ory/oathkeeper/rule"
	"github.com/pkg/errors"
)

type AuthenticatorJWTRegistry interface {
	JWKSFetcher() credentials.Fetcher
}

type AuthenticatorOAuth2JWTConfiguration struct {
	// An array of OAuth 2.0 scopes that are required when accessing an endpoint protected by this handler.
	// If the token used in the Authorization header did not request that specific scope, the request is denied.
	Scopes []string `json:"required_scope"`

	// An array of audiences that are required when accessing an endpoint protected by this handler.
	// If the token used in the Authorization header is not intended for any of the requested audiences, the request is denied.
	Audience []string `json:"target_audience"`

	// The token must have been issued by one of the issuers listed in this array.
	Issuers []string `json:"trusted_issuers"`

	AllowedAlgorithms []string `json:"allowed_algorithms"`
}

type AuthenticatorJWT struct {
	c configuration.Provider
	r AuthenticatorJWTRegistry
}

func NewAuthenticatorJWT(
	r AuthenticatorJWTRegistry,
	c configuration.Provider,
) *AuthenticatorJWT {
	return &AuthenticatorJWT{
		c: c,
		r: r,
	}
}

func (a *AuthenticatorJWT) GetID() string {
	return "jwt"
}

func (a *AuthenticatorJWT) Validate() error {
	if !a.c.AuthenticatorJWTIsEnabled() {
		return errors.WithStack(ErrAuthenticatorNotEnabled.WithReasonf("Authenticator % is disabled per configuration.", a.GetID()))
	}

	if len(a.c.AuthenticatorJWTJWKSURIs()) == 0 {
		return errors.WithStack(ErrAuthenticatorNotEnabled.WithReasonf(`Configuration for authenticator % did not specify any values for configuration key "%s" and is thus disabled.`, a.GetID(), configuration.ViperKeyAuthenticatorJWTJWKSURIs))
	}

	return nil
}

func (a *AuthenticatorJWT) Authenticate(r *http.Request, config json.RawMessage, rl *rule.Rule) (*AuthenticationSession, error) {
	var cf AuthenticatorOAuth2JWTConfiguration
	token := helper.BearerTokenFromRequest(r)
	if token == "" {
		return nil, errors.WithStack(ErrAuthenticatorNotResponsible)
	}

	if len(config) == 0 {
		config = []byte("{}")
	}

	d := json.NewDecoder(bytes.NewBuffer(config))
	d.DisallowUnknownFields()
	if err := d.Decode(&cf); err != nil {
		return nil, errors.WithStack(err)
	}

	if len(cf.AllowedAlgorithms) == 0 {
		cf.AllowedAlgorithms = []string{"RS256"}
	}

	// Parse the token.
	parsedToken, err := jwt.ParseWithClaims(token, jwt.MapClaims{}, func(token *jwt.Token) (interface{}, error) {
		if !stringslice.Has(cf.AllowedAlgorithms, fmt.Sprintf("%s", token.Header["alg"])) {
			return nil, errors.WithStack(helper.ErrUnauthorized.WithReason(fmt.Sprintf(`JSON Web Token used signing method "%s" which is not allowed.`, token.Header["alg"])))
		}

		kid, ok := token.Header["kid"].(string)
		if !ok  || kid == "" {
			return nil, errors.WithStack(helper.ErrUnauthorized.WithReason("The JSON Web Token must contain a kid header value but did not."))
		}

		key, err := a.r.JWKSFetcher().ResolveKey(r.Context(), a.c.AuthenticatorJWTJWKSURIs(), kid, "sig")
		if err != nil {
			return nil, helper.ErrUnauthorized.WithTrace(err).WithDebugf("%s", err)
		}

		switch token.Method.(type) {
		case *jwt.SigningMethodRSA:
			if k, ok := key.Key.(*rsa.PublicKey); ok {
				return k, nil
			}
		case *jwt.SigningMethodECDSA:
			if k, ok := key.Key.(*ecdsa.PublicKey); ok {
				return k, nil
			}
		case *jwt.SigningMethodRSAPSS:
			if k, ok := key.Key.(*rsa.PublicKey); ok {
				return k, nil
			}
		case *jwt.SigningMethodHMAC:
			if k, ok := key.Key.([]byte); ok {
				return k, nil
			}
		default:
			return nil, errors.WithStack(helper.ErrUnauthorized.WithReasonf(`This request object uses unsupported signing algorithm "%s".`, token.Header["alg"]))
		}

		return nil, errors.WithStack(helper.ErrUnauthorized.WithReasonf(`The signing key algorithm does not match the algorithm from the token header.`))
	})

	if err != nil {
		return nil, err
	} else if !parsedToken.Valid {
		return nil, errors.WithStack(fosite.ErrInactiveToken)
	}

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.Errorf("unable to type assert jwt claims to jwt.MapClaims")
	}

	parsedClaims := jwtx.ParseMapStringInterfaceClaims(claims)
	for _, audience := range cf.Audience {
		if !stringslice.Has(parsedClaims.Audience, audience) {
			return nil, errors.WithStack(helper.ErrForbidden.WithReason(fmt.Sprintf("Token audience %v is not intended for target audience %s.", parsedClaims.Audience, audience)))
		}
	}

	if len(cf.Issuers) > 0 {
		if !stringslice.Has(cf.Issuers, parsedClaims.Issuer) {
			return nil, errors.WithStack(helper.ErrForbidden.WithReason(fmt.Sprintf("Token issuer does not match any trusted issuer.")))
		}
	}

	if scopeClaim, err := mapx.GetString(map[interface{}]interface{}{"scope": claims["scope"]}, "scope"); err == nil {
		scopeStrings := strings.Split(scopeClaim, " ")
		scopeInterfaces := make([]interface{}, len(scopeStrings))

		for i := range scopeStrings {
			scopeInterfaces[i] = scopeStrings[i]
		}
		claims["scope"] = scopeInterfaces
	}

	if ss := a.c.AuthenticatorJWTScopeStrategy(); ss != nil {
		for _, scope := range cf.Scopes {
			if !ss(getScopeClaim(claims), scope) {
				return nil, errors.WithStack(helper.ErrForbidden.WithReason(fmt.Sprintf(`JSON Web Token is missing required scope "%s".`, scope)))
			}
		}
	} else {
		if len(cf.Scopes) > 0 {
			return nil, errors.WithStack(helper.ErrRuleFeatureDisabled.WithReason("Scope validation was requested but scope strategy is set to \"none\"."))
		}
	}

	return &AuthenticationSession{
		Subject: parsedClaims.Subject,
		Extra:   claims,
	}, nil
}
