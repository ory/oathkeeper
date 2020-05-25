package template

import (
	"net/http"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"

	"github.com/ory/oathkeeper/pipeline/authn"
	"github.com/ory/oathkeeper/x"
)

type RenderEngine string

const (
	RenderEngineJsonNet RenderEngine = "jsonnet"
	RenderEngineStdLib  RenderEngine = "stdlib"
)

type rendererDependencies interface {
	x.LoggingProvider
}

type Renderer struct {
	d rendererDependencies
}

type renderer interface {
	Render(template string, engine RenderEngine, session *authn.AuthenticationSession, opts renderOptions) (string, error)
}

type RenderProvider interface {
	Renderer() *Renderer
}

func NewRenderer(d rendererDependencies) *Renderer {
	return &Renderer{d: d}
}

type RenderOption func(options *renderOptions)

type renderOptions struct {
	expectJSON bool
}

func ExpectJSON() RenderOption {
	return func(o *renderOptions) {
		o.expectJSON = true
	}
}

func (r *Renderer) Render(template string, engine RenderEngine, session *authn.AuthenticationSession, options ...RenderOption) (string, error) {
	var opts renderOptions
	for _, f := range options {
		f(&opts)
	}

	var value string

	if opts.expectJSON && !gjson.Valid(value) {
		err := errors.New("template output is not valid JSON but should have been")
		r.d.Logger().
			WithField("template", template).
			WithField("engine", engine).
			WithField("output", value).
			WithError(err).
			Error("Template render engine returned a payload that is not valid JSON.")
		return "", err
	}

	return value, nil
}

func (r *Renderer) RenderHeaders(templates map[string]string, destination http.Header, session *authn.AuthenticationSession) error {
	for name, source := range templates {
		// Force stdlib renderer because headers aren't JSON formatted.
		value, err := r.Render(source, RenderEngineStdLib, session)
		if err != nil {
			r.d.Logger().
				WithFields(logrus.Fields{"template": source, "http_header_key": name}).
				WithError(err).Error("Unable to render HTTP header template")
			return err
		}

		// Don't send empty headers
		if len(value) == 0 {
			continue
		}

		destination.Set(name, value)
	}

	r.d.Logger().
		WithFields(logrus.Fields{"http_header_templates": templates, "http_header": destination}).
		Trace("HTTP header rendered successfully")
	return nil
}
