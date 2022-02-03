package errors

import (
	"encoding/json"
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

	http.Redirect(w, r, a.RedirectURL(a.extractRedirectURL(r), c), c.Code)
	return nil
}

func (a *ErrorRedirect) extractRedirectURL(req *http.Request) *url.URL {
	var scheme, host, path string
	if scheme = req.Header.Get(xForwardedProto); scheme == "" {
		scheme = req.URL.Scheme
	}
	if host = req.Header.Get(xForwardedHost); host == "" {
		host = req.URL.Host
	}
	if path = req.Header.Get(xForwardedUri); path == "" {
		path = req.URL.Path
	}

	u := *req.URL
	u.Scheme = scheme
	u.Host = host
	u.Path = path

	return &u
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
