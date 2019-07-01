package authz

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"text/template"

	"github.com/asaskevich/govalidator"

	"github.com/pkg/errors"

	"github.com/ory/oathkeeper/driver/configuration"
	"github.com/ory/oathkeeper/helper"
	"github.com/ory/oathkeeper/pipeline"
	"github.com/ory/oathkeeper/pipeline/authn"

	"github.com/ory/x/httpx"
)

type AuthorizerOPAEngineConfiguration struct {
	RequiredAction   string `json:"required_action" valid:",required"`
	RequiredResource string `json:"required_resource" valid:",required"`
	Subject          string `json:"subject"`
	Group            string `json:"group" valid:",required"`
}

type AuthorizerOPAEngine struct {
	c configuration.Provider

	client *http.Client
}

func NewAuthorizerOPAEngine(c configuration.Provider) *AuthorizerOPAEngine {
	return &AuthorizerOPAEngine{
		c:      c,
		client: httpx.NewResilientClientLatencyToleranceSmall(nil),
	}
}

func (a *AuthorizerOPAEngine) GetID() string {
	return "opa_engine"
}

type AuthorizerOPAEngineRequestBody struct {
	Action   string `json:"action"`
	Resource string `json:"resource"`
	Subject  string `json:"subject"`
	Group    string `json:"group"`
}

func (a *AuthorizerOPAEngine) Authorize(r *http.Request, session *authn.AuthenticationSession, config json.RawMessage, rule pipeline.Rule) error {
	var cf AuthorizerOPAEngineConfiguration

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
		return errors.New("Unable to valid opa warden configuration")
	}

	compiled, err := rule.CompileURL()
	if err != nil {
		return errors.WithStack(err)
	}

	subject := session.Subject
	if cf.Subject != "" {
		templateId := fmt.Sprintf("%s:%s", rule.GetID(), "subject")
		subject, err = a.ParseSubject(session, templateId, cf.Subject)
		if err != nil {
			return errors.WithStack(err)
		}
	}

	var b bytes.Buffer
	if err := json.NewEncoder(&b).Encode(&AuthorizerOPAEngineRequestBody{
		Action:   compiled.ReplaceAllString(r.URL.String(), cf.RequiredAction),
		Resource: compiled.ReplaceAllString(r.URL.String(), cf.RequiredResource),
		Group:    compiled.ReplaceAllString(r.URL.String(), cf.Group),
		Subject:  subject,
	}); err != nil {
		return errors.WithStack(err)
	}

	req, err := http.NewRequest("POST", a.c.AuthorizerOPAEngineBaseURL().String(), &b)
	if err != nil {
		return errors.WithStack(err)
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := a.client.Do(req)
	if err != nil {
		return errors.WithStack(err)
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusForbidden {
		return errors.WithStack(helper.ErrForbidden)
	} else if res.StatusCode != http.StatusOK {
		return errors.Errorf("Expected status code %d but got %d", http.StatusOK, res.StatusCode)
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

func (a *AuthorizerOPAEngine) ParseSubject(session *authn.AuthenticationSession, templateId, templateString string) (string, error) {
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

func (a *AuthorizerOPAEngine) Validate() error {
	if !a.c.AuthorizerOPAEngineIsEnabled() {
		return errors.WithStack(ErrAuthorizerNotEnabled.WithReasonf(`Authorizer "%s" is disabled per configuration.`, a.GetID()))
	}

	if a.c.AuthorizerOPAEngineBaseURL() == nil {
		return errors.WithStack(ErrAuthorizerNotEnabled.WithReasonf(`Configuration for authorizer "%s" did not specify any values for configuration key "%s" and is thus disabled.`, a.GetID(), configuration.ViperKeyAuthorizerOPAEngineBaseURL))
	}

	return nil
}
