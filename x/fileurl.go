package x

import (
	"net/url"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/ory/x/logrusx"
)

// winPathRegex is a regex for [DRIVE-LETTER]:
var winPathRegex = regexp.MustCompile("^[A-Za-z]:.*")

// GetURLFilePath returns the path of a URL that is compatible with the runtime os filesystem
func GetURLFilePath(u *url.URL) string {
	if u == nil {
		return ""
	}
	if !(u.Scheme == "file" || u.Scheme == "") {
		return u.Path
	}

	fPath := u.Path
	if runtime.GOOS == "windows" {
		if u.Host != "" {
			// Make UNC Path
			fPath = "\\\\" + u.Host + filepath.FromSlash(fPath)
			return fPath
		}
		fPathTrimmed := strings.TrimLeft(fPath, "/")
		if winPathRegex.MatchString(fPathTrimmed) {
			// On Windows we should remove the initial path separator in case this
			// is a normal path (for example: "\c:\" -> "c:\"")
			fPath = fPathTrimmed
		}
	}
	return filepath.FromSlash(fPath)
}

// ParseURL parses rawURL into a URL structure with special handling for file:// URLs
func ParseURL(rawURL string) (*url.URL, error) {
	lcRawURL := strings.ToLower(rawURL)
	if strings.HasPrefix(lcRawURL, "file:///") {
		return url.Parse(rawURL)
	}

	// Normally the first part after file:// is a hostname, but since
	// this is often misused we interpret the URL like a normal path
	// by removing the "file://" from the beginning (if it exists)
	rawURL = trimPrefixIC(rawURL, "file://")

	if winPathRegex.MatchString(rawURL) {
		// Windows path
		return url.Parse("file:///" + rawURL)
	}

	if strings.HasPrefix(lcRawURL, "\\\\") {
		// Windows UNC path
		// We extract the hostname and create an appropriate file:// URL
		// based on the hostname and the path
		host, path := extractUNCPathParts(rawURL)
		// It is safe to replace the \ with / here because this is POSIX style path
		return url.Parse("file://" + host + strings.ReplaceAll(path, "\\", "/"))
	}

	return url.Parse(rawURL)
}

// ParseOrPanic parses a url or panics.
func ParseOrPanic(in string) *url.URL {
	out, err := ParseURL(in)
	if err != nil {
		panic(err.Error())
	}
	return out
}

// ParseOrFatal parses a url or fatals.
func ParseOrFatal(l *logrusx.Logger, in string) *url.URL {
	out, err := ParseURL(in)
	if err != nil {
		l.WithError(err).Fatalf("Unable to parse url: %s", in)
	}
	return out
}

func extractUNCPathParts(uncPath string) (host, path string) {
	parts := strings.Split(strings.TrimPrefix(uncPath, "\\\\"), "\\")
	host = parts[0]
	if len(parts) > 0 {
		path = "\\" + strings.Join(parts[1:], "\\")
	}
	return host, path
}

// trimPrefixIC returns s without the provided leading prefix string using
// case insensitive matching.
// If s doesn't start with prefix, s is returned unchanged.
func trimPrefixIC(s, prefix string) string {
	if strings.HasPrefix(strings.ToLower(s), prefix) {
		return s[len(prefix):]
	}
	return s
}
