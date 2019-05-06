package proxy

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/ory/oathkeeper/driver/configuration"
	"github.com/ory/x/httpx"
	"golang.org/x/oauth2"
	"net/http"
	"net/url"

	"github.com/pkg/errors"
	"golang.org/x/oauth2/clientcredentials"

	"github.com/ory/oathkeeper/helper"
	"github.com/ory/oathkeeper/rule"
)

type AuthenticatorOAuth2Configuration struct {
	// Scopes is an array of OAuth 2.0 scopes that are required when accessing an endpoint protected by this rule.
	// If the token used in the Authorization header did not request that specific scope, the request is denied.
	Scopes []string `json:"required_scope"`
}

type AuthenticatorOAuth2ClientCredentials struct {
	c configuration.Provider
}

func NewAuthenticatorOAuth2ClientCredentials(c configuration.Provider) *AuthenticatorOAuth2ClientCredentials {
	return &AuthenticatorOAuth2ClientCredentials{c: c}
}

func (a *AuthenticatorOAuth2ClientCredentials) GetID() string {
	return "oauth2_client_credentials"
}

func (a *AuthenticatorOAuth2ClientCredentials) Validate() error {
	if !a.c.AuthenticatorOAuth2ClientCredentialsIsEnabled() {
		return errors.WithStack(ErrAuthenticatorNotEnabled.WithReasonf("Authenticator % is disabled per configuration.", a.GetID()))
	}

	if a.c.AuthenticatorOAuth2ClientCredentialsTokenURL() == nil {
		return errors.WithStack(ErrAuthenticatorNotEnabled.WithReasonf(`Configuration for authenticator % did not specify any values for configuration key "%s" and is thus disabled.`, a.GetID(), configuration.ViperKeyAuthenticatorClientCredentialsTokenURL))
	}

	return nil
}

func (a *AuthenticatorOAuth2ClientCredentials) Authenticate(r *http.Request, config json.RawMessage, rl *rule.Rule) (*AuthenticationSession, error) {
	var cf AuthenticatorOAuth2Configuration
	if len(config) == 0 {
		config = []byte("{}")
	}

	d := json.NewDecoder(bytes.NewBuffer(config))
	d.DisallowUnknownFields()
	if err := d.Decode(&cf); err != nil {
		return nil, errors.WithStack(err)
	}

	user, password, ok := r.BasicAuth()
	if !ok {
		return nil, errors.WithStack(ErrAuthenticatorNotResponsible)
	}

	var err error
	user, err = url.QueryUnescape(user)
	if !ok {
		return nil, errors.Wrapf(helper.ErrUnauthorized, err.Error())
	}

	password, err = url.QueryUnescape(password)
	if !ok {
		return nil, errors.Wrapf(helper.ErrUnauthorized, err.Error())
	}

	c := &clientcredentials.Config{
		ClientID:     user,
		ClientSecret: password,
		Scopes:       cf.Scopes,
		TokenURL:     a.c.AuthenticatorOAuth2ClientCredentialsTokenURL().String(),
	}

	token, err := c.Token(context.WithValue(
		context.Background(),
		oauth2.HTTPClient,
		httpx.NewResilientClientLatencyToleranceMedium(nil),
	))
	if err != nil {
		return nil, errors.Wrapf(helper.ErrUnauthorized, err.Error())
	}

	if token.AccessToken == "" {
		return nil, errors.WithStack(helper.ErrUnauthorized)
	}

	return &AuthenticationSession{
		Subject: user,
	}, nil
}
