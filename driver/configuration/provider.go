package configuration

import (
	"bytes"
	"encoding/json"
	"net/url"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/rs/cors"
	"golang.org/x/oauth2/clientcredentials"

	"github.com/ory/fosite"
)

type AccessRuleRepository struct {
	URL   *url.URL
	Watch bool
	TTL   time.Duration
}

func (r *AccessRuleRepository) UnmarshalJSON(raw []byte) error {
	var simple struct {
		URL   string `json:"url"`
		Watch bool   `json:"watch"`
		TTL   string `json:"ttl"`
	}
	d := json.NewDecoder(bytes.NewBuffer(raw))
	d.DisallowUnknownFields()
	if err := d.Decode(&simple); err != nil {
		return errors.WithStack(err)
	}

	u, err := url.ParseRequestURI(simple.URL)
	if err != nil {
		return errors.WithStack(err)
	}

	switch u.Scheme {
	case "http":
		fallthrough
	case "https":
		if len(simple.TTL) > 0 && !simple.Watch {
			return errors.Errorf("access rule repository sets ttl but watch is disabled: %s", u.String())
		}
		if len(simple.TTL) == 0 && simple.Watch {
			simple.TTL = "30s"
		}
	case "file":
		if len(simple.TTL) > 0 && !simple.Watch {
			return errors.Errorf("access rule repository sets ttl but watch is disabled: %s", u.String())
		}
	case "inline":
		if simple.Watch {
			return errors.Errorf("access rule repository url inline does not support enabling watch: %s", u.String())
		}
		if len(simple.TTL) > 0 {
			return errors.Errorf("access rule repository url inline does not support enabling ttl: %s", u.String())
		}
	default:
		return errors.Errorf("access rule repository uses invalid scheme: %s", u.String())
	}

	var ttl time.Duration
	if len(simple.TTL) > 0 {
		ttl, err = time.ParseDuration(simple.TTL)
		if err != nil {
			return errors.WithStack(err)
		}
	}

	r.TTL = ttl
	r.Watch = simple.Watch
	r.URL = u

	return nil
}

type Provider interface {
	CORSEnabled(iface string) bool
	CORSOptions(iface string) cors.Options

	ProviderAuthenticators
	ProviderAuthorizers
	ProviderMutators

	ProxyReadTimeout() time.Duration
	ProxyWriteTimeout() time.Duration
	ProxyIdleTimeout() time.Duration

	AccessRuleRepositories() []AccessRuleRepository

	ProxyServeAddress() string
	APIServeAddress() string
}

type ProviderAuthenticators interface {
	AuthenticatorAnonymousIsEnabled() bool
	AuthenticatorAnonymousIdentifier() string

	AuthenticatorNoopIsEnabled() bool

	AuthenticatorJWTIsEnabled() bool
	AuthenticatorJWTJWKSURIs() []url.URL
	AuthenticatorJWTScopeStrategy() fosite.ScopeStrategy

	AuthenticatorOAuth2ClientCredentialsIsEnabled() bool
	AuthenticatorOAuth2ClientCredentialsTokenURL() *url.URL

	AuthenticatorOAuth2TokenIntrospectionIsEnabled() bool
	AuthenticatorOAuth2TokenIntrospectionScopeStrategy() fosite.ScopeStrategy
	AuthenticatorOAuth2TokenIntrospectionIntrospectionURL() *url.URL
	AuthenticatorOAuth2TokenIntrospectionPreAuthorization() *clientcredentials.Config

	AuthenticatorUnauthorizedIsEnabled() bool
}

type ProviderAuthorizers interface {
	AuthorizerAllowIsEnabled() bool

	AuthorizerDenyIsEnabled() bool

	AuthorizerKetoEngineACPORYIsEnabled() bool
	AuthorizerKetoEngineACPORYBaseURL() *url.URL
}

type ProviderMutators interface {
	MutatorCookieIsEnabled() bool

	MutatorHeaderIsEnabled() bool

	MutatorIDTokenIsEnabled() bool
	MutatorIDTokenIssuerURL() *url.URL
	MutatorIDTokenJWKSURL() *url.URL
	MutatorIDTokenTTL() time.Duration

	MutatorNoopIsEnabled() bool
}

func MustValidate(l logrus.FieldLogger, p Provider) {
}
