package credentials

import (
	"context"
	"crypto/ecdsa"
	"crypto/rsa"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/ory/fosite"
	"github.com/ory/herodot"
	"github.com/ory/oathkeeper/helper"
	"github.com/ory/x/jwtx"
	"github.com/ory/x/mapx"
	"github.com/ory/x/stringslice"
	"github.com/ory/x/stringsx"
	"github.com/pkg/errors"
	"net/url"
	"strings"
)

type VerifierDefault struct {
	r FetchRegistry
}

func NewVerifierDefault(f FetchRegistry) *VerifierDefault {
	return &VerifierDefault{
		r: f,
	}
}

func (v *VerifierDefault) Verify(
	ctx context.Context,
	token string,
	locations []url.URL,
	algorithms []string,
	issuers []string,
	audiences []string,
	scopes []string,
	ss fosite.ScopeStrategy,
) (*jwt.Token, error) {
	// Parse the token.
	t, err := jwt.ParseWithClaims(token, jwt.MapClaims{}, func(token *jwt.Token) (interface{}, error) {
		if !stringslice.Has(algorithms, fmt.Sprintf("%s", token.Header["alg"])) {
			return nil, errors.WithStack(herodot.ErrInternalServerError.WithReason(fmt.Sprintf(`JSON Web Token used signing method "%s" which is not allowed.`, token.Header["alg"])))
		}

		kid, ok := token.Header["kid"].(string)
		if !ok || kid == "" {
			return nil, errors.WithStack(herodot.ErrInternalServerError.WithReason("The JSON Web Token must contain a kid header value but did not."))
		}

		key, err := v.r.CredentialsFetcher().ResolveKey(ctx, locations, kid, "sig")
		if err != nil {
			return nil, herodot.ErrInternalServerError.WithTrace(err).WithDebugf("%s", err)
		}

		// Transform to public key
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
			return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf(`This request object uses unsupported signing algorithm "%s".`, token.Header["alg"]))
		}

		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf(`The signing key algorithm does not match the algorithm from the token header.`))
	})

	if err != nil {
		return nil, err
	} else if !t.Valid {
		return nil, errors.WithStack(fosite.ErrInactiveToken)
	}

	claims, ok := t.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.Errorf("unable to type assert jwt claims to jwt.MapClaims")
	}

	parsedClaims := jwtx.ParseMapStringInterfaceClaims(claims)
	for _, audience := range audiences {
		if !stringslice.Has(parsedClaims.Audience, audience) {
			return nil, errors.WithStack(herodot.ErrInternalServerError.WithReason(fmt.Sprintf("Token audience %v is not intended for target audience %s.", parsedClaims.Audience, audience)))
		}
	}

	if len(issuers) > 0 {
		if !stringslice.Has(issuers, parsedClaims.Issuer) {
			return nil, errors.WithStack(herodot.ErrInternalServerError.WithReason(fmt.Sprintf("Token issuer does not match any trusted issuer.")))
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

	if ss != nil {
		for _, sc := range scopes {
			if !ss(scope(claims), sc) {
				return nil, errors.WithStack(herodot.ErrInternalServerError.WithReason(fmt.Sprintf(`JSON Web Token is missing required scope "%s".`, sc)))
			}
		}
	} else {
		if len(scopes) > 0 {
			return nil, errors.WithStack(helper.ErrRuleFeatureDisabled.WithReason("Scope validation was requested but scope strategy is set to \"none\"."))
		}
	}
	
	return t, nil
}

func scope(claims map[string]interface{}) []string {
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
