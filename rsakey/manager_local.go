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
	"crypto/rand"
	"crypto/rsa"

	"github.com/pkg/errors"
)

type LocalManager struct {
	key         *rsa.PrivateKey
	KeyStrength int
}

func (m *LocalManager) Refresh() error {
	if m.KeyStrength == 0 {
		m.KeyStrength = 4096
	}

	key, err := rsa.GenerateKey(rand.Reader, m.KeyStrength)
	if err != nil {
		return errors.WithStack(err)
	}

	m.key = key
	return nil
}

func (m *LocalManager) PublicKey() (*rsa.PublicKey, error) {
	if m.key == nil {
		if err := m.Refresh(); err != nil {
			return nil, err
		}
	}
	return &m.key.PublicKey, nil
}

func (m *LocalManager) PrivateKey() (*rsa.PrivateKey, error) {
	if m.key == nil {
		if err := m.Refresh(); err != nil {
			return nil, err
		}
	}
	return m.key, nil
}

func (m *LocalManager) PublicKeyID() string {
	return "id-token:public"
}

func (m *LocalManager) Algorithm() string {
	return "RS256"
}
