package proxy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/ory/oathkeeper/driver/configuration"
	"net/http"
	"text/template"

	"github.com/pkg/errors"

	"github.com/ory/oathkeeper/rule"
)

type CredentialsCookiesConfig struct {
	Cookies map[string]string `json:"cookies"`
}

type TransformerCookie struct {
	templates *template.Template
	c configuration.Provider
}

func NewTransformerCookies(c configuration.Provider) *TransformerCookie {
	return &TransformerCookie{
		c: c,
		templates: newTemplate("cookie"),
	}
}

func (a *TransformerCookie) GetID() string {
	return "cookie"
}

func (a *TransformerCookie) Transform(r *http.Request, session *AuthenticationSession, config json.RawMessage, rl *rule.Rule) (http.Header, error) {
	if len(config) == 0 {
		config = []byte("{}")
	}

	// Cache request cookies
	requestCookies := r.Cookies()

	req := http.Request{Header: map[string][]string{}}

	// Keep track of rule cookies in a map
	cookies := map[string]bool{}

	var cfg CredentialsCookiesConfig
	d := json.NewDecoder(bytes.NewBuffer(config))
	d.DisallowUnknownFields()
	if err := d.Decode(&cfg); err != nil {
		return nil, errors.WithStack(err)
	}

	for cookie, templateString := range cfg.Cookies {
		var tmpl *template.Template
		var err error

		templateId := fmt.Sprintf("%s:%s", rl.ID, cookie)
		tmpl = a.templates.Lookup(templateId)
		if tmpl == nil {
			tmpl, err = a.templates.New(templateId).Parse(templateString)
			if err != nil {
				return nil, errors.Wrapf(err, `error parsing cookie template "%s" in rule "%s"`, templateString, rl.ID)
			}
		}

		cookieValue := bytes.Buffer{}
		err = tmpl.Execute(&cookieValue, session)
		if err != nil {
			return nil, errors.Wrapf(err, `error executing cookie template "%s" in rule "%s"`, templateString, rl.ID)
		}

		req.AddCookie(&http.Cookie{
			Name:  cookie,
			Value: cookieValue.String(),
		})

		cookies[cookie] = true
	}

	// Re-add previously set cookies that do not coincide with rule cookies
	for _, cookie := range requestCookies {
		// Test if cookie is handled by rule
		if _, ok := cookies[cookie.Name]; !ok {
			// Re-add cookie if not handled by rule
			req.AddCookie(cookie)
		}
	}

	return req.Header, nil
}

func (a *TransformerCookie) Validate() error {
	if !a.c.TransformerCookieIsEnabled() {
		return errors.WithStack(ErrAuthenticatorNotEnabled.WithReasonf("Transformer % is disabled per configuration.", a.GetID()))
	}

	return nil
}
