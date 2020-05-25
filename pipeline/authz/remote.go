package authz

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/ory/x/httpx"
	"github.com/pkg/errors"

	"github.com/ory/oathkeeper/driver"
	"github.com/ory/oathkeeper/driver/configuration"
	"github.com/ory/oathkeeper/helper"
	"github.com/ory/oathkeeper/pipeline"
	"github.com/ory/oathkeeper/pipeline/authn"
	"github.com/ory/oathkeeper/template"
	"github.com/ory/oathkeeper/x"
)

// AuthorizerRemoteConfiguration represents a configuration for the remote authorizer.
type AuthorizerRemoteConfiguration struct {
	Remote  string            `json:"remote"`
	Headers map[string]string `json:"headers"`
}

// AuthorizerRemote implements the Authorizer interface.
type AuthorizerRemote struct {
	c configuration.Provider
	d driver.Registry

	client *http.Client
	t      *template.Renderer
}

// NewAuthorizerRemote creates a new AuthorizerRemote.
func NewAuthorizerRemote(c configuration.Provider,
	d driver.Registry) *AuthorizerRemote {
	return &AuthorizerRemote{
		c: c, d: d, client: httpx.NewResilientClientLatencyToleranceSmall(nil),
		t: template.NewRenderer(),
	}
}

// GetID implements the Authorizer interface.
func (a *AuthorizerRemote) GetID() string {
	return "remote"
}

// Authorize implements the Authorizer interface.
func (a *AuthorizerRemote) Authorize(r *http.Request, session *authn.AuthenticationSession, config json.RawMessage, rl pipeline.Rule) error {
	c, err := a.Config(config)
	if err != nil {
		return err
	}

	logger := x.LogAccessRuleContext(a.d, a.c, rl, session).
		WithField("config", string(config)).
		WithField("handler", "authz/remote")

	read, write := io.Pipe()
	go func() {
		err := pipeRequestBody(r, write)
		write.CloseWithError(errors.Wrapf(err, `could not pipe request body in rule "%s"`, rl.GetID()))
	}()

	req, err := http.NewRequest("POST", c.Remote, read)
	if err != nil {
		return errors.WithStack(err)
	}
	req.Header.Add("Content-Type", r.Header.Get("Content-Type"))

	if err := a.t.RenderHeaders(c.Headers, req.Header, session); err != nil {
		return errors.Wrapf(err, `error parsing headers in rule "%s"`, source, rl.GetID())
	}

	logger.WithError(err).
		WithField("header", req.Header).
		Trace("Making request to remote with these HTTP Headers.")

	res, err := a.client.Do(req)
	if err != nil {
		logger.WithError(err).
			Debug("Unable to initiate request due to a network or timeout error.")
		return errors.WithStack(err)
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusForbidden {
		logger.WithField("remote_status_code", res.StatusCode).
			Trace("Remote replied with status code 403 indicating that the request must not be allowed.")
		return errors.WithStack(helper.ErrForbidden)
	} else if res.StatusCode != http.StatusOK {
		logger.WithField("remote_status_code", res.StatusCode).
			Debug("Remote replied with status code 403 indicating that the request must not be allowed.")
		return errors.Errorf("expected status code %d but got %d", http.StatusOK, res.StatusCode)
	}

	logger.WithField("remote_status_code", res.StatusCode).
		Trace("Remote replied with status code 200 indicating that the request should be allowed.")
	return nil
}

// Validate implements the Authorizer interface.
func (a *AuthorizerRemote) Validate(config json.RawMessage) error {
	if !a.c.AuthorizerIsEnabled(a.GetID()) {
		return NewErrAuthorizerNotEnabled(a)
	}

	_, err := a.Config(config)
	return err
}

// Config merges config and the authorizer's configuration and validates the
// resulting configuration. It reports an error if the configuration is invalid.
func (a *AuthorizerRemote) Config(config json.RawMessage) (*AuthorizerRemoteConfiguration, error) {
	var c AuthorizerRemoteConfiguration
	if err := a.c.AuthorizerConfig(a.GetID(), config, &c); err != nil {
		return nil, NewErrAuthorizerMisconfigured(a, err)
	}

	return &c, nil
}
