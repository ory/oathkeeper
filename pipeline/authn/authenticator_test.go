package authn_test

import (
	"fmt"
	"github.com/ory/oathkeeper/pipeline/authn"
	"github.com/stretchr/testify/assert"
	"testing"
)

const (
	key = "key"
	val = "value"
)

func TestSetHeader(t *testing.T) {

	assert := assert.New(t)
	for k, tc := range [] struct {
		a    *authn.AuthenticationSession
		desc string
	}{
		{
			a:    &authn.AuthenticationSession{},
			desc: "should initiate Header field if it is nil",
		},
		{
			a:    &authn.AuthenticationSession{Header: map[string][]string{}},
			desc: "should add a header to AuthenticationSession",
		},
	} {
		t.Run(fmt.Sprintf("case=%d/description=%s", k, tc.desc), func(t *testing.T) {
			tc.a.SetHeader(key, val)

			assert.NotNil(tc.a.Header)
			assert.Len(tc.a.Header, 1)
			assert.Equal(tc.a.Header.Get(key), val)
		})
	}
}
