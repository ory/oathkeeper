package header

import "net/textproto"

const (
	Accept         = "Accept"
	AcceptEncoding = "Accept-Encoding"
	Authorization  = "Authorization"
	Cookie         = "Cookie"
	XForwardedHost = "X-Forwarded-Host"
)

// Canonical returns the canonical format of the
// MIME header key. The canonicalization converts the first
// letter and any letter following a hyphen to upper case;
// the rest are converted to lowercase.
func Canonical(h string) string {
	return textproto.CanonicalMIMEHeaderKey(h)
}
