package credential

import (
	"context"
	"github.com/dgrijalva/jwt-go"
	"net/url"
)

type Signer interface {
	Sign(ctx context.Context, location *url.URL, claims jwt.Claims) (string, error)
}

type SignerRegistry interface {
	CredentialsSigner() Signer
}
