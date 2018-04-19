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

	"github.com/ory/oathkeeper/rule"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type Judge struct {
	Logger  logrus.FieldLogger
	Matcher rule.Matcher
	Jury    map[string]Juror
	Issuer  string
}

func NewJudge(l logrus.FieldLogger, m rule.Matcher, i string, jury Jury) *Judge {
	if l == nil {
		l = logrus.New()
	}

	j := &Judge{Matcher: m, Logger: l, Issuer: i, Jury: map[string]Juror{}}
	for _, juror := range jury {
		j.Jury[juror.GetID()] = juror
	}

	return j
}

func (d *Judge) EvaluateAccessRequest(r *http.Request) (*Session, error) {
	var u = *r.URL
	u.Host = r.Host
	u.Scheme = "http"
	if r.TLS != nil {
		u.Scheme = "https"
	}

	rl, err := d.Matcher.MatchRule(r.Method, &u)
	if err != nil {
		d.Logger.WithError(err).
			WithField("granted", false).
			WithField("user", "").
			WithField("access_url", u.String()).
			WithField("reason", "Unable to match a rule").
			WithField("reason_id", "no_rule_match").
			Warn("Access request denied")
		return nil, err
	}

	if juror, ok := d.Jury[rl.Mode]; ok {
		session, err := juror.Try(r, rl, &u)
		if err != nil {
			return nil, err
		}
		session.Issuer = d.Issuer
		return session, nil
	}

	d.Logger.WithError(err).
		WithField("granted", false).
		WithField("user", "").
		WithField("access_url", u.String()).
		WithField("mode", rl.Mode).
		WithField("reason", "Rule defines a unknown mode").
		WithField("reason_id", "unknown_mode").
		Warn("Access request denied")
	return nil, errors.Errorf("Unknown rule mode \"%s\"", rl.Mode)
}
