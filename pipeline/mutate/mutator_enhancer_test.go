package mutate_test

import (
	"encoding/json"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"github.com/ory/oathkeeper/pipeline/authn"
	"github.com/ory/oathkeeper/pipeline/mutate"
	"github.com/ory/oathkeeper/rule"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ory/viper"

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

func newAuthenticationSession(modifications ...func(a *authn.AuthenticationSession)) *authn.AuthenticationSession {
	a := authn.AuthenticationSession{}
	for _, f := range modifications {
		f(&a)
	}
	return &a
}

func defaultRouterSetup(actions ...func(a *authn.AuthenticationSession)) func(t *testing.T, router *httprouter.Router) {
	return func(t *testing.T, router *httprouter.Router) {
		router.POST("/", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
			body, err := ioutil.ReadAll(r.Body)
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
			// TODO: Debug only
			t.Logf("json response from API: %s", string(jsonData))
		})
	}
}

func defaultConfigForMutator() func(*httptest.Server) json.RawMessage {
	return func(s *httptest.Server) json.RawMessage {
		return []byte(fmt.Sprintf(`{"api": {"url": "%s"}}`, s.URL))
	}
}

func TestMutatorEnhancer(t *testing.T) {
	conf := internal.NewConfigurationWithDefaults()
	reg := internal.NewRegistry(conf)

	a, err := reg.PipelineMutator("enhancer")
	require.NoError(t, err)
	assert.Equal(t, "enhancer", a.GetID())

	t.Run("method=mutate", func(t *testing.T) {
		sampleKey := "foo"
		sampleValue := "bar"
		complexValueKey := "complex"
		sampleComplexValue := struct {
			foo string
			oof int
			bar float32
			rab bool
		}{"hello", 7, 3.14, true}

		var testMap = map[string]struct {
			Setup   func(*testing.T, *httprouter.Router)
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
				Session: newAuthenticationSession(setExtra(sampleKey, sampleValue)),
				Rule:    &rule.Rule{ID: "test-rule"},
				Config:  defaultConfigForMutator(),
				Request: &http.Request{},
				Match:   newAuthenticationSession(setExtra(sampleKey, sampleValue)),
				Err:     nil,
			},
			"Empty Response": {
				Setup: func(t *testing.T, router *httprouter.Router) {
					router.POST("/", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
						w.WriteHeader(http.StatusOK)
						_, err = w.Write([]byte(`{}`))
						require.NoError(t, err)
					})
				},
				Session: newAuthenticationSession(),
				Rule:    &rule.Rule{ID: "test-rule"},
				Config:  defaultConfigForMutator(),
				Request: &http.Request{},
				Match:   newAuthenticationSession(),
				Err:     mutate.ErrMalformedResponseFromUpstreamAPI,
			},
			"Missing API URL": {
				Setup:   defaultRouterSetup(),
				Session: newAuthenticationSession(),
				Rule:    &rule.Rule{ID: "test-rule"},
				Config: func(s *httptest.Server) json.RawMessage {
					return []byte(`{"api": {"foo": "bar"}}`)
				},
				Request: &http.Request{},
				Match:   newAuthenticationSession(),
				Err:     mutate.ErrMissingAPIURL,
			},
			"Server Error": {
				Setup: func(t *testing.T, router *httprouter.Router) {
					router.POST("/", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
						w.WriteHeader(http.StatusInternalServerError)
					})
				},
				Session: newAuthenticationSession(),
				Rule:    &rule.Rule{ID: "test-rule"},
				Config:  defaultConfigForMutator(),
				Request: &http.Request{},
				Match:   newAuthenticationSession(),
				Err:     mutate.ErrNon200ResponseFromAPI,
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
				Err:     mutate.ErrInvalidAPIURL,
			},
		}
		t.Run("caching=off", func(t *testing.T) {
			for testName, specs := range testMap {
				t.Run(testName, func(t *testing.T) {
					router := httprouter.New()
					var ts *httptest.Server
					if specs.Setup != nil {
						specs.Setup(t, router)
					}
					ts = httptest.NewServer(router)
					defer ts.Close()

					_, err := a.Mutate(specs.Request, specs.Session, specs.Config(ts), specs.Rule)
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
		// TODO: add tests with caching
	})

	t.Run("method=validate", func(t *testing.T) {
		for k, testCase := range []struct {
			enabled    bool
			apiUrl     string
			shouldPass bool
		}{
			{enabled: false, shouldPass: false},
			{enabled: true, shouldPass: true},
		} {
			t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
				viper.Set(configuration.ViperKeyMutatorEnhancerIsEnabled, testCase.enabled)

				if testCase.shouldPass {
					require.NoError(t, a.Validate())
				} else {
					require.Error(t, a.Validate())
				}
			})
		}
	})
}
