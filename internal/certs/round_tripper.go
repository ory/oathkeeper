package certs

import (
	"crypto/tls"
	"net/http"
)

var _ http.RoundTripper = (*RoundTripper)(nil)

type RoundTripper struct {
	cm *CertManager
}

func NewRoundTripper(cm *CertManager) *RoundTripper {
	return &RoundTripper{cm: cm}
}

func (rt *RoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	pool, err := rt.cm.CertPool()
	if err != nil {
		return nil, err
	}

	tr := http.DefaultTransport.(*http.Transport).Clone()
	tr.TLSClientConfig = &tls.Config{
		InsecureSkipVerify: false,
		RootCAs:            pool,
	}

	return tr.RoundTrip(r)
}
