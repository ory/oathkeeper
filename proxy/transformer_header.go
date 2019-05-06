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

type CredentialsHeadersConfig struct {
	Headers map[string]string `json:"headers"`
}

type TransformerHeader struct {
	c configuration.Provider
	t *template.Template
}

func NewCredentialsIssuerHeaders(c configuration.Provider) *TransformerHeader {
	return &TransformerHeader{
		t: newTemplate("header"),
	}
}

func (a *TransformerHeader) GetID() string {
	return "header"
}

func (a *TransformerHeader) Transform(r *http.Request, session *AuthenticationSession, config json.RawMessage, rl *rule.Rule) (http.Header, error) {
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
		tmpl = a.t.Lookup(templateId)
		if tmpl == nil {
			tmpl, err = a.t.New(templateId).Parse(templateString)
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

func (a *TransformerHeader) Validate() error {
	if !a.c.TransformerHeaderIsEnabled() {
		return errors.WithStack(ErrAuthenticatorNotEnabled.WithReasonf("Transformer % is disabled per configuration.", a.GetID()))
	}

	return nil
}
