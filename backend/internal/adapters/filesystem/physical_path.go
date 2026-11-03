package filesystem

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

// PhysicalPath resolves raw components in filesystem order, including symlinks
// before subsequent dotdot components and not-yet-created destination suffixes.
// Resolution alone never establishes that a source exists or has valid metadata.
func PhysicalPath(path string) (string, error) {
	separator := string(filepath.Separator)
	target := filepath.FromSlash(path)
	if !filepath.IsAbs(target) {
		if filepath.VolumeName(target) != "" {
			return "", errors.New("drive-relative path requires an absolute destination")
		}
		cwd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		if target != "" && os.IsPathSeparator(target[0]) {
			target = filepath.VolumeName(cwd) + target
		} else {
			// Join/Abs would clean .. before an earlier symlink is followed.
			target = cwd + separator + target
		}
	}
	volume := filepath.VolumeName(target)
	current := volume + separator
	pending := strings.Split(target[len(volume):], separator)
	links := 0
	for len(pending) > 0 {
		part := pending[0]
		pending = pending[1:]
		switch part {
		case "", ".":
			continue
		case "..":
			current = filepath.Dir(current)
			continue
		}
		next := filepath.Join(current, part)
		info, err := os.Lstat(next)
		if errors.Is(err, os.ErrNotExist) {
			current = next
			continue
		}
		if err != nil {
			return "", err
		}
		if info.Mode()&os.ModeSymlink == 0 {
			current = next
			continue
		}
		links++
		if links > 255 {
			return "", errors.New("path has excessive symlink traversal")
		}
		link, err := os.Readlink(next)
		if err != nil {
			return "", err
		}
		link = filepath.FromSlash(link)
		if filepath.IsAbs(link) {
			volume = filepath.VolumeName(link)
			current = volume + separator
			link = link[len(volume):]
		} else if filepath.VolumeName(link) != "" {
			return "", errors.New("drive-relative symlink requires an absolute destination")
		} else if link != "" && os.IsPathSeparator(link[0]) {
			current = filepath.VolumeName(current) + separator
		}
		// Process the target before the remaining suffix. In particular, a
		// symlink within the target must be followed before a subsequent .. .
		pending = append(strings.Split(link, separator), pending...)
	}
	return current, nil
}
