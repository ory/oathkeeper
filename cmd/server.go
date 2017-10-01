package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"net/http/httputil"
	"github.com/ory/graceful"
	"net/http"
	"os"
	"log"
	"github.com/julienschmidt/httprouter"
	"github.com/spf13/viper"
)

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Runs the management server",
	Run: func(cmd *cobra.Command, args []string) {
		router := httprouter.New()

		server := graceful.WithDefaults(&http.Server{
			Addr: fmt.Sprintf("%s:%s", viper.GetString("SERVER_HOST"), viper.GetString("SERVER_PORT")),
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
