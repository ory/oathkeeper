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

package authz

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"text/template"
	"time"

	"github.com/ory/oathkeeper/driver/configuration"
	"github.com/ory/oathkeeper/pipeline"
	"github.com/ory/oathkeeper/pipeline/authn"
	"github.com/ory/x/httpx"

	"github.com/ory/x/urlx"

	"github.com/asaskevich/govalidator"
	"github.com/pkg/errors"
	"github.com/tomasen/realip"

	"github.com/ory/oathkeeper/helper"
)

type AuthorizerKetoWardenConfiguration struct {
	RequiredAction   string `json:"required_action" valid:",required"`
	RequiredResource string `json:"required_resource" valid:",required"`
	Subject          string `json:"subject"`
	Flavor           string `json:"flavor"`
}

type AuthorizerKetoWarden struct {
	c configuration.Provider

	client         *http.Client
	contextCreator authorizerKetoWardenContext
}

func NewAuthorizerKetoWarden(c configuration.Provider) *AuthorizerKetoWarden {
	return &AuthorizerKetoWarden{
		c:              c,
		client:         httpx.NewResilientClientLatencyToleranceSmall(nil),
		contextCreator: contextFromRequest,
	}
}

func (a *AuthorizerKetoWarden) GetID() string {
	return "keto_engine_acp_ory"
}

type authorizerKetoWardenContext func(r *http.Request) map[string]interface{}

func contextFromRequest(r *http.Request) map[string]interface{} {
	return map[string]interface{}{
		"remoteIpAddress": realip.RealIP(r),
		"requestedAt":     time.Now().UTC(),
	}
}

type ketoWardenInput struct {
	Action   string                 `json:"action"`
	Context  map[string]interface{} `json:"context"`
	Resource string                 `json:"resource"`
	Subject  string                 `json:"subject"`
}

func (a *AuthorizerKetoWarden) Authorize(r *http.Request, session *authn.AuthenticationSession, config json.RawMessage, rl pipeline.Rule) error {
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
		templateId := fmt.Sprintf("%s:%s", rl.GetID(), "subject")
		subject, err = a.ParseSubject(session, templateId, cf.Subject)
		if err != nil {
			return errors.WithStack(err)
		}
	}

	flavor := "regex"
	if len(cf.Flavor) > 0 {
		flavor = cf.Flavor
	}

	var b bytes.Buffer
	if err := json.NewEncoder(&b).Encode(&ketoWardenInput{
		Action:   compiled.ReplaceAllString(r.URL.String(), cf.RequiredAction),
		Resource: compiled.ReplaceAllString(r.URL.String(), cf.RequiredResource),
		Context:  a.contextCreator(r),
		Subject:  subject,
	}); err != nil {
		return errors.WithStack(err)
	}
	req, err := http.NewRequest("POST", urlx.AppendPaths(a.c.AuthorizerKetoEngineACPORYAuthorizedURL(), "/engines/acp/ory", flavor, "/allowed").String(), &b)
	if err != nil {
		return errors.WithStack(err)
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

	var result struct {
		Allowed bool `json:"allowed"`
	}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return errors.WithStack(err)
	}

	if !result.Allowed {
		return errors.WithStack(helper.ErrForbidden)
	}

	return nil
}

func (a *AuthorizerKetoWarden) ParseSubject(session *authn.AuthenticationSession, templateId, templateString string) (string, error) {
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

func (a *AuthorizerKetoWarden) Validate() error {
	if !a.c.AuthorizerKetoEngineACPORYIsEnabled() {
		return errors.WithStack(authn.ErrAuthenticatorNotEnabled.WithReasonf("Authorizer % is disabled per configuration.", a.GetID()))
	}

	if a.c.AuthorizerKetoEngineACPORYAuthorizedURL() == nil {
		return errors.WithStack(authn.ErrAuthenticatorNotEnabled.WithReasonf(`Configuration for authorizer % did not specify any values for configuration key "%s" and is thus disabled.`, a.GetID(), configuration.ViperKeyAuthorizerKetoEngineACPORYAuthorizedURL))
	}

	return nil
}
