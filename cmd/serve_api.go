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

// managementCmd represents the management command
var managementCmd = &cobra.Command{
	Use:   "api",
	Short: "Starts the ORY Oathkeeper HTTP API",
	Long: `This starts a HTTP/2 REST API for managing ORY Oathkeeper.

CORE CONTROLS
=============

` + databaseUrl + `


HTTP CONTROLS
==============

- HOST: The host to listen on.
	Default: HOST="" (all interfaces)
- PORT: The port to listen on.
	Default: PORT="4456"

` + corsMessage,
	Run: func(cmd *cobra.Command, args []string) {
		rules, err := newRuleManager(viper.GetString("DATABASE_URL"))
		if err != nil {
			logger.WithError(err).Fatalln("Unable to connect to rule backend")
		}

		sdk := getHydraSDK()

		keyManager := &rsakey.HydraManager{
			SDK: sdk,
			Set: viper.GetString("HYDRA_JWK_SET_ID"),
		}

		writer := herodot.NewJSONWriter(logger)
		ruleHandler := rule.NewHandler(writer, rules, newJury(logger).GetIDs())
		keyHandler := rsakey.NewHandler(writer, keyManager)
		router := httprouter.New()
		ruleHandler.SetRoutes(router)
		keyHandler.SetRoutes(router)

		n := negroni.New()
		n.Use(negronilogrus.NewMiddlewareFromLogger(logger, "oathkeeper-api"))
		n.UseHandler(router)

		ch := cors.New(parseCorsOptions("")).Handler(n)

		go refreshKeys(keyManager, 0)

		addr := fmt.Sprintf("%s:%s", viper.GetString("HOST"), viper.GetString("PORT"))
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
	},
}

func init() {
	serveCmd.AddCommand(managementCmd)
}
