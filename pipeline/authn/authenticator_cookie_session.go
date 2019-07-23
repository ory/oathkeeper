package authn

import (
	"bytes"
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

func (a *AuthenticatorCookieSession) Validate() error {
	if !a.c.AuthenticatorCookieSessionIsEnabled() {
		return errors.WithStack(ErrAuthenticatorNotEnabled.WithReasonf(`Authenticator "%s" is disabled per configuration.`, a.GetID()))
	}

	if a.c.AuthenticatorCookieSessionCheckSessionURL().String() == "" {
		return errors.WithStack(ErrAuthenticatorNotEnabled.WithReasonf(
			`Configuration for authenticator "%s" did not specify any values for configuration key "%s" and is thus disabled.`,
			a.GetID(),
			configuration.ViperKeyAuthenticatorCookieSessionCheckSessionURL,
		))
	}

	return nil
}

func (a *AuthenticatorCookieSession) Authenticate(r *http.Request, config json.RawMessage, _ pipeline.Rule) (*AuthenticationSession, error) {
	var cf AuthenticatorCookieSessionConfiguration
	if len(config) == 0 {
		config = []byte("{}")
	}
	d := json.NewDecoder(bytes.NewBuffer(config))
	d.DisallowUnknownFields()
	if err := d.Decode(&cf); err != nil {
		return nil, errors.WithStack(err)
	}

	only := cf.Only
	if len(only) == 0 {
		only = a.c.AuthenticatorCookieSessionOnly()
	}
	if !cookieSessionResponsible(r, only) {
		return nil, errors.WithStack(ErrAuthenticatorNotResponsible)
	}

	origin := cf.CheckSessionURL
	if origin == "" {
		origin = a.c.AuthenticatorCookieSessionCheckSessionURL().String()
	}

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
