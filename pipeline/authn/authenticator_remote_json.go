package authn

import (
	"bytes"
	"encoding/json"
	"io"
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

type AuthenticatorRemoteJSONFilter struct {
}

type AuthenticatorRemoteJSONConfiguration struct {
	ServiceURL   string `json:"service_url"`
	PreservePath bool   `json:"preserve_path"`
	ExtraFrom    string `json:"extra_from"`
	SubjectFrom  string `json:"subject_from"`
	Method       string `json:"method"`
}

type AuthenticatorRemoteJSON struct {
	c configuration.Provider
}

func NewAuthenticatorRemoteJSON(c configuration.Provider) *AuthenticatorRemoteJSON {
	return &AuthenticatorRemoteJSON{
		c: c,
	}
}

func (a *AuthenticatorRemoteJSON) GetID() string {
	return "remote_json"
}

func (a *AuthenticatorRemoteJSON) Validate(config json.RawMessage) error {
	if !a.c.AuthenticatorIsEnabled(a.GetID()) {
		return NewErrAuthenticatorNotEnabled(a)
	}

	_, err := a.Config(config)
	return err
}

func (a *AuthenticatorRemoteJSON) Config(config json.RawMessage) (*AuthenticatorRemoteJSONConfiguration, error) {
	var c AuthenticatorRemoteJSONConfiguration
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

func (a *AuthenticatorRemoteJSON) Authenticate(r *http.Request, session *AuthenticationSession, config json.RawMessage, _ pipeline.Rule) error {
	cf, err := a.Config(config)
	if err != nil {
		return err
	}

	method := forwardMethod(r, cf.Method)

	body, err := forwardRequestToAuthenticator(r, method, cf.ServiceURL, cf.PreservePath)
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

func forwardMethod(r *http.Request, method string) string {
	if len(method) == 0 {
		return r.Method
	}
	return method
}

func forwardRequestToAuthenticator(r *http.Request, method string, serviceURL string, preservePath bool) (json.RawMessage, error) {
	reqUrl, err := url.Parse(serviceURL)
	if err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("Unable to parse remote URL: %s", err))
	}

	if !preservePath {
		reqUrl.Path = r.URL.Path
	}

	var forwardRequestBody io.ReadCloser = nil
	if r.Body != nil {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return nil, helper.ErrBadRequest.WithReason(err.Error()).WithTrace(err)
		}

		// Unfortunately the body reader needs to be read once to forward the request,
		// thus the upstream request will fail miserably without recreating a fresh ReaderCloser
		forwardRequestBody = ioutil.NopCloser(bytes.NewReader(body))
		r.Body = ioutil.NopCloser(bytes.NewReader(body))
	}

	req := http.Request{
		Method: method,
		URL:    reqUrl,
		Header: r.Header,
		Body:   forwardRequestBody,
	}
	res, err := http.DefaultClient.Do(req.WithContext(r.Context()))
	if err != nil {
		return nil, helper.ErrForbidden.WithReason(err.Error()).WithTrace(err)
	}

	return handleResponse(res)
}

func handleResponse(r *http.Response) (json.RawMessage, error) {
	if r.StatusCode == http.StatusOK {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return json.RawMessage{}, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("Remote server returned error: %+v", err))
		}
		return body, nil
	} else {
		return json.RawMessage{}, errors.WithStack(helper.ErrUnauthorized)
	}
}
