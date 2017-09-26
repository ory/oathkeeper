package main

import (
	"net/http"
	"net/http/httputil"
	"fmt"
	"github.com/ory-am/common/env"
	"github.com/ory-am/editor-platform/services/exposed/proxies/firewall-reverse-proxy/director"
	"net/url"
	"github.com/ory-am/hydra/sdk"
	"github.com/Sirupsen/logrus"
)

func main() {
	u, err := url.Parse(env.Getenv("BACKEND_URL", "http://localhost:7000"))
	if err != nil {
		panic(err)
	}

	var hydra *sdk.Client
	id := env.Getenv("HYDRA_CLIENT_ID", "ae3a8b6a-d011-4837-9346-6e5e38fc658a")
	secret := env.Getenv("HYDRA_CLIENT_SECRET", "w6ZXOyndAbeV")
	cluster := env.Getenv("HYDRA_URL", "http://localhost:9101")
	hydra, err = sdk.Connect(
		sdk.ClientID(id),
		sdk.ClientSecret(secret),
		sdk.ClusterURL(cluster),
	)
	if err != nil {
		logrus.WithError(err).Fatal("Could not connect to hydra with")
	}

	startProxyServer(u, hydra)
}

func startProxyServer(u *url.URL, hydra *sdk.Client) {
	d := director.NewDirector(u, hydra.Warden, hydra.Introspection)
	proxy := &httputil.ReverseProxy{
		Director: d.Allowed,
		Transport: d,
	}

	server := http.Server{
		Handler: proxy,
		Addr: fmt.Sprintf("%s:%s", env.Getenv("HOST", ""), env.Getenv("PORT", "3000")),
	}

	if err := server.ListenAndServe(); err != nil {
		panic(err)
	}
}
