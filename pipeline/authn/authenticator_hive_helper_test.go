package authn

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsAPIRequest(t *testing.T) {
	for k, tc := range []struct {
		h string
		e bool
	}{
		{
			e: false,
			h: `Host: prometheus.io
User-Agent: Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:64.0) Gecko/20100101 Firefox/64.0
Accept: text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8
Accept-Language: en-US,en;q=0.5
Accept-Encoding: gzip, deflate, br
Referer: https://www.google.com/
DNT: 1
Connection: keep-alive
Cookie: foo=bar
Upgrade-Insecure-Requests: 1
Cache-Control: max-age=0`,
		},
		{
			e: true,
			h: `Host: ogs.google.com
User-Agent: Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:64.0) Gecko/20100101 Firefox/64.0
Accept: */*
Accept-Language: en-US,en;q=0.5
Accept-Encoding: gzip, deflate, br
Referer: https://www.google.com/
Content-Type: application/x-www-form-urlencoded;charset=utf-8
Content-Length: 103
Origin: https://www.google.com
DNT: 1
Connection: keep-alive
Cookie: foo=bar
Cache-Control: max-age=0`,
		},
		{
			e: true,
			h: `Host: backoffice.oryapis.sh
User-Agent: Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:64.0) Gecko/20100101 Firefox/64.0
Accept: */*
Accept-Language: en-US,en;q=0.5
Accept-Encoding: gzip, deflate, br
Referer: https://console.ory.sh/
origin: https://console.ory.sh
DNT: 1
Connection: keep-alive`,
		},
		{
			e: true,
			h: `Host: backoffice.oryapis.sh
User-Agent: Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:64.0) Gecko/20100101 Firefox/64.0
Accept: */*
Accept-Language: en-US,en;q=0.5
Accept-Encoding: gzip, deflate, br
Referer: https://console.ory.sh/
origin: https://console.ory.sh
DNT: 1
Connection: keep-alive`,
		},
		{
			e: false,
			h: `Host: console.ory.sh
User-Agent: Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:64.0) Gecko/20100101 Firefox/64.0
Accept: text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8
Accept-Language: en-US,en;q=0.5
Accept-Encoding: gzip, deflate, br
DNT: 1
Connection: keep-alive
Upgrade-Insecure-Requests: 1`,
		},
	} {
		t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
			headers := http.Header{}
			for _, l := range strings.Split(tc.h, "\n") {
				ll := strings.Split(l, ":")
				headers.Set(ll[0], strings.TrimSpace(ll[1]))
			}

			assert.Equal(t, tc.e, isAPIRequest(&http.Request{
				Header: headers,
			}))
		})
	}
}
