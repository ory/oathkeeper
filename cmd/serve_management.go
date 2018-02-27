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
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/meatballhat/negroni-logrus"
	"github.com/ory/graceful"
	"github.com/ory/herodot"
	"github.com/ory/oathkeeper/rsakey"
	"github.com/ory/oathkeeper/rule"
	"github.com/rs/cors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/urfave/negroni"
)

type managementConfig struct {
	rules      rule.Manager
	corsPrefix string
}

func runManagement(c *managementConfig) {
	sdk := getHydraSDK()

	keyManager := &rsakey.HydraManager{
		SDK: sdk,
		Set: viper.GetString("HYDRA_JWK_SET_ID"),
	}

	rules := rule.Handler{H: herodot.NewJSONWriter(logger), M: c.rules}
	keys := rsakey.Handler{H: herodot.NewJSONWriter(logger), M: keyManager}
	router := httprouter.New()
	rules.SetRoutes(router)
	keys.SetRoutes(router)

	n := negroni.New()
	n.Use(negronilogrus.NewMiddlewareFromLogger(logger, "oathkeeper-management"))
	n.UseHandler(router)

	ch := cors.New(parseCorsOptions(c.corsPrefix)).Handler(n)

	go refreshKeys(keyManager, 0)

	addr := fmt.Sprintf("%s:%s", viper.GetString("MANAGEMENT_HOST"), viper.GetString("MANAGEMENT_PORT"))
	server := graceful.WithDefaults(&http.Server{
		Addr:    addr,
		Handler: ch,
	})

	logger.Printf("Listening on %s.\n", addr)
	if err := graceful.Graceful(server.ListenAndServe, server.Shutdown); err != nil {
		logger.Fatalf("Unable to gracefully shutdown HTTP server because %s.\n", err)
		return
	}
	logger.Println("HTTP server was shutdown gracefully")
}

// managementCmd represents the management command
var managementCmd = &cobra.Command{
	Use:   "management",
	Short: "Starts the ORY Oathkeeper management REST API",
	Long: `This starts a HTTP/2 REST API for managing ORY Oathkeeper.

CORE CONTROLS
=============

` + databaseUrl + `


HTTP CONTROLS
==============

- MANAGEMENT_HOST: The host to listen on.
	Default: PROXY_HOST="" (all interfaces)
- MANAGEMENT_PORT: The port to listen on.
	Default: PROXY_PORT="4456"


` + corsMessage,
	Run: func(cmd *cobra.Command, args []string) {
		rules, err := newRuleManager(viper.GetString("DATABASE_URL"))
		if err != nil {
			logger.WithError(err).Fatalln("Unable to connect to rule backend")
		}

		config := &managementConfig{rules: rules}
		runManagement(config)
	},
}

func init() {
	serveCmd.AddCommand(managementCmd)
}
