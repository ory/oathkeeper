// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package mutate_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/julienschmidt/httprouter"

	"github.com/ory/oathkeeper/pipeline/authn"
	"github.com/ory/oathkeeper/pipeline/mutate"
	"github.com/ory/oathkeeper/rule"
	"github.com/ory/x/configx"

	"github.com/ory/oathkeeper/driver/configuration"
	"github.com/ory/oathkeeper/internal"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setExtra(key string, value interface{}) func(a *authn.AuthenticationSession) {
	return func(a *authn.AuthenticationSession) {
		if a.Extra == nil {
			a.Extra = make(map[string]interface{})
		}
		a.Extra[key] = value
	}
}

func setSubject(subject string) func(a *authn.AuthenticationSession) {
	return func(a *authn.AuthenticationSession) {
		a.Subject = subject
	}
}

func setMatchContext(groups []string) func(a *authn.AuthenticationSession) {
	return func(a *authn.AuthenticationSession) {
		a.MatchContext = authn.MatchContext{
			RegexpCaptureGroups: groups,
		}
	}
}

func newAuthenticationSession(modifications ...func(a *authn.AuthenticationSession)) *authn.AuthenticationSession {
	a := authn.AuthenticationSession{}
	for _, f := range modifications {
		f(&a)
	}
	return &a
}

type routerSetupFunction func(t *testing.T) http.Handler

func defaultRouterSetup(actions ...func(a *authn.AuthenticationSession)) routerSetupFunction {
	return func(t *testing.T) http.Handler {
		router := httprouter.New()
		router.POST("/", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
			body, err := io.ReadAll(r.Body)
			require.NoError(t, err)
			var data authn.AuthenticationSession
			err = json.Unmarshal(body, &data)
			require.NoError(t, err)
			for _, f := range actions {
				f(&data)
			}
			jsonData, err := json.Marshal(data)
			require.NoError(t, err)
			w.WriteHeader(http.StatusOK)
			_, err = w.Write(jsonData)
			require.NoError(t, err)
		})
		return router
	}
}

func withBasicAuth(f routerSetupFunction, user, password string) routerSetupFunction {
	return func(t *testing.T) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			u, p, ok := r.BasicAuth()
			if !ok || u != user || p != password {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			h := f(t)
			h.ServeHTTP(w, r)
		})
	}
}

func withInitialErrors(f routerSetupFunction, numberOfErrorResponses, httpStatusCode int) routerSetupFunction {
	return func(t *testing.T) http.Handler {
		counter := 0
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if counter < numberOfErrorResponses {
				w.WriteHeader(httpStatusCode)
				counter++
				return
			}
			h := f(t)
			h.ServeHTTP(w, r)
		})
	}
}

func defaultConfigForMutator() func(*httptest.Server) json.RawMessage {
	return func(s *httptest.Server) json.RawMessage {
		return []byte(fmt.Sprintf(`{"api": {"url": "%s"}}`, s.URL))
	}
}

func configWithBasicAuthnForMutator(user, password string) func(*httptest.Server) json.RawMessage {
	return func(s *httptest.Server) json.RawMessage {
		return []byte(fmt.Sprintf(`{"api": {"url": "%s", "auth": {"basic": {"username": "%s", "password": "%s"}}}}`, s.URL, user, password))
	}
}

func configWithRetriesForMutator(giveUpAfter, retryDelay string) func(*httptest.Server) json.RawMessage {
	return func(s *httptest.Server) json.RawMessage {
		return []byte(fmt.Sprintf(`{"api": {"url": "%s", "retry": {"give_up_after": "%s", "max_delay": "%s"}}}`, s.URL, giveUpAfter, retryDelay))
	}
}

func TestMutatorHydrator(t *testing.T) {
	t.Parallel()
	conf := internal.NewConfigurationWithDefaults(configx.SkipValidation())
	reg := internal.NewRegistry(conf)

	a, err := reg.PipelineMutator("hydrator")
	require.NoError(t, err)
	assert.Equal(t, "hydrator", a.GetID())

	t.Run("method=mutate", func(t *testing.T) {
		sampleSubject := "sub"
		sampleKey := "foo"
		sampleValue := "bar"
		complexValueKey := "complex"
		sampleComplexValue := map[string]interface{}{
			"foo": "hello",
			"bar": 3.14,
		}
		sampleCaptureGroups := []string{"resource", "context"}
		sampleUserId := "user"
		sampleValidPassword := "passwd1"
		sampleNotValidPassword := "passwd7"
		var testMap = map[string]struct {
			Setup   func(*testing.T) http.Handler
			Session *authn.AuthenticationSession
			Rule    *rule.Rule
			Config  func(*httptest.Server) json.RawMessage
			Request *http.Request
			Match   *authn.AuthenticationSession
			Err     error
		}{
			"Extras From API": {
				Setup:   defaultRouterSetup(setExtra(sampleKey, sampleValue)),
				Session: newAuthenticationSession(),
				Rule:    &rule.Rule{ID: "test-rule"},
				Config:  defaultConfigForMutator(),
				Request: &http.Request{},
				Match:   newAuthenticationSession(setExtra(sampleKey, sampleValue)),
				Err:     nil,
			},
			"Override Extras": {
				Setup:   defaultRouterSetup(setExtra(sampleKey, sampleValue)),
				Session: newAuthenticationSession(setExtra(sampleKey, "initialValue")),
				Rule:    &rule.Rule{ID: "test-rule"},
				Config:  defaultConfigForMutator(),
				Request: &http.Request{},
				Match:   newAuthenticationSession(setExtra(sampleKey, sampleValue)),
				Err:     nil,
			},
			"Multiple Nested Extras": {
				Setup:   defaultRouterSetup(setExtra(sampleKey, sampleValue), setExtra(complexValueKey, sampleComplexValue)),
				Session: newAuthenticationSession(),
				Rule:    &rule.Rule{ID: "test-rule"},
				Config:  defaultConfigForMutator(),
				Request: &http.Request{},
				Match:   newAuthenticationSession(setExtra(sampleKey, sampleValue), setExtra(complexValueKey, sampleComplexValue)),
				Err:     nil,
			},
			"No Changes": {
				Setup:   defaultRouterSetup(),
				Session: newAuthenticationSession(setExtra(sampleKey, sampleValue), setMatchContext(sampleCaptureGroups)),
				Rule:    &rule.Rule{ID: "test-rule"},
				Config:  defaultConfigForMutator(),
				Request: &http.Request{},
				Match:   newAuthenticationSession(setExtra(sampleKey, sampleValue), setMatchContext(sampleCaptureGroups)),
				Err:     nil,
			},
			"No Extra Before And After": {
				Setup:   defaultRouterSetup(),
				Session: newAuthenticationSession(setSubject(sampleSubject)),
				Rule:    &rule.Rule{ID: "test-rule"},
				Config:  defaultConfigForMutator(),
				Request: &http.Request{},
				Match:   newAuthenticationSession(setSubject(sampleSubject)),
				Err:     nil,
			},
			"Empty Response": {
				Setup: func(t *testing.T) http.Handler {
					router := httprouter.New()
					router.POST("/", func(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
						w.WriteHeader(http.StatusOK)
						_, err = w.Write([]byte(`{}`))
						require.NoError(t, err)
					})
					return router
				},
				Session: newAuthenticationSession(setSubject(sampleSubject)),
				Rule:    &rule.Rule{ID: "test-rule"},
				Config:  defaultConfigForMutator(),
				Request: &http.Request{},
				Match:   newAuthenticationSession(setSubject(sampleSubject)),
				Err:     errors.New(mutate.ErrMalformedResponseFromUpstreamAPI),
			},
			"Missing API URL": {
				Setup:   defaultRouterSetup(),
				Session: newAuthenticationSession(),
				Rule:    &rule.Rule{ID: "test-rule"},
				Config: func(s *httptest.Server) json.RawMessage {
					return []byte(`{"api": {}}`)
				},
				Request: &http.Request{},
				Match:   newAuthenticationSession(),
				Err:     errors.New("mutator matching this route is misconfigured or disabled"),
			},
			"Improper Config": {
				Setup:   defaultRouterSetup(),
				Session: newAuthenticationSession(),
				Rule:    &rule.Rule{ID: "test-rule"},
				Config: func(s *httptest.Server) json.RawMessage {
					return []byte(`{"api": {"foo": "bar"}}`)
				},
				Request: &http.Request{},
				Match:   newAuthenticationSession(),
				Err:     errors.New("mutator matching this route is misconfigured or disabled"),
			},
			"Not Found": {
				Setup: func(_ *testing.T) http.Handler {
					router := httprouter.New()
					router.POST("/", func(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
						w.WriteHeader(http.StatusNotFound)
					})
					return router
				},
				Session: newAuthenticationSession(),
				Rule:    &rule.Rule{ID: "test-rule"},
				Config:  defaultConfigForMutator(),
				Request: &http.Request{},
				Match:   newAuthenticationSession(),
				Err:     errors.New("The call to an external API returned a non-200 HTTP response"),
			},
			"Wrong API URL": {
				Setup:   defaultRouterSetup(),
				Session: newAuthenticationSession(),
				Rule:    &rule.Rule{ID: "test-rule"},
				Config: func(s *httptest.Server) json.RawMessage {
					return []byte(`{"api": {"url": "ZGVmaW5pdGVseU5vdFZhbGlkVXJs"}}`)
				},
				Request: &http.Request{},
				Match:   newAuthenticationSession(),
				Err:     errors.New("mutator matching this route is misconfigured or disabled"),
			},
			"Successful Basic Authentication": {
				Setup:   withBasicAuth(defaultRouterSetup(setExtra(sampleKey, sampleValue)), sampleUserId, sampleValidPassword),
				Session: newAuthenticationSession(),
				Rule:    &rule.Rule{ID: "test-rule"},
				Config:  configWithBasicAuthnForMutator(sampleUserId, sampleValidPassword),
				Request: &http.Request{},
				Match:   newAuthenticationSession(setExtra(sampleKey, sampleValue)),
				Err:     nil,
			},
			"Invalid Basic Credentials": {
				Setup:   withBasicAuth(defaultRouterSetup(setExtra(sampleKey, sampleValue)), sampleUserId, sampleValidPassword),
				Session: newAuthenticationSession(),
				Rule:    &rule.Rule{ID: "test-rule"},
				Config:  configWithBasicAuthnForMutator(sampleUserId, sampleNotValidPassword),
				Request: &http.Request{},
				Match:   newAuthenticationSession(),
				Err:     errors.New(mutate.ErrInvalidCredentials),
			},
			"No Basic Credentials": {
				Setup:   withBasicAuth(defaultRouterSetup(setExtra(sampleKey, sampleValue)), sampleUserId, sampleValidPassword),
				Session: newAuthenticationSession(),
				Rule:    &rule.Rule{ID: "test-rule"},
				Config:  defaultConfigForMutator(),
				Request: &http.Request{},
				Match:   newAuthenticationSession(),
				Err:     errors.New(mutate.ErrNoCredentialsProvided),
			},
			"Should Replace Authn Header": {
				Setup: func(t *testing.T) http.Handler {
					router := httprouter.New()
					router.POST("/", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
						authnHeaders := r.Header["Authentication"]
						assert.Equal(t, len(authnHeaders), 1)
						user, passwd, ok := r.BasicAuth()
						assert.True(t, ok)
						assert.Equal(t, user, sampleUserId)
						assert.Equal(t, passwd, sampleValidPassword)
						h := defaultRouterSetup(setExtra(sampleKey, sampleValue))(t)
						h.ServeHTTP(w, r)
					})
					return router
				},
				Session: newAuthenticationSession(),
				Rule:    &rule.Rule{ID: "test-rule"},
				Config:  configWithBasicAuthnForMutator(sampleUserId, sampleValidPassword),
				Request: &http.Request{Header: http.Header{"Authentication": []string{"Bearer sample"}}},
				Match:   newAuthenticationSession(setExtra(sampleKey, sampleValue)),
				Err:     nil,
			},
			"Third Time Lucky": {
				Setup:   withInitialErrors(defaultRouterSetup(setExtra(sampleKey, sampleValue)), 2, http.StatusInternalServerError),
				Session: newAuthenticationSession(),
				Rule:    &rule.Rule{ID: "test-rule"},
				Config:  configWithRetriesForMutator("1s", "100ms"),
				Request: &http.Request{},
				Match:   newAuthenticationSession(setExtra(sampleKey, sampleValue)),
				Err:     nil,
			},
			"Pass Query Parameters": {
				Setup: func(t *testing.T) http.Handler {
					router := httprouter.New()
					router.POST("/", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
						q := r.URL.Query()
						assert.Equal(t, len(q), 2)
						assert.Equal(t, q["a"], []string{"b"})
						assert.Equal(t, q["c"], []string{"&12"})

						_, err = w.Write([]byte(`{}`))
						require.NoError(t, err)
					})
					return router
				},
				Session: newAuthenticationSession(),
				Rule:    &rule.Rule{ID: "test-rule"},
				Config:  defaultConfigForMutator(),
				Request: &http.Request{URL: &url.URL{RawQuery: "a=b&c=%2612"}},
				Match:   newAuthenticationSession(),
				Err:     nil,
			},
		}

		for testName, specs := range testMap {
			t.Run(testName, func(t *testing.T) {
				var router http.Handler
				var ts *httptest.Server
				if specs.Setup != nil {
					router = specs.Setup(t)
				}
				ts = httptest.NewServer(router)
				defer ts.Close()

				err := a.Mutate(specs.Request, specs.Session, specs.Config(ts), specs.Rule)
				if specs.Err == nil {
					// Issuer must run without error
					require.NoError(t, err)
				} else {
					assert.EqualError(t, err, specs.Err.Error())
				}

				assert.Equal(t, specs.Match, specs.Session)
			})
		}

	})

	t.Run("method=validate", func(t *testing.T) {
		for k, testCase := range []struct {
			enabled    bool
			apiUrl     string
			shouldPass bool
		}{
			{enabled: false, shouldPass: false},
			{enabled: true, shouldPass: true, apiUrl: "http://api/bar"},
		} {
			t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
				conf.SetForTest(t, configuration.MutatorHydratorIsEnabled, testCase.enabled)

				err := a.Validate(json.RawMessage(`{"api":{"url":"` + testCase.apiUrl + `"}}`))
				if testCase.shouldPass {
					require.NoError(t, err)
				} else {
					require.Error(t, err)
				}
			})
		}
	})
}
