package certs

import (
	"crypto/tls"
	"net/http"
)

var _ http.RoundTripper = (*RoundTripper)(nil)

type RoundTripper struct {
	cm *CertManager
	tr *http.Transport
}

func NewRoundTripper(cm *CertManager) *RoundTripper {
	return &RoundTripper{
		cm: cm,
		tr: http.DefaultTransport.(*http.Transport).Clone(),
	}
}

func (rt *RoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	pool, err := rt.cm.CertPool()
	if err != nil {
		return nil, err
	}

	rt.tr.TLSClientConfig = &tls.Config{
		InsecureSkipVerify: false,
		RootCAs:            pool,
	}

	return rt.tr.RoundTrip(r)
}
