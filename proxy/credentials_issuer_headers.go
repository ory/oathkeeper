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

type CredentialsHeadersConfig map[string]string

type CredentialsHeaders struct {
	rulesCache *template.Template
}

func NewCredentialsIssuerHeaders() *CredentialsHeaders {
	return &CredentialsHeaders{
		rulesCache: template.New("rules").Option("missingkey=zero"),
	}
}

func (a *CredentialsHeaders) GetID() string {
	return "headers"
}

func (a *CredentialsHeaders) Issue(r *http.Request, session *AuthenticationSession, config json.RawMessage, rl *rule.Rule) error {
	if len(config) == 0 {
		config = []byte("{}")
	}

	var cfg CredentialsHeadersConfig
	d := json.NewDecoder(bytes.NewBuffer(config))
	if err := d.Decode(&cfg); err != nil {
		return errors.WithStack(err)
	}

	convertedSession := convertSession(session)

	for hdr, templateString := range cfg {
		var tmpl *template.Template
		var err error

		templateId := fmt.Sprintf("%s:%s", rl.ID, hdr)
		tmpl = a.rulesCache.Lookup(templateId)
		if tmpl == nil {
			tmpl, err = a.rulesCache.New(templateId).Parse(templateString)
			if err != nil {
				return errors.Wrapf(err, `error parsing header template "%s" in rule "%s"`, templateString, rl.ID)
			}
		}

		headerValue := bytes.Buffer{}
		err = tmpl.Execute(&headerValue, convertedSession)
		if err != nil {
			return errors.Wrapf(err, `error executing header template "%s" in rule "%s"`, templateString, rl.ID)
		}
		r.Header.Set(hdr, headerValue.String())
	}

	return nil
}

type authSession struct {
	Subject string
	Extra   map[string]string
}

func convertSession(in *AuthenticationSession) *authSession {
	out := authSession{
		Subject: in.Subject,
		Extra:   map[string]string{},
	}

	for k, v := range in.Extra {
		out.Extra[k] = fmt.Sprintf("%s", v)
	}

	return &out
}
