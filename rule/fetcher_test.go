package rule_test

import (
	"testing"

	"github.com/ory/oathkeeper/internal"
)

func TestFetcher(t *testing.T) {
	conf := internal.NewConfigurationWithDefaults()
	r := internal.NewRegistry(conf)

	f := r.RuleFetcher()
}
