// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package api_test

import (
	"context"
	"crypto/rsa"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-jose/go-jose/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/x/configx"

	"github.com/ory/oathkeeper/driver/configuration"
	"github.com/ory/oathkeeper/internal"
	"github.com/ory/oathkeeper/rule"
	"github.com/ory/oathkeeper/x"
)

func TestCredentialsHandler(t *testing.T) {
	conf := internal.NewConfigurationWithDefaults(configx.SkipValidation())
	conf.SetForTest(t, configuration.MutatorIDTokenIsEnabled, true)
	conf.SetForTest(t, configuration.MutatorIDTokenJWKSURL, "file://../test/stub/jwks-rsa-multiple.json")
	conf.SetForTest(t, configuration.MutatorIDTokenIssuerURL, "https://example.com")
	conf.SetForTest(t, configuration.AuthenticatorAnonymousIsEnabled, true)
	conf.SetForTest(t, configuration.AuthorizerAllowIsEnabled, true)

	r := internal.NewRegistry(conf)

	require.NoError(t, r.RuleRepository().Set(
		context.Background(),
		[]rule.Rule{{
			Match:          &rule.Match{URL: "http://example.com/*"},
			Authenticators: []rule.Handler{{Handler: "anonymous"}},
			Authorizer:     rule.Handler{Handler: "allow"},
			Mutators: []rule.Handler{{
				Handler: "id_token",
				Config: json.RawMessage(`
{
	"jwks_url": "file://../test/stub/jwks-rsa-single.json",
	"issuer_url": "https://example.com"
}
`)}}}}),
	)

	router := x.NewAPIRouter()
	r.CredentialHandler().SetRoutes(router)
	server := httptest.NewServer(router)
	defer server.Close()

	res, err := server.Client().Get(server.URL + "/.well-known/jwks.json")
	require.NoError(t, err)
	defer res.Body.Close()
	require.Equal(t, http.StatusOK, res.StatusCode)

	var j jose.JSONWebKeySet
	require.NoError(t, json.NewDecoder(res.Body).Decode(&j))
	require.Len(t, j.Key("3e0edde4-12ad-425d-a783-135f46eac57e"), 1, "The public key should be broadcasted")
	require.Len(t, j.Key("81be3441-5303-4c52-b00d-bbdfadc75633"), 1, "The public key should be broadcasted")
	require.Len(t, j.Key("f4190122-ae96-4c29-8b79-56024e459d80"), 1, "The public key generated from the private key should be broadcasted")
	assert.IsType(t, new(rsa.PublicKey), j.Key("3e0edde4-12ad-425d-a783-135f46eac57e")[0].Key, "Ensure a public key")
	assert.IsType(t, new(rsa.PublicKey), j.Key("f4190122-ae96-4c29-8b79-56024e459d80")[0].Key, "Ensure a public key")
	assert.IsType(t, new(rsa.PublicKey), j.Key("81be3441-5303-4c52-b00d-bbdfadc75633")[0].Key, "Ensure a public key")
	assert.Len(t, j.Keys, 3, "There should not be any unexpected keys")
}
