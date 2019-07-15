package configuration

import (
	"fmt"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/ory/viper"
)

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
