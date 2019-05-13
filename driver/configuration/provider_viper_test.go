package configuration

import (
	"bytes"
	"fmt"
	"testing"
	"time"

	"github.com/ory/x/urlx"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAccessRuleRepositories(t *testing.T) {
	v := NewViperProvider(logrus.New())

	for k, tc := range []struct {
		in     string
		expect []AccessRuleRepository
	}{
		// this should not return anything because ttl is used without watch
		{
			in: `
access_rules:
  repositories:
    - url: "file://path/to/rules.json"
      ttl: 60s
`,
			expect: []AccessRuleRepository{},
		},
		// this should not return anything because ttl is used without watch
		{
			in: `
access_rules:
  repositories:
    - url: "http://host/path/rules.json"
      ttl: 60s
`,
			expect: []AccessRuleRepository{},
		},
		// this should pass and set the default ttl for https when watch true
		{
			in: `
access_rules:
  repositories:
    - url: "http://host/path/rules.json"
      watch: true
`,
			expect: []AccessRuleRepository{{URL: urlx.ParseOrPanic("http://host/path/rules.json"), Watch: true, TTL: time.Second*30}},
		},
		// this should not return anything because inline is misconfigured (watch not supported)
		{
			in: `
access_rules:
  repositories:
    - url: "inline://W10="
      watch: true
`,
			expect: []AccessRuleRepository{},
		},
		// this should not return anything because inline is misconfigured (ttl not supported)
		{
			in: `
access_rules:
  repositories:
    - url: "inline://W10="
      ttl: 60s
`,
			expect: []AccessRuleRepository{},
		},
		// this should pass
		{
			in: `
access_rules:
  repositories:
    - url: "file://path/to/rules.json"
      watch: true
      ttl: 60s
    - url: "http://host/path/rules.json"
      watch: true
      ttl: 60s
    - url: "inline://W10="
`,
			expect: []AccessRuleRepository{
				{URL: urlx.ParseOrPanic("file://path/to/rules.json"), Watch: true, TTL: time.Minute},
				{URL: urlx.ParseOrPanic("http://host/path/rules.json"), Watch: true, TTL: time.Minute},
				{URL: urlx.ParseOrPanic("inline://W10=")},
			},
		},
		{
			in: ``,
		},
	} {
		t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
			viper.SetConfigType("yml")
			require.NoError(t, viper.ReadConfig(bytes.NewBufferString(tc.in)))
			assert.EqualValues(t, tc.expect, v.AccessRuleRepositories())
		})
	}
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
