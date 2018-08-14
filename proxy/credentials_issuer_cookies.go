package proxy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"text/template"

	"github.com/ory/oathkeeper/rule"
	"github.com/pkg/errors"
)

type CredentialsCookiesConfig struct {
	Cookies map[string]string `json:"cookies"`
}

type CredentialsCookies struct {
	RulesCache *template.Template
}

func NewCredentialsIssuerCookies() *CredentialsCookies {
	return &CredentialsCookies{
		RulesCache: template.New("rules").
			Option("missingkey=zero").
			Funcs(template.FuncMap{
				"print": func(i interface{}) string {
					if i == nil {
						return ""
					}
					return fmt.Sprintf("%v", i)
				},
			}),
	}
}

func (a *CredentialsCookies) GetID() string {
	return "cookies"
}

func (a *CredentialsCookies) Issue(r *http.Request, session *AuthenticationSession, config json.RawMessage, rl *rule.Rule) error {
	if len(config) == 0 {
		config = []byte("{}")
	}

	// Cache request cookies
	requestCookies := r.Cookies()

	// Remove existing cookies
	r.Header.Del("Cookie")

	// Keep track of rule cookies in a map
	cookies := map[string]bool{}

	var cfg CredentialsCookiesConfig
	d := json.NewDecoder(bytes.NewBuffer(config))
	d.DisallowUnknownFields()
	if err := d.Decode(&cfg); err != nil {
		return errors.WithStack(err)
	}

	for cookie, templateString := range cfg.Cookies {
		var tmpl *template.Template
		var err error

		templateId := fmt.Sprintf("%s:%s", rl.ID, cookie)
		tmpl = a.RulesCache.Lookup(templateId)
		if tmpl == nil {
			tmpl, err = a.RulesCache.New(templateId).Parse(templateString)
			if err != nil {
				return errors.Wrapf(err, `error parsing cookie template "%s" in rule "%s"`, templateString, rl.ID)
			}
		}

		cookieValue := bytes.Buffer{}
		err = tmpl.Execute(&cookieValue, session)
		if err != nil {
			return errors.Wrapf(err, `error executing cookie template "%s" in rule "%s"`, templateString, rl.ID)
		}

		r.AddCookie(&http.Cookie{
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
			r.AddCookie(cookie)
		}
	}

	return nil
}
