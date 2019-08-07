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

type MutatorHeaderConfig struct {
	Headers map[string]string `json:"headers"`
}

type MutatorHeader struct {
	c configuration.Provider
	t *template.Template
}

func NewMutatorHeader(c configuration.Provider) *MutatorHeader {
	return &MutatorHeader{c: c, t: newTemplate("header")}
}

func (a *MutatorHeader) GetID() string {
	return "header"
}

func (a *MutatorHeader) WithCache(t *template.Template) {
	a.t = t
}

func (a *MutatorHeader) Mutate(r *http.Request, session *authn.AuthenticationSession, config json.RawMessage, rl pipeline.Rule) error {
	if len(config) == 0 {
		config = []byte("{}")
	}

	var cfg MutatorHeaderConfig
	d := json.NewDecoder(bytes.NewBuffer(config))
	d.DisallowUnknownFields()
	if err := d.Decode(&cfg); err != nil {
		return errors.WithStack(err)
	}

	for hdr, templateString := range cfg.Headers {
		var tmpl *template.Template
		var err error

		templateId := fmt.Sprintf("%s:%s", rl.GetID(), hdr)
		tmpl = a.t.Lookup(templateId)
		if tmpl == nil {
			tmpl, err = a.t.New(templateId).Parse(templateString)
			if err != nil {
				return errors.Wrapf(err, `error parsing headers template "%s" in rule "%s"`, templateString, rl.GetID())
			}
		}

		headerValue := bytes.Buffer{}
		err = tmpl.Execute(&headerValue, session)
		if err != nil {
			return errors.Wrapf(err, `error executing headers template "%s" in rule "%s"`, templateString, rl.GetID())
		}
		session.SetHeader(hdr, headerValue.String())
	}

	return nil
}

func (a *MutatorHeader) Validate() error {
	if !a.c.MutatorHeaderIsEnabled() {
		return errors.WithStack(ErrMutatorNotEnabled.WithReasonf(`Mutator "%s" is disabled per configuration.`, a.GetID()))
	}

	return nil
}
