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
	shouldSucceed       bool
}

var testURLs = []testData{
	{"file:///home/test/file1.txt", "/home/test/file1.txt", "home\\test\\file1.txt", true},                                                    // Unix style
	{"file:/home/test/file2.txt", "/home/test/file2.txt", "home\\test\\file2.txt", true},                                                      // File variant without slashes
	{"file:///../test/update/file3.txt", "../test/update/file3.txt", "..\\test\\update\\file3.txt", true},                                     // Special case relative path
	{"file://../test/update/file4.txt", "../test/update/file4.txt", "..\\test\\update\\file4.txt", true},                                      // Invalid relative path
	{"file://C:/users/test/file5.txt", "C:/users/test/file5.txt", "C:\\users\\test\\file5.txt", true},                                         // Non standard Windows style
	{"file:///C:/users/test/file5.txt", "/C:/users/test/file5.txt", "C:\\users\\test\\file5.txt", true},                                       // Windows style
	{"file://anotherhost/share/users/test/file6.txt", "/share/users/test/file6.txt", "\\\\anotherhost\\share\\users\\test\\file6.txt", false}, // Windows style with hostname, this is not supported
	{"file://file7.txt", "file7.txt", "file7.txt", true},                                                                                      // Invalid relative path
	{"file://path/file8.txt", "path/file8.txt", "path\\file8.txt", true},                                                                      // Invalid relative path
}

func TestFileURL(t *testing.T) {
	for _, td := range testURLs {
		u, err := url.Parse(td.urlStr)
		assert.NoError(t, err)
		p := GetURLFilePath(*u)
		if runtime.GOOS == "windows" {
			if td.shouldSucceed {
				assert.Equal(t, td.expectedPathWindows, p)
			} else {
				assert.NotEqual(t, td.expectedPathWindows, p)
			}
		} else {
			if td.shouldSucceed {
				assert.Equal(t, td.expectedPathUnix, p)
			} else {
				assert.NotEqual(t, td.expectedPathUnix, p)
			}
		}
	}
}
