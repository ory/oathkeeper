// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package authz

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"text/template"
	"time"

	"github.com/pkg/errors"

	"github.com/ory/x/httpx"
	"github.com/ory/x/otelx"

	"go.opentelemetry.io/otel/trace"

	"github.com/ory/oathkeeper/driver/configuration"
	"github.com/ory/oathkeeper/helper"
	"github.com/ory/oathkeeper/pipeline"
	"github.com/ory/oathkeeper/pipeline/authn"
	"github.com/ory/oathkeeper/x"
)

// AuthorizerRemoteConfiguration represents a configuration for the remote authorizer.
type AuthorizerRemoteConfiguration struct {
	Remote                           string                              `json:"remote"`
	Headers                          map[string]string                   `json:"headers"`
	ForwardResponseHeadersToUpstream []string                            `json:"forward_response_headers_to_upstream"`
	Retry                            *AuthorizerRemoteRetryConfiguration `json:"retry"`
}

type AuthorizerRemoteRetryConfiguration struct {
	Timeout string `json:"max_delay"`
	MaxWait string `json:"give_up_after"`
}

// AuthorizerRemote implements the Authorizer interface.
type AuthorizerRemote struct {
	c configuration.Provider

	client *http.Client
	t      *template.Template
	tracer trace.Tracer
}

// NewAuthorizerRemote creates a new AuthorizerRemote.
func NewAuthorizerRemote(c configuration.Provider, d interface{ Tracer() trace.Tracer }) *AuthorizerRemote {
	return &AuthorizerRemote{
		c:      c,
		client: httpx.NewResilientClient().StandardClient(),
		t:      x.NewTemplate("remote"),
		tracer: d.Tracer(),
	}
}

// GetID implements the Authorizer interface.
func (a *AuthorizerRemote) GetID() string {
	return "remote"
}

// Authorize implements the Authorizer interface.
func (a *AuthorizerRemote) Authorize(r *http.Request, session *authn.AuthenticationSession, config json.RawMessage, rl pipeline.Rule) (err error) {
	ctx, span := a.tracer.Start(r.Context(), "pipeline.authz.AuthorizerRemote.Authorize")
	defer otelx.End(span, &err)

	c, err := a.Config(config)
	if err != nil {
		return err
	}

	read, write := io.Pipe()
	go func() {
		err := pipeRequestBody(r, write)
		write.CloseWithError(errors.Wrapf(err, `could not pipe request body in rule "%s"`, rl.GetID()))
	}()

	req, err := http.NewRequestWithContext(ctx, "POST", c.Remote, read)
	if err != nil {
		return errors.WithStack(err)
	}
	req.Header.Add("Content-Type", r.Header.Get("Content-Type"))
	authz := r.Header.Get("Authorization")
	if authz != "" {
		req.Header.Add("Authorization", authz)
	}

	for hdr, templateString := range c.Headers {
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
		// Don't send empty headers
		if headerValue.String() == "" {
			continue
		}

		req.Header.Set(hdr, headerValue.String())
	}

	res, err := a.client.Do(req)

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
	const (
		defaultTimeout = "500ms"
		defaultMaxWait = "1s"
	)
	var c AuthorizerRemoteConfiguration
	if err := a.c.AuthorizerConfig(a.GetID(), config, &c); err != nil {
		return nil, NewErrAuthorizerMisconfigured(a, err)
	}

	if c.Retry == nil {
		c.Retry = &AuthorizerRemoteRetryConfiguration{Timeout: defaultTimeout, MaxWait: defaultMaxWait}
	} else {
		if c.Retry.Timeout == "" {
			c.Retry.Timeout = defaultTimeout
		}
		if c.Retry.MaxWait == "" {
			c.Retry.MaxWait = defaultMaxWait
		}
	}
	duration, err := time.ParseDuration(c.Retry.Timeout)
	if err != nil {
		return nil, err
	}

	maxWait, err := time.ParseDuration(c.Retry.MaxWait)
	if err != nil {
		return nil, err
	}
	timeout := time.Millisecond * duration
	a.client = httpx.NewResilientClient(
		httpx.ResilientClientWithMaxRetryWait(maxWait),
		httpx.ResilientClientWithConnectionTimeout(timeout),
	).StandardClient()

	return &c, nil
}
