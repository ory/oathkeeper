package proxy

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
)

type transport struct {
	base      *http.Transport
	dialer    *net.Dialer
	tlsDialer *tls.Dialer
}

func (t *transport) handleUnixAddr(addr string) (string, error) {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return "", err
	}
	path, err := url.PathUnescape(host)
	if err != nil {
		return "", err
	}
	return path, nil
}

func (t *transport) dialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	if path, err := t.handleUnixAddr(addr); err != nil {
		return nil, err
	} else {
		return t.dialer.DialContext(ctx, "unix", path)
	}
}

func (t *transport) dialTlsContext(ctx context.Context, network, addr string) (net.Conn, error) {
	if path, err := t.handleUnixAddr(addr); err != nil {
		return nil, err
	} else {
		return t.tlsDialer.DialContext(ctx, "unix", path)
	}
}

func (t *transport) RoundTrip(r *http.Request) (*http.Response, error) {
	if r.URL != nil {
		switch r.URL.Scheme {
		case "http", "https":
			return http.DefaultTransport.RoundTrip(r)
		case "unix":
			if u, err := url.Parse(r.URL.Query().Get("url")); err != nil {
				return nil, err
			} else {
				req := r.Clone(r.Context())
				req.URL.Scheme = u.Scheme
				req.URL.Host = url.QueryEscape(strings.TrimRight(r.URL.Path, "/"))
				req.URL.Path = u.Path
				v := req.URL.Query()
				v.Del("url")
				req.URL.RawQuery = v.Encode()
				return t.base.RoundTrip(req)
			}
		default:
		}
	}
	return nil, fmt.Errorf("invalid request")
}

func NewRoundTripper() http.RoundTripper {
	base := http.DefaultTransport.(*http.Transport)
	dialer := &net.Dialer{}
	t := &transport{
		base:   base,
		dialer: dialer,
		tlsDialer: &tls.Dialer{
			NetDialer: dialer,
			Config:    base.TLSClientConfig,
		},
	}
	t.base.DialContext = t.dialContext
	t.base.DialTLSContext = t.dialTlsContext
	return t
}
