// Copyright © 2017-2018 Aeneas Rekkas <aeneas+oss@aeneas.io>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package cmd

import (
	"strconv"
	"strings"
	"time"

	"github.com/ory/hydra/sdk/go/hydra"
	"github.com/ory/oathkeeper/rsakey"
	"github.com/ory/oathkeeper/rule"
	"github.com/rs/cors"
	"github.com/spf13/viper"
)

func getHydraSDK() hydra.SDK {
	sdk, err := hydra.NewSDK(&hydra.Configuration{
		ClientID:     viper.GetString("HYDRA_CLIENT_ID"),
		ClientSecret: viper.GetString("HYDRA_CLIENT_SECRET"),
		EndpointURL:  viper.GetString("HYDRA_URL"),
		Scopes:       []string{"hydra.introspect", "hydra.warden", "hydra.keys.*"},
	})

	if err != nil {
		logger.WithError(err).Fatalln("Unable to connect to Hydra SDK")
		return nil
	}
	return sdk
}

func refreshRules(c *proxyConfig, m *rule.CachedMatcher, fails int) {
	duration, _ := time.ParseDuration(viper.GetString("RULES_REFRESH_INTERVAL"))
	if duration == 0 {
		duration = time.Second * 30
	}

	if err := m.Refresh(); err != nil {
		logger.WithError(err).WithField("retry", fails).Errorln("Unable to refresh rules")
		if fails > 15 {
			logger.WithError(err).WithField("retry", fails).Fatalf("Terminating after retry %d\n", fails)
		}

		time.Sleep(time.Second * time.Duration(fails+1))
		refreshRules(c, m, fails+1)
		return
	}

	time.Sleep(duration)

	refreshRules(c, m, 0)
}

func refreshKeys(k rsakey.Manager, fails int) {
	duration, _ := time.ParseDuration(viper.GetString("JWK_REFRESH_INTERVAL"))
	if duration == 0 {
		duration = time.Minute * 5
	}

	if err := k.Refresh(); err != nil {
		logger.WithError(err).WithField("retry", fails).Errorln("Unable to refresh RSA keys for JWK signing")
		if fails > 15 {
			logger.WithError(err).WithField("retry", fails).Fatalf("Terminating after retry %d\n", fails)
		}

		time.Sleep(time.Second * time.Duration(fails+1))
		refreshKeys(k, fails+1)
		return
	}

	time.Sleep(duration)

	refreshKeys(k, 0)
}

func parseCorsOptions(prefix string) cors.Options {
	if prefix != "" {
		prefix = prefix + "_"
	}

	allowCredentials, _ := strconv.ParseBool(viper.GetString(prefix + "CORS_ALLOWED_CREDENTIALS"))
	debug, _ := strconv.ParseBool(viper.GetString(prefix + "CORS_DEBUG"))
	maxAge, _ := strconv.Atoi(viper.GetString(prefix + "CORS_MAX_AGE"))
	return cors.Options{
		AllowedOrigins:   strings.Split(viper.GetString(prefix+"CORS_ALLOWED_ORIGINS"), ","),
		AllowedMethods:   strings.Split(viper.GetString(prefix+"CORS_ALLOWED_METHODS"), ","),
		AllowedHeaders:   strings.Split(viper.GetString(prefix+"CORS_ALLOWED_HEADERS"), ","),
		ExposedHeaders:   strings.Split(viper.GetString(prefix+"CORS_EXPOSED_HEADERS"), ","),
		AllowCredentials: allowCredentials,
		MaxAge:           maxAge,
		Debug:            debug,
	}
}
