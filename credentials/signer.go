package credentials

import (
	"context"
	"net/url"

	"github.com/form3tech-oss/jwt-go"
)

type Signer interface {
	Sign(ctx context.Context, location *url.URL, claims jwt.Claims) (string, error)
}

type SignerRegistry interface {
	CredentialsSigner() Signer
}
