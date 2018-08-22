package proxy

import (
	"encoding/json"
	"net/http"

	"github.com/ory/oathkeeper/rule"
	"github.com/pkg/errors"
	"github.com/ory/fosite"
	"gopkg.in/square/go-jose.v2"
	"bytes"
	"github.com/ory/oathkeeper/helper"
	"github.com/dgrijalva/jwt-go"
	"fmt"
	"github.com/ory/go-convenience/stringslice"
	"crypto/rsa"
	"crypto/ecdsa"
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
	jwksURL       string
	fetcher       jwksFetcher
	scopeStrategy fosite.ScopeStrategy
}

func NewAuthenticatorJWT(jwksURL string, scopeStrategy fosite.ScopeStrategy) *AuthenticatorJWT {
	return &AuthenticatorJWT{
		jwksURL:       jwksURL,
		fetcher:       fosite.NewDefaultJWKSFetcherStrategy(),
		scopeStrategy: scopeStrategy,
	}
}

func (a *AuthenticatorJWT) GetID() string {
	return "jwt"
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
	parsedToken, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		if !stringslice.Has(cf.AllowedAlgorithms, fmt.Sprintf("%s", t.Header["alg"])) {
			return nil, errors.WithStack(helper.ErrUnauthorized.WithReason(fmt.Sprintf(`JSON Web Token used signing method "%s" which is not allowed.`, t.Header["alg"])))
		}

		switch t.Method.(type) {
		case *jwt.SigningMethodRSA:
			return a.findRSAPublicKey(t)
		case *jwt.SigningMethodECDSA:
			return a.findECDSAPublicKey(t)
		case *jwt.SigningMethodRSAPSS:
			return a.findRSAPublicKey(t)
		case *jwt.SigningMethodHMAC:
			return a.findSharedKey(t)
		default:
			return nil, errors.WithStack(helper.ErrUnauthorized.WithReason(fmt.Sprintf(`This request object uses unsupported signing algorithm "%s"."`, t.Header["alg"])))
		}
	})

	if err != nil {
		return nil, errors.WithStack(err)
	} else if !parsedToken.Valid {
		return nil, errors.WithStack(fosite.ErrInactiveToken)
	}

	if len(cf.Scopes) > 0 {

	}

		if !stringslice.Has(cf.Audience, parsedToken.Claims["aud"]) {
			return nil, errors.WithStack(helper.ErrForbidden.WithReason(fmt.Sprintf("Token audience is not intended for target audience %s", audience)))
		}

	if !stringslice.Has(cf.Issuers, parsedToken.Claims["iss"]) {
		return nil, errors.WithStack(helper.ErrForbidden.WithReason(fmt.Sprintf("Token issuer does not match any trusted issuer")))
	}

	for _, scope := range cf.Scopes {
		if !a.scopeStrategy(parsedToken.Claims["scope"], scope) {

			// TO BE DONE
			// TO BE DONE
			// TO BE DONE
			// TO BE DONE
			// TO BE DONE
			// TO BE DONE
			return nil, errors.WithStack(helper.ErrForbidden.WithReason(fmt.Sprintf("Token claims TO BE DONE")))
		}
	}
}

func (a *AuthenticatorJWT) findRSAPublicKey(t *jwt.Token) (*rsa.PublicKey, error) {
	keys, err := a.fetcher.Resolve(a.jwksURL, false)
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

	if key, err := findRSAPublicKey(t, keys); err == nil {
		return key, nil
	}

	keys, err = a.fetcher.Resolve(a.jwksURL, true)
	if err != nil {
		return nil, err
	}

	return findRSAPublicKey(t, keys)
}

func (a *AuthenticatorJWT) findSharedKey(t *jwt.Token) (*rsa.PublicKey, error) {
	keys, err := a.fetcher.Resolve(a.jwksURL, false)
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

	return findSharedKey(t, keys)

}

func findRSAPublicKey(t *jwt.Token, set *jose.JSONWebKeySet) (*rsa.PublicKey, error) {
	kid, ok := t.Header["kid"].(string)
	if !ok {
		return nil, errors.WithStack(helper.ErrForbidden.WithReason("The JSON Web Token must contain a kid header value but did not."))
	}

	keys := set.Key(kid)
	if len(keys) == 0 {
		return nil, errors.WithStack(helper.ErrForbidden.WithReason(fmt.Sprintf("The JSON Web Token uses signing key with kid \"%s\", which could not be found.", kid)))
	}

	for _, key := range keys {
		if key.Use != "sig" {
			continue
		}
		if k, ok := key.Key.(*rsa.PublicKey); ok {
			return k, nil
		}
	}

	return nil, errors.WithStack(helper.ErrForbidden.WithReason(fmt.Sprintf("Unable to find RSA public key with use=\"sig\" for kid \"%s\" in JSON Web Key Set.", kid)))
}

func findECDSAPublicKey(t *jwt.Token, set *jose.JSONWebKeySet) (*ecdsa.PublicKey, error) {
	kid, ok := t.Header["kid"].(string)
	if !ok {
		return nil, errors.WithStack(helper.ErrForbidden.WithReason("The JSON Web Token must contain a kid header value but did not."))
	}

	keys := set.Key(kid)
	if len(keys) == 0 {
		return nil, errors.WithStack(helper.ErrForbidden.WithReason(fmt.Sprintf("The JSON Web Token uses signing key with kid \"%s\", which could not be found.", kid)))
	}

	for _, key := range keys {
		if key.Use != "sig" {
			continue
		}
		if k, ok := key.Key.(*ecdsa.PublicKey); ok {
			return k, nil
		}
	}

	return nil, errors.WithStack(helper.ErrForbidden.WithReason(fmt.Sprintf("Unable to find RSA public key with use=\"sig\" for kid \"%s\" in JSON Web Key Set.", kid)))
}

func findSharedKey(t *jwt.Token, set *jose.JSONWebKeySet) ([]byte, error) {
	kid, ok := t.Header["kid"].(string)
	if !ok {
		return nil, errors.WithStack(helper.ErrForbidden.WithReason("The JSON Web Token must contain a kid header value but did not."))
	}

	keys := set.Key(kid)
	if len(keys) == 0 {
		return nil, errors.WithStack(helper.ErrForbidden.WithReason(fmt.Sprintf("The JSON Web Token uses signing key with kid \"%s\", which could not be found.", kid)))
	}

	for _, key := range keys {
		if key.Use != "sig" {
			continue
		}
		if k, ok := key.Key.([]byte); ok {
			return k, nil
		}
	}

	return nil, errors.WithStack(helper.ErrForbidden.WithReason(fmt.Sprintf("Unable to find shared key with use=\"sig\" for kid \"%s\" in JSON Web Key Set.", kid)))
}
