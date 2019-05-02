package proxy

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"github.com/ory/oathkeeper/driver/configuration"
	"net/http"
	"net/url"
	"strings"

	"github.com/ory/x/stringsx"

	"github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"
	"gopkg.in/square/go-jose.v2"

	"github.com/ory/fosite"
	"github.com/ory/go-convenience/jwtx"
	"github.com/ory/go-convenience/mapx"
	"github.com/ory/go-convenience/stringslice"
	"github.com/ory/oathkeeper/helper"
	"github.com/ory/oathkeeper/rule"
)

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

type jwksFetcher interface {
	Resolve(location string, forceRefresh bool) (*jose.JSONWebKeySet, error)
}

type AuthenticatorJWT struct {
	c configuration.Provider

	fetcher       jwksFetcher
}

func NewAuthenticatorJWT(c configuration.Provider) *AuthenticatorJWT {
	return &AuthenticatorJWT{
		c:             c,
		fetcher:       fosite.NewDefaultJWKSFetcherStrategy(),
	}
}

func (a *AuthenticatorJWT) GetID() string {
	return "jwt"
}

type tracer interface {
	StackTrace() errors.StackTrace
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

		switch token.Method.(type) {
		case *jwt.SigningMethodRSA:
			return a.findRSAPublicKey(token)
		case *jwt.SigningMethodECDSA:
			return a.findECDSAPublicKey(token)
		case *jwt.SigningMethodRSAPSS:
			return a.findRSAPublicKey(token)
		case *jwt.SigningMethodHMAC:
			return a.findSharedKey(token)
		default:
			return nil, errors.WithStack(helper.ErrUnauthorized.WithReason(fmt.Sprintf(`This request object uses unsupported signing algorithm "%s".`, token.Header["alg"])))
		}
	})

	if err != nil {
		if _, ok := err.(tracer); ok {
			return nil, err
		} else {
			return nil, errors.WithStack(err)
		}
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


	if  ss := a.c.AuthenticatorJWTScopeStrategy(); ss != nil {
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

func getScopeClaim(claims map[string]interface{}) []string {
	var ok bool
	var interim interface{}

	for _, k := range []string{"scp", "scope", "scopes"} {
		if interim, ok = claims[k]; ok {
			break
		}
	}

	if !ok {
		return []string{}
	}

	switch i := interim.(type) {
	case []string:
		return i
	case []interface{}:
		vs := make([]string, len(i))
		for k, v := range i {
			if vv, ok := v.(string); ok {
				vs[k] = vv
			}
		}
		return vs
	case string:
		return stringsx.Splitx(i, " ")
	default:
		return []string{}
	}
}

func (a *AuthenticatorJWT) findRSAPublicKey(t *jwt.Token) (*rsa.PublicKey, error) {
	keys, err := a.fetcher.Resolve(a.c.AuthenticatorJWTJWKSURIs(), false)
	if err != nil {
		return nil, err
	}

	if key, err := findRSAPublicKey(t, keys); err == nil {
		return key, nil
	}

	keys, err = a.fetcher.Resolve(a.jwksURL, true)
	if err != nil {
		return nil, err
	}

	return findRSAPublicKey(t, keys)
}

func (a *AuthenticatorJWT) findECDSAPublicKey(t *jwt.Token) (*ecdsa.PublicKey, error) {
	keys, err := a.fetcher.Resolve(a.jwksURL, false)
	if err != nil {
		return nil, err
	}

	if key, err := findECDSAPublicKey(t, keys); err == nil {
		return key, nil
	}

	keys, err = a.fetcher.Resolve(a.jwksURL, true)
	if err != nil {
		return nil, err
	}

	return findECDSAPublicKey(t, keys)
}

func (a *AuthenticatorJWT) findSharedKey(t *jwt.Token) ([]byte, error) {
	keys, err := a.fetcher.Resolve(a.jwksURL, false)
	if err != nil {
		return nil, err
	}

	if key, err := findSharedKey(t, keys); err == nil {
		return key, nil
	}

	keys, err = a.fetcher.Resolve(a.jwksURL, true)
	if err != nil {
		return nil, err
	}

	return findSharedKey(t, keys)

}

func findRSAPublicKey(t *jwt.Token, set *jose.JSONWebKeySet) (*rsa.PublicKey, error) {
	kid, ok := t.Header["kid"].(string)
	if !ok {
		return nil, errors.WithStack(helper.ErrUnauthorized.WithReason("The JSON Web Token must contain a kid header value but did not."))
	}

	keys := set.Key(kid)
	if len(keys) == 0 {
		return nil, errors.WithStack(helper.ErrUnauthorized.WithReason(fmt.Sprintf("The JSON Web Token uses signing key with kid \"%s\", which could not be found.", kid)))
	}

	for _, key := range keys {
		if key.Use != "sig" {
			continue
		}
		if k, ok := key.Key.(*rsa.PublicKey); ok {
			return k, nil
		}
	}

	return nil, errors.WithStack(helper.ErrUnauthorized.WithReason(fmt.Sprintf("Unable to find RSA public key with use=\"sig\" for kid \"%s\" in JSON Web Key Set.", kid)))
}

func findECDSAPublicKey(t *jwt.Token, set *jose.JSONWebKeySet) (*ecdsa.PublicKey, error) {
	kid, ok := t.Header["kid"].(string)
	if !ok {
		return nil, errors.WithStack(helper.ErrUnauthorized.WithReason("The JSON Web Token must contain a kid header value but did not."))
	}

	keys := set.Key(kid)
	if len(keys) == 0 {
		return nil, errors.WithStack(helper.ErrUnauthorized.WithReason(fmt.Sprintf("The JSON Web Token uses signing key with kid \"%s\", which could not be found.", kid)))
	}

	for _, key := range keys {
		if key.Use != "sig" {
			continue
		}
		if k, ok := key.Key.(*ecdsa.PublicKey); ok {
			return k, nil
		}
	}

	return nil, errors.WithStack(helper.ErrUnauthorized.WithReason(fmt.Sprintf("Unable to find RSA public key with use=\"sig\" for kid \"%s\" in JSON Web Key Set.", kid)))
}

func findSharedKey(t *jwt.Token, set *jose.JSONWebKeySet) ([]byte, error) {
	kid, ok := t.Header["kid"].(string)
	if !ok {
		return nil, errors.WithStack(helper.ErrUnauthorized.WithReason("The JSON Web Token must contain a kid header value but did not."))
	}

	keys := set.Key(kid)
	if len(keys) == 0 {
		return nil, errors.WithStack(helper.ErrUnauthorized.WithReason(fmt.Sprintf("The JSON Web Token uses signing key with kid \"%s\", which could not be found.", kid)))
	}

	for _, key := range keys {
		if key.Use != "sig" {
			continue
		}
		if k, ok := key.Key.([]byte); ok {
			return k, nil
		}
	}

	return nil, errors.WithStack(helper.ErrUnauthorized.WithReason(fmt.Sprintf("Unable to find shared key with use=\"sig\" for kid \"%s\" in JSON Web Key Set.", kid)))
}
