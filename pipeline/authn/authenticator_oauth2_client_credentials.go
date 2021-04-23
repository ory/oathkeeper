package authn

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/oauth2"

	"github.com/ory/x/httpx"

	"github.com/ory/oathkeeper/driver/configuration"

	"github.com/ory/oathkeeper/pipeline"

	"github.com/pkg/errors"
	"golang.org/x/oauth2/clientcredentials"

	"github.com/ory/oathkeeper/helper"
)

type AuthenticatorOAuth2Configuration struct {
	Scopes   []string `json:"required_scope"`
	TokenURL string   `json:"token_url"`
	Retry    *AuthenticatorOAuth2ClientCredentialsRetryConfiguration
}

type AuthenticatorOAuth2ClientCredentials struct {
	c      configuration.Provider
	client *http.Client
}

type AuthenticatorOAuth2ClientCredentialsRetryConfiguration struct {
	Timeout string `json:"max_delay"`
	MaxWait string `json:"give_up_after"`
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
	const (
		defaultTimeout = "1s"
		defaultMaxWait = "2s"
	)
	var c AuthenticatorOAuth2Configuration
	if err := a.c.AuthenticatorConfig(a.GetID(), config, &c); err != nil {
		return nil, NewErrAuthenticatorMisconfigured(a, err)
	}

	if c.Retry == nil {
		c.Retry = &AuthenticatorOAuth2ClientCredentialsRetryConfiguration{Timeout: defaultTimeout, MaxWait: defaultMaxWait}
	} else {
		if c.Retry.Timeout == "" {
			c.Retry.Timeout = defaultTimeout
		}
		if c.Retry.MaxWait == "" {
			c.Retry.MaxWait = defaultMaxWait
		}
	}
	duration, err := time.ParseDuration(c.Retry.Timeout)
	if err != nil {
		return nil, err
	}

	maxWait, err := time.ParseDuration(c.Retry.MaxWait)
	if err != nil {
		return nil, err
	}
	timeout := time.Millisecond * duration
	a.client = httpx.NewResilientClientLatencyToleranceConfigurable(nil, timeout, maxWait)

	return &c, nil
}

func (a *AuthenticatorOAuth2ClientCredentials) Authenticate(r *http.Request, session *AuthenticationSession, config json.RawMessage, _ pipeline.Rule) error {
	cf, err := a.Config(config)
	if err != nil {
		return err
	}

	user, password, ok := r.BasicAuth()
	if !ok {
		return errors.WithStack(ErrAuthenticatorNotResponsible)
	}

	user, err = url.QueryUnescape(user)
	if err != nil {
		return errors.Wrapf(helper.ErrUnauthorized, err.Error())
	}

	password, err = url.QueryUnescape(password)
	if err != nil {
		return errors.Wrapf(helper.ErrUnauthorized, err.Error())
	}

	c := &clientcredentials.Config{
		ClientID:     user,
		ClientSecret: password,
		Scopes:       cf.Scopes,
		TokenURL:     cf.TokenURL,
		AuthStyle:    oauth2.AuthStyleInHeader,
	}

	token, err := c.Token(context.WithValue(
		r.Context(),
		oauth2.HTTPClient,
		c.Client,
	))

	if err != nil {
		if rErr, ok := err.(*oauth2.RetrieveError); ok {
			switch httpStatusCode := rErr.Response.StatusCode; httpStatusCode {
			case http.StatusServiceUnavailable:
				return errors.Wrapf(helper.ErrUpstreamServiceNotAvailable, err.Error())
			case http.StatusInternalServerError:
				return errors.Wrapf(helper.ErrUpstreamServiceInternalServerError, err.Error())
			case http.StatusGatewayTimeout:
				return errors.Wrapf(helper.ErrUpstreamServiceTimeout, err.Error())
			case http.StatusNotFound:
				return errors.Wrapf(helper.ErrUpstreamServiceNotFound, err.Error())
			default:
				return errors.Wrapf(helper.ErrUnauthorized, err.Error())
			}
		} else {
			return errors.Wrapf(helper.ErrUpstreamServiceNotAvailable, err.Error())
		}
	}

	if token.AccessToken == "" {
		return errors.WithStack(helper.ErrUnauthorized)
	}

	session.Subject = user
	return nil
}
