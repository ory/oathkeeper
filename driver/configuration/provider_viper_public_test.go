package configuration_test

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"testing"

	"github.com/rs/cors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/x/logrusx"

	"github.com/ory/x/urlx"
	"github.com/ory/x/viperx"

	"github.com/ory/viper"

	_ "github.com/ory/jsonschema/v3/fileloader"
	_ "github.com/ory/jsonschema/v3/httploader"

	. "github.com/ory/oathkeeper/driver/configuration"
	"github.com/ory/oathkeeper/pipeline/authn"
	"github.com/ory/oathkeeper/pipeline/authz"
	"github.com/ory/oathkeeper/pipeline/mutate"
	"github.com/ory/oathkeeper/x"
)

func setup(t *testing.T) *ViperProvider {
	l := logrusx.New("", "")
	viper.Reset()
	viperx.InitializeConfig(
		"oathkeeper",
		"./../../internal/config/",
		l,
	)

	err := viperx.ValidateFromURL("file://../../.schema/config.schema.json")
	if err != nil {
		l.WithError(err).Error("unable to validate")
	}
	require.NoError(t, err)

	return NewViperProvider(l)
}

func TestPipelineConfig(t *testing.T) {
	t.Run("case=should use config from environment variables", func(t *testing.T) {
		var res json.RawMessage
		require.NoError(t, os.Setenv("AUTHENTICATORS_OAUTH2_INTROSPECTION_CONFIG_INTROSPECTION_URL", "https://override/path"))
		p := setup(t)

		require.NoError(t, p.PipelineConfig("authenticators", "oauth2_introspection", nil, &res))
		assert.JSONEq(t, `{"cache":{"enabled":false},"introspection_url":"https://override/path","pre_authorization":{"client_id":"some_id","client_secret":"some_secret","enabled":true,"scope":["foo","bar"],"token_url":"https://my-website.com/oauth2/token"},"retry":{"max_delay":"100ms", "give_up_after":"1s"},"scope_strategy":"exact"}`, string(res), "%s", res)

		// Cleanup
		require.NoError(t, os.Setenv("AUTHENTICATORS_OAUTH2_INTROSPECTION_CONFIG_INTROSPECTION_URL", ""))

	})

	t.Run("case=should fail when invalid value is used in override", func(t *testing.T) {
		p := setup(t)
		res := json.RawMessage{}
		require.Error(t, p.PipelineConfig("mutators", "hydrator", json.RawMessage(`{"not-api":"invalid"}`), &res))
		assert.JSONEq(t, `{"cache":{"enabled":false,"ttl":"1m"},"api":{"url":"https://some-url/","retry":{"give_up_after":"1s","max_delay":"100ms"}},"not-api":"invalid"}`, string(res))

		require.Error(t, p.PipelineConfig("mutators", "hydrator", json.RawMessage(`{"api":{"this-key-does-not-exist":true}}`), &res))
		assert.JSONEq(t, `{"cache":{"enabled":false,"ttl":"1m"},"api":{"url":"https://some-url/","this-key-does-not-exist":true,"retry":{"give_up_after":"1s","max_delay":"100ms"}}}`, string(res))

		require.Error(t, p.PipelineConfig("mutators", "hydrator", json.RawMessage(`{"api":{"url":"not-a-url"}}`), &res))
		assert.JSONEq(t, `{"cache":{"enabled":false,"ttl":"1m"},"api":{"url":"not-a-url","retry":{"give_up_after":"1s","max_delay":"100ms"}}}`, string(res))
	})

	t.Run("case=should pass and override values", func(t *testing.T) {
		var dec mutate.MutatorHydratorConfig
		p := setup(t)
		require.NoError(t, p.PipelineConfig("mutators", "hydrator", json.RawMessage(``), &dec))
		assert.Equal(t, "https://some-url/", dec.Api.URL)

		require.NoError(t, p.PipelineConfig("mutators", "hydrator", json.RawMessage(`{"api":{"url":"http://override-url/foo","retry":{"give_up_after":"15s"}}}`), &dec))
		assert.Equal(t, "http://override-url/foo", dec.Api.URL)
		assert.Equal(t, "15s", dec.Api.Retry.GiveUpAfter)
	})

	t.Run("case=should pass array values", func(t *testing.T) {
		var dec authn.AuthenticatorOAuth2JWTConfiguration
		p := setup(t)
		require.NoError(t, p.PipelineConfig("authenticators", "jwt", json.RawMessage(`{}`), &dec))
		assert.Equal(t,
			[]string{"https://my-website.com/.well-known/jwks.json", "https://my-other-website.com/.well-known/jwks.json", "file://path/to/local/jwks.json"},
			dec.JWKSURLs,
		)

		require.NoError(t, p.PipelineConfig("authenticators", "jwt", json.RawMessage(`{"jwks_urls":["http://foo"]}`), &dec))
		assert.Equal(t,
			[]string{"http://foo"},
			dec.JWKSURLs,
		)
	})
}

/*
go test ./... -v -bench=. -run BenchmarkPipelineConfig -benchtime=10s

v0.35.1
594	  20119202 ns/op

v0.35.2
3048037	      3908 ns/op
*/

func BenchmarkPipelineConfig(b *testing.B) {
	viper.Reset()
	l := logrusx.New("", "")
	viperx.InitializeConfig(
		"oathkeeper",
		"./../../internal/config/",
		l,
	)

	err := viperx.ValidateFromURL("file://../../.schema/config.schema.json")
	if err != nil {
		l.WithError(err).Error("unable to validate")
	}
	require.NoError(b, err)

	p := NewViperProvider(logrusx.New("", ""))

	for n := 0; n < b.N; n++ {
		res := json.RawMessage{}
		p.PipelineConfig("authenticators", "oauth2_introspection", nil, &res)
	}
}

/*
go test ./... -v -bench=. -run BenchmarkPipelineEnabled -benchtime=10s

v0.35.4
11708	   1009975 ns/op

v0.35.5
18848694	       603 ns/op
*/

func BenchmarkPipelineEnabled(b *testing.B) {
	viper.Reset()
	logger := logrusx.New("", "")
	viperx.InitializeConfig(
		"oathkeeper",
		"./../../internal/config/",
		logger,
	)

	err := viperx.ValidateFromURL("file://../../.schema/config.schema.json")
	if err != nil {
		logger.WithError(err).Error("unable to validate")
	}
	require.NoError(b, err)

	p := NewViperProvider(logrusx.New("", ""))

	for n := 0; n < b.N; n++ {
		p.AuthorizerIsEnabled("allow")
		p.AuthenticatorIsEnabled("noop")
		p.MutatorIsEnabled("noop")
	}
}

func TestViperProvider(t *testing.T) {
	viper.Reset()
	logger := logrusx.New("", "")
	viperx.InitializeConfig(
		"oathkeeper",
		"./../../internal/config/",
		logger,
	)

	err := viperx.ValidateFromURL("file://../../.schema/config.schema.json")
	if err != nil {
		logger.WithError(err).Error("unable to validate")
	}
	p := NewViperProvider(logrusx.New("", ""))

	t.Run("group=serve", func(t *testing.T) {
		assert.Equal(t, "127.0.0.1:1234", p.ProxyServeAddress())
		assert.Equal(t, "127.0.0.2:1235", p.APIServeAddress())

		t.Run("group=prometheus", func(t *testing.T) {
			assert.Equal(t, "localhost:9000", p.PrometheusServeAddress())
			assert.Equal(t, "/metrics", p.PrometheusMetricsPath())
			assert.Equal(t, true, p.PrometheusCollapseRequestPaths())
		})

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
			a := authn.NewAuthenticatorAnonymous(p)
			assert.True(t, p.AuthenticatorIsEnabled(a.GetID()))
			require.NoError(t, a.Validate(nil))

			config, err := a.Config(nil)
			require.NoError(t, err)
			assert.Equal(t, "guest", config.Subject)
		})

		t.Run("authenticator=noop", func(t *testing.T) {
			a := authn.NewAuthenticatorNoOp(p)
			assert.True(t, p.AuthenticatorIsEnabled(a.GetID()))
			require.NoError(t, a.Validate(nil))
		})

		t.Run("authenticator=cookie_session", func(t *testing.T) {
			a := authn.NewAuthenticatorCookieSession(p)
			assert.True(t, p.AuthenticatorIsEnabled(a.GetID()))
			require.NoError(t, a.Validate(nil))

			config, err := a.Config(nil)
			require.NoError(t, err)

			assert.Equal(t, []string{"sessionid"}, config.Only)
			assert.Equal(t, "https://session-store-host", config.CheckSessionURL)
		})

		t.Run("authenticator=jwt", func(t *testing.T) {
			a := authn.NewAuthenticatorJWT(p, nil)
			assert.True(t, p.AuthenticatorIsEnabled(a.GetID()))
			require.NoError(t, a.Validate(nil))

			config, err := a.Config(nil)
			require.NoError(t, err)

			assert.Equal(t, "wildcard", config.ScopeStrategy)
			assert.Equal(t, []string{
				"https://my-website.com/.well-known/jwks.json",
				"https://my-other-website.com/.well-known/jwks.json",
				"file://path/to/local/jwks.json",
			}, config.JWKSURLs)
		})

		t.Run("authenticator=oauth2_client_credentials", func(t *testing.T) {
			a := authn.NewAuthenticatorOAuth2ClientCredentials(p)
			assert.True(t, p.AuthenticatorIsEnabled(a.GetID()))
			require.NoError(t, a.Validate(nil))

			config, err := a.Config(nil)
			require.NoError(t, err)
			assert.Equal(t, "https://my-website.com/oauth2/token", config.TokenURL)
		})

		t.Run("authenticator=oauth2_introspection", func(t *testing.T) {
			a := authn.NewAuthenticatorOAuth2Introspection(p)
			assert.True(t, p.AuthenticatorIsEnabled(a.GetID()))
			require.NoError(t, a.Validate(nil))

			config, err := a.Config(nil)
			require.NoError(t, err)
			assert.Equal(t, "https://my-website.com/oauth2/introspection", config.IntrospectionURL)
			assert.Equal(t, "exact", config.ScopeStrategy)
			assert.Equal(t, &authn.AuthenticatorOAuth2IntrospectionPreAuthConfiguration{
				ClientID:     "some_id",
				ClientSecret: "some_secret",
				TokenURL:     "https://my-website.com/oauth2/token",
				Scope:        []string{"foo", "bar"},
				Enabled:      true,
			}, config.PreAuth)
		})

		t.Run("authenticator=unauthorized", func(t *testing.T) {
			a := authn.NewAuthenticatorUnauthorized(p)
			assert.True(t, p.AuthenticatorIsEnabled(a.GetID()))
		})
	})

	t.Run("group=authorizers", func(t *testing.T) {
		t.Run("authorizer=allow", func(t *testing.T) {
			a := authz.NewAuthorizerAllow(p)
			assert.True(t, p.AuthorizerIsEnabled(a.GetID()))
			require.NoError(t, a.Validate(nil))
		})

		t.Run("authorizer=deny", func(t *testing.T) {
			a := authz.NewAuthorizerDeny(p)
			assert.True(t, p.AuthorizerIsEnabled(a.GetID()))
			require.NoError(t, a.Validate(nil))
		})

		t.Run("authorizer=keto_engine_acp_ory", func(t *testing.T) {
			a := authz.NewAuthorizerKetoEngineACPORY(p)
			assert.True(t, p.AuthorizerIsEnabled(a.GetID()))
			require.NoError(t, a.Validate(nil))

			config, err := a.Config(nil)
			require.NoError(t, err)

			assert.EqualValues(t, "http://my-keto/", config.BaseURL)
		})

		t.Run("authorizer=remote_json", func(t *testing.T) {
			a := authz.NewAuthorizerRemoteJSON(p)
			assert.True(t, p.AuthorizerIsEnabled(a.GetID()))
			require.NoError(t, a.Validate(nil))

			config, err := a.Config(nil)
			require.NoError(t, err)

			assert.EqualValues(t, "https://host/path", config.Remote)
			assert.EqualValues(t, "{}", config.Payload)
		})
	})

	t.Run("group=mutators", func(t *testing.T) {
		t.Run("mutator=noop", func(t *testing.T) {
			a := mutate.NewMutatorNoop(p)
			assert.True(t, p.MutatorIsEnabled(a.GetID()))
			require.NoError(t, a.Validate(nil))
		})

		t.Run("mutator=cookie", func(t *testing.T) {
			a := mutate.NewMutatorCookie(p)
			assert.True(t, p.MutatorIsEnabled(a.GetID()))
			require.NoError(t, a.Validate(nil))
		})

		t.Run("mutator=header", func(t *testing.T) {
			a := mutate.NewMutatorHeader(p)
			assert.True(t, p.MutatorIsEnabled(a.GetID()))
			require.NoError(t, a.Validate(nil))
		})

		t.Run("mutator=hydrator", func(t *testing.T) {
			a := mutate.NewMutatorHydrator(p, new(x.TestLoggerProvider))
			assert.True(t, p.MutatorIsEnabled(a.GetID()))
			require.NoError(t, a.Validate(nil))
		})

		t.Run("mutator=id_token", func(t *testing.T) {
			a := mutate.NewMutatorIDToken(p, nil)
			assert.True(t, p.MutatorIsEnabled(a.GetID()))
			require.NoError(t, a.Validate(nil))

			config, err := a.Config(nil)
			require.NoError(t, err)

			assert.EqualValues(t, "https://my-oathkeeper/", config.IssuerURL)
			assert.EqualValues(t, "https://fetch-keys/from/this/location.json", config.JWKSURL)
			assert.EqualValues(t, "1h", config.TTL)
		})
	})
}

func TestToScopeStrategy(t *testing.T) {
	v := NewViperProvider(logrusx.New("", ""))

	assert.True(t, v.ToScopeStrategy("exact", "foo")([]string{"foo"}, "foo"))
	assert.True(t, v.ToScopeStrategy("hierarchic", "foo")([]string{"foo"}, "foo.bar"))
	assert.True(t, v.ToScopeStrategy("wildcard", "foo")([]string{"foo.*"}, "foo.bar"))
	assert.Nil(t, v.ToScopeStrategy("none", "foo"))
	assert.Nil(t, v.ToScopeStrategy("whatever", "foo"))
}

func TestAuthenticatorOAuth2TokenIntrospectionPreAuthorization(t *testing.T) {
	viper.Reset()
	v := NewViperProvider(logrusx.New("", ""))
	viper.Set("authenticators.oauth2_introspection.enabled", true)
	viper.Set("authenticators.oauth2_introspection.config.introspection_url", "http://some-url/")

	for k, tc := range []struct {
		enabled bool
		id      string
		secret  string
		turl    string
		err     bool
	}{
		{enabled: true, id: "", secret: "", turl: "", err: true},
		{enabled: true, id: "a", secret: "", turl: "", err: true},
		{enabled: true, id: "", secret: "b", turl: "", err: true},
		{enabled: true, id: "", secret: "", turl: "c", err: true},
		{enabled: true, id: "a", secret: "b", turl: "", err: true},
		{enabled: true, id: "", secret: "b", turl: "c", err: true},
		{enabled: true, id: "a", secret: "", turl: "c", err: true},
		{enabled: false, id: "a", secret: "b", turl: "c", err: true},
		{enabled: true, id: "a", secret: "b", turl: "https://some-url", err: false},
	} {
		t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
			a := authn.NewAuthenticatorOAuth2Introspection(v)

			config, err := a.Config(json.RawMessage(fmt.Sprintf(`{
	"pre_authorization": {
		"enabled": %v,
		"client_id": "%v",
		"client_secret": "%v",
		"token_url": "%v"
	}
}`, tc.enabled, tc.id, tc.secret, tc.turl)))

			if tc.err {
				assert.Error(t, err, "%+v", config)
			} else {
				assert.NoError(t, err, "%+v", config)
			}
		})
	}
}
