package authz

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/pkg/errors"
	"github.com/tidwall/gjson"

	"github.com/ory/x/httpx"

	"github.com/ory/oathkeeper/driver"
	"github.com/ory/oathkeeper/driver/configuration"
	"github.com/ory/oathkeeper/helper"
	"github.com/ory/oathkeeper/pipeline"
	"github.com/ory/oathkeeper/pipeline/authn"
	"github.com/ory/oathkeeper/template"
	"github.com/ory/oathkeeper/x"
)

// AuthorizerRemoteJSONConfiguration represents a configuration for the remote_json authorizer.
type AuthorizerRemoteJSONConfiguration struct {
	Remote   string                `json:"remote"`
	Payload  string                `json:"payload"`
	Renderer template.RenderEngine `json:"template_engine"`
}

// PayloadTemplateID returns a string with which to associate the payload template.
func (c *AuthorizerRemoteJSONConfiguration) PayloadTemplateID() string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(c.Payload)))
}

// AuthorizerRemoteJSON implements the Authorizer interface.
type AuthorizerRemoteJSON struct {
	c configuration.Provider
	d driver.Registry

	client *http.Client
	t      *template.Renderer
}

// NewAuthorizerRemoteJSON creates a new AuthorizerRemoteJSON.
func NewAuthorizerRemoteJSON(c configuration.Provider, d driver.Registry) *AuthorizerRemoteJSON {
	return &AuthorizerRemoteJSON{
		c: c, d: d, client: httpx.NewResilientClientLatencyToleranceSmall(nil), t: template.NewRenderer()}
}

// GetID implements the Authorizer interface.
func (a *AuthorizerRemoteJSON) GetID() string {
	return "remote_json"
}

// Authorize implements the Authorizer interface.
func (a *AuthorizerRemoteJSON) Authorize(_ *http.Request, session *authn.AuthenticationSession, config json.RawMessage, rl pipeline.Rule) error {
	c, err := a.Config(config)
	if err != nil {
		return err
	}

	logger := x.LogAccessRuleContext(a.d, a.c, rl, session).
		WithField("config", string(config)).
		WithField("handler", "authz/remote")

	value, err := a.t.Render(c.Payload, c.Renderer, session,template.ExpectJSON())
	if err != nil {
		return errors.Wrap(err, `template render engine returned an error`)
	}

	req, err := http.NewRequest("POST", c.Remote, strings.NewReader(value))
	if err != nil {
		return errors.WithStack(err)
	}
	req.Header.Add("Content-Type", "application/json")

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
func (a *AuthorizerRemoteJSON) Validate(config json.RawMessage) error {
	if !a.c.AuthorizerIsEnabled(a.GetID()) {
		return NewErrAuthorizerNotEnabled(a)
	}

	_, err := a.Config(config)
	return err
}

// Config merges config and the authorizer's configuration and validates the
// resulting configuration. It reports an error if the configuration is invalid.
func (a *AuthorizerRemoteJSON) Config(config json.RawMessage) (*AuthorizerRemoteJSONConfiguration, error) {
	var c AuthorizerRemoteJSONConfiguration
	if err := a.c.AuthorizerConfig(a.GetID(), config, &c); err != nil {
		return nil, NewErrAuthorizerMisconfigured(a, err)
	}

	return &c, nil
}
