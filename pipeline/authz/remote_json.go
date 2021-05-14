package authz

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"text/template"

	"github.com/pkg/errors"

	"github.com/ory/x/httpx"

	"github.com/ory/oathkeeper/driver/configuration"
	"github.com/ory/oathkeeper/helper"
	"github.com/ory/oathkeeper/pipeline"
	"github.com/ory/oathkeeper/pipeline/authn"
	"github.com/ory/oathkeeper/x"
)

// AuthorizerRemoteJSONConfiguration represents a configuration for the remote_json authorizer.
type AuthorizerRemoteJSONConfiguration struct {
	Remote                           string   `json:"remote"`
	Payload                          string   `json:"payload"`
	ForwardResponseHeadersToUpstream []string `json:"forward_response_headers_to_upstream"`
}

// PayloadTemplateID returns a string with which to associate the payload template.
func (c *AuthorizerRemoteJSONConfiguration) PayloadTemplateID() string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(c.Payload)))
}

// AuthorizerRemoteJSON implements the Authorizer interface.
type AuthorizerRemoteJSON struct {
	c configuration.Provider

	client *http.Client
	t      *template.Template
}

// NewAuthorizerRemoteJSON creates a new AuthorizerRemoteJSON.
func NewAuthorizerRemoteJSON(c configuration.Provider) *AuthorizerRemoteJSON {
	return &AuthorizerRemoteJSON{
		c:      c,
		client: httpx.NewResilientClientLatencyToleranceSmall(nil),
		t:      x.NewTemplate("remote_json"),
	}
}

// GetID implements the Authorizer interface.
func (a *AuthorizerRemoteJSON) GetID() string {
	return "remote_json"
}

// Authorize implements the Authorizer interface.
func (a *AuthorizerRemoteJSON) Authorize(r *http.Request, session *authn.AuthenticationSession, config json.RawMessage, _ pipeline.Rule) error {
	c, err := a.Config(config)
	if err != nil {
		return err
	}

	templateID := c.PayloadTemplateID()
	t := a.t.Lookup(templateID)
	if t == nil {
		var err error
		t, err = a.t.New(templateID).Parse(c.Payload)
		if err != nil {
			return errors.WithStack(err)
		}
	}

	var body bytes.Buffer
	if err := t.Execute(&body, session); err != nil {
		return errors.WithStack(err)
	}

	var j json.RawMessage
	if err := json.Unmarshal(body.Bytes(), &j); err != nil {
		return errors.Wrap(err, "payload is not a JSON text")
	}

	req, err := http.NewRequest("POST", c.Remote, &body)
	if err != nil {
		return errors.WithStack(err)
	}
	req.Header.Add("Content-Type", "application/json")
	authz := r.Header.Get("Authorization")
	if authz != "" {
		req.Header.Add("Authorization", authz)
	}

	res, err := a.client.Do(req.WithContext(r.Context()))
	if err != nil {
		return errors.WithStack(err)
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusForbidden {
		return errors.WithStack(helper.ErrForbidden)
	} else if res.StatusCode != http.StatusOK {
		return errors.Errorf("expected status code %d but got %d", http.StatusOK, res.StatusCode)
	}

	for _, allowedHeader := range c.ForwardResponseHeadersToUpstream {
		session.SetHeader(allowedHeader, res.Header.Get(allowedHeader))
	}

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
