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

	"github.com/ory/x/logrusx"
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
	SetHeaders      map[string]string `json:"additional_headers"`
	ForceMethod     string            `json:"force_method"`
}

type AuthenticatorCookieSession struct {
	c      configuration.Provider
	logger *logrusx.Logger
}

func NewAuthenticatorCookieSession(c configuration.Provider, logger *logrusx.Logger) *AuthenticatorCookieSession {
	return &AuthenticatorCookieSession{
		c:      c,
		logger: logger,
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

	body, err := forwardRequestToSessionStore(r, cf.CheckSessionURL, cf.PreserveQuery, cf.PreservePath, cf.PreserveHost, cf.SetHeaders, cf.ForceMethod, a.logger)
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

func forwardRequestToSessionStore(r *http.Request, checkSessionURL string, preserveQuery bool, preservePath bool, preserveHost bool, setHeaders map[string]string, m string, logger *logrusx.Logger) (json.RawMessage, error) {
	reqUrl, err := url.Parse(checkSessionURL)
	if err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("Unable to parse session check URL: %s", err))
	}

	if !preservePath {
		reqUrl.Path = r.URL.Path
	}

	if !preserveQuery {
		reqUrl.RawQuery = r.URL.RawQuery
	}

	if m == "" {
		m = r.Method
	}

	req := http.Request{
		Method: m,
		URL:    reqUrl,
		Header: http.Header{},
	}

	// We need to make a COPY of the header, not modify r.Header!
	for k, v := range r.Header {
		// remove Accept-Encoding to let the transport handle gzip
		if k == "Accept-Encoding" {
			continue
		}
		req.Header[k] = v
	}

	for k, v := range setHeaders {
		req.Header.Set(k, v)
	}

	if preserveHost {
		req.Header.Set("X-Forwarded-Host", r.Host)
	}

	res, err := http.DefaultClient.Do(req.WithContext(r.Context()))
	if err != nil {
		return nil, helper.ErrForbidden.WithReason(err.Error()).WithTrace(err)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		logger.Tracef("Error reading response from remote: %v", err)
		return json.RawMessage{}, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("Unable to read response from remote: %s", err))
	}
	defer res.Body.Close()

	if res.StatusCode == 200 {
		return body, nil
	}

	logger.WithField("response_code", res.StatusCode).WithField("body", string(body)).Trace()

	return json.RawMessage{}, errors.WithStack(helper.ErrUnauthorized.WithReasonf("Remote returned non 200 status code: %d", res.StatusCode))
}
