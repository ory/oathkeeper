// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package credentials

import (
	"context"
	"crypto/ecdsa"
	"crypto/rsa"
	"fmt"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/pkg/errors"

	"github.com/ory/fosite"
	"github.com/ory/herodot"
	"github.com/ory/oathkeeper/helper"
	"github.com/ory/x/jwtx"
	"github.com/ory/x/stringslice"
	"github.com/ory/x/stringsx"
)

var _ Verifier = new(VerifierDefault)

type VerifierDefault struct {
	r FetcherRegistry
}

func NewVerifierDefault(f FetcherRegistry) *VerifierDefault {
	return &VerifierDefault{
		r: f,
	}
}

func (v *VerifierDefault) Verify(
	ctx context.Context,
	token string,
	r *ValidationContext,
) (*jwt.Token, error) {
	// Parse the token.
	t, err := jwt.ParseWithClaims(token, jwt.MapClaims{}, func(token *jwt.Token) (interface{}, error) {
		if !stringslice.Has(r.Algorithms, fmt.Sprintf("%s", token.Header["alg"])) {
			return nil, errors.WithStack(herodot.ErrInternalServerError.WithReason(fmt.Sprintf(`JSON Web Token used signing method "%s" which is not allowed.`, token.Header["alg"])))
		}

		kid, ok := token.Header["kid"].(string)
		if !ok || kid == "" {
			return nil, errors.WithStack(herodot.ErrBadRequest.WithReason("The JSON Web Token must contain a kid header value but did not."))
		}

		key, err := v.r.CredentialsFetcher().ResolveKey(ctx, r.KeyURLs, kid, "sig")
		if err != nil {
			return nil, err
		}

		// Mutate to public key
		if _, ok := key.Key.([]byte); !ok && !key.IsPublic() {
			k := key.Public()
			key = &k
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
			return nil, errors.WithStack(herodot.ErrBadRequest.WithReasonf(`This request object uses unsupported signing algorithm "%s".`, token.Header["alg"]))
		}

		return nil, errors.WithStack(herodot.ErrBadRequest.WithReasonf(`The signing key algorithm does not match the algorithm from the token header.`))
	}, jwt.WithIssuedAt())
	if err != nil {
		if errors.Is(err, jwt.ErrTokenUnverifiable) ||
			errors.Is(err, jwt.ErrTokenUnverifiable) ||
			errors.Is(err, jwt.ErrTokenSignatureInvalid) ||
			errors.Is(err, jwt.ErrTokenInvalidClaims) ||
			errors.Is(err, jwt.ErrTokenMalformed) {
			return nil, herodot.ErrInternalServerError.WithError(err.Error()).WithTrace(err)
		}
		return nil, err
	} else if !t.Valid {
		return nil, errors.WithStack(fosite.ErrInactiveToken)
	}

	claims, ok := t.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("Unable to type assert jwt claims to jwt.MapClaims."))
	}

	parsedClaims := jwtx.ParseMapStringInterfaceClaims(claims)
	for _, audience := range r.Audiences {
		if !stringslice.Has(parsedClaims.Audience, audience) {
			return nil, herodot.ErrUnauthorized.WithReasonf("Token audience %v is not intended for target audience %s.", parsedClaims.Audience, audience)
		}
	}

	if len(r.Issuers) > 0 {
		if !stringslice.Has(r.Issuers, parsedClaims.Issuer) {
			return nil, herodot.ErrUnauthorized.WithReasonf("Token issuer does not match any trusted issuer %s.", parsedClaims.Issuer).
				WithDetail("received issuers", strings.Join(r.Issuers, ", "))
		}
	}

	s, k := scope(claims)
	delete(claims, k)
	claims["scp"] = s

	if r.ScopeStrategy != nil {
		for _, sc := range r.Scope {
			if !r.ScopeStrategy(s, sc) {
				return nil, herodot.ErrUnauthorized.WithReasonf(`JSON Web Token is missing required scope "%s".`, sc)
			}
		}
	} else {
		if len(r.Scope) > 0 {
			return nil, errors.WithStack(helper.ErrRuleFeatureDisabled.WithReason("Scope validation was requested but scope strategy is set to \"none\"."))
		}
	}

	return t, nil
}

func scope(claims map[string]interface{}) ([]string, string) {
	var ok bool
	var interim interface{}
	var key string

	for _, k := range []string{"scp", "scope", "scopes"} {
		if interim, ok = claims[k]; ok {
			key = k
			break
		}
	}

	if !ok {
		return []string{}, key
	}

	switch i := interim.(type) {
	case []string:
		return i, key
	case []interface{}:
		vs := make([]string, len(i))
		for k, v := range i {
			if vv, ok := v.(string); ok {
				vs[k] = vv
			}
		}
		return vs, key
	case string:
		return stringsx.Splitx(i, " "), key
	default:
		return []string{}, key
	}
}
