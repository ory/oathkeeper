/*
 * Copyright © 2017-2018 Aeneas Rekkas <aeneas+oss@aeneas.io>
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * @author       Aeneas Rekkas <aeneas+oss@aeneas.io>
 * @copyright  2017-2018 Aeneas Rekkas <aeneas+oss@aeneas.io>
 * @license  	   Apache-2.0
 */

package mutate

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/url"
	"time"

	"github.com/ory/oathkeeper/pipeline/authn"

	"github.com/cenkalti/backoff"

	"github.com/ory/x/httpx"

	"github.com/pkg/errors"

	"github.com/ory/oathkeeper/driver/configuration"
	"github.com/ory/oathkeeper/pipeline"
)

const (
	ErrMalformedResponseFromUpstreamAPI = "The call to an external API returned an invalid JSON object"
	ErrMissingAPIURL                    = "Missing URL in mutator configuration"
	ErrInvalidAPIURL                    = "Invalid URL in mutator configuration"
	ErrNon200ResponseFromAPI            = "The call to an external API returned a non-200 HTTP response"
	ErrInvalidCredentials               = "Invalid credentials were provided in mutator configuration"
	ErrNoCredentialsProvided            = "No credentials were provided in mutator configuration"
	defaultNumberOfRetries              = 3
	defaultDelayInMilliseconds          = 100
	contentTypeHeaderKey                = "Content-Type"
	contentTypeJSONHeaderValue          = "application/json"
)

type MutatorHydrator struct {
	c      configuration.Provider
	client *http.Client
}

type BasicAuth struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type auth struct {
	Basic BasicAuth `json:"basic"`
}

type retryConfig struct {
	NumberOfRetries     int `json:"number_of_retries"`
	DelayInMilliseconds int `json:"delay_in_milliseconds"`
}

type externalAPIConfig struct {
	URL   string       `json:"url"`
	Auth  *auth        `json:"auth"`
	Retry *retryConfig `json:"retry"`
}

type MutatorHydratorConfig struct {
	Api externalAPIConfig `json:"api"`
}

func NewMutatorHydrator(c configuration.Provider) *MutatorHydrator {
	return &MutatorHydrator{c: c, client: httpx.NewResilientClientLatencyToleranceSmall(nil)}
}

func (a *MutatorHydrator) GetID() string {
	return "hydrator"
}

func (a *MutatorHydrator) Mutate(r *http.Request, session *authn.AuthenticationSession, config json.RawMessage, _ pipeline.Rule) error {
	cfg, err := a.Config(config)
	if err != nil {
		return err
	}

	var b bytes.Buffer
	if err := json.NewEncoder(&b).Encode(session); err != nil {
		return errors.WithStack(err)
	}

	if cfg.Api.URL == "" {
		return errors.New(ErrMissingAPIURL)
	} else if _, err := url.ParseRequestURI(cfg.Api.URL); err != nil {
		return errors.New(ErrInvalidAPIURL)
	}
	req, err := http.NewRequest("POST", cfg.Api.URL, &b)
	if err != nil {
		return errors.WithStack(err)
	}
	for key, values := range r.Header {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}
	if cfg.Api.Auth != nil {
		credentials := cfg.Api.Auth.Basic
		req.SetBasicAuth(credentials.Username, credentials.Password)
	}
	req.Header.Set(contentTypeHeaderKey, contentTypeJSONHeaderValue)

	retryConfig := retryConfig{defaultNumberOfRetries, defaultDelayInMilliseconds}
	if cfg.Api.Retry != nil {
		retryConfig = *cfg.Api.Retry
	}
	var res *http.Response
	err = backoff.Retry(func() error {
		res, err = a.client.Do(req)
		if err != nil {
			return errors.WithStack(err)
		}
		switch res.StatusCode {
		case http.StatusOK:
			return nil
		case http.StatusUnauthorized:
			if cfg.Api.Auth != nil {
				return errors.New(ErrInvalidCredentials)
			} else {
				return errors.New(ErrNoCredentialsProvided)
			}
		default:
			return errors.New(ErrNon200ResponseFromAPI)
		}
	}, backoff.WithMaxRetries(backoff.NewConstantBackOff(time.Millisecond*time.Duration(retryConfig.DelayInMilliseconds)), uint64(retryConfig.NumberOfRetries)))
	if err != nil {
		return err
	}

	sessionFromUpstream := authn.AuthenticationSession{}
	err = json.NewDecoder(res.Body).Decode(&sessionFromUpstream)
	if err != nil {
		return errors.WithStack(err)
	}
	if sessionFromUpstream.Subject != session.Subject {
		return errors.New(ErrMalformedResponseFromUpstreamAPI)
	}
	*session = sessionFromUpstream

	return nil
}

func (a *MutatorHydrator) Validate(config json.RawMessage) error {
	if !a.c.MutatorIsEnabled(a.GetID()) {
		return NewErrMutatorNotEnabled(a)
	}

	_, err := a.Config(config)
	return err
}

func (a *MutatorHydrator) Config(config json.RawMessage) (*MutatorHydratorConfig, error) {
	var c MutatorHydratorConfig
	if err := a.c.MutatorConfig(a.GetID(), config, &c); err != nil {
		return nil, NewErrMutatorMisconfigured(a, err)
	}

	return &c, nil
}
