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

package rsakey

import (
	"crypto/rsa"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/ory/x/urlx"

	"github.com/pkg/errors"
	"gopkg.in/square/go-jose.v2"
)

type HydraManager struct {
	key      *rsa.PrivateKey
	Set      string
	client   *http.Client
	adminURL *url.URL
}

func NewHydraManager(
	set string,
	client *http.Client,
	adminURL *url.URL,
) *HydraManager {
	if client == nil {
		client = &http.Client{Timeout: time.Second * 5}
	}
	return &HydraManager{
		client:   client,
		Set:      set,
		adminURL: adminURL,
	}
}

func (m *HydraManager) Refresh() error {
	res, err := m.client.Get(urlx.AppendPaths(m.adminURL, "/keys/", m.Set).String())
	if err != nil {
		return errors.WithStack(err)
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusNotFound {
		req, err := http.NewRequest("POST", urlx.AppendPaths(m.adminURL, "/keys/", m.Set).String(), strings.NewReader(`{"alg":"RS256"}`))
		if err != nil {
			return errors.WithStack(err)
		}
		req.Header.Set("Content-Type", "application/json")

		res, err = m.client.Do(req)
		if err != nil {
			return errors.WithStack(err)
		}
		defer res.Body.Close()

		if res.StatusCode != http.StatusCreated {
			return errors.Errorf("Expected status code %d but got %d", http.StatusOK, res.StatusCode)
		}
	} else if res.StatusCode != http.StatusOK {
		return errors.Errorf("Expected status code %d but got %d", http.StatusOK, res.StatusCode)
	}

	var set jose.JSONWebKeySet
	if err := json.NewDecoder(res.Body).Decode(&set); err != nil {
		return errors.WithStack(err)
	}

	var privateKey *rsa.PrivateKey
	for _, key := range set.Keys {
		if strings.Contains(key.KeyID, "private:") {
			var ok bool
			privateKey, ok = key.Key.(*rsa.PrivateKey)
			if !ok {
				return errors.Errorf("Type assertion to *rsa.PrivateKey failed, make sure you are actually sending a RSA private key")
			}
		}
	}

	if privateKey == nil {
		return errors.New("Expected at least one private key but got none")
	}

	m.key = privateKey
	return nil
}

func (m *HydraManager) PublicKey() (interface{}, error) {
	if m.key == nil {
		if err := m.Refresh(); err != nil {
			return nil, err
		}
	}
	return &m.key.PublicKey, nil
}

func (m *HydraManager) PrivateKey() (interface{}, error) {
	if m.key == nil {
		if err := m.Refresh(); err != nil {
			return nil, err
		}
	}
	return m.key, nil
}

func (m *HydraManager) PublicKeyID() string {
	return m.Set + ":public"
}

func (m *HydraManager) Algorithm() string {
	return "RS256"
}
