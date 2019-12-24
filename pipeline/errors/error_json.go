package errors

import (
	"encoding/json"
	"net/http"

	"github.com/ory/herodot"
	"github.com/ory/x/errorsx"

	"github.com/ory/oathkeeper/driver/configuration"
	"github.com/ory/oathkeeper/rule"
	"github.com/ory/oathkeeper/x"
)

var _ Handler = new(ErrorJSON)

type (
	ErrorJSONConfig struct {
		Verbose bool  `json:"verbose"`
		Whens   Whens `json:"when"`
	}
	ErrorJSON struct {
		c configuration.Provider
		d errorJSONDependencies
	}
	errorJSONDependencies interface {
		x.RegistryWriter
	}
)

func NewErrorJSON(
	c configuration.Provider,
	d errorJSONDependencies,
) *ErrorJSON {
	return &ErrorJSON{c: c, d: d}
}

func (a *ErrorJSON) Handle(w http.ResponseWriter, r *http.Request, config json.RawMessage, _ *rule.Rule, ge error) error {
	c, err := a.Config(config)
	if err != nil {
		return err
	}

	if err := MatchesWhen(c.Whens, r, ge); err != nil {
		return err
	}

	if !c.Verbose {
		if sc, ok := errorsx.Cause(ge).(statusCoder); ok {
			switch sc.StatusCode() {
			case http.StatusInternalServerError:
				ge = herodot.ErrInternalServerError.WithTrace(ge)
			case http.StatusForbidden:
				ge = herodot.ErrForbidden.WithTrace(ge)
			case http.StatusNotFound:
				ge = herodot.ErrNotFound.WithTrace(ge)
			case http.StatusUnauthorized:
				ge = herodot.ErrUnauthorized.WithTrace(ge)
			case http.StatusBadRequest:
				ge = herodot.ErrBadRequest.WithTrace(ge)
			case http.StatusUnsupportedMediaType:
				ge = herodot.ErrUnsupportedMediaType.WithTrace(ge)
			case http.StatusConflict:
				ge = herodot.ErrConflict.WithTrace(ge)
			}
		} else {
			ge = herodot.ErrInternalServerError.WithTrace(ge)
		}
	}

	a.d.Writer().WriteError(w, r, ge)
	return nil
}

func (a *ErrorJSON) Validate(config json.RawMessage) error {
	if !a.c.AuthorizerIsEnabled(a.GetID()) {
		return NewErrErrorHandlerNotEnabled(a)
	}

	_, err := a.Config(config)
	return err
}

func (a *ErrorJSON) Config(config json.RawMessage) (*ErrorJSONConfig, error) {
	var c ErrorJSONConfig
	if err := a.c.ErrorHandlerConfig(a.GetID(), config, &c); err != nil {
		return nil, NewErrErrorHandlerMisconfigured(a, err)
	}

	return &c, nil
}

func (a *ErrorJSON) GetID() string {
	return "json"
}
