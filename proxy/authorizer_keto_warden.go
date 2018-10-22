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
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"text/template"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/ory/keto/sdk/go/keto"
	"github.com/ory/keto/sdk/go/keto/swagger"
	"github.com/ory/oathkeeper/helper"
	"github.com/ory/oathkeeper/rule"
	"github.com/pkg/errors"
	"github.com/tomasen/realip"
)

type AuthorizerKetoWardenConfiguration struct {
	RequiredAction   string `json:"required_action" valid:",required"`
	RequiredResource string `json:"required_resource" valid:",required"`
	Subject          string `json:"subject"`
}

type AuthorizerKetoWarden struct {
	K              keto.WardenSDK
	contextCreator authorizerKetoWardenContext
}

func NewAuthorizerKetoWarden(k keto.WardenSDK) *AuthorizerKetoWarden {
	return &AuthorizerKetoWarden{
		K:              k,
		contextCreator: contextFromRequest,
	}
}

func (a *AuthorizerKetoWarden) GetID() string {
	return "keto_warden"
}

type authorizerKetoWardenContext func(r *http.Request) map[string]interface{}

func contextFromRequest(r *http.Request) map[string]interface{} {
	return map[string]interface{}{
		"remoteIpAddress": realip.RealIP(r),
		"requestedAt":     time.Now().UTC(),
	}
}

func (a *AuthorizerKetoWarden) Authorize(r *http.Request, session *AuthenticationSession, config json.RawMessage, rl *rule.Rule) error {
	var cf AuthorizerKetoWardenConfiguration

	if len(config) == 0 {
		config = []byte("{}")
	}

	d := json.NewDecoder(bytes.NewBuffer(config))
	d.DisallowUnknownFields()
	if err := d.Decode(&cf); err != nil {
		return errors.WithStack(err)
	}

	if result, err := govalidator.ValidateStruct(&cf); err != nil {
		return errors.WithStack(err)
	} else if !result {
		return errors.New("Unable to validate keto warden configuration")
	}

	compiled, err := rl.CompileURL()
	if err != nil {
		return errors.WithStack(err)
	}

	subject := session.Subject
	if cf.Subject != "" {
		templateId := fmt.Sprintf("%s:%s", rl.ID, "subject")
		subject, err = a.ParseSubject(session, templateId, cf.Subject)
		if err != nil {
			return errors.WithStack(err)
		}
	}

	defaultSession, response, err := a.K.IsSubjectAuthorized(swagger.WardenSubjectAuthorizationRequest{
		Action:   compiled.ReplaceAllString(r.URL.String(), cf.RequiredAction),
		Resource: compiled.ReplaceAllString(r.URL.String(), cf.RequiredResource),
		Context:  a.contextCreator(r),
		Subject:  subject,
	})
	if err != nil {
		return errors.WithStack(err)
	}
	if response.StatusCode != http.StatusOK {
		return errors.Errorf("Expected status code %d but got %d", http.StatusOK, response.StatusCode)
	}
	if defaultSession == nil {
		return errors.WithStack(helper.ErrUnauthorized)
	}
	if !defaultSession.Allowed {
		return errors.WithStack(helper.ErrUnauthorized)
	}

	return nil
}
func (a *AuthorizerKetoWarden) ParseSubject(session *AuthenticationSession, templateId, templateString string) (string, error) {
	tmplFn := template.New("rules").
		Option("missingkey=zero").
		Funcs(template.FuncMap{
			"print": func(i interface{}) string {
				if i == nil {
					return ""
				}
				return fmt.Sprintf("%v", i)
			},
		})

	tmpl, err := tmplFn.New(templateId).Parse(templateString)
	if err != nil {
		return "", err
	}

	subject := bytes.Buffer{}
	err = tmpl.Execute(&subject, session)
	if err != nil {
		return "", err
	}
	return subject.String(), nil
}
