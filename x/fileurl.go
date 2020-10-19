package x

import (
	"net/url"
	"path/filepath"
	"runtime"
)

// GetURLFilePath returns the path of a URL that is compatible with the runtime os filesystem
func GetURLFilePath(u url.URL) string {
	fPath := u.Path
	sep := string(filepath.Separator)
	// Special case for malformed file urls (file://../relative/path, file://path)
	// this means we cannot support the hostname part of the URL as an actual hostname
	fPath = u.Host + fPath
	if runtime.GOOS == "windows" {
		// Strip any path initial separator on windows
		fPath = stripFistPathSeparators(fPath)
	}
	fPath = filepath.FromSlash(fPath)

	// Special case for non standard relative paths
	isRelative := (len(fPath) > 1 && fPath[0:1] == ".") ||
		(len(fPath) > 2 && fPath[0:2] == sep+".") ||
		(len(fPath) > 3 && fPath[0:3] == sep+sep+".")
	if isRelative {
		// Strip first path separator so we can return a relative path
		return stripFistPathSeparators(fPath)
	}

	return filepath.Clean(fPath)
}

func stripFistPathSeparators(fPath string) string {
	for len(fPath) > 0 && (fPath[0] == '/' || fPath[0] == '\\') {
		fPath = fPath[1:]
	}
	return fPath
}
