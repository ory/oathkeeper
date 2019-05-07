package mutate

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"text/template"

	"github.com/ory/oathkeeper/driver/configuration"
	"github.com/ory/oathkeeper/pipeline"
	"github.com/ory/oathkeeper/pipeline/authn"

	"github.com/pkg/errors"
)

type CredentialsCookiesConfig struct {
	Cookies map[string]string `json:"cookies"`
}

type MutatorCookie struct {
	templates *template.Template
	c         configuration.Provider
}

func NewMutatorCookie(c configuration.Provider) *MutatorCookie {
	return &MutatorCookie{
		c:         c,
		templates: newTemplate("cookie"),
	}
}

func (a *MutatorCookie) GetID() string {
	return "cookie"
}

func (a *MutatorCookie) Mutate(r *http.Request, session *authn.AuthenticationSession, config json.RawMessage, rl pipeline.Rule) (http.Header, error) {
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

		templateId := fmt.Sprintf("%s:%s", rl.GetID(), cookie)
		tmpl = a.templates.Lookup(templateId)
		if tmpl == nil {
			tmpl, err = a.templates.New(templateId).Parse(templateString)
			if err != nil {
				return nil, errors.Wrapf(err, `error parsing cookie template "%s" in rule "%s"`, templateString, rl.GetID())
			}
		}

		cookieValue := bytes.Buffer{}
		err = tmpl.Execute(&cookieValue, session)
		if err != nil {
			return nil, errors.Wrapf(err, `error executing cookie template "%s" in rule "%s"`, templateString, rl.GetID())
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

func (a *MutatorCookie) Validate() error {
	if !a.c.MutatorCookieIsEnabled() {
		return errors.WithStack(authn.ErrAuthenticatorNotEnabled.WithReasonf("Mutator % is disabled per configuration.", a.GetID()))
	}

	return nil
}
