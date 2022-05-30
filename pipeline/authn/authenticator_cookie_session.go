package authn

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/pkg/errors"
	"github.com/tidwall/gjson"

	"github.com/ory/go-convenience/stringsx"
	"github.com/ory/oathkeeper/x/header"

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
	Only                  []string          `json:"only"`
	CheckSessionURL       string            `json:"check_session_url"`
	PreserveQuery         bool              `json:"preserve_query"`
	PreservePath          bool              `json:"preserve_path"`
	ExtraFrom             string            `json:"extra_from"`
	SubjectFrom           string            `json:"subject_from"`
	PreserveHost          bool              `json:"preserve_host"`
	ForwardHTTPHeaders    []string          `json:"forward_http_headers"`
	SetHeaders            map[string]string `json:"additional_headers"`
	ForceMethod           string            `json:"force_method"`
	ForwardHTTPHeadersMap map[string]string `json:"-"`
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
		CheckSessionURL:    a.CheckSessionURL,
		PreserveQuery:      a.PreserveQuery,
		PreservePath:       a.PreservePath,
		PreserveHost:       a.PreserveHost,
		ForwardHTTPHeaders: a.ForwardHTTPHeadersMap,
		SetHeaders:         a.SetHeaders,
		ForceMethod:        a.ForceMethod,
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

	// Add Authorization and Cookie headers for backward compatibility
	c.ForwardHTTPHeaders = append(c.ForwardHTTPHeaders, []string{header.Authorization, header.Cookie}...)

	c.ForwardHTTPHeadersMap = make(map[string]string)
	for _, h := range c.ForwardHTTPHeaders {
		c.ForwardHTTPHeadersMap[h] = h
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
	req, err := PrepareRequest(r, cf)
	if err != nil {
		return nil, err
	}

	res, err := http.DefaultClient.Do(req.WithContext(r.Context()))
	if err != nil {
		return nil, helper.ErrForbidden.WithReason(err.Error()).WithTrace(err)
	}

	defer res.Body.Close()

	if res.StatusCode == http.StatusOK {
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return json.RawMessage{}, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("Unable to fetch cookie session context from remote: %+v", err))
		}
		return body, nil
	} else {
		return json.RawMessage{}, errors.WithStack(helper.ErrUnauthorized)
	}
}

func PrepareRequest(r *http.Request, cf *AuthenticatorForwardConfig) (http.Request, error) {
	reqURL, err := url.Parse(cf.CheckSessionURL)
	if err != nil {
		return http.Request{}, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("Unable to parse session check URL: %s", err))
	}

	if !cf.PreservePath {
		reqURL.Path = r.URL.Path
	}

	if !cf.PreserveQuery {
		reqURL.RawQuery = r.URL.RawQuery
	}

	if cf.ForceMethod == "" {
		cf.ForceMethod = r.Method
	}

	req := http.Request{
		Method: cf.ForceMethod,
		URL:    reqURL,
		Header: http.Header{},
	}

	// We need to copy only essential and configurable headers
	for requested, v := range r.Header {
		for _, allowed := range cf.ForwardHTTPHeaders {
			if requested == allowed {
				req.Header[k] = v
		}
	}

	for k, v := range cf.SetHeaders {
		req.Header.Set(k, v)
	}

	if cf.PreserveHost {
		req.Header.Set(header.XForwardedHost, r.Host)
	}
	return req, nil
}
