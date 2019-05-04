package proxy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"text/template"

	"github.com/pkg/errors"

	"github.com/ory/oathkeeper/rule"
)

type CredentialsHeadersConfig struct {
	Headers map[string]string `json:"headers"`
}

type CredentialsHeaders struct {
	RulesCache *template.Template
}

func NewCredentialsIssuerHeaders() *CredentialsHeaders {
	return &CredentialsHeaders{
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

func (a *CredentialsHeaders) GetID() string {
	return "headers"
}

func (a *CredentialsHeaders) Issue(r *http.Request, session *AuthenticationSession, config json.RawMessage, rl *rule.Rule) (http.Header, error) {
	if len(config) == 0 {
		config = []byte("{}")
	}

	var cfg CredentialsHeadersConfig
	d := json.NewDecoder(bytes.NewBuffer(config))
	d.DisallowUnknownFields()
	if err := d.Decode(&cfg); err != nil {
		return nil, errors.WithStack(err)
	}

	headers := http.Header{}
	for hdr, templateString := range cfg.Headers {
		var tmpl *template.Template
		var err error

		templateId := fmt.Sprintf("%s:%s", rl.ID, hdr)
		tmpl = a.RulesCache.Lookup(templateId)
		if tmpl == nil {
			tmpl, err = a.RulesCache.New(templateId).Parse(templateString)
			if err != nil {
				return nil, errors.Wrapf(err, `error parsing headers template "%s" in rule "%s"`, templateString, rl.ID)
			}
		}

		headerValue := bytes.Buffer{}
		err = tmpl.Execute(&headerValue, session)
		if err != nil {
			return nil, errors.Wrapf(err, `error executing headers template "%s" in rule "%s"`, templateString, rl.ID)
		}
		headers.Set(hdr, headerValue.String())
	}

	return headers, nil
}
