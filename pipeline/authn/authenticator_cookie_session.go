package authn

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/pkg/errors"
	"github.com/tidwall/gjson"

	"github.com/ory/herodot"

	"github.com/ory/oathkeeper/driver/configuration"
	"github.com/ory/oathkeeper/helper"
	"github.com/ory/oathkeeper/pipeline"
)

func init() {
	gjson.AddModifier("this", func(json, arg string) string {
		return json
	})
}

type AuthenticatorCookieSessionFilter struct {
}

type AuthenticatorCookieSessionConfiguration struct {
	Only            []string `json:"only"`
	CheckSessionURL string   `json:"check_session_url"`
	PreservePath    bool     `json:"preserve_path"`
	ExtraFrom       string   `json:"extra_from"`
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

	if len(c.ExtraFrom) == 0 {
		c.ExtraFrom = "extra"
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
	preservePath := cf.PreservePath
	body, err := forwardRequestToSessionStore(r, origin, preservePath)
	if err != nil {
		return nil, err
	}

	var session struct {
		Subject string `json:"subject"`
	}
	if err = json.Unmarshal(body, &session); err != nil {
		return nil, helper.ErrForbidden.WithReason(err.Error()).WithTrace(err)
	}

	extra := map[string]interface{}{}
	rawExtra := gjson.GetBytes(body, cf.ExtraFrom).Raw
	if rawExtra == "" {
		rawExtra = "null"
	}

	if err = json.Unmarshal([]byte(rawExtra), &extra); err != nil {
		return nil, helper.ErrForbidden.WithReasonf("The configured GJSON path returned an error on JSON output: %s", err.Error()).WithDebugf("GJSON path: %s\nBody: %s\nResult: %s", cf.ExtraFrom, body, rawExtra).WithTrace(err)
	}

	return &AuthenticationSession{
		Subject: session.Subject,
		Extra:   extra,
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

func forwardRequestToSessionStore(r *http.Request, checkSessionURL string, preservePath bool) (json.RawMessage, error) {
	reqUrl, err := url.Parse(checkSessionURL)
	if err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("Unable to parse session check URL: %s", err))
	}

	if !preservePath {
		reqUrl.Path = r.URL.Path
	}

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
			return json.RawMessage{}, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("Unable to fetch cookie session context from remote: %+v", err))
		}
		return body, nil
	} else {
		return json.RawMessage{}, errors.WithStack(helper.ErrUnauthorized)
	}
}
