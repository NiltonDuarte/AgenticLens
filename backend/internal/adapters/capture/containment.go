package capture

import (
	"path/filepath"
	"strings"
)

// relativeContainedPath compares already resolved paths. Different Windows
// volumes are disjoint; filepath.Rel errors within one volume remain errors.
func relativeContainedPath(base, target string) (string, bool, error) {
	if !strings.EqualFold(filepath.VolumeName(base), filepath.VolumeName(target)) {
		return "", false, nil
	}
	relative, err := filepath.Rel(base, target)
	if err != nil {
		return "", false, err
	}
	contained := relative == "." || relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator)) && !filepath.IsAbs(relative)
	return relative, contained, nil
}
