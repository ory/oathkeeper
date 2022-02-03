package errors

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/ory/oathkeeper/driver/configuration"
	"github.com/ory/oathkeeper/pipeline"
	"github.com/ory/oathkeeper/x"
)

var _ Handler = new(ErrorRedirect)

const (
	xForwardedProto = "X-Forwarded-Proto"
	xForwardedHost  = "X-Forwarded-Host"
	xForwardedUri   = "X-Forwarded-Uri"
)

type (
	ErrorRedirectConfig struct {
		To                 string `json:"to"`
		Code               int    `json:"code"`
		ReturnToQueryParam string `json:"return_to_query_param"`
	}
	ErrorRedirect struct {
		c configuration.Provider
		d ErrorRedirectDependencies
	}
	ErrorRedirectDependencies interface {
		x.RegistryWriter
	}
)

func NewErrorRedirect(
	c configuration.Provider,
	d ErrorRedirectDependencies,
) *ErrorRedirect {
	return &ErrorRedirect{c: c, d: d}
}

func (a *ErrorRedirect) Handle(w http.ResponseWriter, r *http.Request, config json.RawMessage, _ pipeline.Rule, _ error) error {
	c, err := a.Config(config)
	if err != nil {
		return err
	}

	var scheme, host, requestUri string
	if scheme = r.Header.Get(xForwardedProto); scheme == "" {
		scheme = r.URL.Scheme
	}
	if host = r.Header.Get(xForwardedHost); host == "" {
		host = r.URL.Host
	}
	if requestUri = r.Header.Get(xForwardedUri); requestUri == "" {
		requestUri = r.URL.RequestURI()
	}

	var uri *url.URL
	if scheme == "" || host == "" {
		// FIXME: I don't think this is applicable for real requests. It is however used by tests.
		uri, err = url.Parse(fmt.Sprintf("%s", requestUri))
	} else {
		uri, err = url.Parse(fmt.Sprintf("%s://%s%s", scheme, host, requestUri))
	}
	if err != nil {
		return err
	}

	http.Redirect(w, r, a.RedirectURL(uri, c), c.Code)
	return nil
}

func (a *ErrorRedirect) Validate(config json.RawMessage) error {
	if !a.c.ErrorHandlerIsEnabled(a.GetID()) {
		return NewErrErrorHandlerNotEnabled(a)
	}
	_, err := a.Config(config)
	return err
}

func (a *ErrorRedirect) Config(config json.RawMessage) (*ErrorRedirectConfig, error) {
	var c ErrorRedirectConfig
	if err := a.c.ErrorHandlerConfig(a.GetID(), config, &c); err != nil {
		return nil, NewErrErrorHandlerMisconfigured(a, err)
	}

	if c.Code < 301 || c.Code > 302 {
		c.Code = http.StatusFound
	}

	return &c, nil
}

func (a *ErrorRedirect) GetID() string {
	return "redirect"
}

func (a *ErrorRedirect) RedirectURL(uri *url.URL, c *ErrorRedirectConfig) string {
	if c.ReturnToQueryParam == "" {
		return c.To
	}

	u, err := url.Parse(c.To)
	if err != nil {
		return c.To
	}

	q := u.Query()
	q.Set(c.ReturnToQueryParam, uri.String())
	u.RawQuery = q.Encode()
	return u.String()
}
