package configuration

import (
	"fmt"
	"net/url"
	"reflect"
	"testing"
	"time"

	"github.com/rs/cors"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2/clientcredentials"

	"github.com/ory/fosite"
	"github.com/ory/gojsonschema"
	"github.com/ory/x/urlx"
	"github.com/ory/x/viperx"

	"github.com/ory/viper"
)

func TestViperProvider(t *testing.T) {
	viper.Reset()
	BindEnvs()
	viperx.InitializeConfig(
		"oathkeeper",
		"./../../docs/",
		logrus.New(),
	)

	require.NoError(t, viperx.Validate(gojsonschema.NewReferenceLoader("file://../../.schemas/config.schema.json")))
	p := NewViperProvider(logrus.New())

	t.Run("group=serve", func(t *testing.T) {
		assert.Equal(t, "127.0.0.1:1234", p.ProxyServeAddress())
		assert.Equal(t, "127.0.0.2:1235", p.APIServeAddress())

		t.Run("group=cors", func(t *testing.T) {
			assert.True(t, p.CORSEnabled("proxy"))
			assert.True(t, p.CORSEnabled("api"))

			assert.Equal(t, cors.Options{
				AllowedOrigins:     []string{"https://example.com", "https://*.example.com"},
				AllowedMethods:     []string{"POST", "GET", "PUT", "PATCH", "DELETE"},
				AllowedHeaders:     []string{"Authorization", "Content-Type"},
				ExposedHeaders:     []string{"Content-Type"},
				MaxAge:             10,
				AllowCredentials:   true,
				OptionsPassthrough: false,
				Debug:              true,
			}, p.CORSOptions("proxy"))

			assert.Equal(t, cors.Options{
				AllowedOrigins:     []string{"https://example.org", "https://*.example.org"},
				AllowedMethods:     []string{"GET", "PUT", "PATCH", "DELETE"},
				AllowedHeaders:     []string{"Authorization", "Content-Type"},
				ExposedHeaders:     []string{"Content-Type"},
				MaxAge:             10,
				AllowCredentials:   true,
				OptionsPassthrough: false,
				Debug:              true,
			}, p.CORSOptions("api"))
		})

		t.Run("group=tls", func(t *testing.T) {
			for _, daemon := range []string{"proxy", "api"} {
				t.Run(fmt.Sprintf("daemon="+daemon), func(t *testing.T) {
					assert.Equal(t, "LS0tLS1CRUdJTiBFTkNSWVBURUQgUFJJVkFURSBLRVktLS0tLVxuTUlJRkRqQkFCZ2txaGtpRzl3MEJCUTB3...", viper.GetString("serve."+daemon+".tls.key.base64"))
					assert.Equal(t, "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tXG5NSUlEWlRDQ0FrMmdBd0lCQWdJRVY1eE90REFOQmdr...", viper.GetString("serve."+daemon+".tls.cert.base64"))
					assert.Equal(t, "/path/to/key.pem", viper.GetString("serve."+daemon+".tls.key.path"))
					assert.Equal(t, "/path/to/cert.pem", viper.GetString("serve."+daemon+".tls.cert.path"))
				})
			}
		})
	})

	t.Run("group=access_rules", func(t *testing.T) {
		assert.Equal(t, []url.URL{
			*urlx.ParseOrPanic("file://path/to/rules.json"),
			*urlx.ParseOrPanic("inline://W3siaWQiOiJmb28tcnVsZSIsImF1dGhlbnRpY2F0b3JzIjpbXX1d"),
			*urlx.ParseOrPanic("https://path-to-my-rules/rules.json"),
		}, p.AccessRuleRepositories())

	})

	t.Run("group=authenticators", func(t *testing.T) {
		t.Run("authenticator=anonymous", func(t *testing.T) {
			assert.True(t, p.AuthenticatorAnonymousIsEnabled())
			assert.Equal(t, "guest", p.AuthenticatorAnonymousIdentifier())
		})

		t.Run("authenticator=noop", func(t *testing.T) {
			assert.True(t, p.AuthenticatorNoopIsEnabled())
		})

		t.Run("authenticator=cookie_session", func(t *testing.T) {
			assert.True(t, p.AuthenticatorCookieSessionIsEnabled())
			assert.Equal(t, []string{"sessionid"}, p.AuthenticatorCookieSessionOnly())
			assert.Equal(t, urlx.ParseOrPanic("https://session-store-host"), p.AuthenticatorCookieSessionCheckSessionURL())
		})

		t.Run("authenticator=jwt", func(t *testing.T) {
			assert.True(t, p.AuthenticatorJWTIsEnabled())
			assert.True(t, reflect.ValueOf(fosite.WildcardScopeStrategy).Pointer() == reflect.ValueOf(p.AuthenticatorJWTScopeStrategy()).Pointer())
			assert.Equal(t, []url.URL{
				*urlx.ParseOrPanic("https://my-website.com/.well-known/jwks.json"),
				*urlx.ParseOrPanic("https://my-other-website.com/.well-known/jwks.json"),
				*urlx.ParseOrPanic("file://path/to/local/jwks.json"),
			}, p.AuthenticatorJWTJWKSURIs())
		})

		t.Run("authenticator=oauth2_client_credentials", func(t *testing.T) {
			assert.True(t, p.AuthenticatorOAuth2ClientCredentialsIsEnabled())
			assert.Equal(t, urlx.ParseOrPanic("https://my-website.com/oauth2/token"), p.AuthenticatorOAuth2ClientCredentialsTokenURL())
		})

		t.Run("authenticator=oauth2_introspection", func(t *testing.T) {
			assert.True(t, p.AuthenticatorOAuth2TokenIntrospectionIsEnabled())
			assert.Equal(t, urlx.ParseOrPanic("https://my-website.com/oauth2/introspection"), p.AuthenticatorOAuth2TokenIntrospectionIntrospectionURL())
			assert.True(t, reflect.ValueOf(fosite.ExactScopeStrategy).Pointer() == reflect.ValueOf(p.AuthenticatorOAuth2TokenIntrospectionScopeStrategy()).Pointer())
			assert.Equal(t, &clientcredentials.Config{
				ClientID:     "some_id",
				ClientSecret: "some_secret",
				TokenURL:     "https://my-website.com/oauth2/token",
				Scopes:       []string{"foo", "bar"},
				AuthStyle:    0,
			}, p.AuthenticatorOAuth2TokenIntrospectionPreAuthorization())
		})
		t.Run("authenticator=unauthorized", func(t *testing.T) {
			assert.True(t, p.AuthenticatorUnauthorizedIsEnabled())
		})
	})

	t.Run("group=authorizers", func(t *testing.T) {
		t.Run("authorizer=allow", func(t *testing.T) {
			assert.True(t, p.AuthorizerAllowIsEnabled())

		})

		t.Run("authorizer=deny", func(t *testing.T) {
			assert.True(t, p.AuthorizerDenyIsEnabled())
		})

		t.Run("authorizer=keto_engine_acp_ory", func(t *testing.T) {
			assert.True(t, p.AuthorizerKetoEngineACPORYIsEnabled())
			assert.EqualValues(t, urlx.ParseOrPanic("http://my-keto/"), p.AuthorizerKetoEngineACPORYBaseURL())
		})
	})

	t.Run("group=mutators", func(t *testing.T) {
		t.Run("mutator=noop", func(t *testing.T) {
			assert.True(t, p.MutatorNoopIsEnabled())
		})

		t.Run("mutator=cookie", func(t *testing.T) {
			assert.True(t, p.MutatorCookieIsEnabled())
		})

		t.Run("mutator=header", func(t *testing.T) {
			assert.True(t, p.MutatorHeaderIsEnabled())
		})

		t.Run("mutator=id_token", func(t *testing.T) {
			assert.True(t, p.MutatorIDTokenIsEnabled())
			assert.EqualValues(t, urlx.ParseOrPanic("https://my-oathkeeper/"), p.MutatorIDTokenIssuerURL())
			assert.EqualValues(t, urlx.ParseOrPanic("https://fetch-keys/from/this/location.json"), p.MutatorIDTokenJWKSURL())
			assert.EqualValues(t, time.Hour, p.MutatorIDTokenTTL())
		})
	})
}

func TestToScopeStrategy(t *testing.T) {
	v := NewViperProvider(logrus.New())

	assert.True(t, v.toScopeStrategy("exact", "foo")([]string{"foo"}, "foo"))
	assert.True(t, v.toScopeStrategy("hierarchic", "foo")([]string{"foo"}, "foo.bar"))
	assert.True(t, v.toScopeStrategy("wildcard", "foo")([]string{"foo.*"}, "foo.bar"))
	assert.Nil(t, v.toScopeStrategy("none", "foo"))
	assert.Nil(t, v.toScopeStrategy("whatever", "foo"))
}

func TestAuthenticatorOAuth2TokenIntrospectionPreAuthorization(t *testing.T) {
	viper.Reset()
	v := NewViperProvider(logrus.New())

	for k, tc := range []struct {
		enabled bool
		id      string
		secret  string
		turl    string
		ok      bool
	}{
		{enabled: true, id: "", secret: "", turl: "", ok: false},
		{enabled: true, id: "a", secret: "", turl: "", ok: false},
		{enabled: true, id: "", secret: "b", turl: "", ok: false},
		{enabled: true, id: "", secret: "", turl: "c", ok: false},
		{enabled: true, id: "a", secret: "b", turl: "", ok: false},
		{enabled: true, id: "", secret: "b", turl: "c", ok: false},
		{enabled: true, id: "a", secret: "", turl: "c", ok: false},
		{enabled: false, id: "a", secret: "b", turl: "c", ok: false},
		{enabled: true, id: "a", secret: "b", turl: "c", ok: true},
	} {
		t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
			viper.Set(ViperKeyAuthenticatorOAuth2TokenIntrospectionPreAuthorizationEnabled, tc.enabled)
			viper.Set(ViperKeyAuthenticatorOAuth2TokenIntrospectionPreAuthorizationClientID, tc.id)
			viper.Set(ViperKeyAuthenticatorOAuth2TokenIntrospectionPreAuthorizationClientSecret, tc.secret)
			viper.Set(ViperKeyAuthenticatorOAuth2TokenIntrospectionPreAuthorizationTokenURL, tc.turl)

			c := v.AuthenticatorOAuth2TokenIntrospectionPreAuthorization()
			if tc.ok {
				assert.NotNil(t, c)
			} else {
				assert.Nil(t, c)
			}
		})
	}
	v.AuthenticatorOAuth2TokenIntrospectionPreAuthorization()
}

func TestGetURL(t *testing.T) {
	v := NewViperProvider(logrus.New())
	assert.Nil(t, v.getURL("", "key"))
	assert.Nil(t, v.getURL("a", "key"))
}
