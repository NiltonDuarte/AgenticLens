package capture

import (
	"errors"
	"net/url"
	"path/filepath"
	"strings"
)

// sqliteFileURI keeps Windows drive letters in the URI path, never the host.
// All caller-controlled filename characters are URL-escaped before DSN options
// are appended. Network/extended Windows namespaces are not supported here.
func sqliteFileURI(path string) (string, error) {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	volume := filepath.VolumeName(absolute)
	if strings.HasPrefix(volume, `\\`) {
		return "", errors.New("UNC and extended Windows capture state paths are unsupported; configure a local drive path")
	}
	slash := filepath.ToSlash(absolute)
	if volume != "" && !strings.HasPrefix(slash, "/") {
		slash = "/" + slash
	}
	return (&url.URL{Scheme: "file", Path: slash}).String(), nil
}
