package proxy

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"net/url"

	"github.com/ory/oathkeeper/helper"
	"github.com/ory/oathkeeper/rule"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2/clientcredentials"
)

type AuthenticatorOAuth2Configuration struct {
	// Scopes is an array of OAuth 2.0 scopes that are required when accessing an endpoint protected by this rule.
	// If the token used in the Authorization header did not request that specific scope, the request is denied.
	Scopes []string `json:"required_scope"`
}

type AuthenticatorOAuth2ClientCredentials struct {
	tokenURL string
}

func NewAuthenticatorOAuth2ClientCredentials(tokenURL string) *AuthenticatorOAuth2ClientCredentials {
	return &AuthenticatorOAuth2ClientCredentials{
		tokenURL: tokenURL,
	}
}

func (a *AuthenticatorOAuth2ClientCredentials) GetID() string {
	return "oauth2_client_credentials"
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

	logrus.New().Printf("Got wow user pw, %s, %s", user, password)
	c := &clientcredentials.Config{
		ClientID:     user,
		ClientSecret: password,
		Scopes:       cf.Scopes,
		TokenURL:     a.tokenURL,
	}

	token, err := c.Token(context.Background())
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
