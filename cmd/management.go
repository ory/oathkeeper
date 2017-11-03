// Copyright © 2017 NAME HERE <EMAIL ADDRESS>
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
	"fmt"

	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/meatballhat/negroni-logrus"
	"github.com/ory/graceful"
	"github.com/ory/herodot"
	"github.com/ory/oathkeeper/rule"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/urfave/negroni"
)

type managementConfig struct {
	rules   rule.Manager
	address string
}

func runManagement(c *managementConfig) {
	handler := rule.Handler{H: herodot.NewJSONWriter(logger), M: c.rules}
	router := httprouter.New()
	handler.SetRoutes(router)

	n := negroni.New()
	n.Use(negronilogrus.NewMiddlewareFromLogger(logger, "oahtkeeper-management"))
	n.UseHandler(router)

	addr := c.address
	server := graceful.WithDefaults(&http.Server{
		Addr:    addr,
		Handler: router,
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
	Use: "management",
	Run: func(cmd *cobra.Command, args []string) {
		rules, err := newRuleManager(viper.GetString("DATABASE_URL"))
		if err != nil {
			logger.WithError(err).Fatalln("Unable to connect to rule backend.")
		}

		config := &managementConfig{
			rules:   rules,
			address: fmt.Sprintf("%s:%s", viper.GetString("MANAGEMENT_HOST"), viper.GetString("MANAGEMENT_PORT")),
		}

		runManagement(config)
	},
}

func init() {
	serveCmd.AddCommand(managementCmd)
}
