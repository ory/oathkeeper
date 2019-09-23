package authn

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/pkg/errors"

	"github.com/ory/oathkeeper/driver/configuration"
	"github.com/ory/oathkeeper/helper"
	"github.com/ory/oathkeeper/pipeline"
)

type AuthenticatorCookieSessionFilter struct {
}

type AuthenticatorCookieSessionConfiguration struct {
	Only            []string `json:"only"`
	CheckSessionURL string   `json:"check_session_url"`
}

type AuthenticatorCookieSession struct {
	c configuration.Provider
}

func NewAuthenticatorCookieSession(c configuration.Provider) *AuthenticatorCookieSession {
	return &AuthenticatorCookieSession{
		c: c,
	}
}

func (a *AuthenticatorCookieSession) GetID() string {
	return "cookie_session"
}

func (a *AuthenticatorCookieSession) Validate(config json.RawMessage) error {
	if !a.c.AuthenticatorIsEnabled(a.GetID()) {
		return NewErrAuthenticatorNotEnabled(a)
	}

	_, err := a.Config(config)
	return err
}

func (a *AuthenticatorCookieSession) Config(config json.RawMessage) (*AuthenticatorCookieSessionConfiguration, error) {
	var c AuthenticatorCookieSessionConfiguration
	if err := a.c.AuthenticatorConfig(a.GetID(), config, &c); err != nil {
		return nil, NewErrAuthenticatorMisconfigured(a, err)
	}

	return &c, nil
}

func (a *AuthenticatorCookieSession) Authenticate(r *http.Request, config json.RawMessage, _ pipeline.Rule) (*AuthenticationSession, error) {
	cf, err := a.Config(config)
	if err != nil {
		return nil, err
	}

	if !cookieSessionResponsible(r, cf.Only) {
		return nil, errors.WithStack(ErrAuthenticatorNotResponsible)
	}

	origin := cf.CheckSessionURL
	body, err := forwardRequestToSessionStore(r, origin)
	if err != nil {
		return nil, helper.ErrForbidden.WithReason(err.Error()).WithTrace(err)
	}

	var session struct {
		Subject string                 `json:"subject"`
		Extra   map[string]interface{} `json:"extra"`
	}
	err = json.Unmarshal(body, &session)
	if err != nil {
		return nil, helper.ErrForbidden.WithReason(err.Error()).WithTrace(err)
	}

	return &AuthenticationSession{
		Subject: session.Subject,
		Extra:   session.Extra,
	}, nil
}

func cookieSessionResponsible(r *http.Request, only []string) bool {
	if len(only) == 0 {
		return true
	}
	for _, cookieName := range only {
		if _, err := r.Cookie(cookieName); err == nil {
			return true
		}
	}
	return false
}

func forwardRequestToSessionStore(r *http.Request, checkSessionURL string) (json.RawMessage, error) {
	reqUrl, err := url.Parse(checkSessionURL)
	if err != nil {
		return nil, helper.ErrForbidden.WithReason(err.Error()).WithTrace(err)
	}
	reqUrl.Path = r.URL.Path

	res, err := http.DefaultClient.Do(&http.Request{
		Method: r.Method,
		URL:    reqUrl,
		Header: r.Header,
	})
	if err != nil {
		return nil, helper.ErrForbidden.WithReason(err.Error()).WithTrace(err)
	}

	if res.StatusCode == 200 {
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return json.RawMessage{}, err
		}
		return json.RawMessage(body), nil
	} else {
		return json.RawMessage{}, errors.WithStack(helper.ErrUnauthorized)
	}
}
