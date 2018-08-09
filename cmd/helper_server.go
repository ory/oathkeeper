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

package cmd

import (
	"strings"
	"time"

	"github.com/ory/fosite"
	"github.com/ory/hydra/sdk/go/hydra"
	"github.com/ory/keto/sdk/go/keto"
	"github.com/ory/oathkeeper/proxy"
	"github.com/ory/oathkeeper/rsakey"
	"github.com/ory/oathkeeper/rule"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func getHydraSDK() hydra.SDK {
	sdk, err := hydra.NewSDK(&hydra.Configuration{
		ClientID:     viper.GetString("CREDENTIALS_ISSUER_ID_TOKEN_HYDRA_CLIENT_ID"),
		ClientSecret: viper.GetString("CREDENTIALS_ISSUER_ID_TOKEN_HYDRA_CLIENT_SECRET"),
		EndpointURL:  viper.GetString("CREDENTIALS_ISSUER_ID_TOKEN_HYDRA_URL"),
		Scopes:       strings.Split(viper.GetString("CREDENTIALS_ISSUER_ID_TOKEN_HYDRA_CLIENT_SCOPES"), ","),
	})

	if err != nil {
		logger.WithError(err).Fatalln("Unable to connect to Hydra SDK")
		return nil
	}
	return sdk
}

func refreshRules(m rule.Refresher, duration time.Duration) {
	if duration == 0 {
		duration, _ = time.ParseDuration(viper.GetString("RULES_REFRESH_INTERVAL"))
		if duration == 0 {
			duration = time.Second * 30
		}
	}

	var fails int
	for {
		if err := m.Refresh(); err != nil {
			logger.WithError(err).WithField("retry", fails).Errorln("Unable to refresh rules")
			if fails > 15 {
				logger.WithError(err).WithField("retry", fails).Fatalf("Terminating after retry %d\n", fails)
			}

			time.Sleep(time.Second * time.Duration(fails+1))

			fails++
		} else {
			time.Sleep(duration)
			fails = 0
		}
	}
}

func refreshKeys(k rsakey.Manager, duration time.Duration) {
	if duration == 0 {
		duration, _ = time.ParseDuration(viper.GetString("CREDENTIALS_ISSUER_ID_TOKEN_JWK_REFRESH_INTERVAL"))
		if duration == 0 {
			duration = time.Minute * 5
		}
	}

	var fails int
	for {
		if err := k.Refresh(); err != nil {
			logger.WithError(err).WithField("retry", fails).Errorln("Unable to refresh keys for signing ID Token, 'id_token' credentials issuer will not work.")
			//if fails > 15 {
			//	logger.WithError(err).WithField("retry", fails).Fatalf("Terminating after retry %d\n", fails)
			//}

			wait := fails
			if wait > 10 {
				wait = 10
			}
			time.Sleep(time.Second * time.Duration(wait^2))

			fails++
		} else {
			fails = 0
			time.Sleep(duration)
		}
	}
}

func keyManagerFactory(l logrus.FieldLogger) (keyManager rsakey.Manager, err error) {
	switch a := strings.ToLower(viper.GetString("CREDENTIALS_ISSUER_ID_TOKEN_ALGORITHM")); a {
	case "hs256":
		secret := []byte(viper.GetString("CREDENTIALS_ISSUER_ID_TOKEN_HS256_SECRET"))
		if len(secret) < 32 {
			return nil, errors.New("The secret set in CREDENTIALS_ISSUER_ID_TOKEN_HS256_SECRET must be 32 characters long.")
		}
		keyManager = rsakey.NewLocalHS256Manager(secret)
		//case "rs256":
		//	keyManager = &rsakey.LocalRS256Manager{KeyStrength: 4096}
	case "ory-hydra":
		sdk := getHydraSDK()
		keyManager = &rsakey.HydraManager{
			SDK: sdk,
			Set: viper.GetString("CREDENTIALS_ISSUER_ID_TOKEN_HYDRA_JWK_SET_ID"),
		}
	default:
		return nil, errors.Errorf("Unknown ID Token singing algorithm %s", a)
	}

	return keyManager, nil
}

func availableHandlerNames() ([]string, []string, []string) {
	return []string{
			new(proxy.AuthenticatorNoOp).GetID(),
			new(proxy.AuthenticatorAnonymous).GetID(),
			new(proxy.AuthenticatorOAuth2Introspection).GetID(),
			new(proxy.AuthenticatorOAuth2ClientCredentials).GetID(),
		},
		[]string{
			new(proxy.AuthorizerAllow).GetID(),
			new(proxy.AuthorizerDeny).GetID(),
			new(proxy.AuthorizerKetoWarden).GetID(),
		},
		[]string{
			new(proxy.CredentialsIssuerNoOp).GetID(),
			new(proxy.CredentialsIDToken).GetID(),
		}
}

func enabledHandlerNames() (d []string, e []string, f []string) {
	a, b, c := handlerFactories(nil)
	for _, i := range a {
		d = append(d, i.GetID())
	}
	for _, i := range b {
		e = append(e, i.GetID())
	}
	for _, i := range c {
		f = append(f, i.GetID())
	}
	return
}

func handlerFactories(keyManager rsakey.Manager) ([]proxy.Authenticator, []proxy.Authorizer, []proxy.CredentialsIssuer) {
	var authorizers = []proxy.Authorizer{
		proxy.NewAuthorizerAllow(),
		proxy.NewAuthorizerDeny(),
	}

	if u := viper.GetString("AUTHORIZER_KETO_WARDEN_KETO_URL"); len(u) > 0 {
		ketoSdk, err := keto.NewCodeGenSDK(&keto.Configuration{
			EndpointURL: viper.GetString("AUTHORIZER_KETO_WARDEN_KETO_URL"),
		})
		if err != nil {
			logger.WithError(err).Fatal("Unable to initialize the ORY Keto SDK")
		}
		authorizers = append(authorizers, proxy.NewAuthorizerKetoWarden(ketoSdk))
	}

	return []proxy.Authenticator{
			proxy.NewAuthenticatorNoOp(),
			proxy.NewAuthenticatorAnonymous(viper.GetString("AUTHENTICATOR_ANONYMOUS_USERNAME")),
			proxy.NewAuthenticatorOAuth2Introspection(
				viper.GetString("AUTHENTICATOR_OAUTH2_INTROSPECTION_CLIENT_ID"),
				viper.GetString("AUTHENTICATOR_OAUTH2_INTROSPECTION_CLIENT_SECRET"),
				viper.GetString("AUTHENTICATOR_OAUTH2_INTROSPECTION_TOKEN_URL"),
				viper.GetString("AUTHENTICATOR_OAUTH2_INTROSPECTION_INTROSPECT_URL"),
				strings.Split(viper.GetString("AUTHENTICATOR_OAUTH2_INTROSPECTION_SCOPE"), ","),
				fosite.WildcardScopeStrategy,
			),
			proxy.NewAuthenticatorOAuth2ClientCredentials(
				viper.GetString("AUTHENTICATOR_OAUTH2_CLIENT_CREDENTIALS_TOKEN_URL"),
			),
		},
		authorizers,
		[]proxy.CredentialsIssuer{
			proxy.NewCredentialsIssuerNoOp(),
			proxy.NewCredentialsIssuerIDToken(
				keyManager,
				logger,
				viper.GetDuration("CREDENTIALS_ISSUER_ID_TOKEN_LIFESPAN"),
				viper.GetString("CREDENTIALS_ISSUER_ID_TOKEN_ISSUER"),
			),
			proxy.NewCredentialsIssuerHeaders(),
		}
}
