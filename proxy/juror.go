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

package proxy

import (
	"net/http"
	"net/url"

	"github.com/ory/oathkeeper/rule"
	"github.com/sirupsen/logrus"
)

type Jury []Juror

func (j Jury) GetIDs() (ids []string) {
	for _, l := range j {
		ids = append(ids, l.GetID())
	}
	return ids
}

type Juror interface {
	GetID() string
	Try(*http.Request, *rule.Rule, *url.URL) (*Session, error)
}

func toLogFields(r *http.Request, u *url.URL, granted bool, rl *rule.Rule, user string) logrus.Fields {
	_, tokenID := getBearerToken(r)
	return map[string]interface{}{
		"access_url": u.String(),
		"granted":    granted,
		"rule":       rl.ID,
		"mode":       rl.Mode,
		"user":       user,
		"token":      tokenID,
	}
}
