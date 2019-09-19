package authn

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/url"

	"golang.org/x/oauth2"

	"github.com/ory/oathkeeper/driver/configuration"
	"github.com/ory/oathkeeper/pipeline"
	"github.com/ory/x/httpx"

	"github.com/pkg/errors"
	"golang.org/x/oauth2/clientcredentials"

	"github.com/ory/oathkeeper/helper"
)

type AuthenticatorOAuth2Configuration struct {
	Scopes []string `json:"required_scope"`
	TokenURL string `json:"token_url"`
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

func (a *AuthenticatorOAuth2ClientCredentials) Validate(config json.RawMessage) error {
	if !a.c.AuthenticatorIsEnabled(a.GetID()) {
		return NewErrAuthenticatorNotEnabled(a)
	}

	_, err := a.Config(config)
	return err
}

func (a *AuthenticatorOAuth2ClientCredentials) Config(config json.RawMessage) (*AuthenticatorOAuth2Configuration, error) {
	var c AuthenticatorOAuth2Configuration
	if err := a.c.AuthenticatorConfig(a.GetID(), config, &c); err != nil {
		return nil, NewErrAuthenticatorMisconfigured(a, err)
	}

	return &c, nil
}

func (a *AuthenticatorOAuth2ClientCredentials) Authenticate(r *http.Request, config json.RawMessage, _ pipeline.Rule) (*AuthenticationSession, error) {
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
	if err != nil {
		return nil, errors.Wrapf(helper.ErrUnauthorized, err.Error())
	}

	password, err = url.QueryUnescape(password)
	if err != nil {
		return nil, errors.Wrapf(helper.ErrUnauthorized, err.Error())
	}

	c := &clientcredentials.Config{
		ClientID:     user,
		ClientSecret: password,
		Scopes:       cf.Scopes,
		TokenURL:     cf.TokenURL,
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
