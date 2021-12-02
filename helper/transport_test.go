package helper

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"io"
	"math/big"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"path"
	"reflect"
	"testing"
	"time"
)

const pattern = "/test/roundtrip"

func generateCertificate() (*tls.Certificate, error) {
	priv, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		return nil, err
	}
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"Test"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(time.Hour * 24 * 180),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}
	crt, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return nil, err
	}
	return &tls.Certificate{
		Certificate: [][]byte{crt},
		PrivateKey:  priv,
	}, nil
}

func TestRoundTrip(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc(pattern, func(rw http.ResponseWriter, r *http.Request) {
		rw.WriteHeader(http.StatusOK)
		rw.Write([]byte(r.URL.RawQuery))
	})
	cert, err := generateCertificate()
	if err != nil {
		t.Error(err)
	}
	dt := http.DefaultTransport.(*http.Transport)
	dt.TLSHandshakeTimeout = time.Second
	dt.IdleConnTimeout = time.Second
	dt.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	tests := []struct {
		name   string
		expect func() (net.Listener, *http.Request, bool)
	}{
		{
			"Invalid request",
			func() (net.Listener, *http.Request, bool) {
				return nil, &http.Request{}, true
			},
		},
		{
			"HTTP : Dial error",
			func() (net.Listener, *http.Request, bool) {
				req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://%s", path.Join("127.0.0.1:12345", pattern)), nil)
				if err != nil {
					t.Error(err)
				}
				return nil, req, true
			},
		},
		{
			"UNIX : Dial error",
			func() (net.Listener, *http.Request, bool) {
				req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://%s", path.Join("unix://path/to/unix.sock?path=%s", pattern)), nil)
				if err != nil {
					t.Error(err)
				}
				return nil, req, true
			},
		},
		{
			"HTTP : OK",
			func() (net.Listener, *http.Request, bool) {
				lis, err := net.Listen("tcp", "127.0.0.1:0")
				if err != nil {
					t.Error(err)
				}
				go http.Serve(lis, mux)
				req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://%s", path.Join(lis.Addr().String(), pattern)), nil)
				if err != nil {
					t.Error(err)
				}
				return lis, req, false
			},
		},
		{
			"UNIX : OK",
			func() (net.Listener, *http.Request, bool) {
				lis, err := net.Listen("unix", path.Join(t.TempDir(), "unix.sock"))
				if err != nil {
					t.Error(err)
				}
				go http.Serve(lis, mux)
				req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("unix://%s?path=%s&a=1&b=2", lis.Addr().String(), pattern), nil)
				if err != nil {
					t.Error(err)
				}
				return lis, req, false
			},
		},
		{
			"HTTP + TLS : OK",
			func() (net.Listener, *http.Request, bool) {
				lis, err := net.Listen("tcp", "127.0.0.1:0")
				if err != nil {
					t.Error(err)
				}
				lis = tls.NewListener(lis, &tls.Config{
					NextProtos:   []string{"http/1.1"},
					Certificates: []tls.Certificate{*cert},
				})
				go http.Serve(lis, mux)
				req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("https://%s", path.Join(lis.Addr().String(), pattern)), nil)
				if err != nil {
					t.Error(err)
				}
				return lis, req, false
			},
		},
		{
			"UNIX + TLS : OK",
			func() (net.Listener, *http.Request, bool) {
				lis, err := net.Listen("unix", path.Join(t.TempDir(), "unix.sock"))
				if err != nil {
					t.Error(err)
				}
				lis = tls.NewListener(lis, &tls.Config{
					NextProtos:   []string{"http/1.1"},
					Certificates: []tls.Certificate{*cert},
				})
				go http.Serve(lis, mux)
				req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("unix://%s?path=%s&tls=true&a=1&b=2", lis.Addr().String(), pattern), nil)
				if err != nil {
					t.Error(err)
				}
				return lis, req, false
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewRoundTripper()
			lis, xreq, wantErr := tt.expect()
			res, err := r.RoundTrip(xreq)
			if wantErr && err == nil || !wantErr && err != nil {
				t.Errorf("want err : %v, got err : %v", wantErr, err)
			}
			if res != nil {
				if res.StatusCode != http.StatusOK {
					t.Errorf("want code : %v, got code : %v", http.StatusOK, res.StatusCode)
					s, _ := httputil.DumpRequest(xreq, false)
					fmt.Println(string(s))
					s, _ = httputil.DumpResponse(res, false)
					fmt.Println(string(s))
				} else {
					testURL := url.URL{}
					b, _ := io.ReadAll(res.Body)
					testURL.RawQuery = string(b)
					for k, v := range xreq.URL.Query() {
						if k != "path" && k != "tls" && !reflect.DeepEqual(v, testURL.Query()[k]) {
							t.Errorf("want query : %v, got query : %v", testURL.Query()[k], v)
						}
					}
				}
			}
			if lis != nil {
				lis.Close()
			}
		})
	}
}
