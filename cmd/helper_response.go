package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/ory/oathkeeper/sdk/go/oathkeepersdk/swagger"
)

func checkResponse(response *swagger.APIResponse, err error, expectedStatusCode int) {
	must(err, "Could not validate token: %s", err)

	if response.StatusCode != expectedStatusCode {
		fmt.Printf("Command failed because status code %d was expeceted but code %d was received", expectedStatusCode, response.StatusCode)
		os.Exit(1)
		return
	}
}

func formatResponse(response interface{}) string {
	out, err := json.MarshalIndent(response, "", "\t")
	must(err, `Command failed because an error ("%s") occurred while prettifying output.`, err)
	return string(out)
}

func must(err error, message string, args ...interface{}) {
	if err == nil {
		return
	}

	fmt.Fprintf(os.Stderr, message+"\n", args...)
	os.Exit(1)
}
