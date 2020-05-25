package mutate

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pkg/errors"

	"github.com/ory/oathkeeper/driver/configuration"
	"github.com/ory/oathkeeper/pipeline"
	"github.com/ory/oathkeeper/pipeline/authn"
	"github.com/ory/oathkeeper/template"
)

type MutatorHeaderConfig struct {
	Headers map[string]string `json:"headers"`
}

type mutatorHeaderDependencies interface {
	template.RenderProvider
}

type MutatorHeader struct {
	c configuration.Provider
	d mutatorHeaderDependencies
}

func NewMutatorHeader(c configuration.Provider, d mutatorHeaderDependencies) *MutatorHeader {
	return &MutatorHeader{c: c, d: d}
}

func (a *MutatorHeader) GetID() string {
	return "header"
}

func (a *MutatorHeader) Mutate(r *http.Request, session *authn.AuthenticationSession, config json.RawMessage, rl pipeline.Rule) error {
	cfg, err := a.config(config)
	if err != nil {
		return err
	}

	dest := http.Header{}
	if err := a.d.Renderer().RenderHeaders(cfg.Headers,dest,session); err != nil {
		return errors.Wrapf(err, `error parsing headers in rule "%s"`, source, rl.GetID())
	}

	for k := range dest {
		session.SetHeader(k, dest.Get(k))
	}

	return nil
}

func (a *MutatorHeader) Validate(config json.RawMessage) error {
	if !a.c.MutatorIsEnabled(a.GetID()) {
		return NewErrMutatorNotEnabled(a)
	}

	_, err := a.config(config)
	return err
}

func (a *MutatorHeader) config(config json.RawMessage) (*MutatorHeaderConfig, error) {
	var c MutatorHeaderConfig
	if err := a.c.MutatorConfig(a.GetID(), config, &c); err != nil {
		return nil, NewErrMutatorMisconfigured(a, err)
	}

	return &c, nil
}
