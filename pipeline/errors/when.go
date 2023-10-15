// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package errors

import (
	"mime"
	"net"
	"net/http"
	"strings"

	"github.com/golang/gddo/httputil/header"
	"github.com/pkg/errors"

	"github.com/ory/x/errorsx"
	"github.com/ory/x/stringsx"
)

type (
	Whens []When
	When  struct {
		Error   []string     `json:"error"`
		Request *WhenRequest `json:"request"`
	}

	WhenRequest struct {
		RemoteIP *WhenRequestRemoteIP `json:"remote_ip"`
		Header   *WhenRequestHeader   `json:"header"`
	}

	WhenRequestRemoteIP struct {
		Match                     []string `json:"match"`
		RespectForwardedForHeader bool     `json:"respect_forwarded_for_header"`
	}

	WhenRequestHeader struct {
		ContentType []string `json:"content_type"`
		Accept      []string `json:"accept"`
	}

	statusCoder interface {
		StatusCode() int
	}
)

var ErrDoesNotMatchWhen = ErrHandlerNotResponsible

func statusText(code int) string {
	return strings.ReplaceAll(
		strings.ReplaceAll(
			strings.ToLower(http.StatusText(code)),
			" ", "_",
		), "'", "",
	)
}

func matches(whens Whens, r *http.Request, err error) error {
	if len(whens) == 0 {
		return nil
	}

	sce, isStatusCoder := errorsx.Cause(err).(statusCoder)
	for _, when := range whens {
		if len(when.Error) == 0 {
			if err := matchesRequest(when, r); err != nil {
				if errorsx.Cause(err) == ErrDoesNotMatchWhen {
					continue
				}
				return err
			}
			return nil
		}

		for _, code := range when.Error {
			if isStatusCoder {
				if statusText(sce.StatusCode()) == code {
					if err := matchesRequest(when, r); err != nil {
						if errorsx.Cause(err) == ErrDoesNotMatchWhen {
							continue
						}
						return err
					}
					return nil
				} else {
					continue
				}
			} else if code == statusText(http.StatusInternalServerError) {
				if err := matchesRequest(when, r); err != nil {
					if errorsx.Cause(err) == ErrDoesNotMatchWhen {
						continue
					}
					return err
				}
				return nil
			} else {
				continue
			}
		}
	}

	return errors.WithStack(ErrDoesNotMatchWhen)
}

// matchesAcceptMIME returns nil if any of the provided handler MIME types matches the HTTP Header string
// (e.g. application/json;q=1.0, application/xml;q=0.5)
func matchesAcceptMIME(request string, handlers []string) bool {
	rspec := header.ParseAccept(http.Header{"Accept": {request}}, "Accept")
	hspec := header.ParseAccept(http.Header{"Accept": {strings.Join(handlers, ",")}}, "Accept")

	for _, actual := range rspec {
		a := strings.Split(actual.Value, ";")[0]
		for _, match := range hspec {
			m := match.Value
			switch {
			case m == "*/*":
				return true
			case strings.HasSuffix(m, "/*") && strings.TrimSuffix(m, "/*") == strings.Split(a, "/")[0]:
				return true

			case a == m:
				return true

			// If the request contains wildcards, we expect the handler / search value to have an exact match! Otherwise
			// we will get a lot of conflicts.
			case a == "*/*" && a == m:
				return true
			case strings.HasSuffix(a, "/*") && a == m:
				return true
			}
		}
	}

	return false
}

func clearMIMEParams(mimes string) (string, error) {
	contentTypes := strings.Split(mimes, ",")
	for k, ct := range contentTypes {
		ct, _, err := mime.ParseMediaType(strings.TrimSpace(ct))
		if err != nil {
			return "", errors.WithStack(err)
		}
		contentTypes[k] = ct
	}
	return strings.Join(contentTypes, ", "), nil
}

func matchesRequest(when When, r *http.Request) error {
	if when.Request == nil {
		return nil
	}

	if when.Request.Header != nil && len(when.Request.Header.ContentType) > 0 {
		contentTypes, err := clearMIMEParams(stringsx.Coalesce(r.Header.Get("Content-Type"), "application/octet-stream"))
		if err != nil {
			return err
		}
		if !matchesAcceptMIME(contentTypes, when.Request.Header.ContentType) {
			return errors.WithStack(ErrDoesNotMatchWhen)
		}
	}

	if when.Request.Header != nil && len(when.Request.Header.Accept) > 0 {
		if !matchesAcceptMIME(
			stringsx.Coalesce(r.Header.Get("Accept"), "application/octet-stream"),
			when.Request.Header.Accept,
		) {
			return errors.WithStack(ErrDoesNotMatchWhen)
		}
	}

	if when.Request.RemoteIP != nil && len(when.Request.RemoteIP.Match) > 0 {
		remoteIP, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			return errors.WithStack(err)
		}

		check := []string{remoteIP}

		if when.Request.RemoteIP.RespectForwardedForHeader {
			for _, fwd := range stringsx.Splitx(r.Header.Get("X-Forwarded-For"), ",") {
				check = append(check, strings.TrimSpace(fwd))
			}
		}

		for _, rn := range when.Request.RemoteIP.Match {
			_, cidr, err := net.ParseCIDR(rn)
			if err != nil {
				return errors.WithStack(err)
			}

			for _, ip := range check {
				addr := net.ParseIP(ip)
				if cidr.Contains(addr) {
					return nil
				}
			}
		}
	} else {
		return nil
	}

	return errors.WithStack(ErrDoesNotMatchWhen)
}

func MatchesWhen(whens Whens, r *http.Request, err error) error {
	if len(whens) == 0 {
		return nil
	}

	if err := matches(whens, r, err); err != nil {
		return err
	}

	return nil
}
