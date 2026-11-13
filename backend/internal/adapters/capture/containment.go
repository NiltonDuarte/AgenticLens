package capture

import (
	"path/filepath"
	"strings"
)

// relativeContainedPath compares already resolved paths.
func relativeContainedPath(base, target string) (string, bool, error) {
	relative, err := filepath.Rel(base, target)
	if err != nil {
		return "", false, err
	}
	contained := relative == "." || relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator)) && !filepath.IsAbs(relative)
	return relative, contained, nil
}
