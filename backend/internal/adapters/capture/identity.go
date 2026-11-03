package capture

import (
	"errors"
	"os"

	"agenticlens/internal/domain"
)

// CanonicalStore resolves existing filesystem aliases. An unavailable store
// retains an explicit absolute path; absence never establishes source evidence.
func CanonicalStore(path string) (string, error) {
	return ProjectDirectory(path)
}

func SameStore(a, b string) (bool, error) {
	first, err := CanonicalStore(a)
	if err != nil {
		return false, err
	}
	second, err := CanonicalStore(b)
	if err != nil {
		return false, err
	}
	if first == second {
		return true, nil
	}
	fa, ea := os.Stat(first)
	fb, eb := os.Stat(second)
	if ea != nil {
		if !errors.Is(ea, os.ErrNotExist) {
			return false, ea
		}
		return false, nil
	}
	if eb != nil {
		if !errors.Is(eb, os.ErrNotExist) {
			return false, eb
		}
		return false, nil
	}
	return os.SameFile(fa, fb), nil
}
func canonicalIdentity(id domain.CaptureIdentity) (domain.CaptureIdentity, error) {
	var err error
	id.Store, err = CanonicalStore(id.Store)
	return id, err
}
