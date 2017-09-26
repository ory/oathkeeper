package main

import (
	"testing"
	"net/http/httptest"
	"fmt"
	"github.com/ory-am/editor-platform/services/exposed/proxies/firewall-reverse-proxy/director"
	"net/http"
	"net/http/httputil"
	"github.com/golang/mock/gomock"
	"net/url"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"github.com/stretchr/testify/assert"
	"github.com/ory-am/fosite"
	"github.com/pkg/errors"
	"github.com/ory-am/hydra/oauth2"
)

func TestProxy(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	fw := director.NewMockFirewall(ctrl)
	it := director.NewMockIntrospector(ctrl)

	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.NotEmpty(t, r.Header.Get("X-Firewall-Method"))
		fmt.Fprint(w, "Hello, client")
	}))
	defer backend.Close()

	t.Log(backend.URL)
	u, _ := url.Parse(backend.URL)
	d := &director.Director{
		Rules: []director.Rule{
			director.Rule{
				PathMatch:director.MustCompileMatch("GET:/protected/<[^/]+>"),
				Scopes:      []string{"protected"},
			},
		},
		TargetURL: u,
		Firewall: fw,
		Introspector: it,
	}
	proxy := httptest.NewServer(&httputil.ReverseProxy{
		Director: d.Allowed,
		Transport: d,
	})
	defer proxy.Close()

	for k, c := range []struct {
		url     string
		code    int
		message string
		init    func()
	}{
		{
			url: proxy.URL + "/invalid",
			code: http.StatusInternalServerError,
		},
		{
			url: proxy.URL + "/protected/foo",
			code: http.StatusUnauthorized,
			init: func() {
				fw.EXPECT().TokenFromRequest(gomock.Any()).Return("token")
				it.EXPECT().IntrospectToken(gomock.Any(), gomock.Eq("token"), "protected").Return(nil, errors.Wrap(fosite.ErrUnauthorizedClient, ""))
			},
		},
		{
			url: proxy.URL + "/protected/foo",
			message: "Hello, client",
			code: http.StatusOK,
			init: func() {
				fw.EXPECT().TokenFromRequest(gomock.Any()).Return("token")
				it.EXPECT().IntrospectToken(gomock.Any(), gomock.Eq("token"), "protected").Return(&oauth2.Introspection{}, nil)
			},
		},
	} {
		t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
			if c.init != nil {
				c.init()
			}

			res, err := http.Get(c.url)
			require.Nil(t, err)

			greeting, err := ioutil.ReadAll(res.Body)
			res.Body.Close()
			require.Nil(t, err)

			assert.Equal(t, c.code, res.StatusCode)
			if c.message != ""{
				assert.Equal(t, c.message, fmt.Sprintf("%s", greeting))
			}
		})
	}
}
