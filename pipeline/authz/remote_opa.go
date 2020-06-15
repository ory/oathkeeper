package authz

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"text/template"

	"github.com/pkg/errors"

	"github.com/ory/x/httpx"

	"github.com/ory/oathkeeper/driver/configuration"
	"github.com/ory/oathkeeper/helper"
	"github.com/ory/oathkeeper/pipeline"
	"github.com/ory/oathkeeper/pipeline/authn"
	"github.com/ory/oathkeeper/x"
)

// AuthorizerRemoteOPAConfiguration represents a configuration for the remote_opa authorizer.
type AuthorizerRemoteOPAConfiguration struct {
	Remote  string `json:"remote"`
	Payload string `json:"payload"`
}

// PayloadTemplateID returns a string with which to associate the payload template.
func (c *AuthorizerRemoteOPAConfiguration) PayloadTemplateID() string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(c.Payload)))
}

// AuthorizerRemoteOPA implements the Authorizer interface.
type AuthorizerRemoteOPA struct {
	c configuration.Provider

	client *http.Client
	t      *template.Template
}

// NewAuthorizerRemoteOPA creates a new AuthorizerRemoteOPA.
func NewAuthorizerRemoteOPA(c configuration.Provider) *AuthorizerRemoteOPA {
	return &AuthorizerRemoteOPA{
		c:      c,
		client: httpx.NewResilientClientLatencyToleranceSmall(nil),
		t:      x.NewTemplate("remote_opa"),
	}
}

// OPA Response
type OPAResponse struct {
	Result struct {
		Allow bool
	}
	Decision_id string
}

// GetID implements the Authorizer interface.
func (a *AuthorizerRemoteOPA) GetID() string {
	return "remote_opa"
}

// Authorize implements the Authorizer interface.
func (a *AuthorizerRemoteOPA) Authorize(_ *http.Request, session *authn.AuthenticationSession, config json.RawMessage, _ pipeline.Rule) error {
	c, err := a.Config(config)
	if err != nil {
		return err
	}

	// Substitute the PathArray data structure into the payload if present in the payload template
	if strings.Contains(c.Payload, "{{ .PathArray }}") {
		b, err := json.Marshal(strings.Split(session.MatchContext.URL.Path, "/")[1:])
		if err != nil {
			return errors.WithStack(err)
		}
		c.Payload = strings.ReplaceAll(c.Payload, "{{ .PathArray }}", string(b))
	}

	// Evaluate the remaining template items against the AuthenticationSession struct
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

	res, err := a.client.Do(req)
	if err != nil {
		return errors.WithStack(err)
	}
	defer res.Body.Close()

	if err != nil {
		return errors.WithStack(err)
	} else if res.StatusCode != http.StatusOK {
		return errors.WithStack(helper.ErrBadAuthorizerResponse)
	} else {
		var oparesp OPAResponse
		decoder := json.NewDecoder(res.Body)
		err = decoder.Decode(&oparesp)

		if err != nil {
			return errors.WithStack(err)
		}

		if oparesp.Result.Allow != true {
			return errors.WithStack(helper.ErrForbidden)
		}

	}

	return nil
}

// Validate implements the Authorizer interface.
func (a *AuthorizerRemoteOPA) Validate(config json.RawMessage) error {
	if !a.c.AuthorizerIsEnabled(a.GetID()) {
		return NewErrAuthorizerNotEnabled(a)
	}

	_, err := a.Config(config)
	return err
}

// Config merges config and the authorizer's configuration and validates the
// resulting configuration. It reports an error if the configuration is invalid.
func (a *AuthorizerRemoteOPA) Config(config json.RawMessage) (*AuthorizerRemoteOPAConfiguration, error) {
	var c AuthorizerRemoteOPAConfiguration
	if err := a.c.AuthorizerConfig(a.GetID(), config, &c); err != nil {
		return nil, NewErrAuthorizerMisconfigured(a, err)
	}

	return &c, nil
}
