package cloudstorage

import (
	"fmt"
)

const (
	prefix  = "ory/oathkeeper"
	version = "0.1.0"
)

// AzureUserAgentPrefix returns a prefix that is used to set Azure SDK User-Agent to help with diagnostics.
func AzureUserAgentPrefix(api string) string {
	return userAgentString(api)
}

func userAgentString(api string) string {
	return fmt.Sprintf("%s/%s/%s", prefix, api, version)
}
