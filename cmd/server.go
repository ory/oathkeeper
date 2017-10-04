package cmd

import (
	"fmt"

	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/ory/graceful"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/ory/oathkeeper/rule"
	"github.com/ory/herodot"
	"github.com/meatballhat/negroni-logrus"
	"github.com/urfave/negroni"
)

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Runs the management server",
	Run: func(cmd *cobra.Command, args []string) {
		rm, err := newRuleManager(viper.GetString("DATABASE_URL"))
		if err != nil {
			logger.WithError(err).Fatalln("Unable to connect to rule backend.")
		}

		handler := rule.Handler{H: herodot.NewJSONWriter(logger), M: rm}
		router := httprouter.New()
		handler.SetRoutes(router)

		n := negroni.New()
		n.Use(negronilogrus.NewMiddlewareFromLogger(logger, "oahtkeeper"))
		n.UseHandler(router)

		server := graceful.WithDefaults(&http.Server{
			Addr:    fmt.Sprintf("%s:%s", viper.GetString("MANAGEMENT_HOST"), viper.GetString("MANAGEMENT_PORT")),
			Handler: router,
		})

		if err := graceful.Graceful(server.ListenAndServe, server.Shutdown); err != nil {
			log.Fatalf("Unable to gracefully shutdown HTTP server becase %s.\n", err)
			return
		}
		log.Println("HTTP server was shutdown gracefully")
	},
}

func init() {
	RootCmd.AddCommand(serverCmd)
}
