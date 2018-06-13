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
	"github.com/pborman/uuid"
)

type LocalHS256Manager struct {
	key []byte
	kid string
}

func NewLocalHS256Manager(key []byte) *LocalHS256Manager {
	return &LocalHS256Manager{
		key: key,
		kid: uuid.New(),
	}
}

func (m *LocalHS256Manager) Refresh() error {
	return nil
}

func (m *LocalHS256Manager) PublicKey() (interface{}, error) {
	return m.key, nil
}

func (m *LocalHS256Manager) PrivateKey() (interface{}, error) {
	return m.key, nil
}

func (m *LocalHS256Manager) PublicKeyID() string {
	return m.kid
}

func (m *LocalHS256Manager) Algorithm() string {
	return "HS256"
}
