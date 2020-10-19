package x

import (
	"net/url"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testData struct {
	urlStr              string
	expectedPathUnix    string
	expectedPathWindows string
}

var testURLs = []testData{
	{"file:///home/test/file1.txt", "/home/test/file1.txt", "home\\test\\file1.txt"},                                                   // Unix style
	{"file:/home/test/file2.txt", "/home/test/file2.txt", "home\\test\\file2.txt"},                                                     // File variant without slashes
	{"file:///../test/update/file3.txt", "../test/update/file3.txt", "..\\test\\update\\file3.txt"},                                    // Special case relative path
	{"file://../test/update/file4.txt", "../test/update/file4.txt", "..\\test\\update\\file4.txt"},                                     // Invalid relative path
	{"file://C:/users/test/file5.txt", "C:/users/test/file5.txt", "C:\\users\\test\\file5.txt"},                                        // Non standard Windows style
	{"file:///C:/users/test/file5.txt", "/C:/users/test/file5.txt", "C:\\users\\test\\file5.txt"},                                      // Windows style
	{"file://anotherhost/share/users/test/file6.txt", "/share/users/test/file6.txt", "\\\\anotherhost\\share\\users\\test\\file6.txt"}, // Windows style with hostname
}

func TestFileURL(t *testing.T) {
	for _, td := range testURLs {
		u, err := url.Parse(td.urlStr)
		assert.NoError(t, err)
		p := GetURLFilePath(*u)
		if runtime.GOOS == "windows" {
			assert.Equal(t, td.expectedPathWindows, p)
		} else {
			assert.Equal(t, td.expectedPathUnix, p)
		}
	}
}
