package configuration

import (
	"github.com/ory/fosite"
	"github.com/rs/cors"
	"net/url"
)

type Provider interface {
	ServesHTTPS() bool

	CORSEnabled(iface string) bool
	CORSOptions(iface string) cors.Options

	AuthenticatorAnonymousEnabled() bool
	AuthenticatorAnonymousIdentifier() string

	AuthenticatorNoopEnabled() bool

	AuthenticatorJWTEnabled() bool
	AuthenticatorJWTJWKSURIs() []*url.URL
	AuthenticatorJWTScopeStrategy() fosite.ScopeStrategy
}
