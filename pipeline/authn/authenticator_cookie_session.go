package authn

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/pkg/errors"
	"github.com/tidwall/gjson"

	"github.com/ory/go-convenience/stringsx"

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
	Only            []string          `json:"only"`
	CheckSessionURL string            `json:"check_session_url"`
	PreserveQuery   bool              `json:"preserve_query"`
	PreservePath    bool              `json:"preserve_path"`
	ExtraFrom       string            `json:"extra_from"`
	SubjectFrom     string            `json:"subject_from"`
	PreserveHost    bool              `json:"preserve_host"`
	ProxyHeaders    []string          `json:"forward_http_headers"`
	SetHeaders      map[string]string `json:"additional_headers"`
	ForceMethod     string            `json:"force_method"`
	ProxyHeadersMap map[string]string `json:"-"`
}

type AuthenticatorCookieSession struct {
	c configuration.Provider
}

func NewAuthenticatorCookieSession(c configuration.Provider) *AuthenticatorCookieSession {
	return &AuthenticatorCookieSession{
		c: c,
	}
}

func (a *AuthenticatorCookieSessionConfiguration) ToAuthenticatorForwardConfig() *AuthenticatorForwardConfig {
	return &AuthenticatorForwardConfig{
		CheckSessionURL: a.CheckSessionURL,
		PreserveQuery:   a.PreserveQuery,
		PreservePath:    a.PreservePath,
		PreserveHost:    a.PreserveHost,
		ProxyHeadersMap: a.ProxyHeadersMap,
		SetHeaders:      a.SetHeaders,
		ForceMethod:     a.ForceMethod,
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

	if len(c.SubjectFrom) == 0 {
		c.SubjectFrom = "subject"
	}
	if len(c.ProxyHeaders) == 0 {
		c.ProxyHeaders = []string{"Authorization", "Cookie"}
	}
	for _, h := range c.ProxyHeaders {
		c.ProxyHeadersMap[h] = h
	}

	return &c, nil
}

func (a *AuthenticatorCookieSession) Authenticate(r *http.Request, session *AuthenticationSession, config json.RawMessage, _ pipeline.Rule) error {
	cf, err := a.Config(config)
	if err != nil {
		return err
	}

	if !cookieSessionResponsible(r, cf.Only) {
		return errors.WithStack(ErrAuthenticatorNotResponsible)
	}

	body, err := forwardRequestToSessionStore(r, cf.ToAuthenticatorForwardConfig())
	if err != nil {
		return err
	}

	var (
		subject string
		extra   map[string]interface{}

		subjectRaw = []byte(stringsx.Coalesce(gjson.GetBytes(body, cf.SubjectFrom).Raw, "null"))
		extraRaw   = []byte(stringsx.Coalesce(gjson.GetBytes(body, cf.ExtraFrom).Raw, "null"))
	)

	if err = json.Unmarshal(subjectRaw, &subject); err != nil {
		return helper.ErrForbidden.WithReasonf("The configured subject_from GJSON path returned an error on JSON output: %s", err.Error()).WithDebugf("GJSON path: %s\nBody: %s\nResult: %s", cf.SubjectFrom, body, subjectRaw).WithTrace(err)
	}

	if err = json.Unmarshal(extraRaw, &extra); err != nil {
		return helper.ErrForbidden.WithReasonf("The configured extra_from GJSON path returned an error on JSON output: %s", err.Error()).WithDebugf("GJSON path: %s\nBody: %s\nResult: %s", cf.ExtraFrom, body, extraRaw).WithTrace(err)
	}

	session.Subject = subject
	session.Extra = extra
	return nil
}

func cookieSessionResponsible(r *http.Request, only []string) bool {
	if len(only) == 0 && len(r.Cookies()) > 0 {
		return true
	}

	for _, cookieName := range only {
		if _, err := r.Cookie(cookieName); err == nil {
			return true
		}
	}

	return false
}

func forwardRequestToSessionStore(r *http.Request, cf *AuthenticatorForwardConfig) (json.RawMessage, error) {
	reqUrl, err := url.Parse(cf.CheckSessionURL)
	if err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("Unable to parse session check URL: %s", err))
	}

	if !cf.PreservePath {
		reqUrl.Path = r.URL.Path
	}

	if !cf.PreserveQuery {
		reqUrl.RawQuery = r.URL.RawQuery
	}

	if cf.ForceMethod == "" {
		cf.ForceMethod = r.Method
	}

	req := http.Request{
		Method: cf.ForceMethod,
		URL:    reqUrl,
		Header: http.Header{},
	}

	// We need to copy only essential and configurable headers
	for k, v := range r.Header {
		if _, ok := cf.ProxyHeadersMap[k]; ok {
			req.Header[k] = v
		}
	}

	for k, v := range cf.SetHeaders {
		req.Header.Set(k, v)
	}

	if cf.PreserveHost {
		req.Header.Set("X-Forwarded-Host", r.Host)
	}

	res, err := http.DefaultClient.Do(req.WithContext(r.Context()))
	if err != nil {
		return nil, helper.ErrForbidden.WithReason(err.Error()).WithTrace(err)
	}

	defer res.Body.Close()

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
