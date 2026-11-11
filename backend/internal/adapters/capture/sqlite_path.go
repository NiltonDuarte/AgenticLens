package capture

import (
	"net/url"
	"path/filepath"
)

// sqliteFileURI URL-escapes caller-controlled filename characters before DSN
// options are appended.
func sqliteFileURI(path string) (string, error) {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	return (&url.URL{Scheme: "file", Path: absolute}).String(), nil
}
