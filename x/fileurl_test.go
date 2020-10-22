package x

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetURLFilePath(t *testing.T) {
	type testData struct {
		urlStr          string
		expectedUnix    string
		expectedWindows string
		shouldSucceed   bool
	}
	var testURLs = []testData{
		{"File:///home/test/file1.txt", "/home/test/file1.txt", "\\home\\test\\file1.txt", true},
		{"fIle:/home/test/file2.txt", "/home/test/file2.txt", "\\home\\test\\file2.txt", true},
		{"fiLe:///../test/update/file3.txt", "/../test/update/file3.txt", "\\..\\test\\update\\file3.txt", true},
		{"filE://../test/update/file4.txt", "../test/update/file4.txt", "..\\test\\update\\file4.txt", true},
		{"file://C:/users/test/file5.txt", "/C:/users/test/file5.txt", "C:\\users\\test\\file5.txt", true},
		{"file:///C:/users/test/file5b.txt", "/C:/users/test/file5b.txt", "C:\\users\\test\\file5b.txt", true},
		{"file://anotherhost/share/users/test/file6.txt", "/share/users/test/file6.txt", "\\\\anotherhost\\share\\users\\test\\file6.txt", false}, // this is not supported
		{"file://file7.txt", "file7.txt", "file7.txt", true},
		{"file://path/file8.txt", "path/file8.txt", "path\\file8.txt", true},
		{"file://C:\\Users\\RUNNER~1\\AppData\\Local\\Temp\\9ccf9f68-121c-451a-8a73-2aa360925b5a386398343/access-rules.json", "/C:\\Users\\RUNNER~1\\AppData\\Local\\Temp\\9ccf9f68-121c-451a-8a73-2aa360925b5a386398343/access-rules.json", "C:\\Users\\RUNNER~1\\AppData\\Local\\Temp\\9ccf9f68-121c-451a-8a73-2aa360925b5a386398343\\access-rules.json", true},
		{"file:///C:\\Users\\RUNNER~1\\AppData\\Local\\Temp\\9ccf9f68-121c-451a-8a73-2aa360925b5a386398343/access-rules.json", "/C:\\Users\\RUNNER~1\\AppData\\Local\\Temp\\9ccf9f68-121c-451a-8a73-2aa360925b5a386398343/access-rules.json", "C:\\Users\\RUNNER~1\\AppData\\Local\\Temp\\9ccf9f68-121c-451a-8a73-2aa360925b5a386398343\\access-rules.json", true},
		{"file8.txt", "file8.txt", "file8.txt", true},
		{"../file9.txt", "../file9.txt", "..\\file9.txt", true},
		{"./file9b.txt", "./file9b.txt", ".\\file9b.txt", true},
		{"file://./file9c.txt", "./file9c.txt", ".\\file9c.txt", true},
		{"file://./folder/.././file9d.txt", "./folder/.././file9d.txt", ".\\folder\\..\\.\\file9d.txt", true},
		{"..\\file10.txt", "..\\file10.txt", "..\\file10.txt", true},
		{"C:\\file11.txt", "/C:\\file11.txt", "C:\\file11.txt", true},
		{"\\\\hostname\\share\\file12.txt", "/share/file12.txt", "\\\\hostname\\share\\file12.txt", true},
		{"file:///home/test/file 13.txt", "/home/test/file 13.txt", "\\home\\test\\file 13.txt", true},
		{"file:///home/test/file%2014.txt", "/home/test/file 14.txt", "\\home\\test\\file 14.txt", true},
		{"http://server:80/test/file%2015.txt", "/test/file 15.txt", "/test/file 15.txt", true},
		{"file:///dir/file\\ with backslash", "/dir/file\\ with backslash", "\\dir\\file\\ with backslash", true},
		{"file://dir/file\\ with backslash", "dir/file\\ with backslash", "dir\\file\\ with backslash", true},
		{"file:///dir/file with windows path forbidden chars \\<>:\"|%3F*", "/dir/file with windows path forbidden chars \\<>:\"|?*", "\\dir\\file with windows path forbidden chars \\<>:\"|?*", true},
		{"file://dir/file with windows path forbidden chars \\<>:\"|%3F*", "dir/file with windows path forbidden chars \\<>:\"|?*", "dir\\file with windows path forbidden chars \\<>:\"|?*", true},
		{"file:///path/file?query=1", "/path/file", "\\path\\file", true},
		{"http://host:80/path/file?query=1", "/path/file", "/path/file", true},
		{"file://////C:/file.txt", "////C:/file.txt", "C:\\file.txt", true},
		{"file://////C:\\file.txt", "////C:\\file.txt", "C:\\file.txt", true},
	}
	for _, td := range testURLs {
		u, err := ParseURL(td.urlStr)
		assert.NoError(t, err)
		if err != nil {
			continue
		}
		p := GetURLFilePath(u)
		if runtime.GOOS == "windows" {
			if td.shouldSucceed {
				assert.Equal(t, td.expectedWindows, p)
			} else {
				assert.NotEqual(t, td.expectedWindows, p)
			}
		} else {
			if td.shouldSucceed {
				assert.Equal(t, td.expectedUnix, p)
			} else {
				assert.NotEqual(t, td.expectedUnix, p)
			}
		}
	}
	assert.Empty(t, GetURLFilePath(nil))
}

func TestParseURL(t *testing.T) {
	type testData struct {
		urlStr       string
		expectedPath string
		expectedStr  string
	}
	var testURLs = []testData{
		{"File:///home/test/file1.txt", "/home/test/file1.txt", "file:///home/test/file1.txt"},
		{"fIle:/home/test/file2.txt", "/home/test/file2.txt", "file:///home/test/file2.txt"},
		{"fiLe:///../test/update/file3.txt", "/../test/update/file3.txt", "file:///../test/update/file3.txt"},
		{"filE://../test/update/file4.txt", "../test/update/file4.txt", "../test/update/file4.txt"},
		{"file://C:/users/test/file5.txt", "/C:/users/test/file5.txt", "file:///C:/users/test/file5.txt"},  // We expect a initial / in the path because this is a Windows absolute path
		{"file:///C:/users/test/file6.txt", "/C:/users/test/file6.txt", "file:///C:/users/test/file6.txt"}, // --//--
		{"file://file7.txt", "file7.txt", "file7.txt"},
		{"file://path/file8.txt", "path/file8.txt", "path/file8.txt"},
		{"file://C:\\Users\\RUNNER~1\\AppData\\Local\\Temp\\9ccf9f68-121c-451a-8a73-2aa360925b5a386398343/access-rules.json", "/C:\\Users\\RUNNER~1\\AppData\\Local\\Temp\\9ccf9f68-121c-451a-8a73-2aa360925b5a386398343/access-rules.json", "file:///C:%5CUsers%5CRUNNER~1%5CAppData%5CLocal%5CTemp%5C9ccf9f68-121c-451a-8a73-2aa360925b5a386398343/access-rules.json"},
		{"file:///C:\\Users\\RUNNER~1\\AppData\\Local\\Temp\\9ccf9f68-121c-451a-8a73-2aa360925b5a386398343/access-rules.json", "/C:\\Users\\RUNNER~1\\AppData\\Local\\Temp\\9ccf9f68-121c-451a-8a73-2aa360925b5a386398343/access-rules.json", "file:///C:%5CUsers%5CRUNNER~1%5CAppData%5CLocal%5CTemp%5C9ccf9f68-121c-451a-8a73-2aa360925b5a386398343/access-rules.json"},
		{"file://C:\\Users\\path with space\\file.txt", "/C:\\Users\\path with space\\file.txt", "file:///C:%5CUsers%5Cpath%20with%20space%5Cfile.txt"},
		{"file8b.txt", "file8b.txt", "file8b.txt"},
		{"../file9.txt", "../file9.txt", "../file9.txt"},
		{"./file9b.txt", "./file9b.txt", "./file9b.txt"},
		{"file://./file9c.txt", "./file9c.txt", "./file9c.txt"},
		{"file://./folder/.././file9d.txt", "./folder/.././file9d.txt", "./folder/.././file9d.txt"},
		{"..\\file10.txt", "..\\file10.txt", "..%5Cfile10.txt"},
		{"C:\\file11.txt", "/C:\\file11.txt", "file:///C:%5Cfile11.txt"},
		{"\\\\hostname\\share\\file12.txt", "/share/file12.txt", "file://hostname/share/file12.txt"},
		{"\\\\", "/", "file:///"},
		{"\\\\hostname", "/", "file://hostname/"},
		{"\\\\hostname\\", "/", "file://hostname/"},
		{"file:///home/test/file 13.txt", "/home/test/file 13.txt", "file:///home/test/file%2013.txt"},
		{"file:///home/test/file%2014.txt", "/home/test/file 14.txt", "file:///home/test/file%2014.txt"},
		{"http://server:80/test/file%2015.txt", "/test/file 15.txt", "http://server:80/test/file%2015.txt"},
		{"file:///dir/file\\ with backslash", "/dir/file\\ with backslash", "file:///dir/file%5C%20with%20backslash"},
		{"file://dir/file\\ with backslash", "dir/file\\ with backslash", "dir/file%5C%20with%20backslash"},
		{"file:///dir/file with windows path forbidden chars \\<>:\"|%3F*", "/dir/file with windows path forbidden chars \\<>:\"|?*", "file:///dir/file%20with%20windows%20path%20forbidden%20chars%20%5C%3C%3E:%22%7C%3F%2A"},
		{"file://dir/file with windows path forbidden chars \\<>:\"|%3F*", "dir/file with windows path forbidden chars \\<>:\"|?*", "dir/file%20with%20windows%20path%20forbidden%20chars%20%5C%3C%3E:%22%7C%3F%2A"},
		{"file:///path/file?query=1", "/path/file", "file:///path/file?query=1"},
		{"http://host:80/path/file?query=1", "/path/file", "http://host:80/path/file?query=1"},
		{"file://////C:/file.txt", "////C:/file.txt", "file://////C:/file.txt"},
		{"file://////C:\\file.txt", "////C:\\file.txt", "file://////C:%5Cfile.txt"},
	}

	for _, td := range testURLs {
		u, err := ParseURL(td.urlStr)
		assert.NoError(t, err)
		if err != nil {
			continue
		}
		assert.Equal(t, td.expectedPath, u.Path, "expected path for %s", td.urlStr)
		assert.Equal(t, td.expectedStr, u.String(), "expected URL string for %s", td.urlStr)
	}
	_, err := ParseURL("://")
	assert.Error(t, err)
	_, err = ParseURL("://host:80/file")
	assert.Error(t, err)
	_, err = ParseURL(":///path/file")
	assert.Error(t, err)
}
