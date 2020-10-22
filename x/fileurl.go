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

	if u.Scheme != "file" && u.Scheme != "" {
		return u.Path
	}
	fPath := u.Path
	if runtime.GOOS == "windows" {
		if u.Host != "" {
			// Make UNC Path
			s := string(filepath.Separator)
			fPath = s + s + u.Host + filepath.FromSlash(fPath)
			return fPath
		}
		if winPathRegex.MatchString(fPath[1:]) {
			// On Windows we should remove the initial path separator in case this
			// is a normal path (for example: "\c:\" -> "c:\"")
			fPath = stripFistPathSeparators(fPath)
		}
	}
	return filepath.FromSlash(fPath)
}

// ParseURL parses rawurl into a URL structure with special handling for file:// URLs
func ParseURL(rawURL string) (*url.URL, error) {
	lcRawurl := strings.ToLower(rawurl)
	if strings.HasPrefix(lcRawurl, "file:///") {
		return url.Parse("file:///" + toSlash(rawurl[8:]))
	}
	if strings.Index(lcRawurl, "file://") == 0 {
		// Normally the first part after file:// is a hostname, but since
		// this is often misused we interpret the URL like a normal path
		// by removing the "file://" from the beginning
		rawurl = rawurl[7:]
	}
	if winPathRegex.MatchString(rawurl) {
		// Windows path
		return url.Parse("file:///" + toSlash(rawurl))
	}
	if strings.Index(lcRawurl, "\\\\") == 0 {
		// Windows UNC path
		// We extract the hostname and creates an appropriate file:// URL
		// based on the hostname and the path
		parts := strings.Split(filepath.FromSlash(rawurl), "\\")
		host := ""
		if len(parts) > 2 {
			host = parts[2]
		}
		p := "/"
		if len(parts) > 4 {
			p += strings.Join(parts[3:], "/")
		}
		return url.Parse("file://" + host + p)
	}
	u, err := url.Parse(rawurl)
	if err != nil {
		return nil, err
	}
	if u.Scheme == "file" || u.Scheme == "" {
		u.Path = toSlash(u.Path)
	}
	return u, nil
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

func stripFistPathSeparators(fPath string) string {
	for len(fPath) > 0 && (fPath[0] == '/' || fPath[0] == '\\') {
		fPath = fPath[1:]
	}
	return fPath
}

func toSlash(path string) string {
	return strings.ReplaceAll(path, "\\", "/")
}
