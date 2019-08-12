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
	"github.com/cenkalti/backoff"
	"github.com/ory/x/httpx"
	"net/http"
	"net/url"
	"time"

	"github.com/pkg/errors"

	"github.com/ory/oathkeeper/driver/configuration"
	"github.com/ory/oathkeeper/pipeline"
	"github.com/ory/oathkeeper/pipeline/authn"
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
)

type MutatorHydrator struct {
	c      configuration.Provider
	client *http.Client
}

type BasicAuthn struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Authentication struct {
	Basic BasicAuthn `json:"basic"`
}

type RetryConfig struct {
	NumberOfRetries     int `json:"number"`
	DelayInMilliseconds int `json:"delayInMilliseconds"`
}

type externalAPIConfig struct {
	Url   string          `json:"url"`
	Authn *Authentication `json:"authn,omitempty"`
	Retry *RetryConfig    `json:"retry,omitempty"`
}

type MutatorHydratorConfig struct {
	Api externalAPIConfig `json:"api"`
}

func NewMutatorHydrator(c configuration.Provider) *MutatorHydrator {
	return &MutatorHydrator{c: c, client: httpx.NewResilientClientLatencyToleranceSmall(nil)}
}

func (a *MutatorHydrator) GetID() string {
	return "Hydrator"
}

func (a *MutatorHydrator) Mutate(r *http.Request, session *authn.AuthenticationSession, config json.RawMessage, _ pipeline.Rule) error {
	if len(config) == 0 {
		config = []byte("{}")
	}
	var cfg MutatorHydratorConfig
	d := json.NewDecoder(bytes.NewBuffer(config))
	d.DisallowUnknownFields()
	if err := d.Decode(&cfg); err != nil {
		return errors.WithStack(err)
	}

	var b bytes.Buffer
	err := json.NewEncoder(&b).Encode(session)
	if err != nil {
		return errors.WithStack(err)
	}

	if cfg.Api.Url == "" {
		return errors.New(ErrMissingAPIURL)
	} else if _, err := url.ParseRequestURI(cfg.Api.Url); err != nil {
		return errors.New(ErrInvalidAPIURL)
	}
	req, err := http.NewRequest("POST", cfg.Api.Url, &b)
	if err != nil {
		return errors.WithStack(err)
	}
	for key, values := range r.Header {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}
	if cfg.Api.Authn != nil {
		credentials := cfg.Api.Authn.Basic
		req.SetBasicAuth(credentials.Username, credentials.Password)
	}

	retryConfig := RetryConfig{defaultNumberOfRetries, defaultDelayInMilliseconds}
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
			if cfg.Api.Authn != nil {
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
	if sessionFromUpstream.Extra == nil || sessionFromUpstream.Subject != session.Subject {
		return errors.New(ErrMalformedResponseFromUpstreamAPI)
	}
	*session = sessionFromUpstream

	return nil
}

func (a *MutatorHydrator) Validate() error {
	if !a.c.MutatorHydratorIsEnabled() {
		return errors.WithStack(ErrMutatorNotEnabled.WithReasonf(`Mutator "%s" is disabled per configuration.`, a.GetID()))
	}
	return nil
}
