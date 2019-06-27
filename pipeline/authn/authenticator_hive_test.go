package authn_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/julienschmidt/httprouter"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/herodot"
	"github.com/ory/hive-cloud/hive/identity"
	"github.com/ory/hive-cloud/hive/session"

	"github.com/ory/oathkeeper/driver/configuration"
	"github.com/ory/oathkeeper/helper"
	"github.com/ory/oathkeeper/internal"
	"github.com/ory/oathkeeper/pipeline/authn"
)

func TestAuthenticatorHive(t *testing.T) {

	writer := herodot.NewJSONWriter(logrus.New())
	var sessionHandler func(w http.ResponseWriter, r *http.Request, _ httprouter.Params)

	h := httprouter.New()
	h.GET("/session", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		sessionHandler(w, r, ps)
	})
	h.GET("/login", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		_, _ = fmt.Fprint(w, "login")
	})
	ts := httptest.NewServer(h)
	defer ts.Close()

	conf := internal.NewConfigurationWithDefaults()
	viper.Set(configuration.ViperKeyAuthenticatorHiveIsEnabled, true)
	viper.Set(configuration.ViperKeyAuthenticatorHiveSessionCheckURL, ts.URL + "/session")
	viper.Set(configuration.ViperKeyAuthenticatorHiveLoginURL, ts.URL + "/login")
	reg := internal.NewRegistry(conf)
	pa, _ := reg.PipelineAuthenticator("hive")
	a := pa.(*authn.AuthenticatorHive)

	t.Run("sub=FindSession", func(t *testing.T) {
		for k, tc := range []struct {
			sh            func(w http.ResponseWriter, r *http.Request, _ httprouter.Params)
			cv            string
			expectErr     error
			expectSession *session.Session
		}{
			{
				sh: func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
					w.WriteHeader(http.StatusInternalServerError)
				},
				cv:        "foobar",
				expectErr: &herodot.ErrInternalServerError,
			},
			{
				sh: func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
					w.WriteHeader(http.StatusUnauthorized)
				},
				cv:        "foobar",
				expectErr: helper.ErrUnauthorized,
			},
			{
				sh: func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
					writer.Write(w, r, &session.Session{
						SID: "1-session-id",
					})
				},
				cv: "1-session-id",
				expectSession: &session.Session{
					SID: "1-session-id",
				},
			},
			{
				sh: func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
					// This request should not be called because we've cached it already. We check if the request
					// is being called again.
					panic("This should not have been called")
				},
				cv: "1-session-id",
				expectSession: &session.Session{
					SID: "1-session-id",
				},
			},
		} {
			t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
				sessionHandler = tc.sh
				s, err := a.FindSession(&http.Cookie{Value: tc.cv})
				if tc.expectErr != nil {
					require.EqualError(t, tc.expectErr, err.Error())
				} else {
					require.NoError(t, err)
					assert.EqualValues(t, tc.expectSession, s)
				}
			})
		}
	})

	t.Run("sub=Authenticate", func(t *testing.T) {
		for k, tc := range []struct {
			d          string
			sh         func(w http.ResponseWriter, r *http.Request, _ httprouter.Params)
			cookie     *http.Cookie
			config     json.RawMessage
			expectErr  error
			expectSess *authn.AuthenticationSession
		}{
			{
				d:      "returns an authentication session when the session is found in hive",
				cookie: &http.Cookie{Value: "2-session-id", Name: session.DefaultSessionCookieName},
				sh: func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
					writer.Write(w, r, &session.Session{
						SID: "2-session-id",
						Identity: &identity.Identity{
							ID: "some-subject",
						},
					})
				},
				expectSess: &authn.AuthenticationSession{
					Subject: "some-subject",
				},
			},
			{
				d:      "when no cookie is set, return an unauthorized error",
				cookie: &http.Cookie{Value: "not-a-session", Name: "some-other-cookie"},
				sh: func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
					panic("should not have been called")
				},
				expectErr: helper.ErrForceResponse,
			},
		} {
			t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
				sessionHandler = tc.sh
				r := &http.Request{Header: http.Header{}}
				if tc.cookie != nil {
					r.AddCookie(tc.cookie)
				}

				s, err := a.Authenticate(r, json.RawMessage(tc.config), nil)

				if tc.expectErr != nil {
					require.EqualError(t, tc.expectErr, err.Error())
				} else {
					require.NoError(t, err, "%#v", err)
				}

				if tc.expectSess != nil {
					assert.Equal(t, tc.expectSess.Subject, s.Subject)
				}
			})
		}
	})

	t.Run("method=validate", func(t *testing.T) {
		viper.Set(configuration.ViperKeyAuthenticatorHiveIsEnabled, false)
		require.Error(t, a.Validate())

		viper.Set(configuration.ViperKeyAuthenticatorHiveIsEnabled, true)
		viper.Set(configuration.ViperKeyAuthenticatorHiveLoginURL, "")
		viper.Set(configuration.ViperKeyAuthenticatorHiveSessionCheckURL, "http://localhost/")
		require.Error(t, a.Validate())

		viper.Set(configuration.ViperKeyAuthenticatorHiveIsEnabled, true)
		viper.Set(configuration.ViperKeyAuthenticatorHiveLoginURL, "http://localhost/")
		viper.Set(configuration.ViperKeyAuthenticatorHiveSessionCheckURL, "")
		require.Error(t, a.Validate())

		viper.Set(configuration.ViperKeyAuthenticatorHiveIsEnabled, true)
		viper.Set(configuration.ViperKeyAuthenticatorHiveLoginURL, "http://localhost/")
		viper.Set(configuration.ViperKeyAuthenticatorHiveSessionCheckURL, "http://localhost/")
		require.NoError(t, a.Validate())
	})
}
